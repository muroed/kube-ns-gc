package main

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestTelegramConfig(t *testing.T) {
	config := &TelegramConfig{
		Enabled:   true,
		BotToken:  "test-token",
		ChatID:    "test-chat-id",
		ParseMode: "Markdown",
		Notifications: TelegramNotifications{
			Startup:            true,
			NamespaceDeleted:   true,
			HelmReleaseDeleted: true,
			CleanupSummary:     true,
			Errors:             true,
		},
	}

	if !config.Enabled {
		t.Errorf("Expected Enabled to be true, got %v", config.Enabled)
	}

	if config.BotToken != "test-token" {
		t.Errorf("Expected BotToken to be 'test-token', got %s", config.BotToken)
	}

	if config.ChatID != "test-chat-id" {
		t.Errorf("Expected ChatID to be 'test-chat-id', got %s", config.ChatID)
	}

	if config.ParseMode != "Markdown" {
		t.Errorf("Expected ParseMode to be 'Markdown', got %s", config.ParseMode)
	}

	if !config.Notifications.Startup {
		t.Errorf("Expected Startup notification to be true, got %v", config.Notifications.Startup)
	}

	if !config.Notifications.NamespaceDeleted {
		t.Errorf("Expected NamespaceDeleted notification to be true, got %v", config.Notifications.NamespaceDeleted)
	}
}

func TestNewTelegramClient(t *testing.T) {
	config := &TelegramConfig{
		Enabled:   true,
		BotToken:  "test-token",
		ChatID:    "test-chat-id",
		ParseMode: "Markdown",
	}

	logger := logrus.New()
	client := NewTelegramClient(config, logger)

	if client == nil {
		t.Error("Expected client to be created, got nil")
		return // Exit early if client is nil to avoid nil pointer dereference
	}

	if client.config != config {
		t.Error("Expected client config to match provided config")
	}

	if client.logger != logger {
		t.Error("Expected client logger to match provided logger")
	}

	if client.client == nil {
		t.Error("Expected HTTP client to be initialized")
	}
}

func TestTelegramMessageFormatting(t *testing.T) {
	config := &TelegramConfig{
		Enabled:   false, // Disabled for testing
		BotToken:  "test-token",
		ChatID:    "test-chat-id",
		ParseMode: "Markdown",
	}

	logger := logrus.New()
	client := NewTelegramClient(config, logger)

	// Test namespace deletion message
	err := client.SendNamespaceDeleted("test-namespace", 2*time.Hour)
	if err != nil {
		t.Errorf("Expected no error for disabled client, got %v", err)
	}

	// Test Helm release deletion message
	err = client.SendHelmReleaseDeleted("test-release", "test-namespace")
	if err != nil {
		t.Errorf("Expected no error for disabled client, got %v", err)
	}

	// Test cleanup summary message
	err = client.SendCleanupSummary(10, 3, 30*time.Second)
	if err != nil {
		t.Errorf("Expected no error for disabled client, got %v", err)
	}

	// Test error message
	err = client.SendError("Test error", &testError{message: "test error message"})
	if err != nil {
		t.Errorf("Expected no error for disabled client, got %v", err)
	}

	// Test startup message
	err = client.SendStartupMessage()
	if err != nil {
		t.Errorf("Expected no error for disabled client, got %v", err)
	}
}

func TestTelegramClientWithEmptyConfig(t *testing.T) {
	config := &TelegramConfig{
		Enabled:   true,
		BotToken:  "", // Empty token
		ChatID:    "", // Empty chat ID
		ParseMode: "Markdown",
		Notifications: TelegramNotifications{
			Startup:            true,
			NamespaceDeleted:   true,
			HelmReleaseDeleted: true,
			CleanupSummary:     true,
			Errors:             true,
		},
	}

	logger := logrus.New()
	client := NewTelegramClient(config, logger)

	// Should not send message due to empty token/chat ID
	err := client.SendMessage("test message")
	if err != nil {
		t.Errorf("Expected no error for empty config, got %v", err)
	}
}

func TestTelegramClientWithDisabledNotifications(t *testing.T) {
	config := &TelegramConfig{
		Enabled:   true,
		BotToken:  "test-token",
		ChatID:    "test-chat-id",
		ParseMode: "Markdown",
		Notifications: TelegramNotifications{
			Startup:            false, // Disabled
			NamespaceDeleted:   false, // Disabled
			HelmReleaseDeleted: false, // Disabled
			CleanupSummary:     false, // Disabled
			Errors:             false, // Disabled
		},
	}

	logger := logrus.New()
	client := NewTelegramClient(config, logger)

	// All notification methods should return nil without sending
	err := client.SendStartupMessage()
	if err != nil {
		t.Errorf("Expected no error for disabled startup notification, got %v", err)
	}

	err = client.SendNamespaceDeleted("test-ns", time.Hour)
	if err != nil {
		t.Errorf("Expected no error for disabled namespace notification, got %v", err)
	}

	err = client.SendHelmReleaseDeleted("test-release", "test-ns")
	if err != nil {
		t.Errorf("Expected no error for disabled helm release notification, got %v", err)
	}

	err = client.SendCleanupSummary(10, 3, time.Minute)
	if err != nil {
		t.Errorf("Expected no error for disabled cleanup summary notification, got %v", err)
	}

	err = client.SendError("test error", &testError{message: "test"})
	if err != nil {
		t.Errorf("Expected no error for disabled error notification, got %v", err)
	}
}

// Helper struct for testing error messages
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
