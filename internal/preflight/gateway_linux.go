//go:build linux

package preflight

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func detectGateway(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "ip", "route", "show", "default")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ip route command failed: %w", err)
	}

	// Format: default via 192.168.1.1 dev eth0 ...
	fields := strings.Fields(string(out))
	for i, f := range fields {
		if f == "via" && i+1 < len(fields) {
			return fields[i+1], nil
		}
	}

	return "", fmt.Errorf("no gateway found in ip route output")
}
