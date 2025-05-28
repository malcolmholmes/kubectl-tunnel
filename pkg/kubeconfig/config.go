package kubeconfig

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type KubeConfig struct {
	rawConfig api.Config
	config    *api.Config
}

// New creates a new KubeConfig instance
func New() (*KubeConfig, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting raw config: %w", err)
	}

	return &KubeConfig{
		rawConfig: rawConfig,
	}, nil
}

// GetCurrentContext returns the name of the current context
func (k *KubeConfig) GetCurrentContext() (string, error) {
	if k.rawConfig.CurrentContext == "" {
		return "", fmt.Errorf("no current context found")
	}
	return k.rawConfig.CurrentContext, nil
}

// GetServerURL returns the API server URL for the current context
func (k *KubeConfig) GetServerURL() (string, string, error) {
	ctxName := k.rawConfig.CurrentContext
	if ctxName == "" {
		return "", "", fmt.Errorf("no current context")
	}

	ctx, exists := k.rawConfig.Contexts[ctxName]
	if !exists {
		return "", "", fmt.Errorf("context %s not found", ctxName)
	}

	cluster, exists := k.rawConfig.Clusters[ctx.Cluster]
	if !exists {
		return "", "", fmt.Errorf("cluster %s not found", ctx.Cluster)
	}

	return cluster.Server, ctx.Cluster, nil
}

// CreateTempConfig creates a temporary kubeconfig with the given server URL
func (k *KubeConfig) CreateTempConfig(serverURL string) (string, func(), error) {
	// Create a deep copy of the config
	config := k.rawConfig.DeepCopy()

	// Update the server URL for the current context
	ctxName := config.CurrentContext
	if ctxName == "" {
		return "", nil, fmt.Errorf("no current context")
	}

	ctx, exists := config.Contexts[ctxName]
	if !exists {
		return "", nil, fmt.Errorf("context %s not found", ctxName)
	}

	cluster, exists := config.Clusters[ctx.Cluster]
	if !exists {
		return "", nil, fmt.Errorf("cluster %s not found", ctx.Cluster)
	}

	// Update the server URL
	cluster.Server = serverURL

	// Create a temp file for the kubeconfig
	tempFile, err := os.CreateTemp("", "kubeconfig-*")
	if err != nil {
		return "", nil, fmt.Errorf("error creating temp file: %w", err)
	}

	cleanup := func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}

	// Write the modified config to the temp file
	if err := clientcmd.WriteToFile(*config, tempFile.Name()); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("error writing temp kubeconfig: %w", err)
	}

	return tempFile.Name(), cleanup, nil
}
