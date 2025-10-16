package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jyxjjj/Monitor/pkg/agent"
	"github.com/jyxjjj/Monitor/pkg/config"
	"github.com/jyxjjj/Monitor/pkg/models"
)

func main() {
	configPath := flag.String("config", "agent-config.json", "Path to agent configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadAgentConfig(*configPath)
	if err != nil {
		log.Printf("Failed to load config from %s: %v", *configPath, err)
		log.Println("Creating default configuration...")

		// Create default config
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown"
		}

		cfg = &models.AgentConfig{
			ServerURL:      "https://localhost:8443",
			AgentID:        hostname,
			AgentName:      hostname,
			ReportInterval: 5,
			TLSSkipVerify:  true,
		}

		if err := config.SaveConfig(*configPath, cfg); err != nil {
			log.Fatalf("Failed to save default config: %v", err)
		}
		log.Printf("Default configuration saved to %s", *configPath)
		log.Println("Please edit the configuration and restart the agent.")
		os.Exit(0)
	}

	// Create agent
	a := agent.NewAgent(cfg)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down agent...")
		os.Exit(0)
	}()

	// Run agent
	log.Fatal(a.Run())
}
