package tui

import (
	"github.com/allenan/brr/internal/preflight"
	"github.com/allenan/brr/internal/speedtest"
)

// Phase transition messages
type metaCompleteMsg struct {
	server speedtest.ServerInfo
}

type latencyCompleteMsg struct {
	result speedtest.LatencyResult
}

type downloadCompleteMsg struct {
	result  speedtest.PhaseResult
	latency speedtest.LatencyResult
	grade   speedtest.BufferbloatGrade
}

type uploadCompleteMsg struct {
	result  speedtest.PhaseResult
	latency speedtest.LatencyResult
	grade   speedtest.BufferbloatGrade
}

type testCompleteMsg struct {
	result *speedtest.Result
}

// Sample messages (streamed during phases)
type downloadSampleMsg struct {
	sample speedtest.Sample
}

type uploadSampleMsg struct {
	sample speedtest.Sample
}

type idleLatencySampleMsg struct {
	sample speedtest.LatencySample
}

type loadedLatencySampleMsg struct {
	sample speedtest.LatencySample
}

// Preflight messages
type preflightCheckMsg struct {
	result preflight.CheckResult
}

type preflightCompleteMsg struct {
	result *preflight.Result
}

// Animation tick
type animTickMsg struct{}

// Error
type errMsg struct {
	err error
}
