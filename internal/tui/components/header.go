package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Header renders the logo and border bar.
type Header struct {
	Width int

	titleStyle  lipgloss.Style
	borderStyle lipgloss.Style
}

// NewHeader creates a header component.
func NewHeader(titleStyle, borderStyle lipgloss.Style) Header {
	return Header{
		titleStyle:  titleStyle,
		borderStyle: borderStyle,
	}
}

// View renders the header.
func (h Header) View() string {
	w := h.Width
	if w < 40 {
		w = 60
	}

	logo := h.titleStyle.Render("⚡ brr")
	border := h.borderStyle.Render(repeatChar('─', w))

	return lipgloss.JoinVertical(lipgloss.Left, logo, border)
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
