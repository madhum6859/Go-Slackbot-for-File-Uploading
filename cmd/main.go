package main

import (
	"log"

	"github.com/trae/slackbot/internal/bot"
	"github.com/trae/slackbot/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create and start the bot
	slackBot := bot.New(cfg.SlackBotToken, cfg.SlackAppToken, cfg.UploadDir)
	log.Println("Starting Slack bot...")
	if err := slackBot.Start(); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}
}