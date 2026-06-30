package main

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestNewDiscordBot_Validation(t *testing.T) {
	// Test empty token
	_, err := NewDiscordBot("", "123456")
	if err == nil {
		t.Error("Expected error when token is empty, got nil")
	}

	// Test empty channel ID
	_, err = NewDiscordBot("token", "")
	if err == nil {
		t.Error("Expected error when channel ID is empty, got nil")
	}

	// Test successful setup configuration (doesn't connect)
	bot, err := NewDiscordBot("fake_token", "fake_channel")
	if err != nil {
		t.Errorf("Unexpected error during init: %v", err)
	}
	if bot == nil {
		t.Fatal("Expected non-nil bot instance")
	}
}

func TestDiscordBot_Integration(t *testing.T) {
	// Load environment/dot-env variables for integration testing
	_ = godotenv.Load()

	token := os.Getenv("DISCORD_TOKEN")
	channelID := os.Getenv("DISCORD_CHANNEL_ID")

	if token == "" || channelID == "" {
		t.Skip("Skipping live Discord test; DISCORD_TOKEN or DISCORD_CHANNEL_ID not set")
	}

	bot, err := NewDiscordBot(token, channelID)
	if err != nil {
		t.Fatalf("Failed to initialize bot: %v", err)
	}

	t.Log("Connecting to Discord...")
	if err := bot.Start(); err != nil {
		t.Fatalf("Failed to start bot: %v", err)
	}
	defer bot.Stop()

	// Try sending a test update notification report
	report := &UpdateReport{
		AppName:        "Snooter Test Stack",
		CurrentVersion: "v1.0.0",
		NewVersion:     "v1.1.0",
		ReleaseURL:     "https://github.com/brightwaters/snooter/releases",
		UpdateMethod:   "Docker Compose",
		Analysis: &AIAnalysisResult{
			AutoUpdateRisk: "Safe",
			ChangeSummary:  "This is a test notification from the Snooter Discord bot integration test.",
			SecurityReport: "No security issues found.",
		},
	}

	t.Log("Sending test notification to Discord...")
	msgID, err := bot.SendNotification(report)
	if err != nil {
		t.Fatalf("Failed to send notification: %v", err)
	}
	if msgID == "" {
		t.Error("Expected non-empty message ID")
	}

	t.Logf("Notification sent successfully! Message ID: %s", msgID)
}
