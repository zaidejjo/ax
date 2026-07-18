package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ─── Request Pane ────────────────────────────────────────────────────────────

// requestPane holds the request builder form: method selector + URL input.
// In Phase 2 it will gain headers and body textareas.
type requestPane struct {
	// method is the currently selected HTTP method (GET, POST, etc.).
	method string

	// urlInput is the focused text input for the request URL.
	urlInput textinput.Model

	// hasFocus tracks whether this pane is the active pane in the root model.
	hasFocus bool
}

// newRequestPane creates the request builder pane with default values.
func newRequestPane() requestPane {
	ui := textinput.New()
	ui.Placeholder = "https://api.example.com/endpoint"
	ui.Prompt = "▌ "
	ui.SetWidth(80) // initial width; resized on each render
	ui.Blur()

	return requestPane{
		method:   "GET",
		urlInput: ui,
	}
}

func (p requestPane) Init() tea.Cmd {
	return nil
}

// Focus gives keyboard focus to the URL input. Returns a Cmd for the
// textinput's focus animation if any.
func (p requestPane) Focus() (requestPane, tea.Cmd) {
	p.hasFocus = true
	cmd := p.urlInput.Focus()
	return p, cmd
}

// Blur removes keyboard focus from all inputs in this pane.
func (p requestPane) Blur() requestPane {
	p.hasFocus = false
	p.urlInput.Blur()
	return p
}

// Value returns the URL entered in the request pane.
func (p requestPane) Value() string {
	return p.urlInput.Value()
}

// SetValue sets the URL text input to the given value, replacing any existing
// text. This is used when loading a history entry or pasting from clipboard.
func (p *requestPane) SetValue(val string) {
	p.urlInput.SetValue(val)
	p.urlInput.SetCursor(len(val))
}

// Method returns the currently selected HTTP method.
func (p requestPane) Method() string {
	return p.method
}

// Update handles events for the request pane.
func (p requestPane) Update(msg tea.Msg) (requestPane, tea.Cmd) {
	// Only process input events when this pane has focus.
	if !p.hasFocus {
		return p, nil
	}

	// Delegate to the URL text input.
	var cmd tea.Cmd
	p.urlInput, cmd = p.urlInput.Update(msg)
	return p, cmd
}

// View renders the request builder within the given dimensions.
// The caller wraps this string with the pane border style.
func (p requestPane) View(width, height int) string {
	// Set the textinput width to fill available space.
	p.urlInput.SetWidth(width - len(p.urlInput.Prompt))

	// ── Method line ──────────────────────────────────────────────────────
	label := LabelStyle.Render("Method")

	cycleHint := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Render("")

	methodDisplay := MethodColor(p.method).Render(p.method)
	methodLine := lipgloss.JoinHorizontal(
		lipgloss.Center,
		label,
		lipgloss.NewStyle().Padding(0, 1).Render(methodDisplay),
		cycleHint,
	)

	// ── URL line ─────────────────────────────────────────────────────────
	urlLabel := LabelStyle.Render("URL")
	urlLine := lipgloss.JoinHorizontal(
		lipgloss.Center,
		urlLabel,
		lipgloss.NewStyle().Padding(0, 1).Render(p.urlInput.View()),
	)

	// ── Instructions ─────────────────────────────────────────────────────
	hint := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Padding(0, 0).
		Render("Enter or Ctrl+R to send  •  Ctrl+P to paste  •  m to cycle method")

	// ── Assemble ─────────────────────────────────────────────────────────
	s := PaneTitleStyle.Render("Request Builder") + "\n\n"
	s += methodLine + "\n"
	s += urlLine + "\n\n"
	s += hint

	// Fill remaining space.
	lines := stringsCount(s, "\n")
	if lines < height {
		s += stringsRepeat("\n", height-lines)
	}

	return s
}

// cycleMethod rotates the HTTP method through common verbs.
func (p *requestPane) cycleMethod() {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	for i, m := range methods {
		if m == p.method {
			p.method = methods[(i+1)%len(methods)]
			return
		}
	}
	p.method = methods[0]
}
