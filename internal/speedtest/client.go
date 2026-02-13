package speedtest

import (
	"crypto/tls"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://speed.cloudflare.com"

// NewHTTPClient creates an HTTP/2 client optimized for speed testing.
func NewHTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 16,
		MaxConnsPerHost:     0, // unlimited
		DisableCompression:  true,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		IdleConnTimeout:     90 * time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}
}

// parseServerTiming extracts cfRequestDuration from the Server-Timing header.
// Returns the duration in milliseconds, or 0 if not found.
func parseServerTiming(header string) float64 {
	// Format: cfRequestDuration;dur=X.Y
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "cfRequestDuration") {
			for _, kv := range strings.Split(part, ";") {
				kv = strings.TrimSpace(kv)
				if strings.HasPrefix(kv, "dur=") {
					val, err := strconv.ParseFloat(strings.TrimPrefix(kv, "dur="), 64)
					if err == nil {
						return val
					}
				}
			}
		}
	}
	return 0
}
