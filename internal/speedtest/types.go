package speedtest

import "time"

// Phase represents a stage of the speed test.
type Phase int

const (
	PhaseMeta Phase = iota
	PhaseLatency
	PhaseDownload
	PhaseUpload
	PhaseDone
)

func (p Phase) String() string {
	switch p {
	case PhaseMeta:
		return "meta"
	case PhaseLatency:
		return "latency"
	case PhaseDownload:
		return "download"
	case PhaseUpload:
		return "upload"
	case PhaseDone:
		return "done"
	default:
		return "unknown"
	}
}

// Sample is a single throughput measurement point.
type Sample struct {
	Timestamp time.Time `json:"timestamp"`
	Mbps      float64   `json:"mbps"`
}

// LatencySample is a single latency measurement.
type LatencySample struct {
	Timestamp time.Time `json:"timestamp"`
	RTT       float64   `json:"rtt_ms"` // round-trip time in milliseconds
}

// PhaseResult holds the outcome of a download or upload phase.
type PhaseResult struct {
	Mbps    float64  `json:"mbps"`    // P90 speed
	Samples []Sample `json:"samples"` // all raw samples
}

// LatencyResult holds latency measurement outcomes.
type LatencyResult struct {
	Min     float64         `json:"min_ms"`
	Max     float64         `json:"max_ms"`
	Avg     float64         `json:"avg_ms"`
	Jitter  float64         `json:"jitter_ms"`
	Samples []LatencySample `json:"samples"`
}

// ServerInfo holds metadata about the test server / client.
type ServerInfo struct {
	IP       string `json:"ip"`
	Colo     string `json:"colo"`      // IATA airport code
	ColoCity string `json:"colo_city"` // human-readable city name
	Location string `json:"location"`  // country code
}

// BufferbloatGrade represents the quality grade for bufferbloat.
type BufferbloatGrade string

const (
	GradeAPlus BufferbloatGrade = "A+"
	GradeA     BufferbloatGrade = "A"
	GradeB     BufferbloatGrade = "B"
	GradeC     BufferbloatGrade = "C"
	GradeD     BufferbloatGrade = "D"
	GradeF     BufferbloatGrade = "F"
)

// Result is the complete outcome of a speed test run.
type Result struct {
	Timestamp       time.Time        `json:"timestamp"`
	Server          ServerInfo       `json:"server"`
	Download        PhaseResult      `json:"download"`
	Upload          PhaseResult      `json:"upload"`
	IdleLatency     LatencyResult    `json:"idle_latency"`
	DownloadLatency LatencyResult    `json:"download_latency"`
	UploadLatency   LatencyResult    `json:"upload_latency"`
	BufferbloatDL   BufferbloatGrade `json:"bufferbloat_download"`
	BufferbloatUL   BufferbloatGrade `json:"bufferbloat_upload"`
	ContextLine     string           `json:"context_line"`
}

// ProgressCallback receives updates as the test progresses.
type ProgressCallback interface {
	OnPhase(phase Phase)
	OnDownloadSample(s Sample)
	OnUploadSample(s Sample)
	OnIdleLatencySample(s LatencySample)
	OnLoadedLatencySample(s LatencySample)
}
