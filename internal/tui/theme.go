package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Theme holds all color styles for the TUI.
type Theme struct {
	Name string

	// Base styles
	Title     lipgloss.Style
	Subtitle  lipgloss.Style
	Muted     lipgloss.Style
	Bold      lipgloss.Style
	SpeedNum  lipgloss.Style
	SpeedUnit lipgloss.Style

	// Phase indicators
	Download lipgloss.Style
	Upload   lipgloss.Style
	Latency  lipgloss.Style

	// Bufferbloat grades
	GradeGood lipgloss.Style // A+, A
	GradeOk   lipgloss.Style // B
	GradeWarn lipgloss.Style // C
	GradeBad  lipgloss.Style // D, F

	// Sparkline
	SparkStyle lipgloss.Style

	// Progress bar colors
	ProgressDL lipgloss.Style
	ProgressUL lipgloss.Style

	// Footer
	FooterKey    lipgloss.Style
	FooterAction lipgloss.Style

	// Border
	Border lipgloss.Style
}

// DefaultTheme returns a vivid color theme.
func DefaultTheme() Theme {
	return Theme{
		Name:         "default",
		Title:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700")),
		Subtitle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#B0B0B0")),
		Muted:        lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		Bold:         lipgloss.NewStyle().Bold(true),
		SpeedNum:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")),
		SpeedUnit:    lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Download:     lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA")),
		Upload:       lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B9D")),
		Latency:      lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")),
		GradeGood:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D4AA")),
		GradeOk:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700")),
		GradeWarn:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8C00")),
		GradeBad:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4444")),
		SparkStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA")),
		ProgressDL:   lipgloss.NewStyle().Foreground(lipgloss.Color("#00D4AA")),
		ProgressUL:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B9D")),
		FooterKey:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD700")),
		FooterAction: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Border:       lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")),
	}
}

// ColorBlindTheme returns an Okabe-Ito friendly theme.
func ColorBlindTheme() Theme {
	return Theme{
		Name:         "colorblind",
		Title:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E69F00")),
		Subtitle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#B0B0B0")),
		Muted:        lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")),
		Bold:         lipgloss.NewStyle().Bold(true),
		SpeedNum:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")),
		SpeedUnit:    lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Download:     lipgloss.NewStyle().Foreground(lipgloss.Color("#0072B2")),
		Upload:       lipgloss.NewStyle().Foreground(lipgloss.Color("#D55E00")),
		Latency:      lipgloss.NewStyle().Foreground(lipgloss.Color("#CC79A7")),
		GradeGood:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#009E73")),
		GradeOk:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E69F00")),
		GradeWarn:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D55E00")),
		GradeBad:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#CC79A7")),
		SparkStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#0072B2")),
		ProgressDL:   lipgloss.NewStyle().Foreground(lipgloss.Color("#0072B2")),
		ProgressUL:   lipgloss.NewStyle().Foreground(lipgloss.Color("#D55E00")),
		FooterKey:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E69F00")),
		FooterAction: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		Border:       lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")),
	}
}

// MonochromeTheme returns a theme suitable for NO_COLOR environments.
func MonochromeTheme() Theme {
	return Theme{
		Name:         "mono",
		Title:        lipgloss.NewStyle().Bold(true),
		Subtitle:     lipgloss.NewStyle(),
		Muted:        lipgloss.NewStyle().Faint(true),
		Bold:         lipgloss.NewStyle().Bold(true),
		SpeedNum:     lipgloss.NewStyle().Bold(true),
		SpeedUnit:    lipgloss.NewStyle(),
		Download:     lipgloss.NewStyle(),
		Upload:       lipgloss.NewStyle(),
		Latency:      lipgloss.NewStyle(),
		GradeGood:    lipgloss.NewStyle().Bold(true),
		GradeOk:      lipgloss.NewStyle().Bold(true),
		GradeWarn:    lipgloss.NewStyle().Bold(true),
		GradeBad:     lipgloss.NewStyle().Bold(true),
		SparkStyle:   lipgloss.NewStyle(),
		ProgressDL:   lipgloss.NewStyle(),
		ProgressUL:   lipgloss.NewStyle(),
		FooterKey:    lipgloss.NewStyle().Bold(true),
		FooterAction: lipgloss.NewStyle(),
		Border:       lipgloss.NewStyle(),
	}
}

// ThemeFromName returns the theme for the given name, auto-detecting NO_COLOR.
func ThemeFromName(name string) Theme {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return MonochromeTheme()
	}
	switch name {
	case "colorblind":
		return ColorBlindTheme()
	case "mono":
		return MonochromeTheme()
	default:
		return DefaultTheme()
	}
}
