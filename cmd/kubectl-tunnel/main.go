package main

import (
	"os"

	"github.com/yourusername/kubectl-tunnel/pkg/cmd"
	"k8s.io/klog/v2"
)

func main() {
	if err := cmd.Run(); err != nil {
		klog.ErrorS(err, "Error executing command")
		os.Exit(1)
	}
}
