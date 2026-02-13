package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	"github.com/allenan/brr/internal/speedtest"
)

// ToJSON writes the result as formatted JSON.
func ToJSON(w io.Writer, result *speedtest.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// ToCSV writes the result (or multiple results) as CSV.
func ToCSV(w io.Writer, results []speedtest.Result) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{
		"timestamp", "server_colo", "server_city", "location",
		"download_mbps", "upload_mbps",
		"latency_avg_ms", "latency_jitter_ms",
		"bufferbloat_dl", "bufferbloat_ul",
		"context",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, r := range results {
		row := []string{
			r.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			r.Server.Colo,
			r.Server.ColoCity,
			r.Server.Location,
			fmt.Sprintf("%.1f", r.Download.Mbps),
			fmt.Sprintf("%.1f", r.Upload.Mbps),
			fmt.Sprintf("%.1f", r.IdleLatency.Avg),
			fmt.Sprintf("%.1f", r.IdleLatency.Jitter),
			string(r.BufferbloatDL),
			string(r.BufferbloatUL),
			r.ContextLine,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
