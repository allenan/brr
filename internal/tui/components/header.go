package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/allenan/brr/internal/speedtest"
)

// Header renders the logo and server info bar.
type Header struct {
	Server  speedtest.ServerInfo
	Latency float64 // idle latency avg in ms
	Width   int

	titleStyle    lipgloss.Style
	subtitleStyle lipgloss.Style
	mutedStyle    lipgloss.Style
	borderStyle   lipgloss.Style
}

// NewHeader creates a header component.
func NewHeader(titleStyle, subtitleStyle, mutedStyle, borderStyle lipgloss.Style) Header {
	return Header{
		titleStyle:    titleStyle,
		subtitleStyle: subtitleStyle,
		mutedStyle:    mutedStyle,
		borderStyle:   borderStyle,
	}
}

// View renders the header.
func (h Header) View() string {
	w := h.Width
	if w < 40 {
		w = 60
	}

	logo := h.titleStyle.Render("⚡ brr")

	var serverLine string
	if h.Server.Colo != "" {
		serverLine = h.subtitleStyle.Render(fmt.Sprintf("Server: %s", h.Server.ColoCity))
	} else {
		serverLine = h.subtitleStyle.Render("Connecting...")
	}

	topRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		logo,
		lipgloss.NewStyle().Width(w-lipgloss.Width(logo)-lipgloss.Width(serverLine)).Render(""),
		serverLine,
	)

	var routeLine string
	if h.Server.Colo != "" {
		latStr := ""
		if h.Latency > 0 {
			latStr = fmt.Sprintf(" (%.0fms)", h.Latency)
		}
		routeLine = h.mutedStyle.Render(fmt.Sprintf("  %s → %s%s",
			h.Server.Location, h.Server.ColoCity, latStr))
	}

	border := h.borderStyle.Render(repeatChar('─', w))

	if routeLine != "" {
		return lipgloss.JoinVertical(lipgloss.Left, topRow, routeLine, border)
	}
	return lipgloss.JoinVertical(lipgloss.Left, topRow, border)
}

func repeatChar(ch rune, n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]rune, n)
	for i := range buf {
		buf[i] = ch
	}
	return string(buf)
}
