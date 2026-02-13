package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/allenan/brr/internal/history"
	"github.com/allenan/brr/internal/speedtest"
	"github.com/allenan/brr/internal/tui/components"
)

// state represents the current TUI phase.
type state int

const (
	stateInit state = iota
	statePreflight
	stateMeta
	stateLatency
	stateDownload
	stateUpload
	stateDone
	stateError
	stateHistory
	stateHelp
)

// programRef is a shared reference that survives model copies.
type programRef struct {
	p *tea.Program
}

// Model is the main Bubble Tea model.
type Model struct {
	state  state
	width  int
	height int
	err    error

	// Engine
	engine *speedtest.Engine
	ctx    context.Context
	cancel context.CancelFunc

	// Results
	result *speedtest.Result
	server speedtest.ServerInfo

	// History
	store          *history.Store
	historyEntries []speedtest.Result
	statusMsg      string

	// Live speed tracking
	currentDLMbps float64
	currentULMbps float64

	// Components
	spinner        spinner.Model
	header         components.Header
	preflightPanel components.PreflightPanel
	dlGauge        components.SpeedGauge
	ulGauge        components.SpeedGauge
	latencyPanel   components.LatencyPanel
	footer         components.Footer
	theme          Theme

	// Program reference for p.Send() — shared across model copies
	pref *programRef
}

// NewModel creates a new TUI model.
func NewModel(themeName string, store *history.Store) Model {
	theme := ThemeFromName(themeName)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))

	header := components.NewHeader(theme.Title, theme.Border)
	preflightPanel := components.NewPreflightPanel(theme.GradeGood, theme.GradeBad, theme.Muted, theme.Bold, theme.Subtitle)
	dlGauge := components.NewSpeedGauge("  ↓ DOWNLOAD", theme.Download, theme.SpeedNum, theme.SpeedUnit, theme.SparkStyle, theme.ProgressDL)
	ulGauge := components.NewSpeedGauge("  ↑ UPLOAD", theme.Upload, theme.SpeedNum, theme.SpeedUnit,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B9D")), theme.ProgressUL)
	latencyPanel := components.NewLatencyPanel(theme.Latency, theme.GradeGood, theme.GradeOk, theme.GradeWarn, theme.GradeBad, theme.Muted, theme.Bold, theme.SparkStyle)
	footer := components.NewFooter(theme.FooterKey, theme.FooterAction, theme.Muted)

	return Model{
		state:          stateInit,
		engine:         speedtest.NewEngine(),
		store:          store,
		spinner:        s,
		header:         header,
		preflightPanel: preflightPanel,
		dlGauge:        dlGauge,
		ulGauge:        ulGauge,
		latencyPanel:   latencyPanel,
		footer:         footer,
		theme:          theme,
		width:          80,
		height:         24,
		pref:           &programRef{},
	}
}

// SetProgram sets the tea.Program reference for p.Send().
func (m *Model) SetProgram(p *tea.Program) {
	m.pref.p = p
}

// Init starts the TUI.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		animTick(),
	)
}
