package agent

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/jyxjjj/Monitor/pkg/collector"
	"github.com/jyxjjj/Monitor/pkg/compress"
	"github.com/jyxjjj/Monitor/pkg/models"
)

// Agent represents the monitoring agent
type Agent struct {
	config    *models.AgentConfig
	collector *collector.Collector
	client    *http.Client
}

// NewAgent creates a new agent
func NewAgent(config *models.AgentConfig) *Agent {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.TLSSkipVerify,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 10 * time.Second,
	}

	return &Agent{
		config:    config,
		collector: collector.NewCollector(config.AgentID),
		client:    client,
	}
}

// Run starts the agent monitoring loop
func (a *Agent) Run() error {
	fmt.Printf("Agent %s starting, reporting to %s\n", a.config.AgentID, a.config.ServerURL)
	
	ticker := time.NewTicker(time.Duration(a.config.ReportInterval) * time.Second)
	defer ticker.Stop()

	// Send initial report immediately
	if err := a.report(); err != nil {
		fmt.Printf("Failed to send initial report: %v\n", err)
	}

	for range ticker.C {
		if err := a.report(); err != nil {
			fmt.Printf("Failed to report metrics: %v\n", err)
		}
	}

	return nil
}

// report collects and sends metrics to server
func (a *Agent) report() error {
	metrics, err := a.collector.Collect()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Add platform information
	metrics.AgentID = a.config.AgentID

	// Serialize to JSON
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	// Compress with Brotli
	compressed, err := compress.CompressBrotli(jsonData)
	if err != nil {
		return fmt.Errorf("failed to compress metrics: %w", err)
	}

	// Send to server
	url := fmt.Sprintf("%s/api/metrics/report", a.config.ServerURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "br")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %d - %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Reported metrics: CPU=%.1f%%, Memory=%d/%d, Disk=%d/%d\n",
		metrics.CPUPercent,
		metrics.MemoryUsed,
		metrics.MemoryTotal,
		metrics.DiskUsed,
		metrics.DiskTotal,
	)

	return nil
}

// GetPlatform returns the current platform
func GetPlatform() string {
	return runtime.GOOS
}
