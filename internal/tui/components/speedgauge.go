package components

import (
	"fmt"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

const (
	sparkHeight    = 1
	minProgressW   = 20
	minSparkW      = 10
	labelReservedW = 36 // "  ↓ DOWNLOAD" + speed number + " Mbps" + padding
)

// SpeedGauge displays a progress bar, sparkline, and spring-animated speed number.
type SpeedGauge struct {
	Label   string
	MaxMbps float64
	Active  bool
	Done    bool
	Width   int // terminal width — set by parent

	// Current actual measured value
	TargetMbps float64

	// Spring-animated display value
	shownMbps float64
	velocity  float64
	spring    harmonica.Spring

	// Progress bar
	prog progress.Model

	// Sparkline
	spark sparkline.Model

	// Styles
	labelStyle     lipgloss.Style
	speedNumStyle  lipgloss.Style
	speedUnitStyle lipgloss.Style
	sparkStyle     lipgloss.Style
	progressColor  string
}

// NewSpeedGauge creates a new speed gauge component.
func NewSpeedGauge(label string, labelStyle, speedNumStyle, speedUnitStyle, sparkStyle, progressStyle lipgloss.Style) SpeedGauge {
	colorStr := fmt.Sprintf("%s", progressStyle.GetForeground())

	p := progress.New(
		progress.WithGradient(colorStr, colorStr),
		progress.WithWidth(minProgressW),
		progress.WithoutPercentage(),
	)

	sp := sparkline.New(minSparkW, sparkHeight)
	sp.AutoMaxValue = true
	sp.Style = sparkStyle

	spring := harmonica.NewSpring(harmonica.FPS(60), 6.0, 1.0)

	return SpeedGauge{
		Label:          label,
		MaxMbps:        500,
		Width:          80,
		spring:         spring,
		prog:           p,
		spark:          sp,
		labelStyle:     labelStyle,
		speedNumStyle:  speedNumStyle,
		speedUnitStyle: speedUnitStyle,
		sparkStyle:     sparkStyle,
		progressColor:  colorStr,
	}
}

// PushSample adds a new speed sample to the sparkline.
func (g *SpeedGauge) PushSample(mbps float64) {
	g.TargetMbps = mbps
	g.spark.Push(mbps)
	g.spark.Draw()
	if mbps > g.MaxMbps {
		g.MaxMbps = mbps * 1.2
	}
}

// Tick updates the spring animation. Call at 60fps.
func (g *SpeedGauge) Tick() {
	g.shownMbps, g.velocity = g.spring.Update(g.shownMbps, g.velocity, g.TargetMbps)
	if g.shownMbps < 0 {
		g.shownMbps = 0
	}
}

// Resize updates internal widths to fill the terminal.
func (g *SpeedGauge) Resize(width int) {
	g.Width = width
	pw, sw := g.barWidths()

	g.prog = progress.New(
		progress.WithGradient(g.progressColor, g.progressColor),
		progress.WithWidth(pw),
		progress.WithoutPercentage(),
	)

	g.spark.Resize(sw, sparkHeight)
}

// barWidths computes the progress bar and sparkline widths from the terminal width.
// Layout:  "  " + progress + "  " + sparkline + padding
// We give ~55% to the progress bar and ~45% to the sparkline.
func (g SpeedGauge) barWidths() (int, int) {
	usable := g.Width - 6 // leading "  " + gap "  " + trailing margin
	if usable < minProgressW+minSparkW {
		return minProgressW, minSparkW
	}
	pw := usable * 55 / 100
	sw := usable - pw
	return pw, sw
}

// View renders the speed gauge.
func (g SpeedGauge) View() string {
	label := g.labelStyle.Render(g.Label)

	if !g.Active && !g.Done {
		speed := lipgloss.NewStyle().Faint(true).Render("  —")
		headerLine := lipgloss.JoinHorizontal(lipgloss.Top, label, speed)
		// Render two lines (header + blank bar line) so height is stable
		barLine := ""
		return lipgloss.JoinVertical(lipgloss.Left, headerLine, barLine)
	}

	// Format speed
	displayed := g.shownMbps
	if g.Done {
		displayed = g.TargetMbps
	}
	speedStr := g.speedNumStyle.Render(fmt.Sprintf("%.1f", displayed))
	unitStr := g.speedUnitStyle.Render(" Mbps")

	// Right-align speed number: fill the gap between label and speed
	speedFull := speedStr + unitStr
	gap := g.Width - lipgloss.Width(label) - lipgloss.Width(speedFull) - 2
	if gap < 1 {
		gap = 1
	}
	headerLine := label + repeatChar(' ', gap) + speedFull

	// Progress bar
	pct := displayed / g.MaxMbps
	if pct > 1 {
		pct = 1
	}
	if pct < 0 {
		pct = 0
	}
	progressStr := g.prog.ViewAs(pct)

	// Sparkline
	sparkStr := g.spark.View()

	barLine := "  " + progressStr + "  " + sparkStr

	return lipgloss.JoinVertical(lipgloss.Left, headerLine, barLine)
}

