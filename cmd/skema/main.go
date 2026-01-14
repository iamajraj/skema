package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/iamajraj/skema/internal/config"
	"github.com/iamajraj/skema/internal/db"
	"github.com/iamajraj/skema/internal/docs"
	"github.com/iamajraj/skema/internal/server"
)

func main() {
	configPath := flag.String("config", "skema.yml", "Path to the configuration file")
	flag.Parse()

	// Check if config exists
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		fmt.Printf("Config file %s not found.\n", *configPath)
		fmt.Println("Please create a skema.yml file or specify one with --config.")
		fmt.Println("Example skema.yml:")
		fmt.Print(`
server:
  port: 8080
  name: "Skema API"
entities:
  - name: User
    fields:
      - name: name
        type: string
`)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("🚀 Starting %s...\n", cfg.Server.Name)

	// Initialize DB
	database, err := db.InitDB(cfg, "skema.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create server
	srv := server.NewServer(cfg, database)

	// Register Docs
	docs.RegisterSwagger(srv.Router, cfg)

	fmt.Printf("📚 Docs available at http://localhost:%d/docs\n", cfg.Server.Port)

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
