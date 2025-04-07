package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Bot represents our Slack bot
type Bot struct {
	client      *slack.Client
	socketClient *socketmode.Client
	uploadDir   string
}

// New creates a new Bot instance
func New(botToken, appToken, uploadDir string) *Bot {
	client := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
	)

	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
	)

	return &Bot{
		client:      client,
		socketClient: socketClient,
		uploadDir:   uploadDir,
	}
}

// Start starts the bot
func (b *Bot) Start() error {
	go b.handleEvents()
	return b.socketClient.Run()
}

// handleEvents processes incoming Slack events
func (b *Bot) handleEvents() {
	for evt := range b.socketClient.Events {
		switch evt.Type {
		case socketmode.EventTypeConnecting:
			log.Println("Connecting to Slack with Socket Mode...")
		case socketmode.EventTypeConnectionError:
			log.Println("Connection failed. Retrying...")
		case socketmode.EventTypeConnected:
			log.Println("Connected to Slack with Socket Mode.")
		case socketmode.EventTypeEventsAPI:
			b.socketClient.Ack(*evt.Request)
			eventsAPIEvent, ok := evt.Data.(slack.EventsAPIEvent)
			if !ok {
				log.Printf("Ignored %v", evt)
				continue
			}

			log.Printf("Event received: %v", eventsAPIEvent.Type)
			b.handleEventAPI(eventsAPIEvent)
		case socketmode.EventTypeSlashCommand:
			cmd, ok := evt.Data.(slack.SlashCommand)
			if !ok {
				log.Printf("Ignored %v", evt)
				continue
			}
			b.socketClient.Ack(*evt.Request)
			b.handleSlashCommand(cmd)
		}
	}
}

// handleEventAPI handles Slack Events API events
func (b *Bot) handleEventAPI(event slack.EventsAPIEvent) {
	switch event.Type {
	case slack.EventTypeMessage:
		b.handleMessageEvent(event)
	}
}

// handleMessageEvent handles message events
func (b *Bot) handleMessageEvent(event slack.EventsAPIEvent) {
	ev, ok := event.InnerEvent.Data.(*slack.MessageEvent)
	if !ok {
		return
	}

	// Ignore messages from the bot itself
	info, err := b.client.GetBotInfo(b.client.GetBotID())
	if err != nil {
		log.Printf("Error getting bot info: %v", err)
		return
	}
	if ev.User == info.ID {
		return
	}

	// Check if the message has files
	if ev.Files != nil && len(ev.Files) > 0 {
		b.handleFileUpload(ev)
	}
}

// handleFileUpload processes file uploads in messages
func (b *Bot) handleFileUpload(ev *slack.MessageEvent) {
	for _, file := range ev.Files {
		// Download the file
		err := b.downloadFile(file)
		if err != nil {
			log.Printf("Error downloading file: %v", err)
			b.client.PostMessage(ev.Channel, slack.MsgOptionText(fmt.Sprintf("Error downloading file: %v", err), false))
			continue
		}

		// Send confirmation message
		b.client.PostMessage(
			ev.Channel,
			slack.MsgOptionText(fmt.Sprintf("File `%s` has been successfully downloaded and saved.", file.Name), false),
		)
	}
}

// downloadFile downloads a file from Slack
func (b *Bot) downloadFile(file slack.File) error {
	// Create a safe filename
	safeFilename := strings.ReplaceAll(file.Name, "/", "_")
	filePath := filepath.Join(b.uploadDir, safeFilename)

	// Get file info
	fileInfo, _, _, err := b.client.GetFileInfo(file.ID, 0, 0)
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	// Download the file
	resp, err := http.Get(fileInfo.URLPrivateDownload)
	if err != nil {
		return fmt.Errorf("error downloading file: %w", err)
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	// Write the file content
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	log.Printf("File downloaded: %s", filePath)
	return nil
}

// handleSlashCommand handles slash commands
func (b *Bot) handleSlashCommand(cmd slack.SlashCommand) {
	switch cmd.Command {
	case "/upload":
		// This is a placeholder for a slash command that might trigger file upload
		b.client.PostMessage(
			cmd.ChannelID,
			slack.MsgOptionText("To upload a file, simply attach it to a message in this channel.", false),
		)
	}
}