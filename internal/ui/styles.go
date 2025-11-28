package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	Primary   = lipgloss.Color("#F59E0B") // Amber/Gold - represents speed/light
	Secondary = lipgloss.Color("#3B82F6") // Blue
	Success   = lipgloss.Color("#10B981") // Green
	Warning   = lipgloss.Color("#F59E0B") // Amber
	Error     = lipgloss.Color("#EF4444") // Red
	Muted     = lipgloss.Color("#6B7280") // Gray
	White     = lipgloss.Color("#FFFFFF") // White
)

// Styles
var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	KeyStyle = lipgloss.NewStyle().
			Foreground(Muted)

	ValueStyle = lipgloss.NewStyle().
			Foreground(White)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
)

// Banner returns the ASCII art banner for Lightspeed
func Banner() string {
	banner := `
 █   ▀█▀ █▀▀▀ █  █ ▀▀█▀▀ █▀▀ █▀▀█ █▀▀ █▀▀ █▀▀▄
 █    █  █ ▀█ █▀▀█   █   ▀▀█ █  █ █▀▀ █▀▀ █  █
 ▀▀▀ ▀▀▀ ▀▀▀▀ ▀  ▀   ▀   ▀▀▀ █▀▀▀ ▀▀▀ ▀▀▀ ▀▀▀ `
	return TitleStyle.Render(banner)
}

// Divider returns a styled divider line
func Divider() string {
	return MutedStyle.Render(strings.Repeat("─", 50))
}

// VersionLine returns a styled version string
func VersionLine(version string) string {
	return MutedStyle.Render("  version ") + HighlightStyle.Render(version)
}

// PrintHeader prints the full header with banner, dividers, and version
func PrintHeader(version string) {
	fmt.Println(Divider())
	fmt.Println(Banner())
	fmt.Println(VersionLine(version))
	fmt.Println()
	fmt.Println(Divider())
}

// Header returns a styled section header
func Header(text string) string {
	return BoldStyle.Render("▸ " + text)
}

// PrintSuccess prints a success message with checkmark
func PrintSuccess(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(SuccessStyle.Render("✓ " + msg))
}

// PrintError prints an error message with X mark
func PrintError(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(ErrorStyle.Render("✗ " + msg))
}

// PrintWarning prints a warning message
func PrintWarning(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(WarningStyle.Render("⚠ " + msg))
}

// PrintInfo prints an info message
func PrintInfo(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(InfoStyle.Render("• " + msg))
}

// PrintKeyValue prints a formatted key-value pair
func PrintKeyValue(key, value string) {
	fmt.Printf("%s: %s\n", KeyStyle.Render(key), ValueStyle.Render(value))
}

// Highlight returns highlighted text
func Highlight(s string) string {
	return HighlightStyle.Render(s)
}

// Muted returns muted/gray text
func Muted(s string) string {
	return MutedStyle.Render(s)
}

// Bold returns bold text
func Bold(s string) string {
	return BoldStyle.Render(s)
}
