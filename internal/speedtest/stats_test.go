package speedtest

import (
	"math"
	"testing"
)

func TestPercentile(t *testing.T) {
	tests := []struct {
		name string
		data []float64
		p    float64
		want float64
	}{
		{"empty", nil, 0.5, 0},
		{"single", []float64{42}, 0.5, 42},
		{"median_odd", []float64{1, 2, 3, 4, 5}, 0.5, 3},
		{"median_even", []float64{1, 2, 3, 4}, 0.5, 2.5},
		{"p90", []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 0.90, 9.1},
		{"p0", []float64{5, 3, 1}, 0, 1},
		{"p100", []float64{5, 3, 1}, 1, 5},
		{"unsorted", []float64{5, 1, 3, 2, 4}, 0.5, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Percentile(tt.data, tt.p)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("Percentile(%v, %v) = %v, want %v", tt.data, tt.p, got, tt.want)
			}
		})
	}
}

func TestMedian(t *testing.T) {
	got := Median([]float64{1, 5, 3, 2, 4})
	if got != 3 {
		t.Errorf("Median = %v, want 3", got)
	}
}

func TestJitter(t *testing.T) {
	tests := []struct {
		name string
		data []float64
		want float64
	}{
		{"empty", nil, 0},
		{"single", []float64{5}, 0},
		{"constant", []float64{3, 3, 3}, 0},
		{"varied", []float64{10, 20, 30}, 8.165}, // stddev of 10,20,30
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Jitter(tt.data)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("Jitter(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestBufferbloatGrading(t *testing.T) {
	makeLatency := func(rtts ...float64) *LatencyResult {
		samples := make([]LatencySample, len(rtts))
		for i, r := range rtts {
			samples[i] = LatencySample{RTT: r}
		}
		return &LatencyResult{Samples: samples}
	}

	tests := []struct {
		name   string
		idle   *LatencyResult
		loaded *LatencyResult
		want   BufferbloatGrade
	}{
		{"no_bloat", makeLatency(10, 12, 11), makeLatency(12, 13, 14), GradeAPlus},
		{"minor_bloat", makeLatency(10, 12, 11), makeLatency(30, 35, 40), GradeA},
		{"moderate_bloat", makeLatency(10, 12, 11), makeLatency(60, 65, 70), GradeB},
		{"bad_bloat", makeLatency(10, 12, 11), makeLatency(150, 200, 250), GradeC},
		{"severe_bloat", makeLatency(10, 12, 11), makeLatency(350, 400, 450), GradeD},
		{"terrible_bloat", makeLatency(10, 12, 11), makeLatency(500, 600, 700), GradeF},
		{"nil_idle", nil, makeLatency(100), GradeF},
		{"no_loaded_samples", makeLatency(10), &LatencyResult{}, GradeAPlus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BufferbloatGrading(tt.idle, tt.loaded)
			if got != tt.want {
				t.Errorf("BufferbloatGrading() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextLine(t *testing.T) {
	r := &Result{
		Download:    PhaseResult{Mbps: 200},
		Upload:      PhaseResult{Mbps: 50},
		BufferbloatDL: GradeA,
	}
	line := ContextLine(r)
	if line == "" {
		t.Error("ContextLine returned empty string")
	}
}
