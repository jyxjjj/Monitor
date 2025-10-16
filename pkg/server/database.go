package server

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jyxjjj/Monitor/pkg/models"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Database handles database operations with multi-driver support
type Database struct {
	db     *sql.DB
	driver string
}

// NewDatabase creates and initializes the database
func NewDatabase(config models.DatabaseConfig) (*Database, error) {
	var dsn string
	var err error

	// Build DSN based on driver
	switch config.Driver {
	case "mysql":
		if config.Charset == "" {
			config.Charset = "utf8mb4"
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true",
			config.Username, config.Password, config.Host, config.Port, config.Database, config.Charset)
	case "postgres":
		if config.SSLMode == "" {
			config.SSLMode = "disable"
		}
		dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode)
	case "sqlite", "":
		dsn = config.Database
		config.Driver = "sqlite3"
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	database := &Database{
		db:     db,
		driver: config.Driver,
	}

	return database, nil
}

// InitSchema initializes database schema with Laravel-style naming
func (d *Database) InitSchema() error {
	var schema string

	switch d.driver {
	case "mysql":
		schema = d.getMySQLSchema()
	case "postgres":
		schema = d.getPostgreSQLSchema()
	default: // sqlite3
		schema = d.getSQLiteSchema()
	}

	// Execute schema in transactions
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Split and execute each statement
	statements := strings.Split(schema, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := tx.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %w\n%s", err, stmt)
		}
	}

	return tx.Commit()
}

// getSQLiteSchema returns SQLite schema with Laravel-style naming
func (d *Database) getSQLiteSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		host TEXT NOT NULL,
		last_seen_at DATETIME(3) NOT NULL,
		status TEXT NOT NULL,
		platform TEXT NOT NULL,
		version TEXT NOT NULL,
		created_at DATETIME(3) DEFAULT (datetime('now','localtime')),
		updated_at DATETIME(3) NOT NULL
	);

	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id TEXT NOT NULL,
		cpu_percent REAL NOT NULL,
		cpu_cores INTEGER NOT NULL DEFAULT 1,
		memory_used BIGINT NOT NULL,
		memory_total BIGINT NOT NULL,
		disk_used BIGINT NOT NULL,
		disk_total BIGINT NOT NULL,
		network_rx BIGINT NOT NULL,
		network_tx BIGINT NOT NULL,
		load_avg_1 REAL NOT NULL,
		load_avg_5 REAL NOT NULL,
		load_avg_15 REAL NOT NULL,
		created_at DATETIME(3) DEFAULT (datetime('now','localtime')),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_agent_created ON metrics(agent_id, created_at);

	CREATE TABLE IF NOT EXISTS alert_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id TEXT NOT NULL,
		metric_type TEXT NOT NULL,
		threshold REAL NOT NULL,
		operator TEXT NOT NULL,
		duration INTEGER NOT NULL,
		enabled INTEGER NOT NULL DEFAULT 1,
		description TEXT NOT NULL,
		created_at DATETIME(3) DEFAULT (datetime('now','localtime')),
		updated_at DATETIME(3) DEFAULT (datetime('now','localtime')),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		rule_id INTEGER NOT NULL,
		agent_id TEXT NOT NULL,
		message TEXT NOT NULL,
		value REAL NOT NULL,
		resolved INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME(3) DEFAULT (datetime('now','localtime')),
		updated_at DATETIME(3) DEFAULT (datetime('now','localtime')),
		FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_alerts_agent_created ON alerts(agent_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_alerts_agent_created ON alerts(agent_id, created_at);
	`
}

// getMySQLSchema returns MySQL schema with Laravel-style naming
func (d *Database) getMySQLSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS agents (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		host VARCHAR(255) NOT NULL,
		last_seen_at DATETIME(3) NOT NULL,
		status VARCHAR(50) NOT NULL,
		platform VARCHAR(50) NOT NULL,
		version VARCHAR(50) NOT NULL,
	created_at DATETIME(3) NOT NULL,
		updated_at DATETIME(3) NOT NULL,
		INDEX idx_status (status),
		INDEX idx_last_seen (last_seen_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS metrics (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		cpu_percent DOUBLE NOT NULL,
		cpu_cores INT NOT NULL DEFAULT 1,
		memory_used BIGINT UNSIGNED NOT NULL,
		memory_total BIGINT UNSIGNED NOT NULL,
		disk_used BIGINT UNSIGNED NOT NULL,
		disk_total BIGINT UNSIGNED NOT NULL,
		network_rx BIGINT UNSIGNED NOT NULL,
		network_tx BIGINT UNSIGNED NOT NULL,
		load_avg_1 DOUBLE NOT NULL,
		load_avg_5 DOUBLE NOT NULL,
		load_avg_15 DOUBLE NOT NULL,
	created_at DATETIME(3) NOT NULL,
		INDEX idx_metrics_agent_created (agent_id, created_at),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS alert_rules (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		metric_type VARCHAR(50) NOT NULL,
		threshold DOUBLE NOT NULL,
		operator VARCHAR(10) NOT NULL,
		duration INT NOT NULL,
		enabled TINYINT(1) NOT NULL DEFAULT 1,
		description TEXT NOT NULL,
	created_at DATETIME(3) NOT NULL,
		updated_at DATETIME(3) NOT NULL,
		INDEX idx_enabled (enabled),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

	CREATE TABLE IF NOT EXISTS alerts (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		rule_id BIGINT UNSIGNED NOT NULL,
		agent_id VARCHAR(255) NOT NULL,
		message TEXT NOT NULL,
		value DOUBLE NOT NULL,
		resolved TINYINT(1) NOT NULL DEFAULT 0,
	created_at DATETIME(3) NOT NULL,
		updated_at DATETIME(3) NOT NULL,
		INDEX idx_alerts_agent_created (agent_id, created_at),
		INDEX idx_resolved (resolved),
		FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`
}

