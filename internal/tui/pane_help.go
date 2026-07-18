package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// ─── Help Overlay ────────────────────────────────────────────────────────────

// helpOverlay renders a full keybinding reference and xh/httpie syntax
// examples as a centered modal overlay on top of the normal TUI view.
type helpOverlay struct {
	visible bool
}

func newHelpOverlay() helpOverlay {
	return helpOverlay{}
}

func (h *helpOverlay) Toggle()      { h.visible = !h.visible }
func (h *helpOverlay) Show()        { h.visible = true }
func (h *helpOverlay) Hide()        { h.visible = false }
func (h helpOverlay) Visible() bool { return h.visible }

// View renders the help overlay centered on screen. It returns the overlay
// string (which should be layered on top of the normal view) or an empty
// string if the overlay is hidden.
func (h helpOverlay) View(screenW, screenH int) string {
	if !h.visible {
		return ""
	}

	// ── Build content lines ────────────────────────────────────────────
	content := h.buildContent()

	// ── Box styling ────────────────────────────────────────────────────
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(content.width + 4).
		Align(lipgloss.Left)

	rendered := boxStyle.Render(content.text)

	// ── Semi-transparent backdrop ──────────────────────────────────────
	// Render a dim overlay covering the full screen, then place the box
	// on top. Use a dark background with the box centered.
	overlayBg := lipgloss.NewStyle().
		Width(screenW).
		Height(screenH).
		Background(lipgloss.Color("#000000")).
		Foreground(lipgloss.Color("#000000")).
		Render(strings.Repeat(" ", screenW*screenH))

	// Position the box in the center of the screen.
	boxLines := strings.Split(rendered, "\n")
	bgLines := strings.Split(overlayBg, "\n")

	startY := (screenH - len(boxLines)) / 2
	if startY < 0 {
		startY = 0
	}
	startX := (screenW - content.width - 6) / 2
	if startX < 0 {
		startX = 0
	}

	var result strings.Builder
	for y := 0; y < screenH && y < len(bgLines); y++ {
		if y >= startY && y-startY < len(boxLines) {
			boxLine := boxLines[y-startY]
			// Concatenate background before box, box content, then background after.
			prefix := ""
			if startX > 0 && startX <= len(bgLines[y]) {
				prefix = bgLines[y][:startX]
			}
			suffix := ""
			boxEnd := startX + lipgloss.Width(boxLine)
			if boxEnd < len(bgLines[y]) {
				suffix = bgLines[y][boxEnd:]
			}
			result.WriteString(prefix + boxLine + suffix)
		} else {
			result.WriteString(bgLines[y])
		}
		if y < screenH-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

type helpContent struct {
	text  string
	width int
}

func (h helpOverlay) buildContent() helpContent {
	var b strings.Builder

	// ── Title ──────────────────────────────────────────────────────────
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Width(50).
		Align(lipgloss.Center).
		Render("Help & Keybindings")
	b.WriteString(title)
	b.WriteString("\n\n")

	// ── Navigation ─────────────────────────────────────────────────────
	b.WriteString(sectionStyle.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString(helpRow("Tab", "Cycle focus between panes (forward)"))
	b.WriteString(helpRow("Shift+Tab", "Cycle focus between panes (backward)"))
	b.WriteString("\n")

	// ── Request Building ───────────────────────────────────────────────
	b.WriteString(sectionStyle.Render("Request Building"))
	b.WriteString("\n")
	b.WriteString(helpRow("Enter / Ctrl+R", "Execute the current request"))
	b.WriteString(helpRow("m", "Cycle HTTP method (when URL input not focused)"))
	b.WriteString(helpRow("Ctrl+P", "Paste from clipboard into URL"))
	b.WriteString(helpRow("Ctrl+Y", "Copy response body to clipboard"))
	b.WriteString("\n")

	// ── Sidebar ────────────────────────────────────────────────────────
	b.WriteString(sectionStyle.Render("Sidebar (History)"))
	b.WriteString("\n")
	b.WriteString(helpRow("↑/↓", "Navigate history entries"))
	b.WriteString(helpRow("Enter", "Load selected entry into request"))
	b.WriteString(helpRow("Ctrl+D", "Delete selected entry"))
	b.WriteString(helpRow("Ctrl+H", "Toggle sidebar visibility"))
	b.WriteString("\n")

	// ── Response Pane ──────────────────────────────────────────────────
	b.WriteString(sectionStyle.Render("Response Viewer"))
	b.WriteString("\n")
	b.WriteString(helpRow("↑/↓/PgUp/PgDn", "Scroll response body"))
	b.WriteString(helpRow("g / G", "Go to top / bottom"))
	b.WriteString(helpRow("h", "Toggle headers / body-only view"))
	b.WriteString("\n")

	// ── xh / httpie Syntax ─────────────────────────────────────────────
	b.WriteString(sectionStyle.Render("Single-Line Syntax (xh/httpie style)"))
	b.WriteString("\n")
	b.WriteString(helpRow(":PORT/path", "Shorthand for http://localhost:PORT/path"))
	b.WriteString(helpRow("/path", "Shorthand for http://localhost/path"))
	b.WriteString(helpRow("Key:Value", "HTTP request header"))
	b.WriteString(helpRow("key==value", "JSON body field"))
	b.WriteString(helpRow("key=value", "Form-encoded body field"))
	b.WriteString("\n")

	// ── Examples ───────────────────────────────────────────────────────
	b.WriteString(sectionStyle.Render("Examples"))
	b.WriteString("\n")
	b.WriteString(helpRow(":8080/api/users", "GET localhost:8080"))
	b.WriteString(helpRow(`POST :8080/api/users name=="John"`, "POST with JSON body"))
	b.WriteString(helpRow("https://api.example.com Authorization:token123", "Custom header GET"))
	b.WriteString(helpRow(`PATCH example.com/res/1 name=="Updated"`, "PATCH with JSON"))
	b.WriteString("\n")

	// ── General ────────────────────────────────────────────────────────
	b.WriteString(sectionStyle.Render("General"))
	b.WriteString("\n")
	b.WriteString(helpRow("?", "Toggle this help overlay"))
	b.WriteString(helpRow("q / Ctrl+C", "Quit ax"))
	b.WriteString("\n")

	// ── Dismiss hint ───────────────────────────────────────────────────
	dismiss := lipgloss.NewStyle().
		Foreground(subtleColor).
		Italic(true).
		Width(52).
		Align(lipgloss.Center).
		Render("Press any key to close")
	b.WriteString(dismiss)

	// Determine the widest line for box sizing.
	text := b.String()
	maxWidth := 0
	for _, line := range strings.Split(text, "\n") {
		w := lipgloss.Width(line)
		if w > maxWidth {
			maxWidth = w
		}
	}
	if maxWidth < 52 {
		maxWidth = 52
	}

	return helpContent{text: text, width: maxWidth}
}

// helpRow formats a key + description pair with aligned columns.
func helpRow(key, desc string) string {
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(successColor).
		Width(18).
		Align(lipgloss.Left)

	descStyle := lipgloss.NewStyle().
		Foreground(textPrimary)

	return fmt.Sprintf("  %s %s\n", keyStyle.Render(key), descStyle.Render(desc))
}

// sectionStyle is a lipgloss style for section headers in the help overlay.
var sectionStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(primaryColor).
	Underline(true)
