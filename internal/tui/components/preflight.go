package components

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/allenan/brr/internal/preflight"
)

// PreflightPanel renders the horizontal network path flow.
type PreflightPanel struct {
	checks  []preflight.CheckResult
	message string // diagnostic message on failure
	Active  bool   // true during preflight (shows pending/spinner nodes)

	passStyle   lipgloss.Style
	failStyle   lipgloss.Style
	mutedStyle  lipgloss.Style
	boldStyle   lipgloss.Style
	detailStyle lipgloss.Style
}

// NewPreflightPanel creates a preflight panel component.
func NewPreflightPanel(passStyle, failStyle, mutedStyle, boldStyle, detailStyle lipgloss.Style) PreflightPanel {
	return PreflightPanel{
		passStyle:   passStyle,
		failStyle:   failStyle,
		mutedStyle:  mutedStyle,
		boldStyle:   boldStyle,
		detailStyle: detailStyle,
	}
}

// PushResult adds a completed check to the panel.
func (p *PreflightPanel) PushResult(r preflight.CheckResult) {
	p.checks = append(p.checks, r)
}

// SetMessage sets the diagnostic message shown on failure.
func (p *PreflightPanel) SetMessage(msg string) {
	p.message = msg
}

// HasChecks returns true if any check results have been pushed.
func (p PreflightPanel) HasChecks() bool {
	return len(p.checks) > 0
}

// nodeLabels are the default labels for the 5 network path nodes.
// Node 0 (Device) is always pass; nodes 1–4 map to preflight checks 0–3.
var nodeLabels = [5]string{"Device", "Router", "Internet", "DNS", "Server"}

// nodeCount is the number of nodes in the flow.
const nodeCount = 5

// arrow is the glyph between nodes: 4 horizontal box-drawing chars + triangle.
const arrow = "────▸"

// sepWidth is the rendered width of "  ────▸  " (2 + 5 + 2).
const sepWidth = 9

// ViewFlow renders the two-line horizontal network path.
// spinnerFrame is the current braille spinner frame for the active hop.
func (p PreflightPanel) ViewFlow(width int, spinnerFrame string) string {
	styledSep := "  " + p.mutedStyle.Render(arrow) + "  "
	spaceSep := strings.Repeat(" ", sepWidth)

	// Determine content for each node
	type nodeContent struct {
		top    string // styled icon + " " + label
		bottom string // styled detail
	}
	var nodes [nodeCount]nodeContent
	for i := 0; i < nodeCount; i++ {
		icon, _, detail := p.nodeState(i, spinnerFrame)
		label := nodeLabels[i]
		// Device node (index 0): use hostname as label
		if i == 0 {
			if h, err := os.Hostname(); err == nil {
				label = strings.TrimSuffix(h, ".local")
			}
		}
		// Server node (index 4): use city name as label when available
		if i == 4 && len(p.checks) >= 4 && p.checks[3].Passed && p.checks[3].Detail != "" {
			label = p.checks[3].Detail
		}
		nodes[i] = nodeContent{
			top:    icon + " " + label,
			bottom: detail,
		}
	}

	// Column widths sized to content (max of top/bottom), min 6
	var colWidths [nodeCount]int
	for i := 0; i < nodeCount; i++ {
		topW := lipgloss.Width(nodes[i].top)
		botW := lipgloss.Width(nodes[i].bottom)
		w := topW
		if botW > w {
			w = botW
		}
		if w < 6 {
			w = 6
		}
		colWidths[i] = w
	}

	// Build top and bottom lines
	var topParts, bottomParts []string
	for i := 0; i < nodeCount; i++ {
		topParts = append(topParts,
			padRight(nodes[i].top, lipgloss.Width(nodes[i].top), colWidths[i]))
		bottomParts = append(bottomParts,
			padRight(nodes[i].bottom, lipgloss.Width(nodes[i].bottom), colWidths[i]))
	}

	topLine := "  " + strings.Join(topParts, styledSep)
	bottomLine := "  " + strings.Join(bottomParts, spaceSep)

	return topLine + "\n" + bottomLine
}

// ViewDiagnostic renders the failure diagnostic message.
func (p PreflightPanel) ViewDiagnostic() string {
	if p.message == "" {
		return ""
	}
	var lines []string
	for _, line := range wordWrap(p.message, 60) {
		lines = append(lines, "  "+p.failStyle.Render(line))
	}
	return strings.Join(lines, "\n")
}

// nodeState returns the icon, label, and detail text for a given node index.
// Node 0 (Device) is always pass. Nodes 1–4 map to checks[0]–checks[3].
func (p PreflightPanel) nodeState(index int, spinnerFrame string) (icon, label, detail string) {
	label = nodeLabels[index]

	// Device node — always green, show local IP
	if index == 0 {
		icon = p.passStyle.Render("●")
		if ip := localIP(); ip != "" {
			detail = p.mutedStyle.Render(ip)
		}
		return icon, label, detail
	}

	checkIdx := index - 1 // map node index to check index

	if checkIdx < len(p.checks) {
		c := p.checks[checkIdx]
		if c.Passed {
			icon = p.passStyle.Render("●")
			detail = p.nodeDetail(index, c)
		} else {
			icon = p.failStyle.Render("●")
			detail = p.failStyle.Render("timeout")
		}
	} else if checkIdx == len(p.checks) && p.Active {
		// Currently being checked — show spinner
		if spinnerFrame != "" {
			icon = spinnerFrame
		} else {
			icon = p.mutedStyle.Render("○")
		}
	} else {
		// Pending
		icon = p.mutedStyle.Render("○")
	}

	return icon, label, detail
}

// nodeDetail returns the subtext for a completed check.
// index is the node index (1–4), not the check index.
func (p PreflightPanel) nodeDetail(index int, c preflight.CheckResult) string {
	switch index {
	case 1: // Router — gateway IP
		if c.Detail != "" {
			return p.mutedStyle.Render(c.Detail)
		}
	case 2: // Internet — IP
		if c.Detail != "" {
			return p.mutedStyle.Render(c.Detail)
		}
	case 3: // DNS — latency
		return p.mutedStyle.Render(fmt.Sprintf("%dms", int(c.Latency)))
	case 4: // Server — latency
		return p.mutedStyle.Render(fmt.Sprintf("%dms", int(c.Latency)))
	}
	return ""
}

// padRight pads a string with spaces to reach the target rendered width.
func padRight(s string, renderedWidth, targetWidth int) string {
	if renderedWidth >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-renderedWidth)
}

// localIP returns the preferred outbound local IP address.
func localIP() string {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

func checkLabel(name preflight.CheckName) string {
	switch name {
	case preflight.CheckGateway:
		return "Router"
	case preflight.CheckInternet:
		return "Internet"
	case preflight.CheckDNS:
		return "DNS"
	case preflight.CheckTestServer:
		return "Server"
	default:
		return string(name)
	}
}

func wordWrap(s string, width int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{s}
	}

	var lines []string
	current := words[0]

	for _, word := range words[1:] {
		if len(current)+1+len(word) > width {
			lines = append(lines, current)
			current = word
		} else {
			current += " " + word
		}
	}
	lines = append(lines, current)

	return lines
}
