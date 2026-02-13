package speedtest

import "time"

// TransferSpec defines a single transfer size and repeat count.
type TransferSpec struct {
	Bytes int
	Count int
}

// Config holds all tunable parameters for a speed test run.
type Config struct {
	DownloadSequence []TransferSpec
	UploadSequence   []TransferSpec
	MaxConnections   int
	SampleInterval   time.Duration
	LatencyProbes    int
	LatencyInterval  time.Duration // interval for loaded latency probes
}

// DefaultConfig returns the default speed test configuration.
func DefaultConfig() Config {
	return Config{
		DownloadSequence: []TransferSpec{
			{Bytes: 100_000, Count: 10},
			{Bytes: 1_000_000, Count: 8},
			{Bytes: 10_000_000, Count: 6},
			{Bytes: 25_000_000, Count: 4},
		},
		UploadSequence: []TransferSpec{
			{Bytes: 11_000, Count: 10},
			{Bytes: 101_000, Count: 10},
			{Bytes: 1_000_000, Count: 8},
		},
		MaxConnections:  16,
		SampleInterval:  100 * time.Millisecond,
		LatencyProbes:   20,
		LatencyInterval: 400 * time.Millisecond,
	}
}
