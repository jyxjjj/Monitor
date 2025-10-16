package server

import "github.com/jyxjjj/Monitor/pkg/models"

// DefaultConfig provides default server configuration
var DefaultConfig = models.Config{
	ServerAddr:  ":8443",
	TLSCertFile: "",
	TLSKeyFile:  "",
	Database: models.DatabaseConfig{
		Driver:   "sqlite3",
		Database: "./monitor.db",
	},
	AdminPassword: "admin123",
	SMTPHost:      "",
	SMTPPort:      587,
	SMTPUser:      "",
	SMTPPassword:  "",
	EmailFrom:     "",
	AlertEmail:    "",
	Installed:     false,
}
