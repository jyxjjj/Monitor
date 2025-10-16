package config

import (
	"encoding/json"
	"os"

	"github.com/jyxjjj/Monitor/pkg/models"
)

// LoadServerConfig loads server configuration from file
func LoadServerConfig(path string) (*models.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config models.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.ServerAddr == "" {
		config.ServerAddr = ":8443"
	}
	if config.DBPath == "" {
		config.DBPath = "./monitor.db"
	}

	return &config, nil
}

// LoadAgentConfig loads agent configuration from file
func LoadAgentConfig(path string) (*models.AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config models.AgentConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.ReportInterval == 0 {
		config.ReportInterval = 5 // 5 seconds
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(path string, config interface{}) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
