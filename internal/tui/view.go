package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/allenan/brr/internal/speedtest"
)

// View renders the TUI.
func (m Model) View() string {
	if m.state == stateHistory {
		return m.viewHistoryScreen()
	}

	var sections []string

	// Header (3 lines: top row, route, border)
	sections = append(sections, m.header.View())
	sections = append(sections, "")

	// Status line — what's currently happening
	sections = append(sections, m.statusLine())
	sections = append(sections, "")

	// Download gauge (always rendered, 2 lines)
	sections = append(sections, m.dlGauge.View())
	sections = append(sections, "")

	// Upload gauge (always rendered, 2 lines)
	sections = append(sections, m.ulGauge.View())
	sections = append(sections, "")

	// Latency panel (always rendered, 3 lines)
	sections = append(sections, m.latencyPanel.View())
	sections = append(sections, "")

	// Context line / status message
	if m.state == stateDone && m.result != nil && m.result.ContextLine != "" {
		contextStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#B0B0B0"))
		sections = append(sections, "  "+contextStyle.Render(m.result.ContextLine))
	} else {
		sections = append(sections, "")
	}

	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
		sections = append(sections, "  "+statusStyle.Render(m.statusMsg))
	} else {
		sections = append(sections, "")
	}

	sections = append(sections, "")

	// Footer
	sections = append(sections, m.footer.View())

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) statusLine() string {
	switch m.state {
	case stateInit, stateMeta:
		return "  " + m.spinner.View() + " Connecting..."
	case stateLatency:
		return "  " + m.spinner.View() + " Measuring latency..."
	case stateDownload:
		return "  " + m.spinner.View() + " Testing download..."
	case stateUpload:
		return "  " + m.spinner.View() + " Testing upload..."
	case stateDone:
		if m.result != nil {
			checkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA"))
			dlStr := m.theme.Download.Render(fmt.Sprintf("%.0f↓", m.result.Download.Mbps))
			ulStr := m.theme.Upload.Render(fmt.Sprintf("%.0f↑", m.result.Upload.Mbps))
			unit := m.theme.SpeedUnit.Render(" Mbps")
			grade := m.renderGrade(m.result.BufferbloatDL)
			return "  " + checkStyle.Render("✓") + " " + dlStr + "  " + ulStr + unit + "  · Bufferbloat " + grade
		}
		doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA"))
		return "  " + doneStyle.Render("✓") + " Test complete"
	case stateError:
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
		return "  " + errStyle.Render(fmt.Sprintf("✗ Error: %v", m.err))
	default:
		return ""
	}
}

func (m Model) viewHistoryScreen() string {
	var sections []string

	sections = append(sections, m.header.View())
	sections = append(sections, "")

	if len(m.historyEntries) == 0 {
		sections = append(sections, "  No history yet.")
	} else {
		boldStyle := lipgloss.NewStyle().Bold(true)
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

		header := fmt.Sprintf("  %-18s  %-6s  %10s  %10s  %8s  %5s",
			boldStyle.Render("Date"),
			boldStyle.Render("Server"),
			boldStyle.Render("Download"),
			boldStyle.Render("Upload"),
			boldStyle.Render("Latency"),
			boldStyle.Render("Grade"))

		sections = append(sections, header)
		sections = append(sections, mutedStyle.Render("  "+strings.Repeat("─", m.width-4)))

		for i, e := range m.historyEntries {
			date := e.Timestamp.Format("2006-01-02 15:04")
			server := e.Server.Colo
			if server == "" {
				server = "—"
			}

			dlArrow, ulArrow := " ", " "
			if i < len(m.historyEntries)-1 {
				prev := m.historyEntries[i+1]
				dlArrow = trendArrow(e.Download.Mbps, prev.Download.Mbps)
				ulArrow = trendArrow(e.Upload.Mbps, prev.Upload.Mbps)
			}

			line := fmt.Sprintf("  %-18s  %-6s  %7.1f%s Mbps  %7.1f%s Mbps  %5.0fms  %5s",
				date, server,
				e.Download.Mbps, dlArrow,
				e.Upload.Mbps, ulArrow,
				e.IdleLatency.Avg,
				e.BufferbloatDL)
			sections = append(sections, line)
		}
	}

	sections = append(sections, "")
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	sections = append(sections, mutedStyle.Render("  Press ESC to go back"))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderGrade(grade speedtest.BufferbloatGrade) string {
	switch grade {
	case speedtest.GradeAPlus, speedtest.GradeA:
		return m.theme.GradeGood.Render(string(grade))
	case speedtest.GradeB:
		return m.theme.GradeOk.Render(string(grade))
	case speedtest.GradeC:
		return m.theme.GradeWarn.Render(string(grade))
	default:
		return m.theme.GradeBad.Render(string(grade))
	}
}

func trendArrow(current, previous float64) string {
	diff := current - previous
	pct := diff / previous * 100
	if pct > 5 {
		return "↑"
	} else if pct < -5 {
		return "↓"
	}
	return "→"
}
