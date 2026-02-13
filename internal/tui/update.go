package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenan/brr/internal/export"
)

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case "h":
			if m.state == stateDone {
				m.state = stateHistory
				entries, err := m.store.Last(10)
				if err == nil {
					m.historyEntries = entries
				}
				return m, nil
			}
		case "e":
			if m.state == stateDone && m.result != nil {
				f, err := os.Create("brr-result.json")
				if err == nil {
					export.ToJSON(f, m.result)
					f.Close()
					m.statusMsg = "Exported to brr-result.json"
				}
				return m, nil
			}
		case "c":
			if m.state == stateDone {
				// Compare with last run
				entries, err := m.store.Last(2)
				if err == nil && len(entries) >= 2 {
					prev := entries[1] // entries[0] is the current run we just saved
					m.statusMsg = fmt.Sprintf("vs last: ↓ %.1f→%.1f  ↑ %.1f→%.1f  ⏱ %.0f→%.0fms",
						prev.Download.Mbps, m.result.Download.Mbps,
						prev.Upload.Mbps, m.result.Upload.Mbps,
						prev.IdleLatency.Avg, m.result.IdleLatency.Avg,
					)
				} else {
					m.statusMsg = "No previous run to compare"
				}
				return m, nil
			}
		case "r":
			if m.state == stateDone || m.state == stateHistory {
				if m.cancel != nil {
					m.cancel()
				}
				fresh := NewModel(m.theme.Name, m.store)
				fresh.width = m.width
				fresh.height = m.height
				fresh.pref = m.pref
				fresh.header.Width = m.width
				fresh.footer.Width = m.width
				fresh.dlGauge.Resize(m.width)
				fresh.ulGauge.Resize(m.width)
				fresh.latencyPanel.Resize(m.width)
				return fresh, tea.Batch(fresh.spinner.Tick, animTick())
			}
		case "esc":
			if m.state == stateHistory {
				m.state = stateDone
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.header.Width = msg.Width
		m.footer.Width = msg.Width
		m.dlGauge.Resize(msg.Width)
		m.ulGauge.Resize(msg.Width)
		m.latencyPanel.Resize(msg.Width)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case animTickMsg:
		m.dlGauge.Tick()
		m.ulGauge.Tick()

		// On first tick with program reference, start the test
		if m.state == stateInit && m.pref != nil && m.pref.p != nil {
			m.state = stateMeta
			m.ctx, m.cancel = context.WithCancel(context.Background())
			return m, tea.Batch(
				animTick(),
				runFullTest(m.ctx, m.engine, m.pref.p),
			)
		}
		return m, animTick()

	// Sample messages
	case downloadSampleMsg:
		m.state = stateDownload
		m.currentDLMbps = msg.sample.Mbps
		m.dlGauge.Active = true
		m.dlGauge.PushSample(msg.sample.Mbps)
		return m, nil

	case uploadSampleMsg:
		m.state = stateUpload
		m.currentULMbps = msg.sample.Mbps
		m.ulGauge.Active = true
		m.ulGauge.PushSample(msg.sample.Mbps)
		return m, nil

	case idleLatencySampleMsg:
		m.state = stateLatency
		m.latencyPanel.Active = true
		m.latencyPanel.PushLatency(msg.sample.RTT)
		return m, nil

	case loadedLatencySampleMsg:
		m.latencyPanel.PushLatency(msg.sample.RTT)
		return m, nil

	// Test complete
	case testCompleteMsg:
		m.state = stateDone
		m.result = msg.result
		m.server = msg.result.Server
		m.header.Server = msg.result.Server
		m.header.Latency = msg.result.IdleLatency.Avg

		// Set final values
		m.dlGauge.TargetMbps = msg.result.Download.Mbps
		m.dlGauge.Active = true
		m.dlGauge.Done = true
		m.ulGauge.TargetMbps = msg.result.Upload.Mbps
		m.ulGauge.Active = true
		m.ulGauge.Done = true

		m.latencyPanel.Active = true
		m.latencyPanel.Done = true
		m.latencyPanel.IdleLatency = msg.result.IdleLatency.Avg
		m.latencyPanel.Jitter = msg.result.IdleLatency.Jitter
		if len(msg.result.DownloadLatency.Samples) > 0 {
			m.latencyPanel.LoadedLatencyDL = msg.result.DownloadLatency.Avg
		}

		m.latencyPanel.BBGradeDL = msg.result.BufferbloatDL
		m.latencyPanel.BBGradeUL = msg.result.BufferbloatUL
		if len(msg.result.DownloadLatency.Samples) > 0 {
			m.latencyPanel.BBDeltaDL = msg.result.DownloadLatency.Avg - msg.result.IdleLatency.Avg
		}
		if len(msg.result.UploadLatency.Samples) > 0 {
			m.latencyPanel.BBDeltaUL = msg.result.UploadLatency.Avg - msg.result.IdleLatency.Avg
		}

		m.footer.Done = true

		// Save to history
		if m.store != nil {
			m.store.Save(msg.result)
		}

		return m, nil

	case errMsg:
		m.state = stateError
		m.err = msg.err
		return m, nil
	}

	return m, nil
}
