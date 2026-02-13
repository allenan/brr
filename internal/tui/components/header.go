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

	var routeLine string
	if h.Server.Colo != "" {
		latStr := ""
		if h.Latency > 0 {
			latStr = fmt.Sprintf(" (%.0fms)", h.Latency)
		}
		origin := h.Server.Location
		if h.Server.ClientCity != "" {
			origin = h.Server.ClientCity
		}
		routeLeft := h.mutedStyle.Render(fmt.Sprintf("  %s → %s%s",
			origin, h.Server.ColoCity, latStr))
		serverRight := h.subtitleStyle.Render(fmt.Sprintf("Server: %s", h.Server.ColoCity))
		gap := w - lipgloss.Width(routeLeft) - lipgloss.Width(serverRight)
		if gap < 2 {
			gap = 2
		}
		routeLine = routeLeft + repeatChar(' ', gap) + serverRight
	} else {
		routeLine = h.subtitleStyle.Render("  Connecting...")
	}

	border := h.borderStyle.Render(repeatChar('─', w))

	return lipgloss.JoinVertical(lipgloss.Left, logo, routeLine, border)
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
