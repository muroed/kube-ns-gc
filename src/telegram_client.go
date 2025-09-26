package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type TelegramConfig struct {
	Enabled       bool                  `json:"enabled"`
	BotToken      string                `json:"bot_token"`
	ChatID        string                `json:"chat_id"`
	ParseMode     string                `json:"parse_mode"`
	Notifications TelegramNotifications `json:"notifications"`
}

type TelegramNotifications struct {
	Startup            bool `json:"startup"`
	NamespaceDeleted   bool `json:"namespace_deleted"`
	HelmReleaseDeleted bool `json:"helm_release_deleted"`
	CleanupSummary     bool `json:"cleanup_summary"`
	Errors             bool `json:"errors"`
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type TelegramClient struct {
	config *TelegramConfig
	logger *logrus.Logger
	client *http.Client
}

func NewTelegramClient(config *TelegramConfig, logger *logrus.Logger) *TelegramClient {
	return &TelegramClient{
		config: config,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (tc *TelegramClient) SendMessage(text string) error {
	if tc.config == nil {
		tc.logger.Warn("Telegram config is nil")
		return nil
	}

	if !tc.config.Enabled {
		tc.logger.Debug("Telegram notifications are disabled")
		return nil
	}

	if tc.config.BotToken == "" || tc.config.ChatID == "" {
		tc.logger.Warn("Telegram bot token or chat ID is not configured")
		return nil
	}

	message := TelegramMessage{
		ChatID:    tc.config.ChatID,
		Text:      text,
		ParseMode: tc.config.ParseMode,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %v", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tc.config.BotToken)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create telegram request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := tc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	tc.logger.Debug("Telegram message sent successfully")
	return nil
}

func (tc *TelegramClient) SendNamespaceDeleted(namespace string, age time.Duration) error {
	if tc.config == nil || !tc.config.Notifications.NamespaceDeleted {
		tc.logger.Debug("Namespace deletion notifications are disabled")
		return nil
	}

	text := fmt.Sprintf("🗑️ *Namespace Deleted*\n\n"+
		"📦 Namespace: `%s`\n"+
		"⏰ Age: %s\n"+
		"🕐 Time: %s",
		namespace,
		age.Round(time.Minute),
		time.Now().Format("2006-01-02 15:04:05 MST"))

	return tc.SendMessage(text)
}

func (tc *TelegramClient) SendHelmReleaseDeleted(releaseName, namespace string) error {
	if tc.config == nil || !tc.config.Notifications.HelmReleaseDeleted {
		tc.logger.Debug("Helm release deletion notifications are disabled")
		return nil
	}

	text := fmt.Sprintf("🧹 *Helm Release Deleted*\n\n"+
		"📦 Release: `%s`\n"+
		"🏠 Namespace: `%s`\n"+
		"🕐 Time: %s",
		releaseName,
		namespace,
		time.Now().Format("2006-01-02 15:04:05 MST"))

	return tc.SendMessage(text)
}

func (tc *TelegramClient) SendCleanupSummary(totalNamespaces, cleanedNamespaces int, duration time.Duration) error {
	if tc.config == nil || !tc.config.Notifications.CleanupSummary {
		tc.logger.Debug("Cleanup summary notifications are disabled")
		return nil
	}

	text := fmt.Sprintf("📊 *Cleanup Summary*\n\n"+
		"🔍 Total namespaces checked: %d\n"+
		"🗑️ Namespaces deleted: %d\n"+
		"⏱️ Cleanup duration: %s\n"+
		"🕐 Time: %s",
		totalNamespaces,
		cleanedNamespaces,
		duration.Round(time.Second),
		time.Now().Format("2006-01-02 15:04:05 MST"))

	return tc.SendMessage(text)
}

func (tc *TelegramClient) SendError(message string, err error) error {
	if tc.config == nil || !tc.config.Notifications.Errors {
		tc.logger.Debug("Error notifications are disabled")
		return nil
	}

	text := fmt.Sprintf("❌ *Error*\n\n"+
		"📝 Message: %s\n"+
		"🔍 Error: `%s`\n"+
		"🕐 Time: %s",
		message,
		err.Error(),
		time.Now().Format("2006-01-02 15:04:05 MST"))

	return tc.SendMessage(text)
}

func (tc *TelegramClient) SendStartupMessage() error {
	if tc.config == nil || !tc.config.Notifications.Startup {
		tc.logger.Debug("Startup notifications are disabled")
		return nil
	}

	text := fmt.Sprintf("🚀 *kube-ns-gc Started*\n\n"+
		"🕐 Time: %s\n"+
		"📋 Service is now monitoring namespaces for cleanup",
		time.Now().Format("2006-01-02 15:04:05 MST"))

	return tc.SendMessage(text)
}
