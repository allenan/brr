package components

import (
	"fmt"

	"github.com/NimbleMarkets/ntcharts/sparkline"
	"github.com/charmbracelet/lipgloss"

	"github.com/allenan/brr/internal/speedtest"
)

// LatencyPanel displays latency, jitter, and bufferbloat information.
type LatencyPanel struct {
	IdleLatency float64 // avg idle latency ms
	Jitter      float64 // jitter ms
	BBGradeDL   speedtest.BufferbloatGrade
	BBGradeUL   speedtest.BufferbloatGrade
	BBDeltaDL   float64 // loaded - idle median delta
	BBDeltaUL   float64
	Active      bool
	Width       int // terminal width — set by parent

	latencySparkline sparkline.Model
	jitterSparkline  sparkline.Model

	latencyStyle lipgloss.Style
	gradeGood    lipgloss.Style
	gradeOk      lipgloss.Style
	gradeWarn    lipgloss.Style
	gradeBad     lipgloss.Style
	mutedStyle   lipgloss.Style
	boldStyle    lipgloss.Style
	sparkStyle   lipgloss.Style
}

// NewLatencyPanel creates a new latency panel component.
func NewLatencyPanel(latencyStyle, gradeGood, gradeOk, gradeWarn, gradeBad, mutedStyle, boldStyle, sparkStyle lipgloss.Style) LatencyPanel {
	ls := sparkline.New(15, 1)
	ls.AutoMaxValue = true
	ls.Style = sparkStyle

	js := sparkline.New(15, 1)
	js.AutoMaxValue = true
	js.Style = sparkStyle

	return LatencyPanel{
		Width:            80,
		latencySparkline: ls,
		jitterSparkline:  js,
		latencyStyle:     latencyStyle,
		gradeGood:        gradeGood,
		gradeOk:          gradeOk,
		gradeWarn:        gradeWarn,
		gradeBad:         gradeBad,
		mutedStyle:       mutedStyle,
		boldStyle:        boldStyle,
		sparkStyle:       sparkStyle,
	}
}

// Resize updates sparkline widths based on terminal width.
func (p *LatencyPanel) Resize(width int) {
	p.Width = width
	sw := p.sparkWidth()
	p.latencySparkline.Resize(sw, 1)
	p.jitterSparkline.Resize(sw, 1)
}

func (p LatencyPanel) sparkWidth() int {
	// 3 columns with gaps, each column gets ~1/3 of width
	colW := (p.Width - 8) / 3 // 8 = gaps + margin
	if colW < 8 {
		colW = 8
	}
	return colW
}

// PushLatency adds a latency sample to the sparkline.
func (p *LatencyPanel) PushLatency(rtt float64) {
	p.IdleLatency = rtt
	p.latencySparkline.Push(rtt)
	p.latencySparkline.Draw()
}

// PushJitter adds a jitter value to the sparkline.
func (p *LatencyPanel) PushJitter(jitter float64) {
	p.Jitter = jitter
	p.jitterSparkline.Push(jitter)
	p.jitterSparkline.Draw()
}

// View renders the latency panel.
func (p LatencyPanel) View() string {
	if !p.Active {
		// Render placeholder so height stays stable
		return p.viewPlaceholder()
	}

	colW := (p.Width - 8) / 3
	if colW < 12 {
		colW = 12
	}
	colStyle := lipgloss.NewStyle().Width(colW)

	// Latency column
	latLabel := p.latencyStyle.Render("Latency")
	latValue := p.boldStyle.Render(fmt.Sprintf("%.0fms", p.IdleLatency))
	latSpark := p.latencySparkline.View()
	latCol := colStyle.Render(lipgloss.JoinVertical(lipgloss.Left, latLabel, latValue, latSpark))

	// Jitter column
	jitLabel := p.latencyStyle.Render("Jitter")
	jitValue := p.boldStyle.Render(fmt.Sprintf("%.1fms", p.Jitter))
	jitSpark := p.jitterSparkline.View()
	jitCol := colStyle.Render(lipgloss.JoinVertical(lipgloss.Left, jitLabel, jitValue, jitSpark))

	// Bufferbloat column
	bbLabel := p.latencyStyle.Render("Bufferbloat")
	bbGrade := p.renderGrade(p.BBGradeDL)
	var bbDetail string
	if p.BBDeltaDL > 0 {
		bbDetail = p.mutedStyle.Render(fmt.Sprintf("+%.0fms", p.BBDeltaDL))
	} else {
		bbDetail = " "
	}
	bbCol := colStyle.Render(lipgloss.JoinVertical(lipgloss.Left, bbLabel, bbGrade, bbDetail))

	return lipgloss.JoinHorizontal(lipgloss.Top, latCol, "  ", jitCol, "  ", bbCol)
}

func (p LatencyPanel) viewPlaceholder() string {
	colW := (p.Width - 8) / 3
	if colW < 12 {
		colW = 12
	}
	colStyle := lipgloss.NewStyle().Width(colW)
	muted := p.mutedStyle

	latCol := colStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		muted.Render("Latency"), muted.Render("—"), " "))
	jitCol := colStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		muted.Render("Jitter"), muted.Render("—"), " "))
	bbCol := colStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		muted.Render("Bufferbloat"), muted.Render("—"), " "))

	return lipgloss.JoinHorizontal(lipgloss.Top, latCol, "  ", jitCol, "  ", bbCol)
}

func (p LatencyPanel) renderGrade(grade speedtest.BufferbloatGrade) string {
	switch grade {
	case speedtest.GradeAPlus, speedtest.GradeA:
		return p.gradeGood.Render(string(grade))
	case speedtest.GradeB:
		return p.gradeOk.Render(string(grade))
	case speedtest.GradeC:
		return p.gradeWarn.Render(string(grade))
	default:
		return p.gradeBad.Render(string(grade))
	}
}
