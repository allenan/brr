//go:build !darwin && !linux

package preflight

import (
	"context"
	"fmt"
)

func detectGateway(ctx context.Context) (string, error) {
	return "", fmt.Errorf("gateway detection not supported on this platform")
}
