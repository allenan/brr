package speedtest

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Engine orchestrates a complete speed test sequence.
type Engine struct {
	Client *http.Client
	Config Config
}

// NewEngine creates a new speed test engine with default config.
func NewEngine() *Engine {
	return &Engine{
		Client: NewHTTPClient(),
		Config: DefaultConfig(),
	}
}

// Run executes the full speed test sequence, calling cb for progress updates.
func (e *Engine) Run(ctx context.Context, cb ProgressCallback) (*Result, error) {
	result := &Result{
		Timestamp: time.Now(),
	}

	// Phase 1: Metadata
	cb.OnPhase(PhaseMeta)
	meta, err := FetchMeta(ctx, e.Client)
	if err != nil {
		return nil, fmt.Errorf("metadata: %w", err)
	}
	result.Server = *meta

	measID := fmt.Sprintf("%d", time.Now().UnixNano())

	// Phase 2: Idle Latency
	cb.OnPhase(PhaseLatency)
	idleLatency, err := MeasureIdleLatency(ctx, e.Client, e.Config.LatencyProbes, cb.OnIdleLatencySample)
	if err != nil {
		return nil, fmt.Errorf("idle latency: %w", err)
	}
	result.IdleLatency = *idleLatency

	// Phase 3: Download + Loaded Latency
	cb.OnPhase(PhaseDownload)
	cancelDLLatency, dlLatencyCh := MeasureLoadedLatency(ctx, e.Client, e.Config.LatencyInterval, cb.OnLoadedLatencySample)

	dlResult, err := MeasureDownload(ctx, e.Client, e.Config, measID, cb.OnDownloadSample)
	cancelDLLatency()
	dlLatency := <-dlLatencyCh
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	result.Download = *dlResult
	result.DownloadLatency = *dlLatency
	result.BufferbloatDL = BufferbloatGrading(idleLatency, dlLatency)

	// Phase 4: Upload + Loaded Latency
	cb.OnPhase(PhaseUpload)
	cancelULLatency, ulLatencyCh := MeasureLoadedLatency(ctx, e.Client, e.Config.LatencyInterval, cb.OnLoadedLatencySample)

	ulResult, err := MeasureUpload(ctx, e.Client, e.Config, measID, cb.OnUploadSample)
	cancelULLatency()
	ulLatency := <-ulLatencyCh
	if err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}
	result.Upload = *ulResult
	result.UploadLatency = *ulLatency
	result.BufferbloatUL = BufferbloatGrading(idleLatency, ulLatency)

	result.ContextLine = ContextLine(result)

	// Done
	cb.OnPhase(PhaseDone)
	return result, nil
}
