package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type HelmClient struct {
	actionConfig *action.Configuration
	logger       *logrus.Logger
}

type HelmRelease struct {
	Name      string
	Namespace string
	Status    string
	Version   int
}

func NewHelmClient() (*HelmClient, error) {
	logger := logrus.New()

	// Initialize Helm action configuration
	actionConfig := new(action.Configuration)

	// Get Kubernetes config
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fallback to kubeconfig
		config, err = getKubeConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get Kubernetes config: %v", err)
		}
	}

	// Create Kubernetes client (not used directly but required for Helm)
	_, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Initialize action configuration
	if err := actionConfig.Init(cli.New().RESTClientGetter(), "", "secrets", func(format string, v ...interface{}) {
		logger.Debugf(format, v...)
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action config: %v", err)
	}

	return &HelmClient{
		actionConfig: actionConfig,
		logger:       logger,
	}, nil
}

func (hc *HelmClient) ListReleases(namespace string) ([]HelmRelease, error) {
	// Create list action
	listAction := action.NewList(hc.actionConfig)
	listAction.AllNamespaces = false
	listAction.StateMask = action.ListAll

	// List releases
	releases, err := listAction.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list Helm releases: %v", err)
	}

	var helmReleases []HelmRelease
	for _, release := range releases {
		// Filter by namespace
		if release.Namespace == namespace {
			helmReleases = append(helmReleases, HelmRelease{
				Name:      release.Name,
				Namespace: release.Namespace,
				Status:    string(release.Info.Status),
				Version:   release.Version,
			})
		}
	}

	return helmReleases, nil
}

func (hc *HelmClient) UninstallRelease(releaseName, namespace string, timeout time.Duration) error {
	// Create uninstall action
	uninstallAction := action.NewUninstall(hc.actionConfig)
	uninstallAction.Timeout = timeout
	uninstallAction.Wait = true

	// Uninstall release
	_, err := uninstallAction.Run(releaseName)
	if err != nil {
		return fmt.Errorf("failed to uninstall Helm release %s: %v", releaseName, err)
	}

	hc.logger.Infof("Successfully uninstalled Helm release: %s", releaseName)
	return nil
}

func (hc *HelmClient) GetReleaseStatus(releaseName, namespace string) (*release.Release, error) {
	// Create get action
	getAction := action.NewGet(hc.actionConfig)

	// Get release
	release, err := getAction.Run(releaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to get Helm release %s: %v", releaseName, err)
	}

	return release, nil
}

func getKubeConfig() (*rest.Config, error) {
	// Try to get kubeconfig from environment
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	// Build config from kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig: %v", err)
	}

	return config, nil
}
