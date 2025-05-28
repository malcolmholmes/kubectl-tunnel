# kubectl-tunnel

A kubectl plugin that creates an SSH tunnel to your Kubernetes API server and forwards traffic through it.

## Features

- Automatically detects the current Kubernetes context
- Creates an SSH tunnel to the API server
- Executes kubectl commands through the tunnel
- Cleans up the tunnel when done

## Installation

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/kubectl-tunnel.git
   cd kubectl-tunnel
   ```

2. Build the plugin:
   ```bash
   go build -o kubectl-tunnel ./cmd/kubectl-tunnel
   ```

3. Make it executable:
   ```bash
   chmod +x kubectl-tunnel
   ```

4. Move it to a directory in your PATH:
   ```bash
   sudo mv kubectl-tunnel /usr/local/bin/kubectl-tunnel
   ```

### Using go install

```bash
go install github.com/yourusername/kubectl-tunnel/cmd/kubectl-tunnel@latest
```

## Usage

```bash
# Basic usage
kubectl tunnel get pods

# With any kubectl command
kubectl tunnel get nodes
kubectl tunnel get pods -n kube-system
kubectl tunnel apply -f deployment.yaml

# Keep the tunnel open without running a command
kubectl tunnel
# This will give you a KUBECONFIG export. Use this in another terminal for any k8s access. The
# tunnel is removed when you terminate the kubectl tunnel process.
```

## How It Works

1. Reads your current kubeconfig
2. Extracts the API server URL from the current context
3. Creates an SSH tunnel from 127.0.0.2:<random-port> to the API server
4. Creates a temporary kubeconfig using the local tunnel
5. Executes the specified kubectl command through the tunnel
6. Cleans up the tunnel when done

## Requirements

- Go 1.21 or later
- kubectl installed and configured
- SSH access to the Kubernetes API server
- Proper SSH configuration in ~/.ssh/config for the API server

## Note

This plugin does not accept config parameters for SSH access. These can be provided via your `~/.ssh/config` file. The hostname will be taken from the `CurrentContext` in your kubeconfig file.

## Development

To build and test changes:

```bash
# Build
make build

# Run tests
make test

# Install to $GOPATH/bin
make install
```
