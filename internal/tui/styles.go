// Package tui provides the Bubble Tea TUI for ax — a terminal-based API client.
package tui

import "charm.land/lipgloss/v2"

// ─── Catppuccin Macchiato Palette ────────────────────────────────────────────
//
// Reference: https://github.com/catppuccin/catppuccin
//
//	Base      #1E2030    Surface2  #6C7086    Text      #CAD3F5
//	Mantle    #181926    Overlay0  #8087A2    Lavender  #B7BDF8
//	Crust     #11111B    Overlay1  #939AB7    Mauve     #C6A0F6
//	Surface0  #363A4F    Overlay2  #A5ADCB    Blue      #8AADF4
//	Surface1  #494D64    Subtext0  #B8C0E0    Green     #A6DA95
//	                                  Yellow    #EED49F
//	                                  Peach     #F5A97F
//	                                  Red       #ED8796
//	                                  Maroon    #EE99A0
// ──────────────────────────────────────────────────────────────────────────────

// ─── Catppuccin Macchiato Colors ─────────────────────────────────────────────

var (
	// Base and surface colors.
	baseColor     = lipgloss.Color("#1E2030")
	mantleColor   = lipgloss.Color("#181926")
	crustColor    = lipgloss.Color("#11111B")
	surface0Color = lipgloss.Color("#363A4F")
	surface1Color = lipgloss.Color("#494D64")
	surface2Color = lipgloss.Color("#6C7086")

	// Text and overlay colors.
	textColor     = lipgloss.Color("#CAD3F5")
	subtext0Color = lipgloss.Color("#B8C0E0")
	subtext1Color = lipgloss.Color("#CAD3F5")
	overlay0Color = lipgloss.Color("#8087A2")
	overlay1Color = lipgloss.Color("#939AB7")
	overlay2Color = lipgloss.Color("#A5ADCB")

	// Accent colors.
	lavenderColor = lipgloss.Color("#B7BDF8") // Active borders, highlights
	mauveColor    = lipgloss.Color("#C6A0F6") // Pane titles, primary accent
	blueColor     = lipgloss.Color("#8AADF4") // PUT method
	sapphireColor = lipgloss.Color("#7BD4F4") // informational

	// Semantic colors.
	greenColor  = lipgloss.Color("#A6DA95") // 2xx, GET
	yellowColor = lipgloss.Color("#EED49F") // 3xx/4xx, POST
	peachColor  = lipgloss.Color("#F5A97F") // subtle warnings
	redColor    = lipgloss.Color("#ED8796") // 5xx, DELETE, errors
	maroonColor = lipgloss.Color("#EE99A0") // alternative error
)

// ─── Pointers for convenience ─────────────────────────────────────────────────

var (
	// primaryColor is the main accent for active borders.
	primaryColor = lavenderColor

	// successColor is used for success states.
	successColor = greenColor

	// warningColor is used for warning states.
	warningColor = yellowColor

	// errorColor is used for error states.
	errorColor = redColor

	// mutedColor is used for secondary text and inactive borders.
	mutedColor = surface2Color

	// subtleColor is used for meta information.
	subtleColor = overlay0Color
)

// ─── Border Styles ───────────────────────────────────────────────────────────

var (
	// ActiveBorder is used for the currently focused pane.
	// Uses a clean Lavender thick border.
	ActiveBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lavenderColor).
			Padding(0, 1)

	// InactiveBorder is used for non-focused panes.
	// Uses a very muted Surface2 rounded border.
	InactiveBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(surface2Color).
			Padding(0, 1)
)

// ─── Header & Footer ─────────────────────────────────────────────────────────

var (
	// HeaderStyle is minimal — bold Mauve text, no background.
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 2).
			Foreground(mauveColor)

	// FooterStyle uses muted overlay colors.
	FooterStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Foreground(overlay1Color)
)

// ─── Pane Title ───────────────────────────────────────────────────────────────

var (
	// PaneTitleStyle uses Mauve without bold for a cleaner look.
	PaneTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(mauveColor).
			Padding(0, 0)

	// PlaceholderStyle uses muted surface2.
	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(surface2Color).
				Italic(true)
)

// ─── Request Pane Styles ──────────────────────────────────────────────────────

var (
	LabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(overlay2Color).
			Width(8).
			Align(lipgloss.Right)

	InputStyle = lipgloss.NewStyle().
			Padding(0, 0)

	MethodStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(greenColor)
)

// MethodColor returns a style for the given HTTP method using Catppuccin colors.
func MethodColor(method string) lipgloss.Style {
	switch method {
	case "GET":
		return lipgloss.NewStyle().Bold(true).Foreground(greenColor)
	case "POST":
		return lipgloss.NewStyle().Bold(true).Foreground(yellowColor)
	case "PUT":
		return lipgloss.NewStyle().Bold(true).Foreground(blueColor)
	case "PATCH":
		return lipgloss.NewStyle().Bold(true).Foreground(mauveColor)
	case "DELETE":
		return lipgloss.NewStyle().Bold(true).Foreground(redColor)
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(overlay2Color)
	}
}

// ─── Response Pane Styles ─────────────────────────────────────────────────────

var (
	Status2xxStyle = lipgloss.NewStyle().Bold(true).Foreground(greenColor)
	Status3xxStyle = lipgloss.NewStyle().Bold(true).Foreground(yellowColor)
	Status4xxStyle = lipgloss.NewStyle().Bold(true).Foreground(peachColor)
	Status5xxStyle = lipgloss.NewStyle().Bold(true).Foreground(redColor)

	MetaStyle    = lipgloss.NewStyle().Foreground(overlay1Color)
	BodyStyle    = lipgloss.NewStyle().Padding(0, 0)
	ErrorStyle   = lipgloss.NewStyle().Foreground(redColor)
	SuccessStyle = lipgloss.NewStyle().Foreground(greenColor)
)

// StatusStyle returns a style for the given HTTP status code.
func StatusStyle(code int) lipgloss.Style {
	switch {
	case code < 200:
		return Status2xxStyle // informational
	case code < 300:
		return Status2xxStyle
	case code < 400:
		return Status3xxStyle
	case code < 500:
		return Status4xxStyle
	default:
		return Status5xxStyle
	}
}

// ─── Sidebar Styles ──────────────────────────────────────────────────────────

var (
	SidebarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(mauveColor).
				Underline(true)

	SidebarItemStyle = lipgloss.NewStyle().
				Padding(0, 1)

	SidebarSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(lavenderColor).
				Foreground(crustColor)

	SidebarHelpStyle = lipgloss.NewStyle().
				Foreground(surface2Color).
				Italic(true).
				Padding(0, 1)
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

// PaneBorder returns the appropriate border style based on whether the pane
// is the active (focused) pane.
func PaneBorder(active bool, width int) lipgloss.Style {
	if active {
		return ActiveBorder.Width(width)
	}
	return InactiveBorder.Width(width)
}
