package sshtunnel

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"k8s.io/klog/v2"
)

type SSHTunnel struct {
	localAddr  string
	LocalPort  int // Exported so it can be read by other packages
	remoteAddr string
	remotePort int
	remoteUser string
	cmd        *exec.Cmd
}

// New creates a new SSH tunnel
func New(remoteAddr string, remotePort int) *SSHTunnel {
	// Create a listener to find a free port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		// If we can't find a free port, fall back to 0 (let OS choose)
		klog.Warningf("Failed to find free port, letting OS choose: %v", err)
		return &SSHTunnel{
			localAddr:  "localhost",
			LocalPort:  0, // 0 means let OS choose
			remoteAddr: remoteAddr,
			remotePort: remotePort,
		}
	}
	// Get the port before closing the listener
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	return &SSHTunnel{
		localAddr:  "localhost",
		LocalPort:  port,
		remoteAddr: remoteAddr,
		remotePort: remotePort,
	}
}

// Start starts the SSH tunnel
func (t *SSHTunnel) Start() error {
	if t.cmd != nil {
		return fmt.Errorf("tunnel is already running")
	}

	t.cmd = exec.Command("ssh", "-NfL", fmt.Sprintf("%s:%d:localhost:%d", t.localAddr, t.LocalPort, t.remotePort), t.remoteAddr)
	t.cmd.Stdout = os.Stdout
	t.cmd.Stderr = os.Stderr
	klog.Infof("Starting SSH tunnel: %s -> %s", t.localAddr, t.remoteAddr)
	if err := t.cmd.Start(); err != nil {
		return fmt.Errorf("error starting SSH tunnel: %w", err)
	}

	// Check if the process started successfully
	if t.cmd.Process == nil {
		return fmt.Errorf("failed to start SSH tunnel: process is nil")
	}

	time.Sleep(time.Second)
	return nil
}

// Stop stops the SSH tunnel
func (t *SSHTunnel) Stop() error {
	if t.cmd == nil || t.cmd.Process == nil {
		return nil
	}

	klog.Info("Stopping SSH tunnel")
	if err := t.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("error sending SIGTERM to SSH tunnel: %w", err)
	}

	// Wait for the process to exit
	if err := t.cmd.Wait(); err != nil {
		// Ignore process already finished error
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != -1 {
			return fmt.Errorf("error waiting for SSH tunnel to exit: %w", err)
		}
	}

	t.cmd = nil
	return nil
}

// ParseServerURL parses the server URL and returns host and port
func ParseServerURL(serverURL string) (string, int, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", 0, fmt.Errorf("error parsing server URL: %w", err)
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return "", 0, fmt.Errorf("error parsing port: %w", err)
	}

	return host, portInt, nil
}
