package speedtest

import (
	"math"
	"sort"
)

// Percentile computes the p-th percentile (0..1) of a float64 slice.
// Uses linear interpolation between closest ranks.
func Percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}

	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}

	rank := p * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := lower + 1
	if upper >= len(sorted) {
		return sorted[lower]
	}

	frac := rank - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

// Median returns the median value of a float64 slice.
func Median(data []float64) float64 {
	return Percentile(data, 0.50)
}

// Jitter computes the standard deviation of a float64 slice (used as jitter measure).
func Jitter(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}

	avg := mean(data)
	sumSq := 0.0
	for _, v := range data {
		d := v - avg
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(data)))
}

// BufferbloatGrading returns the bufferbloat grade based on the increase in latency
// under load (loaded median - idle median, in ms).
func BufferbloatGrading(idleLatency, loadedLatency *LatencyResult) BufferbloatGrade {
	if idleLatency == nil || loadedLatency == nil {
		return GradeF
	}
	if len(loadedLatency.Samples) == 0 {
		return GradeAPlus
	}

	idleRTTs := make([]float64, len(idleLatency.Samples))
	for i, s := range idleLatency.Samples {
		idleRTTs[i] = s.RTT
	}
	loadedRTTs := make([]float64, len(loadedLatency.Samples))
	for i, s := range loadedLatency.Samples {
		loadedRTTs[i] = s.RTT
	}

	idleMedian := Median(idleRTTs)
	loadedMedian := Median(loadedRTTs)

	delta := loadedMedian - idleMedian
	if delta < 0 {
		delta = 0
	}

	return gradeFromDelta(delta)
}

func gradeFromDelta(delta float64) BufferbloatGrade {
	switch {
	case delta < 5:
		return GradeAPlus
	case delta < 30:
		return GradeA
	case delta < 60:
		return GradeB
	case delta < 200:
		return GradeC
	case delta < 400:
		return GradeD
	default:
		return GradeF
	}
}

// ContextLine generates a human-readable summary based on test results.
func ContextLine(result *Result) string {
	dl := result.Download.Mbps
	ul := result.Upload.Mbps
	grade := result.BufferbloatDL

	switch {
	case dl >= 100 && ul >= 20 && (grade == GradeAPlus || grade == GradeA):
		return "Excellent for 4K streaming, video calls, and gaming"
	case dl >= 25 && (grade == GradeAPlus || grade == GradeA || grade == GradeB):
		return "Great for 4K streaming and video calls"
	case dl >= 25 && (grade == GradeC || grade == GradeD):
		return "Good speeds, but bufferbloat may affect video calls and gaming"
	case dl >= 5 && (grade == GradeAPlus || grade == GradeA || grade == GradeB):
		return "Good for HD streaming and video calls"
	case dl >= 5 && (grade == GradeC || grade == GradeD):
		return "OK for HD streaming, bufferbloat may affect video calls"
	case dl >= 1:
		return "Sufficient for basic browsing and SD streaming"
	default:
		return "Connection is very slow, may have trouble with most activities"
	}
}
