package preflight

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/allenan/brr/internal/speedtest"
)

// CheckName identifies a preflight check.
type CheckName string

const (
	CheckGateway    CheckName = "gateway"
	CheckInternet   CheckName = "internet"
	CheckDNS        CheckName = "dns"
	CheckTestServer CheckName = "server"
)

// CheckResult is the outcome of a single preflight check.
type CheckResult struct {
	Name    CheckName
	Passed  bool
	Detail  string  // e.g. "192.168.1.1", "Ashburn, VA"
	Latency float64 // ms, 0 if failed
	Err     error
}

// Result is the aggregate outcome of all preflight checks.
type Result struct {
	Checks  []CheckResult
	Passed  bool   // true if test server reachable
	Message string // diagnostic if failed
}

// OnCheck is called after each individual check completes.
type OnCheck func(CheckResult)

// Run executes all 4 preflight checks sequentially, calling onCheck after each.
func Run(ctx context.Context, client *http.Client, onCheck OnCheck) *Result {
	var checks []CheckResult

	// 1. Gateway
	gwResult := checkGateway(ctx)
	checks = append(checks, gwResult)
	onCheck(gwResult)

	// 2. Internet
	inetResult := checkInternet(ctx)
	checks = append(checks, inetResult)
	onCheck(inetResult)

	// 3. DNS
	dnsResult := checkDNS(ctx)
	checks = append(checks, dnsResult)
	onCheck(dnsResult)

	// 4. Test server
	serverResult := checkTestServer(ctx, client)
	checks = append(checks, serverResult)
	onCheck(serverResult)

	result := &Result{Checks: checks}

	if serverResult.Passed {
		result.Passed = true
		return result
	}

	result.Message = diagnose(checks)
	return result
}

func checkGateway(ctx context.Context) CheckResult {
	gw, err := detectGateway(ctx)
	if err != nil || gw == "" {
		return CheckResult{
			Name: CheckGateway,
			Err:  err,
		}
	}

	start := time.Now()
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(gw, "80"))
	latency := time.Since(start).Seconds() * 1000

	if err != nil {
		return CheckResult{
			Name:   CheckGateway,
			Detail: gw,
			Err:    err,
		}
	}
	conn.Close()

	return CheckResult{
		Name:    CheckGateway,
		Passed:  true,
		Detail:  gw,
		Latency: latency,
	}
}

func checkInternet(ctx context.Context) CheckResult {
	start := time.Now()
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", "1.1.1.1:443")
	latency := time.Since(start).Seconds() * 1000

	if err != nil {
		return CheckResult{
			Name:   CheckInternet,
			Detail: "1.1.1.1",
			Err:    err,
		}
	}
	conn.Close()

	return CheckResult{
		Name:    CheckInternet,
		Passed:  true,
		Detail:  "1.1.1.1",
		Latency: latency,
	}
}

func checkDNS(ctx context.Context) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	start := time.Now()
	addrs, err := net.DefaultResolver.LookupHost(ctx, "speed.cloudflare.com")
	latency := time.Since(start).Seconds() * 1000

	if err != nil || len(addrs) == 0 {
		return CheckResult{
			Name: CheckDNS,
			Err:  err,
		}
	}

	return CheckResult{
		Name:    CheckDNS,
		Passed:  true,
		Detail:  addrs[0],
		Latency: latency,
	}
}

func checkTestServer(ctx context.Context, client *http.Client) CheckResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://speed.cloudflare.com/cdn-cgi/trace", nil)
	if err != nil {
		return CheckResult{
			Name: CheckTestServer,
			Err:  err,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Name: CheckTestServer,
			Err:  err,
		}
	}
	defer resp.Body.Close()
	latency := time.Since(start).Seconds() * 1000

	// Parse colo from response
	var colo string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && parts[0] == "colo" {
			colo = parts[1]
			break
		}
	}

	detail := colo
	if city := speedtest.ColoName(colo); city != colo {
		detail = city
	}

	return CheckResult{
		Name:    CheckTestServer,
		Passed:  true,
		Detail:  detail,
		Latency: latency,
	}
}

func diagnose(checks []CheckResult) string {
	gw := checks[0]
	inet := checks[1]
	dns := checks[2]

	if gw.Detail == "" && !gw.Passed {
		return "Could not detect a default gateway — are you connected to a network?"
	}

	if !gw.Passed {
		return fmt.Sprintf("Can't reach your router (%s) — check your WiFi or ethernet connection", gw.Detail)
	}

	if !inet.Passed {
		return "Your router is reachable but the internet connection appears down. This is likely an ISP issue."
	}

	if !dns.Passed {
		return "DNS resolution failed — try using 1.1.1.1 or 8.8.8.8 as your DNS server"
	}

	return "Can't reach speed.cloudflare.com — check firewall settings or try again later"
}
