package speedtest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MeasureIdleLatency performs count sequential latency probes and reports each via onSample.
func MeasureIdleLatency(ctx context.Context, client *http.Client, count int, onSample func(LatencySample)) (*LatencyResult, error) {
	var samples []LatencySample

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		rtt, err := probeLatency(ctx, client)
		if err != nil {
			continue // skip failed probes
		}

		s := LatencySample{
			Timestamp: time.Now(),
			RTT:       rtt,
		}
		samples = append(samples, s)
		if onSample != nil {
			onSample(s)
		}
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("all latency probes failed")
	}

	return computeLatencyResult(samples), nil
}

// MeasureLoadedLatency runs latency probes in the background at the given interval.
// Returns a cancel function and a channel that receives the result when cancelled.
func MeasureLoadedLatency(ctx context.Context, client *http.Client, interval time.Duration, onSample func(LatencySample)) (cancel func(), resultCh <-chan *LatencyResult) {
	ctx, cancelFn := context.WithCancel(ctx)
	ch := make(chan *LatencyResult, 1)

	go func() {
		var samples []LatencySample
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				if len(samples) > 0 {
					ch <- computeLatencyResult(samples)
				} else {
					ch <- &LatencyResult{}
				}
				return
			case <-ticker.C:
				rtt, err := probeLatency(ctx, client)
				if err != nil {
					continue
				}
				s := LatencySample{
					Timestamp: time.Now(),
					RTT:       rtt,
				}
				samples = append(samples, s)
				if onSample != nil {
					onSample(s)
				}
			}
		}
	}()

	return cancelFn, ch
}

// probeLatency makes a single latency measurement to __down?bytes=0.
func probeLatency(ctx context.Context, client *http.Client) (float64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/__down?bytes=0", nil)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	elapsed := time.Since(start).Seconds() * 1000 // ms

	serverTime := parseServerTiming(resp.Header.Get("Server-Timing"))
	rtt := elapsed - serverTime
	if rtt < 0 {
		rtt = elapsed
	}

	return rtt, nil
}

func computeLatencyResult(samples []LatencySample) *LatencyResult {
	rtts := make([]float64, len(samples))
	for i, s := range samples {
		rtts[i] = s.RTT
	}

	result := &LatencyResult{
		Samples: samples,
		Avg:     mean(rtts),
		Jitter:  Jitter(rtts),
	}

	if len(rtts) > 0 {
		result.Min = rtts[0]
		result.Max = rtts[0]
		for _, v := range rtts {
			if v < result.Min {
				result.Min = v
			}
			if v > result.Max {
				result.Max = v
			}
		}
	}

	return result
}

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
