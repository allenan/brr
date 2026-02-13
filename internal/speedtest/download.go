package speedtest

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// MeasureDownload measures download speed using the configured sequence.
func MeasureDownload(ctx context.Context, client *http.Client, cfg Config, measID string, onSample func(Sample)) (*PhaseResult, error) {
	var totalBytes atomic.Int64
	sem := make(chan struct{}, cfg.MaxConnections)

	// Build the work queue
	type job struct {
		bytes int
	}
	var jobs []job
	for _, spec := range cfg.DownloadSequence {
		for i := 0; i < spec.Count; i++ {
			jobs = append(jobs, job{bytes: spec.Bytes})
		}
	}

	// Sampling goroutine
	var samples []Sample
	sampleCtx, sampleCancel := context.WithCancel(ctx)
	sampleDone := make(chan struct{})

	go func() {
		defer close(sampleDone)
		ticker := time.NewTicker(cfg.SampleInterval)
		defer ticker.Stop()
		lastBytes := int64(0)
		lastTime := time.Now()

		for {
			select {
			case <-sampleCtx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				currentBytes := totalBytes.Load()
				deltaBytes := currentBytes - lastBytes
				deltaSeconds := now.Sub(lastTime).Seconds()

				if deltaSeconds > 0 && deltaBytes > 0 {
					mbps := float64(deltaBytes*8) / deltaSeconds / 1e6
					s := Sample{Timestamp: now, Mbps: mbps}
					samples = append(samples, s)
					if onSample != nil {
						onSample(s)
					}
				}

				lastBytes = currentBytes
				lastTime = now
			}
		}
	}()

	// Worker goroutines
	errCh := make(chan error, len(jobs))
	doneCh := make(chan struct{}, len(jobs))

	for _, j := range jobs {
		select {
		case <-ctx.Done():
			sampleCancel()
			<-sampleDone
			return nil, ctx.Err()
		case sem <- struct{}{}:
		}

		go func(size int) {
			defer func() { <-sem }()

			url := fmt.Sprintf("%s/__down?bytes=%d", baseURL, size)
			if measID != "" {
				url += "&measId=" + measID
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				errCh <- err
				doneCh <- struct{}{}
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				errCh <- err
				doneCh <- struct{}{}
				return
			}
			defer resp.Body.Close()

			buf := make([]byte, 64*1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					totalBytes.Add(int64(n))
				}
				if err != nil {
					break
				}
			}
			doneCh <- struct{}{}
		}(j.bytes)
	}

	// Wait for all jobs
	for i := 0; i < len(jobs); i++ {
		<-doneCh
	}

	sampleCancel()
	<-sampleDone

	// Discard first 2 seconds of samples for warmup
	filtered := discardWarmup(samples, 2*time.Second)

	mbps := 0.0
	if len(filtered) > 0 {
		vals := make([]float64, len(filtered))
		for i, s := range filtered {
			vals[i] = s.Mbps
		}
		mbps = Percentile(vals, 0.90)
	}

	return &PhaseResult{Mbps: mbps, Samples: samples}, nil
}

// discardWarmup removes samples from the first `dur` of the test.
func discardWarmup(samples []Sample, dur time.Duration) []Sample {
	if len(samples) == 0 {
		return samples
	}
	cutoff := samples[0].Timestamp.Add(dur)
	for i, s := range samples {
		if !s.Timestamp.Before(cutoff) {
			return samples[i:]
		}
	}
	// If all samples are within warmup, return the last half
	if len(samples) > 2 {
		return samples[len(samples)/2:]
	}
	return samples
}
