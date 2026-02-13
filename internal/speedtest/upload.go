package speedtest

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// countingReader wraps a reader and counts bytes read, adding them to an atomic counter.
type countingReader struct {
	reader  io.Reader
	counter *atomic.Int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	if n > 0 {
		c.counter.Add(int64(n))
	}
	return n, err
}

// MeasureUpload measures upload speed using the configured sequence.
func MeasureUpload(ctx context.Context, client *http.Client, cfg Config, measID string, onSample func(Sample)) (*PhaseResult, error) {
	var totalBytes atomic.Int64
	sem := make(chan struct{}, cfg.MaxConnections)

	// Pre-generate a 1MB random buffer to reuse
	randomBuf := make([]byte, 1_000_000)
	if _, err := rand.Read(randomBuf); err != nil {
		return nil, fmt.Errorf("generating random data: %w", err)
	}

	type job struct {
		bytes int
	}
	var jobs []job
	for _, spec := range cfg.UploadSequence {
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

			// Build upload body from the random buffer
			var body []byte
			if size <= len(randomBuf) {
				body = randomBuf[:size]
			} else {
				body = make([]byte, size)
				for i := 0; i < size; i += len(randomBuf) {
					copy(body[i:], randomBuf)
				}
			}

			url := baseURL + "/__up"
			if measID != "" {
				url += "?measId=" + measID
			}

			reader := &countingReader{
				reader:  bytes.NewReader(body),
				counter: &totalBytes,
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reader)
			if err != nil {
				doneCh <- struct{}{}
				return
			}
			req.ContentLength = int64(size)
			req.Header.Set("Content-Type", "application/octet-stream")

			resp, err := client.Do(req)
			if err != nil {
				doneCh <- struct{}{}
				return
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			doneCh <- struct{}{}
		}(j.bytes)
	}

	for i := 0; i < len(jobs); i++ {
		<-doneCh
	}

	sampleCancel()
	<-sampleDone

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
