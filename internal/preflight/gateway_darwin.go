//go:build darwin

package preflight

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func detectGateway(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "route", "-n", "get", "default")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("route command failed: %w", err)
	}

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "gateway:") {
			gw := strings.TrimSpace(strings.TrimPrefix(line, "gateway:"))
			if gw != "" {
				return gw, nil
			}
		}
	}

	return "", fmt.Errorf("no gateway found in route output")
}
