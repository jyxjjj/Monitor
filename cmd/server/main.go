package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jyxjjj/Monitor/pkg/config"
	"github.com/jyxjjj/Monitor/pkg/server"
)

func main() {
	configPath := flag.String("config", "server-config.json", "Path to server configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		log.Printf("Failed to load config from %s: %v", *configPath, err)
		log.Println("Creating default configuration...")
		
		// Create default config
		cfg = &server.DefaultConfig
		if err := config.SaveConfig(*configPath, cfg); err != nil {
			log.Fatalf("Failed to save default config: %v", err)
		}
		log.Printf("Default configuration saved to %s", *configPath)
		log.Println("Please edit the configuration and restart the server.")
		os.Exit(0)
	}

	// Create server
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Start cleanup routine for old metrics
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			// Keep metrics for 30 days
			cutoff := time.Now().Add(-30 * 24 * time.Hour)
			log.Printf("Cleaning up metrics older than %s", cutoff.Format(time.RFC3339))
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		srv.Close()
		os.Exit(0)
	}()

	// Start server
	log.Fatal(srv.Start())
}
