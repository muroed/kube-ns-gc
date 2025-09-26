package main

import (
	"testing"
	"time"
)

func TestLoadConfigFromEnv(t *testing.T) {
	config := loadConfigFromEnv()

	// Проверяем значения по умолчанию
	if config.CleanupInterval != 24*time.Hour {
		t.Errorf("Expected CleanupInterval to be 24h, got %v", config.CleanupInterval)
	}

	if config.NamespaceMaxAge != 7*24*time.Hour {
		t.Errorf("Expected NamespaceMaxAge to be 168h, got %v", config.NamespaceMaxAge)
	}

	if config.HelmReleaseTimeout != 5*time.Minute {
		t.Errorf("Expected HelmReleaseTimeout to be 5m, got %v", config.HelmReleaseTimeout)
	}

	expectedExcluded := []string{"kube-system", "kube-public", "kube-node-lease", "default"}
	if len(config.ExcludedNamespaces) != len(expectedExcluded) {
		t.Errorf("Expected %d excluded namespaces, got %d", len(expectedExcluded), len(config.ExcludedNamespaces))
	}

	if config.IgnoreLabel != "kube-ns-gc.ignore" {
		t.Errorf("Expected IgnoreLabel to be 'kube-ns-gc.ignore', got %s", config.IgnoreLabel)
	}

	if config.LogLevel != "info" {
		t.Errorf("Expected LogLevel to be 'info', got %s", config.LogLevel)
	}

	if config.Port != 8080 {
		t.Errorf("Expected Port to be 8080, got %d", config.Port)
	}
}

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		key          string
		defaultValue string
		expected     string
	}{
		{"NONEXISTENT_VAR", "default", "default"},
		{"", "default", "default"},
	}

	for _, test := range tests {
		result := getEnvString(test.key, test.defaultValue)
		if result != test.expected {
			t.Errorf("getEnvString(%s, %s) = %s, expected %s", test.key, test.defaultValue, result, test.expected)
		}
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		key          string
		defaultValue int
		expected     int
	}{
		{"NONEXISTENT_VAR", 42, 42},
		{"", 42, 42},
	}

	for _, test := range tests {
		result := getEnvInt(test.key, test.defaultValue)
		if result != test.expected {
			t.Errorf("getEnvInt(%s, %d) = %d, expected %d", test.key, test.defaultValue, result, test.expected)
		}
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		key          string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"NONEXISTENT_VAR", 5 * time.Minute, 5 * time.Minute},
		{"", 5 * time.Minute, 5 * time.Minute},
	}

	for _, test := range tests {
		result := getEnvDuration(test.key, test.defaultValue)
		if result != test.expected {
			t.Errorf("getEnvDuration(%s, %v) = %v, expected %v", test.key, test.defaultValue, result, test.expected)
		}
	}
}

func TestGetEnvStringSlice(t *testing.T) {
	tests := []struct {
		key          string
		defaultValue []string
		expected     []string
	}{
		{"NONEXISTENT_VAR", []string{"a", "b"}, []string{"a", "b"}},
		{"", []string{"a", "b"}, []string{"a", "b"}},
	}

	for _, test := range tests {
		result := getEnvStringSlice(test.key, test.defaultValue)
		if len(result) != len(test.expected) {
			t.Errorf("getEnvStringSlice(%s, %v) = %v, expected %v", test.key, test.defaultValue, result, test.expected)
		}
	}
}
