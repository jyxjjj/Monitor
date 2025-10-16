package server

import (
	"database/sql"
	"time"

	"github.com/jyxjjj/Monitor/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

// Database handles database operations
type Database struct {
	db *sql.DB
}

// NewDatabase creates and initializes the database
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.initSchema(); err != nil {
		return nil, err
	}

	return database, nil
}

// initSchema initializes database schema
func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS agents (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		host TEXT NOT NULL,
		last_seen DATETIME NOT NULL,
		status TEXT NOT NULL,
		platform TEXT NOT NULL,
		version TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		cpu_percent REAL NOT NULL,
		memory_used INTEGER NOT NULL,
		memory_total INTEGER NOT NULL,
		disk_used INTEGER NOT NULL,
		disk_total INTEGER NOT NULL,
		network_rx INTEGER NOT NULL,
		network_tx INTEGER NOT NULL,
		load_avg_1 REAL NOT NULL,
		load_avg_5 REAL NOT NULL,
		load_avg_15 REAL NOT NULL,
		FOREIGN KEY (agent_id) REFERENCES agents(id)
	);

	CREATE INDEX IF NOT EXISTS idx_metrics_agent_time ON metrics(agent_id, timestamp);

	CREATE TABLE IF NOT EXISTS alert_rules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id TEXT NOT NULL,
		metric_type TEXT NOT NULL,
		threshold REAL NOT NULL,
		operator TEXT NOT NULL,
		duration INTEGER NOT NULL,
		enabled INTEGER NOT NULL,
		description TEXT NOT NULL,
		FOREIGN KEY (agent_id) REFERENCES agents(id)
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		rule_id INTEGER NOT NULL,
		agent_id TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		message TEXT NOT NULL,
		value REAL NOT NULL,
		resolved INTEGER NOT NULL,
		FOREIGN KEY (rule_id) REFERENCES alert_rules(id),
		FOREIGN KEY (agent_id) REFERENCES agents(id)
	);

	CREATE INDEX IF NOT EXISTS idx_alerts_agent_time ON alerts(agent_id, timestamp);
	`

	_, err := d.db.Exec(schema)
	return err
}

// SaveMetrics saves metrics to database
func (d *Database) SaveMetrics(m *models.Metrics) error {
	_, err := d.db.Exec(`
		INSERT INTO metrics (agent_id, timestamp, cpu_percent, memory_used, memory_total,
			disk_used, disk_total, network_rx, network_tx, load_avg_1, load_avg_5, load_avg_15)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.AgentID, m.Timestamp, m.CPUPercent, m.MemoryUsed, m.MemoryTotal,
		m.DiskUsed, m.DiskTotal, m.NetworkRx, m.NetworkTx,
		m.LoadAvg1, m.LoadAvg5, m.LoadAvg15,
	)
	return err
}

// UpdateAgent updates or inserts agent information
func (d *Database) UpdateAgent(agent *models.Agent) error {
	_, err := d.db.Exec(`
		INSERT OR REPLACE INTO agents (id, name, host, last_seen, status, platform, version)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		agent.ID, agent.Name, agent.Host, agent.LastSeen, agent.Status, agent.Platform, agent.Version,
	)
	return err
}

// GetAgents retrieves all agents
func (d *Database) GetAgents() ([]*models.Agent, error) {
	rows, err := d.db.Query(`
		SELECT id, name, host, last_seen, status, platform, version FROM agents
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
	rows, err := d.db.Query(`
		SELECT agent_id, timestamp, cpu_percent, memory_used, memory_total,
			disk_used, disk_total, network_rx, network_tx, load_avg_1, load_avg_5, load_avg_15
		FROM metrics
		WHERE agent_id = ? AND timestamp >= ?
		ORDER BY timestamp DESC
		LIMIT 1000
	`, agentID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*models.Metrics
	for rows.Next() {
		m := &models.Metrics{}
		err := rows.Scan(&m.AgentID, &m.Timestamp, &m.CPUPercent, &m.MemoryUsed, &m.MemoryTotal,
			&m.DiskUsed, &m.DiskTotal, &m.NetworkRx, &m.NetworkTx,
			&m.LoadAvg1, &m.LoadAvg5, &m.LoadAvg15)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}

// SaveAlertRule saves an alert rule
func (d *Database) SaveAlertRule(rule *models.AlertRule) error {
	if rule.ID == 0 {
		result, err := d.db.Exec(`
			INSERT INTO alert_rules (agent_id, metric_type, threshold, operator, duration, enabled, description)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			rule.AgentID, rule.MetricType, rule.Threshold, rule.Operator, rule.Duration,
			rule.Enabled, rule.Description,
		)
		if err != nil {
			return err
		}
		id, _ := result.LastInsertId()
		rule.ID = int(id)
	} else {
		_, err := d.db.Exec(`
			UPDATE alert_rules SET agent_id=?, metric_type=?, threshold=?, operator=?, 
				duration=?, enabled=?, description=?
			WHERE id=?`,
			rule.AgentID, rule.MetricType, rule.Threshold, rule.Operator, rule.Duration,
			rule.Enabled, rule.Description, rule.ID,
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
		var enabled int
		err := rows.Scan(&rule.ID, &rule.AgentID, &rule.MetricType, &rule.Threshold,
			&rule.Operator, &rule.Duration, &enabled, &rule.Description)
		if err != nil {
			return nil, err
		}
		rule.Enabled = enabled != 0
		rules = append(rules, rule)
	}

	return rules, nil
}

// SaveAlert saves a triggered alert
func (d *Database) SaveAlert(alert *models.Alert) error {
	result, err := d.db.Exec(`
		INSERT INTO alerts (rule_id, agent_id, timestamp, message, value, resolved)
		VALUES (?, ?, ?, ?, ?, ?)`,
		alert.RuleID, alert.AgentID, alert.Timestamp, alert.Message, alert.Value, alert.Resolved,
	)
	if err != nil {
		return err
	}
	id, _ := result.LastInsertId()
	alert.ID = int(id)
	return nil
}

// GetAlerts retrieves recent alerts
func (d *Database) GetAlerts(limit int) ([]*models.Alert, error) {
	rows, err := d.db.Query(`
		SELECT id, rule_id, agent_id, timestamp, message, value, resolved
		FROM alerts
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*models.Alert
	for rows.Next() {
		alert := &models.Alert{}
		var resolved int
		err := rows.Scan(&alert.ID, &alert.RuleID, &alert.AgentID, &alert.Timestamp,
			&alert.Message, &alert.Value, &resolved)
		if err != nil {
			return nil, err
		}
		alert.Resolved = resolved != 0
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// DeleteOldMetrics deletes metrics older than the specified duration
func (d *Database) DeleteOldMetrics(olderThan time.Time) error {
	_, err := d.db.Exec(`DELETE FROM metrics WHERE timestamp < ?`, olderThan)
	return err
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}
