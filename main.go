package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/phuslu/log"
	"gopkg.in/yaml.v3"
)

func main() {
	// Load .env file from the working directory
	err := godotenv.Load()
	if err != nil {
		log.Info().Msg("No .env file found, relying on environment variables")
	}

	data, err := os.ReadFile("snooter.yaml")
	if err != nil {
		log.Error().Err(err).Msg("Failed to read config file")
		os.Exit(1)
	}

	var config SnooterConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Error().Err(err).Msg("Failed to parse YAML")
		os.Exit(1)
	}

	// 1. Fail fast if no apps are defined
	if len(config.Deployments) == 0 {
		log.Fatal().Msg("No apps defined in snooter.yaml. Add at least one app to continue.")
	}

	// 2. Initialize the SQLite database and create schemas
	db, err := InitDB("data/snooter.db")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize SQLite database")
	}
	defer db.Close()

	// 3. Sync the desired state (config) into the current state (database)
	if err := SyncAppMetadata(db, &config); err != nil {
		log.Fatal().Err(err).Msg("Failed to sync app metadata to database")
	}

	for _, dw := range config.Deployments {
		dep := dw.Deployment
		log.Info().Str("name", dep.GetName()).Str("type", dep.GetType()).Msg("Loaded deployment configuration")

		// Temporarily disabled so it doesn't actually try to run deployments on every start
		// dep.Run()
	}
}
