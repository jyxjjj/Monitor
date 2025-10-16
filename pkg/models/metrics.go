package models

import "time"

// Metrics represents system metrics from an agent
type Metrics struct {
	AgentID       string    `json:"agent_id"`
	Timestamp     time.Time `json:"timestamp"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryUsed    uint64    `json:"memory_used"`
	MemoryTotal   uint64    `json:"memory_total"`
	DiskUsed      uint64    `json:"disk_used"`
	DiskTotal     uint64    `json:"disk_total"`
	NetworkRx     uint64    `json:"network_rx"`
	NetworkTx     uint64    `json:"network_tx"`
	LoadAvg1      float64   `json:"load_avg_1"`
	LoadAvg5      float64   `json:"load_avg_5"`
	LoadAvg15     float64   `json:"load_avg_15"`
}

// Agent represents a monitored agent
type Agent struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"` // online, offline
	Platform    string    `json:"platform"`
	Version     string    `json:"version"`
}

// AlertRule represents an alert rule
type AlertRule struct {
	ID          int     `json:"id"`
	AgentID     string  `json:"agent_id"`
	MetricType  string  `json:"metric_type"` // cpu, memory, disk, load
	Threshold   float64 `json:"threshold"`
	Operator    string  `json:"operator"` // gt, lt, gte, lte
	Duration    int     `json:"duration"` // seconds
	Enabled     bool    `json:"enabled"`
	Description string  `json:"description"`
}

// Alert represents a triggered alert
type Alert struct {
	ID          int       `json:"id"`
	RuleID      int       `json:"rule_id"`
	AgentID     string    `json:"agent_id"`
	Timestamp   time.Time `json:"timestamp"`
	Message     string    `json:"message"`
	Value       float64   `json:"value"`
	Resolved    bool      `json:"resolved"`
}

// Config represents server configuration
type Config struct {
	ServerAddr    string `json:"server_addr"`
	TLSCertFile   string `json:"tls_cert_file"`
	TLSKeyFile    string `json:"tls_key_file"`
	DBPath        string `json:"db_path"`
	AdminPassword string `json:"admin_password"`
	SMTPHost      string `json:"smtp_host"`
	SMTPPort      int    `json:"smtp_port"`
	SMTPUser      string `json:"smtp_user"`
	SMTPPassword  string `json:"smtp_password"`
	EmailFrom     string `json:"email_from"`
	AlertEmail    string `json:"alert_email"`
}

// AgentConfig represents agent configuration
type AgentConfig struct {
	ServerURL     string `json:"server_url"`
	AgentID       string `json:"agent_id"`
	AgentName     string `json:"agent_name"`
	ReportInterval int   `json:"report_interval"` // seconds
	TLSSkipVerify bool   `json:"tls_skip_verify"`
}
