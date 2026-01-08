package main

import (
	"flag"
	"log"

	"go-boilerplate/config"
	"go-boilerplate/internal/app"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Configuration
	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
