package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/allenan/brr/internal/speedtest"
)

// Store manages persisting speed test results.
type Store struct {
	path string
}

// NewStore creates a store using the default config path.
func NewStore() *Store {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return &Store{
		path: filepath.Join(configDir, "brr", "history.json"),
	}
}

// Save appends a result to the history file.
func (s *Store) Save(result *speedtest.Result) error {
	entries, _ := s.Load() // ignore error on first run
	entries = append(entries, *result)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	// Atomic write via temp file
	tmp := s.path + ".tmp"
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Load reads all history entries.
func (s *Store) Load() ([]speedtest.Result, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entries []speedtest.Result
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	return entries, nil
}

// Last returns the n most recent entries.
func (s *Store) Last(n int) ([]speedtest.Result, error) {
	entries, err := s.Load()
	if err != nil {
		return nil, err
	}
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n], nil
}

// Average computes the average download/upload/latency over the last n entries.
func (s *Store) Average(n int) (*speedtest.Result, error) {
	entries, err := s.Last(n)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}

	avg := &speedtest.Result{}
	for _, e := range entries {
		avg.Download.Mbps += e.Download.Mbps
		avg.Upload.Mbps += e.Upload.Mbps
		avg.IdleLatency.Avg += e.IdleLatency.Avg
		avg.IdleLatency.Jitter += e.IdleLatency.Jitter
	}
	n64 := float64(len(entries))
	avg.Download.Mbps /= n64
	avg.Upload.Mbps /= n64
	avg.IdleLatency.Avg /= n64
	avg.IdleLatency.Jitter /= n64

	return avg, nil
}
