package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/allenan/brr/internal/speedtest"
)

// runFullTest runs the entire speed test, sending sample messages via p.Send()
// and returning testCompleteMsg when done.
func runFullTest(ctx context.Context, engine *speedtest.Engine, p *tea.Program) tea.Cmd {
	return func() tea.Msg {
		cb := &tuiCallback{program: p}
		result, err := engine.Run(ctx, cb)
		if err != nil {
			return errMsg{err: err}
		}
		return testCompleteMsg{result: result}
	}
}

// tuiCallback sends TUI messages for each progress event.
type tuiCallback struct {
	program *tea.Program
}

func (c *tuiCallback) OnPhase(phase speedtest.Phase) {
	switch phase {
	case speedtest.PhaseMeta:
		// nothing to send yet
	case speedtest.PhaseLatency:
		// nothing to send yet
	case speedtest.PhaseDownload:
		// nothing to send yet
	case speedtest.PhaseUpload:
		// nothing to send yet
	case speedtest.PhaseDone:
		// nothing to send yet
	}
}

func (c *tuiCallback) OnDownloadSample(s speedtest.Sample) {
	c.program.Send(downloadSampleMsg{sample: s})
}

func (c *tuiCallback) OnUploadSample(s speedtest.Sample) {
	c.program.Send(uploadSampleMsg{sample: s})
}

func (c *tuiCallback) OnIdleLatencySample(s speedtest.LatencySample) {
	c.program.Send(idleLatencySampleMsg{sample: s})
}

func (c *tuiCallback) OnLoadedLatencySample(s speedtest.LatencySample) {
	c.program.Send(loadedLatencySampleMsg{sample: s})
}

// animTick returns a command that sends animTickMsg at 60fps.
func animTick() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return animTickMsg{}
	})
}