// getPostgreSQLSchema returns PostgreSQL schema with Laravel-style naming
func (d *Database) getPostgreSQLSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS agents (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		host VARCHAR(255) NOT NULL,
		last_seen_at TIMESTAMP NOT NULL,
		status VARCHAR(50) NOT NULL,
		platform VARCHAR(50) NOT NULL,
		version VARCHAR(50) NOT NULL,
		created_at TIMESTAMP(3) NOT NULL,
		updated_at TIMESTAMP(3) NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
	CREATE INDEX IF NOT EXISTS idx_agents_last_seen ON agents(last_seen_at);

	CREATE TABLE IF NOT EXISTS metrics (
		id BIGSERIAL PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		cpu_percent DOUBLE PRECISION NOT NULL,
		cpu_cores INTEGER NOT NULL DEFAULT 1,
		memory_used BIGINT NOT NULL,
		memory_total BIGINT NOT NULL,
		disk_used BIGINT NOT NULL,
		disk_total BIGINT NOT NULL,
		network_rx BIGINT NOT NULL,
		network_tx BIGINT NOT NULL,
		load_avg_1 DOUBLE PRECISION NOT NULL,
		load_avg_5 DOUBLE PRECISION NOT NULL,
		load_avg_15 DOUBLE PRECISION NOT NULL,
		created_at TIMESTAMP(3) NOT NULL,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_agent_created ON metrics(agent_id, created_at);

	CREATE TABLE IF NOT EXISTS alert_rules (
		id BIGSERIAL PRIMARY KEY,
		agent_id VARCHAR(255) NOT NULL,
		metric_type VARCHAR(50) NOT NULL,
		threshold DOUBLE PRECISION NOT NULL,
		operator VARCHAR(10) NOT NULL,
		duration INTEGER NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT TRUE,
		description TEXT NOT NULL,
		created_at TIMESTAMP(3) NOT NULL,
		updated_at TIMESTAMP(3) NOT NULL,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_alert_rules_enabled ON alert_rules(enabled);

	CREATE TABLE IF NOT EXISTS alerts (
		id BIGSERIAL PRIMARY KEY,
		rule_id BIGINT NOT NULL,
		agent_id VARCHAR(255) NOT NULL,
		message TEXT NOT NULL,
		value DOUBLE PRECISION NOT NULL,
		resolved BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP(3) NOT NULL,
		updated_at TIMESTAMP(3) NOT NULL,
		FOREIGN KEY (rule_id) REFERENCES alert_rules(id) ON DELETE CASCADE,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_alerts_agent_created ON alerts(agent_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON alerts(resolved);
	`
}

// SaveMetrics saves metrics to database
func (d *Database) SaveMetrics(m *models.Metrics) error {
	_, err := d.db.Exec(`
		INSERT INTO metrics (agent_id, cpu_percent, cpu_cores, memory_used, memory_total,
			disk_used, disk_total, network_rx, network_tx, load_avg_1, load_avg_5, load_avg_15, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.AgentID, m.CPUPercent, m.CPUCores, m.MemoryUsed, m.MemoryTotal,
		m.DiskUsed, m.DiskTotal, m.NetworkRx, m.NetworkTx,
		m.LoadAvg1, m.LoadAvg5, m.LoadAvg15, m.Timestamp,
	)
	return err
}

// UpdateAgent updates or inserts agent information
func (d *Database) UpdateAgent(agent *models.Agent) error {
	now := time.Now()

	// Use UPSERT pattern based on driver
	switch d.driver {
	case "mysql":
		_, err := d.db.Exec(`
			INSERT INTO agents (id, name, host, last_seen_at, status, platform, version, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				name=VALUES(name), host=VALUES(host), last_seen_at=VALUES(last_seen_at),
				status=VALUES(status), platform=VALUES(platform), version=VALUES(version), updated_at=VALUES(updated_at)`,
			agent.ID, agent.Name, agent.Host, agent.LastSeen, agent.Status, agent.Platform, agent.Version, now, now,
		)
		return err
	case "postgres":
		_, err := d.db.Exec(`
			INSERT INTO agents (id, name, host, last_seen_at, status, platform, version, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE SET
				name=EXCLUDED.name, host=EXCLUDED.host, last_seen_at=EXCLUDED.last_seen_at,
				status=EXCLUDED.status, platform=EXCLUDED.platform, version=EXCLUDED.version, updated_at=EXCLUDED.updated_at`,
			agent.ID, agent.Name, agent.Host, agent.LastSeen, agent.Status, agent.Platform, agent.Version, now, now,
		)
		return err
	default: // sqlite3
		_, err := d.db.Exec(`
			INSERT OR REPLACE INTO agents (id, name, host, last_seen_at, status, platform, version, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			agent.ID, agent.Name, agent.Host, agent.LastSeen, agent.Status, agent.Platform, agent.Version, now, now,
		)
		return err
	}
}

// GetAgents retrieves all agents
func (d *Database) GetAgents() ([]*models.Agent, error) {
	rows, err := d.db.Query(`
		SELECT id, name, host, last_seen_at, status, platform, version FROM agents
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*models.Agent
	for rows.Next() {
		agent := &models.Agent{}
		err := rows.Scan(&agent.ID, &agent.Name, &agent.Host, &agent.LastSeen,
			&agent.Status, &agent.Platform, &agent.Version)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// GetMetricsHistory retrieves metrics history for an agent
func (d *Database) GetMetricsHistory(agentID string, since time.Time) ([]*models.Metrics, error) {
	// Some drivers/store formats (e.g. sqlite's CURRENT_TIMESTAMP) store timestamps
	// as 'YYYY-MM-DD HH:MM:SS'. To ensure comparisons work reliably, format the
	// 'since' parameter appropriately for sqlite. For other drivers we pass the
	// time.Time value directly.
	var sinceParam interface{} = since
	var query string
	// Use local formatted string for sqlite; other drivers can accept time.Time
	var layout = "2006-01-02 15:04:05"
	if d.driver == "sqlite3" {
		sinceParam = since.Format(layout)
		query = `
		SELECT agent_id, cpu_percent, cpu_cores, memory_used, memory_total,
			disk_used, disk_total, network_rx, network_tx, load_avg_1, load_avg_5, load_avg_15, created_at
		FROM metrics
		WHERE agent_id = ? AND created_at >= ?
		ORDER BY created_at DESC
		LIMIT 1000
	`
	} else {
		query = `
		SELECT agent_id, cpu_percent, cpu_cores, memory_used, memory_total,
			disk_used, disk_total, network_rx, network_tx, load_avg_1, load_avg_5, load_avg_15, created_at
		FROM metrics
		WHERE agent_id = ? AND created_at >= ?
		ORDER BY created_at DESC
		LIMIT 1000
	`
	}

	rows, err := d.db.Query(query, agentID, sinceParam)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*models.Metrics
	for rows.Next() {
		m := &models.Metrics{}
		err := rows.Scan(&m.AgentID, &m.CPUPercent, &m.CPUCores, &m.MemoryUsed, &m.MemoryTotal,
			&m.DiskUsed, &m.DiskTotal, &m.NetworkRx, &m.NetworkTx,
			&m.LoadAvg1, &m.LoadAvg5, &m.LoadAvg15, &m.Timestamp)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// SaveAlertRule saves an alert rule
func (d *Database) SaveAlertRule(rule *models.AlertRule) error {
	now := time.Now()

	if rule.ID == 0 {
		var query string
		switch d.driver {
		case "postgres":
			query = `INSERT INTO alert_rules (agent_id, metric_type, threshold, operator, duration, enabled, description, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
		default:
			query = `INSERT INTO alert_rules (agent_id, metric_type, threshold, operator, duration, enabled, description, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		}

		if d.driver == "postgres" {
			err := d.db.QueryRow(query,
				rule.AgentID, rule.MetricType, rule.Threshold, rule.Operator, rule.Duration,
				rule.Enabled, rule.Description, now, now,
			).Scan(&rule.ID)
			return err
		} else {
			result, err := d.db.Exec(query,
				rule.AgentID, rule.MetricType, rule.Threshold, rule.Operator, rule.Duration,
				rule.Enabled, rule.Description, now, now,
			)
			if err != nil {
				return err
			}
			id, _ := result.LastInsertId()
			rule.ID = int(id)
		}
	} else {
		_, err := d.db.Exec(`
			UPDATE alert_rules SET agent_id=?, metric_type=?, threshold=?, operator=?,
				duration=?, enabled=?, description=?, updated_at=?
			WHERE id=?`,
			rule.AgentID, rule.MetricType, rule.Threshold, rule.Operator, rule.Duration,
			rule.Enabled, rule.Description, now, rule.ID,
		)
		return err
	}
	return nil
}

// GetAlertRules retrieves all alert rules
func (d *Database) GetAlertRules() ([]*models.AlertRule, error) {
	rows, err := d.db.Query(`
		SELECT id, agent_id, metric_type, threshold, operator, duration, enabled, description
		FROM alert_rules
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.AlertRule
	for rows.Next() {
		rule := &models.AlertRule{}
		var enabled interface{}
		err := rows.Scan(&rule.ID, &rule.AgentID, &rule.MetricType, &rule.Threshold,
			&rule.Operator, &rule.Duration, &enabled, &rule.Description)
		if err != nil {
			return nil, err
		}
		// Handle different enabled types from different databases
		switch v := enabled.(type) {
		case bool:
			rule.Enabled = v
		case int64:
			rule.Enabled = v != 0
		case int:
			rule.Enabled = v != 0
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// SaveAlert saves a triggered alert
func (d *Database) SaveAlert(alert *models.Alert) error {
	now := time.Now()

	var query string
	switch d.driver {
	case "postgres":
		query = `INSERT INTO alerts (rule_id, agent_id, message, value, resolved, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	default:
		query = `INSERT INTO alerts (rule_id, agent_id, message, value, resolved, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`
	}

	if d.driver == "postgres" {
		err := d.db.QueryRow(query,
			alert.RuleID, alert.AgentID, alert.Message, alert.Value, alert.Resolved, alert.Timestamp, now,
		).Scan(&alert.ID)
		return err
	} else {
		result, err := d.db.Exec(query,
			alert.RuleID, alert.AgentID, alert.Message, alert.Value, alert.Resolved, alert.Timestamp, now,
		)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		alert.ID = int(id)
	}
	return nil
}

// GetAlerts retrieves recent alerts
func (d *Database) GetAlerts(limit int) ([]*models.Alert, error) {
	rows, err := d.db.Query(`
		SELECT id, rule_id, agent_id, message, value, resolved, created_at
		FROM alerts
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		var resolved interface{}
		err := rows.Scan(&alert.ID, &alert.RuleID, &alert.AgentID,
			&alert.Message, &alert.Value, &resolved, &alert.Timestamp)
		if err != nil {
			return nil, err
		}
		// Handle different resolved types from different databases
		switch v := resolved.(type) {
		case bool:
			alert.Resolved = v
		case int64:
			alert.Resolved = v != 0
		case int:
			alert.Resolved = v != 0
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// DeleteOldMetrics deletes metrics older than the specified duration
func (d *Database) DeleteOldMetrics(olderThan time.Time) error {
	if d.driver == "sqlite3" {
		_, err := d.db.Exec(`DELETE FROM metrics WHERE created_at < ?`, olderThan.Format("2006-01-02 15:04:05"))
		return err
	}
	_, err := d.db.Exec(`DELETE FROM metrics WHERE created_at < ?`, olderThan)
	return err
}

// CheckInstalled checks if the database schema is installed
func (d *Database) CheckInstalled() (bool, error) {
	var tableName string
	var err error

	switch d.driver {
	case "mysql":
		err = d.db.QueryRow("SHOW TABLES LIKE 'agents'").Scan(&tableName)
	case "postgres":
		err = d.db.QueryRow("SELECT tablename FROM pg_tables WHERE tablename = 'agents'").Scan(&tableName)
	default: // sqlite3
		err = d.db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='agents'").Scan(&tableName)
	}

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return tableName == "agents", nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}
