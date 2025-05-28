package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/yourusername/kubectl-tunnel/pkg/kubeconfig"
	"github.com/yourusername/kubectl-tunnel/pkg/sshtunnel"
	"k8s.io/klog/v2"
)

// Run executes the main command
func Run() error {
	// Initialize klog
	klog.InitFlags(nil)
	// Load kubeconfig
	kubeConfig, err := kubeconfig.New()
	if err != nil {
		return fmt.Errorf("error loading kubeconfig: %w", err)
	}

	// Get current context and server URL
	ctxName, err := kubeConfig.GetCurrentContext()
	if err != nil {
		return err
	}
	klog.Infof("Using context: %s", ctxName)

	serverURL, _, err := kubeConfig.GetServerURL()
	if err != nil {
		return fmt.Errorf("error getting server URL: %w", err)
	}

	// Parse server URL to get host and port
	host, port, err := sshtunnel.ParseServerURL(serverURL)
	if err != nil {
		return fmt.Errorf("error parsing server URL: %w", err)
	}

	// Create SSH tunnel
	tunnel := sshtunnel.New(host, port, "root")

	// Start the tunnel
	if err := tunnel.Start(); err != nil {
		return fmt.Errorf("error starting tunnel: %w", err)
	}
	defer func() {
		if err := tunnel.Stop(); err != nil {
			klog.Errorf("Error stopping tunnel: %v", err)
		}
	}()

	// Create temporary kubeconfig with local tunnel
	localServerURL := fmt.Sprintf("https://localhost:%d", tunnel.LocalPort)
	tempConfigPath, cleanup, err := kubeConfig.CreateTempConfig(localServerURL)
	if err != nil {
		return fmt.Errorf("error creating temp kubeconfig: %w", err)
	}
	defer cleanup()

	// If no command is provided, just keep the tunnel open
	if len(os.Args) < 2 {
		klog.Info("No command provided, keeping tunnel open. Press Ctrl+C to exit.")
		klog.Infof("Use: export KUBECONFIG=%s", tempConfigPath)
		// Wait for interrupt signal
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		return nil
	}

	// Execute the kubectl command with the temporary kubeconfig
	cmdArgs := os.Args[1:]

	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", tempConfigPath))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set up signal forwarding
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		if cmd.Process != nil {
			cmd.Process.Signal(sig)
		}
	}()

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}
