package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	CleanupInterval    time.Duration  `json:"cleanup_interval"`
	NamespaceMaxAge    time.Duration  `json:"namespace_max_age"`
	HelmReleaseTimeout time.Duration  `json:"helm_release_timeout"`
	ExcludedNamespaces []string       `json:"excluded_namespaces"`
	IgnoreLabel        string         `json:"ignore_label"`
	LogLevel           string         `json:"log_level"`
	Port               int            `json:"port"`
	Telegram           TelegramConfig `json:"telegram"`
}

type NamespaceGC struct {
	config         *Config
	clientset      *kubernetes.Clientset
	logger         *logrus.Logger
	helmClient     *HelmClient
	telegramClient *TelegramClient
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration from ConfigMap
	config, err := loadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Initialize Kubernetes client
	clientset, err := initKubernetesClient()
	if err != nil {
		logger.Fatalf("Failed to initialize Kubernetes client: %v", err)
	}

	// Initialize Helm client
	helmClient, err := NewHelmClient()
	if err != nil {
		logger.Fatalf("Failed to initialize Helm client: %v", err)
	}

	// Initialize Telegram client
	telegramClient := NewTelegramClient(&config.Telegram, logger)

	// Create namespace garbage collector
	gc := &NamespaceGC{
		config:         config,
		clientset:      clientset,
		logger:         logger,
		helmClient:     helmClient,
		telegramClient: telegramClient,
	}

	// Setup HTTP server for health checks
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	router.GET("/metrics", gc.getMetrics)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}

	// Start HTTP server
	go func() {
		logger.Infof("Starting HTTP server on port %d", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Send startup notification
	if gc.telegramClient != nil {
		if err := gc.telegramClient.SendStartupMessage(); err != nil {
			logger.Warnf("Failed to send startup notification: %v", err)
		}
	}

	// Start cleanup routine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go gc.startCleanupRoutine(ctx)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down...")
	cancel()

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func loadConfig() (*Config, error) {
	// Try to load from ConfigMap first
	configData, err := os.ReadFile("/etc/config/config.json")
	if err != nil {
		// Fallback to environment variables
		return loadConfigFromEnv(), nil
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

func loadConfigFromEnv() *Config {
	return &Config{
		CleanupInterval:    getEnvDuration("CLEANUP_INTERVAL", 24*time.Hour),
		NamespaceMaxAge:    getEnvDuration("NAMESPACE_MAX_AGE", 7*24*time.Hour),
		HelmReleaseTimeout: getEnvDuration("HELM_RELEASE_TIMEOUT", 5*time.Minute),
		ExcludedNamespaces: getEnvStringSlice("EXCLUDED_NAMESPACES", []string{"kube-system", "kube-public", "kube-node-lease", "default"}),
		IgnoreLabel:        getEnvString("IGNORE_LABEL", "kube-ns-gc.ignore"),
		LogLevel:           getEnvString("LOG_LEVEL", "info"),
		Port:               getEnvInt("PORT", 8080),
		Telegram: TelegramConfig{
			Enabled:   getEnvBool("TELEGRAM_ENABLED", false),
			BotToken:  getEnvString("TELEGRAM_BOT_TOKEN", ""),
			ChatID:    getEnvString("TELEGRAM_CHAT_ID", ""),
			ParseMode: getEnvString("TELEGRAM_PARSE_MODE", "Markdown"),
			Notifications: TelegramNotifications{
				Startup:            getEnvBool("TELEGRAM_NOTIFY_STARTUP", true),
				NamespaceDeleted:   getEnvBool("TELEGRAM_NOTIFY_NAMESPACE_DELETED", true),
				HelmReleaseDeleted: getEnvBool("TELEGRAM_NOTIFY_HELM_RELEASE_DELETED", true),
				CleanupSummary:     getEnvBool("TELEGRAM_NOTIFY_CLEANUP_SUMMARY", true),
				Errors:             getEnvBool("TELEGRAM_NOTIFY_ERRORS", true),
			},
		},
	}
}

func initKubernetesClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return clientset, nil
}

func (gc *NamespaceGC) startCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(gc.config.CleanupInterval)
	defer ticker.Stop()

	// Run initial cleanup
	gc.performCleanup()

	for {
		select {
		case <-ctx.Done():
			gc.logger.Info("Cleanup routine stopped")
			return
		case <-ticker.C:
			gc.performCleanup()
		}
	}
}

func (gc *NamespaceGC) performCleanup() {
	startTime := time.Now()
	gc.logger.Info("Starting namespace cleanup")

	// Get all namespaces
	namespaces, err := gc.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		gc.logger.Errorf("Failed to list namespaces: %v", err)
		if gc.telegramClient != nil {
			if err := gc.telegramClient.SendError("Failed to list namespaces", err); err != nil {
				gc.logger.Warnf("Failed to send error notification: %v", err)
			}
		}
		return
	}

	cutoffTime := time.Now().Add(-gc.config.NamespaceMaxAge)
	cleanedCount := 0

	for _, ns := range namespaces.Items {
		// Check if namespace should be excluded
		if gc.shouldExcludeNamespace(&ns) {
			gc.logger.Debugf("Skipping excluded namespace: %s", ns.Name)
			continue
		}

		// Check if namespace has ignore label
		if gc.hasIgnoreLabel(&ns) {
			gc.logger.Debugf("Skipping namespace with ignore label: %s", ns.Name)
			continue
		}

		// Check if namespace is old enough
		if ns.CreationTimestamp.Time.After(cutoffTime) {
			gc.logger.Debugf("Namespace %s is not old enough (created: %s)", ns.Name, ns.CreationTimestamp.Time)
			continue
		}

		// Clean up Helm releases first
		if err := gc.cleanupHelmReleases(ns.Name); err != nil {
			gc.logger.Errorf("Failed to cleanup Helm releases in namespace %s: %v", ns.Name, err)
			if gc.telegramClient != nil {
				if err := gc.telegramClient.SendError(fmt.Sprintf("Failed to cleanup Helm releases in namespace %s", ns.Name), err); err != nil {
					gc.logger.Warnf("Failed to send error notification: %v", err)
				}
			}
			continue
		}

		// Delete namespace
		if err := gc.deleteNamespace(ns.Name); err != nil {
			gc.logger.Errorf("Failed to delete namespace %s: %v", ns.Name, err)
			if gc.telegramClient != nil {
				if err := gc.telegramClient.SendError(fmt.Sprintf("Failed to delete namespace %s", ns.Name), err); err != nil {
					gc.logger.Warnf("Failed to send error notification: %v", err)
				}
			}
			continue
		}

		// Send notification about deleted namespace
		namespaceAge := time.Since(ns.CreationTimestamp.Time)
		if gc.telegramClient != nil {
			if err := gc.telegramClient.SendNamespaceDeleted(ns.Name, namespaceAge); err != nil {
				gc.logger.Warnf("Failed to send namespace deletion notification: %v", err)
			}
		}

		cleanedCount++
		gc.logger.Infof("Successfully cleaned up namespace: %s", ns.Name)
	}

	duration := time.Since(startTime)
	gc.logger.Infof("Cleanup completed. Cleaned %d namespaces", cleanedCount)

	// Send cleanup summary
	if gc.telegramClient != nil {
		if err := gc.telegramClient.SendCleanupSummary(len(namespaces.Items), cleanedCount, duration); err != nil {
			gc.logger.Warnf("Failed to send cleanup summary: %v", err)
		}
	}
}

func (gc *NamespaceGC) shouldExcludeNamespace(ns *v1.Namespace) bool {
	for _, excluded := range gc.config.ExcludedNamespaces {
		if ns.Name == excluded {
			return true
		}
	}
	return false
}

func (gc *NamespaceGC) hasIgnoreLabel(ns *v1.Namespace) bool {
	if gc.config.IgnoreLabel == "" {
		return false
	}

	_, exists := ns.Labels[gc.config.IgnoreLabel]
	return exists
}

func (gc *NamespaceGC) cleanupHelmReleases(namespace string) error {
	gc.logger.Debugf("Cleaning up Helm releases in namespace: %s", namespace)

	releases, err := gc.helmClient.ListReleases(namespace)
	if err != nil {
		return fmt.Errorf("failed to list Helm releases: %v", err)
	}

	for _, release := range releases {
		gc.logger.Debugf("Uninstalling Helm release: %s in namespace: %s", release.Name, namespace)

		if err := gc.helmClient.UninstallRelease(release.Name, namespace, gc.config.HelmReleaseTimeout); err != nil {
			gc.logger.Errorf("Failed to uninstall Helm release %s: %v", release.Name, err)
			// Continue with other releases
		} else {
			gc.logger.Infof("Successfully uninstalled Helm release: %s", release.Name)

			// Send notification about deleted Helm release
			if gc.telegramClient != nil {
				if err := gc.telegramClient.SendHelmReleaseDeleted(release.Name, namespace); err != nil {
					gc.logger.Warnf("Failed to send Helm release deletion notification: %v", err)
				}
			}
		}
	}

	return nil
}

func (gc *NamespaceGC) deleteNamespace(name string) error {
	gc.logger.Debugf("Deleting namespace: %s", name)

	// Delete namespace
	err := gc.clientset.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %v", err)
	}

	// Wait for namespace to be deleted
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for namespace deletion")
		case <-ticker.C:
			_, err := gc.clientset.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
			if err != nil {
				// Namespace is deleted
				return nil
			}
		}
	}
}

func (gc *NamespaceGC) getMetrics(c *gin.Context) {
	// Simple metrics endpoint
	namespaces, err := gc.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get metrics"})
		return
	}

	cutoffTime := time.Now().Add(-gc.config.NamespaceMaxAge)
	oldNamespaces := 0

	for _, ns := range namespaces.Items {
		if !gc.shouldExcludeNamespace(&ns) && !gc.hasIgnoreLabel(&ns) && ns.CreationTimestamp.Time.Before(cutoffTime) {
			oldNamespaces++
		}
	}

	c.JSON(200, gin.H{
		"total_namespaces":    len(namespaces.Items),
		"old_namespaces":      oldNamespaces,
		"excluded_namespaces": len(gc.config.ExcludedNamespaces),
		"cleanup_interval":    gc.config.CleanupInterval.String(),
		"namespace_max_age":   gc.config.NamespaceMaxAge.String(),
	})
}

// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
