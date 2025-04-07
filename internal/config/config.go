package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	SlackBotToken string
	SlackAppToken string
	UploadDir     string
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	config := &Config{
		SlackBotToken: os.Getenv("SLACK_BOT_TOKEN"),
		SlackAppToken: os.Getenv("SLACK_APP_TOKEN"),
		UploadDir:     os.Getenv("UPLOAD_DIR"),
	}

	// Set default upload directory if not specified
	if config.UploadDir == "" {
		config.UploadDir = "./uploads"
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Validate required configuration
	if config.SlackBotToken == "" {
		log.Fatal("SLACK_BOT_TOKEN is required")
	}
	if config.SlackAppToken == "" {
		log.Fatal("SLACK_APP_TOKEN is required")
	}

	return config
}