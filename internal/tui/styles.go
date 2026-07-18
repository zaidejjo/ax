// Package tui provides the Bubble Tea TUI for ax — a terminal-based API client.
package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

// ─── Premium Dark Palette ────────────────────────────────────────────────────
//
//	Base (bg):       #0A0A0B  — Pitch-black background
//	Surface:          #141416  — Component surface
//	Surface2:         #1C1C1F  — Elevated surface / hover
//	Border (inactive):#3A3A40  — Subtle slate border
//	Border (active):  #22D3EE  — Cyan glow for focused pane
//	Accent primary:   #A855F7  — Vibrant purple (titles, primary accent)
//	Accent secondary: #22D3EE  — Cyan (secondary accent, highlights)
//	Text primary:     #E4E4E7  — Zinc-200 body text
//	Text muted:       #71717A  — Zinc-500 secondary/meta text
//	Text dim:         #52525B  — Zinc-600 status bar hints
//
//	HTTP methods:
//	  GET    #22C55E  Green
//	  POST   #3B82F6  Blue
//	  PUT    #F59E0B  Amber
//	  PATCH  #A855F7  Purple
//	  DELETE #EF4444  Red
// ──────────────────────────────────────────────────────────────────────────────

var (
	// Background and surface.
	bgBase     = lipgloss.Color("#0A0A0B")
	bgSurface  = lipgloss.Color("#141416")
	bgSurface2 = lipgloss.Color("#1C1C1F")

	// Border colors.
	borderInactive = lipgloss.Color("#3A3A40")
	borderActive   = lipgloss.Color("#22D3EE")

	// Accent colors.
	accentPrimary   = lipgloss.Color("#A855F7")
	accentSecondary = lipgloss.Color("#22D3EE")

	// Text colors.
	textPrimary = lipgloss.Color("#E4E4E7")
	textMuted   = lipgloss.Color("#71717A")
	textDim     = lipgloss.Color("#52525B")

	// HTTP method colors (vibrant).
	methodGreen  = lipgloss.Color("#22C55E")
	methodBlue   = lipgloss.Color("#3B82F6")
	methodAmber  = lipgloss.Color("#F59E0B")
	methodPurple = lipgloss.Color("#A855F7")
	methodRed    = lipgloss.Color("#EF4444")

	// Status code colors.
	status2xx = lipgloss.Color("#22C55E")
	status3xx = lipgloss.Color("#3B82F6")
	status4xx = lipgloss.Color("#F97316") // orange
	status5xx = lipgloss.Color("#EF4444")

	// Semantic aliases.
	primaryColor = accentPrimary
	mutedColor   = borderInactive
	subtleColor  = textMuted
	successColor = methodGreen
	warningColor = methodAmber
	errorColor   = methodRed
)

// ─── Border Styles ───────────────────────────────────────────────────────────
//
// Uses sharp NormalBorder (thin lines: ┌┐└┘│─) for a clean, premium feel.
// Active pane glows with cyan border; inactive panes use subtle slate.

var (
	// ActiveBorder is used for the currently focused pane.
	// Sharp thin cyan border.
	ActiveBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(borderActive).
			Padding(0, 1)

	// InactiveBorder is used for non-focused panes.
	// Sharp thin slate border.
	InactiveBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(borderInactive).
			Padding(0, 1)

	// SidebarDivider is a thin horizontal separator line between sidebar entries.
	SidebarDivider = lipgloss.NewStyle().
			Foreground(borderInactive).
			Render("─")
)

// ─── Header & Footer ─────────────────────────────────────────────────────────

var (
	// HeaderStyle — bold purple, no background, minimal padding.
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 2).
			Foreground(accentPrimary)

	// StatusBarStyle — dim text on base background, for the bottom keybinding bar.
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(textDim).
			Padding(0, 2)

	// StatusBarDivider is the thin top border for the status bar.
	StatusBarDivider = lipgloss.NewStyle().
				Foreground(borderInactive).
				Render("─")
)

// ─── Pane Title ───────────────────────────────────────────────────────────────

var (
	// PaneTitleStyle — bold purple pane header.
	PaneTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentPrimary).
			Padding(0, 0)

	// PlaceholderStyle — muted italic for empty-state hints.
	PlaceholderStyle = lipgloss.NewStyle().
				Foreground(textMuted).
				Italic(true)
)

// ─── Request Pane Styles ──────────────────────────────────────────────────────

var (
	LabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(textMuted).
			Width(8).
			Align(lipgloss.Right)

	InputStyle = lipgloss.NewStyle().
			Padding(0, 0)
)

// MethodColor returns a style for the given HTTP method.
func MethodColor(method string) lipgloss.Style {
	switch method {
	case "GET":
		return lipgloss.NewStyle().Bold(true).Foreground(methodGreen)
	case "POST":
		return lipgloss.NewStyle().Bold(true).Foreground(methodBlue)
	case "PUT":
		return lipgloss.NewStyle().Bold(true).Foreground(methodAmber)
	case "PATCH":
		return lipgloss.NewStyle().Bold(true).Foreground(methodPurple)
	case "DELETE":
		return lipgloss.NewStyle().Bold(true).Foreground(methodRed)
	default:
		return lipgloss.NewStyle().Bold(true).Foreground(textMuted)
	}
}

// ─── Response Pane Styles ─────────────────────────────────────────────────────

var (
	Status2xxStyle = lipgloss.NewStyle().Bold(true).Foreground(status2xx)
	Status3xxStyle = lipgloss.NewStyle().Bold(true).Foreground(status3xx)
	Status4xxStyle = lipgloss.NewStyle().Bold(true).Foreground(status4xx)
	Status5xxStyle = lipgloss.NewStyle().Bold(true).Foreground(status5xx)

	MetaStyle    = lipgloss.NewStyle().Foreground(textMuted)
	BodyStyle    = lipgloss.NewStyle().Padding(0, 0)
	ErrorStyle   = lipgloss.NewStyle().Foreground(errorColor)
	SuccessStyle = lipgloss.NewStyle().Foreground(successColor)

	// MetadataLineStyle formats the compact status | time | size line.
	MetadataLineStyle = lipgloss.NewStyle().Foreground(textMuted)
)

// StatusStyle returns a style for the given HTTP status code.
func StatusStyle(code int) lipgloss.Style {
	switch {
	case code < 200:
		return Status2xxStyle
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
				Foreground(accentPrimary)

	SidebarItemStyle = lipgloss.NewStyle().
				Padding(0, 1)

	SidebarSelectedStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Background(accentPrimary).
				Foreground(lipgloss.Color("#0A0A0B"))

	SidebarHelpStyle = lipgloss.NewStyle().
				Foreground(textMuted).
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

// formatBytes returns a human-readable byte size string (e.g., "1.2 KB").
func formatBytes(b int64) string {
	switch {
	case b < 1024:
		return fmt.Sprintf("%d B", b)
	case b < 1024*1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
	}
}
