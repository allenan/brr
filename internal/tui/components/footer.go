package components

import (
	"github.com/charmbracelet/lipgloss"
)

// Footer renders the action bar at the bottom.
type Footer struct {
	Done  bool
	Width int

	keyStyle    lipgloss.Style
	actionStyle lipgloss.Style
	mutedStyle  lipgloss.Style
}

// NewFooter creates a new footer component.
func NewFooter(keyStyle, actionStyle, mutedStyle lipgloss.Style) Footer {
	return Footer{
		keyStyle:    keyStyle,
		actionStyle: actionStyle,
		mutedStyle:  mutedStyle,
	}
}

// View renders the footer.
func (f Footer) View() string {
	if !f.Done {
		return f.mutedStyle.Render("  Press q to cancel")
	}

	items := []struct {
		key    string
		action string
	}{
		{"r", "Run Again"},
		{"h", "History"},
		{"e", "Export JSON"},
		{"c", "Compare"},
		{"q", "Quit"},
	}

	var parts []string
	for _, item := range items {
		k := f.keyStyle.Render(item.key)
		a := f.actionStyle.Render(item.action)
		parts = append(parts, "  "+k+" "+a)
	}

	result := ""
	for _, p := range parts {
		result += p
	}
	return result
}
