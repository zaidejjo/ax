package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/chroma/v2/quick"

	"github.com/zaidejjo/ax/internal/client"
)

// ─── Response Pane ───────────────────────────────────────────────────────────

// responsePane displays the HTTP response status, headers, and body in a
// scrollable viewport. It supports syntax-highlighted JSON with automatic
// pretty-printing, a loading spinner, and a body-only toggle.
type responsePane struct {
	// viewport provides scrollable text rendering.
	viewport viewport.Model

	// The raw response data currently displayed.
	status     string
	proto      string
	statusCode int
	headers    map[string][]string
	body       string
	duration   string
	bodySize   int64 // raw body size in bytes

	// loaded is true once a response or error has been received.
	loaded bool

	// err stores the last error, if any.
	err error

	// hasFocus tracks whether this pane is the active pane.
	hasFocus bool

	// loading is true while an HTTP request is in-flight.
	loading bool

	// spinnerText holds the current animation frame of the root model's
	// spinner. Updated on each spinner.TickMsg from the root model.
	spinnerText string

	// bodyOnly toggles between "Full Response (Headers + Body)" and
	// "Body Only" view. Toggled with the 'h' key when focused.
	bodyOnly bool
}

// newResponsePane creates the response viewer pane with a placeholder message.
func newResponsePane() responsePane {
	vp := viewport.New(
		viewport.WithWidth(80),
		viewport.WithHeight(10),
	)

	return responsePane{
		viewport: vp,
		loaded:   false,
		loading:  false,
	}
}

func (p responsePane) Init() tea.Cmd {
	return nil
}

// Focus enables keyboard scrolling for the viewport.
func (p responsePane) Focus() (responsePane, tea.Cmd) {
	p.hasFocus = true
	return p, nil
}

// Blur disables nothing — the viewport always accepts scroll events from the
// root model when this pane is active.
func (p responsePane) Blur() responsePane {
	p.hasFocus = false
	return p
}

// ─── Body Accessors ──────────────────────────────────────────────────────────

// HasBody returns true if a response body has been received and stored.
func (p responsePane) HasBody() bool {
	return p.body != ""
}

// BodyText returns the raw response body text.
func (p responsePane) BodyText() string {
	return p.body
}

// ─── State Setters ───────────────────────────────────────────────────────────

// SetResponse stores a successful HTTP response, syntax-highlights the body,
// and updates the viewport content.
func (p *responsePane) SetResponse(resp *client.Response) {
	p.loading = false
	p.loaded = true
	p.err = nil
	p.status = resp.Status
	p.proto = resp.Proto
	p.statusCode = resp.StatusCode
	p.headers = resp.Headers
	p.body = string(resp.Body)
	p.duration = resp.Duration.Round(time.Millisecond).String()
	p.bodySize = resp.BodySize

	p.viewport.SetContent(p.formatResponse())
	p.viewport.GotoTop()
}

// SetError stores an error and displays it in the viewport with a clean
// red-styled error card.
func (p *responsePane) SetError(err error) {
	p.loading = false
	p.loaded = true
	p.err = err
	p.status = ""
	p.proto = ""
	p.body = ""
	p.duration = ""
	p.spinnerText = ""

	p.viewport.SetContent(p.formatError(err))
	p.viewport.GotoTop()
}

// SetLoading marks the pane as in-flight and displays the spinner animation.
// The root model must call SetSpinnerText on each spinner.TickMsg.
func (p *responsePane) SetLoading(spinnerText string) {
	p.loading = true
	p.loaded = false
	p.err = nil
	p.spinnerText = spinnerText
	p.viewport.SetContent(p.formatLoading())
}

// SetSpinnerText updates the spinner animation frame in the loading display.
// Called from the root model on each spinner.TickMsg.
func (p *responsePane) SetSpinnerText(text string) {
	p.spinnerText = text
	if p.loading {
		p.viewport.SetContent(p.formatLoading())
	}
}

// ─── Body/Header Toggle ──────────────────────────────────────────────────────

// ToggleBodyView switches between full response view and body-only view.
func (p *responsePane) ToggleBodyView() {
	p.bodyOnly = !p.bodyOnly
	if p.loaded && p.err == nil {
		p.viewport.SetContent(p.formatResponse())
		p.viewport.GotoTop()
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (p responsePane) Update(msg tea.Msg) (responsePane, tea.Cmd) {
	// Only process events when this pane has focus.
	if !p.hasFocus {
		return p, nil
	}

	// Handle key events before delegating to viewport.
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "h":
			p.ToggleBodyView()
			return p, nil
		}
	}

	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)
	return p, cmd
}

// ─── View ────────────────────────────────────────────────────────────────────

func (p responsePane) View(width, height int) string {
	p.viewport.SetWidth(width)
	p.viewport.SetHeight(height)

	// ── Loading state: show spinner + message in viewport ───────────────
	if p.loading {
		return p.viewport.View()
	}

	// ── Initial state: show placeholder text (no viewport border) ──────
	if !p.loaded {
		s := PaneTitleStyle.Render("Response") + "\n\n"
		s += PlaceholderStyle.Width(width).Render("Press Ctrl+R to send a request")
		lines := stringsCount(s, "\n")
		if lines < height {
			s += stringsRepeat("\n", height-lines)
		}
		return s
	}

	// ── Response or error: show viewport content ───────────────────────
	return p.viewport.View()
}

// ─── Formatting Helpers ──────────────────────────────────────────────────────

// formatLoading builds the loading animation content for the viewport.
func (p responsePane) formatLoading() string {
	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("  %s  Sending request...\n\n", p.spinnerText))
	b.WriteString(MetaStyle.Render("  Press Ctrl+R to re-send or q to quit"))
	return b.String()
}

// formatResponse builds the full response text for the viewport, including
// syntax-highlighted JSON body when applicable. When bodyOnly is true, only
// the body (with highlighting) is shown.
func (p responsePane) formatResponse() string {
	var b strings.Builder

	if p.bodyOnly {
		// ── Body-only mode — just the body content ─────────────────────
		if p.body != "" {
			b.WriteString(p.renderBody())
		}
		// Ensure trailing newline.
		if !strings.HasSuffix(b.String(), "\n") {
			b.WriteString("\n")
		}
		return b.String()
	}

	// ── Full response mode — metadata line, headers, body ─────────────

	// Compact metadata line:  HTTP/1.1 200 OK  │  142ms  │  1.2 KB
	statusColor := StatusStyle(p.statusCode)
	metaStatus := statusColor.Render(p.proto + " " + p.status)
	metaDuration := MetaStyle.Render(p.duration)
	metaSize := MetaStyle.Render(formatBytes(p.bodySize))
	metaLine := fmt.Sprintf("%s  │  %s  │  %s", metaStatus, metaDuration, metaSize)
	b.WriteString(metaLine)
	b.WriteString("\n")

	// Thin divider.
	b.WriteString(MetadataLineStyle.Render(stringsRepeat("─", 40)))
	b.WriteString("\n\n")

	// Headers.
	for key, values := range p.headers {
		for _, v := range values {
			b.WriteString(MetaStyle.Render(fmt.Sprintf("%s: %s", key, v)))
			b.WriteString("\n")
		}
	}

	// Blank line before body.
	if len(p.headers) > 0 {
		b.WriteString("\n")
	}

	// Body with optional syntax highlighting.
	if p.body != "" {
		b.WriteString(p.renderBody())
	}

	// Ensure trailing newline.
	if !strings.HasSuffix(b.String(), "\n") {
		b.WriteString("\n")
	}

	return b.String()
}

// renderBody returns the response body, pretty-printed if it is valid JSON,
// and optionally syntax-highlighted.
func (p responsePane) renderBody() string {
	// Step 1: pretty-print if valid JSON.
	formatted := p.prettifyJSON(p.body)

	// Step 2: syntax-highlight if the content looks like JSON.
	if p.shouldHighlight() {
		highlighted := p.highlightJSON(formatted)
		return highlighted
	}
	return formatted
}

// prettifyJSON attempts to indent/minify the body into a readable format.
// If the body is not valid JSON, it is returned unchanged.
func (p responsePane) prettifyJSON(body string) string {
	trimmed := strings.TrimSpace(body)
	if len(trimmed) == 0 {
		return body
	}

	first := trimmed[0]
	if first != '{' && first != '[' {
		return body // not JSON-like
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(body), "", "  "); err != nil {
		return body // not valid JSON, return as-is
	}

	return pretty.String()
}

// shouldHighlight returns true if the response body appears to be JSON
// (detected either by Content-Type or by content inspection).
func (p responsePane) shouldHighlight() bool {
	// Check Content-Type header first.
	if ct, ok := p.headers["Content-Type"]; ok {
		for _, v := range ct {
			if strings.Contains(strings.ToLower(v), "json") {
				return true
			}
		}
	}

	// Fall back to content detection: starts with '{' or '[' and is valid JSON.
	trimmed := strings.TrimSpace(p.body)
	if len(trimmed) == 0 {
		return false
	}
	first := trimmed[0]
	if first == '{' || first == '[' {
		return json.Valid([]byte(p.body))
	}
	return false
}

// highlightJSON applies chroma syntax highlighting to a JSON string for
// terminal output using the catppuccin-macchiato dark theme.
func (p responsePane) highlightJSON(body string) string {
	var buf strings.Builder
	err := quick.Highlight(&buf, body, "json", "tty256", "catppuccin-macchiato")
	if err != nil {
		// Fall back to raw text if highlighting fails.
		return body
	}
	return buf.String()
}

// formatError builds an error display for the viewport.
func (p responsePane) formatError(err error) string {
	var b strings.Builder

	// Error header with icon.
	b.WriteString(ErrorStyle.Bold(true).Render("✗  Request Failed"))
	b.WriteString("\n\n")

	// Error details.
	b.WriteString(ErrorStyle.Render(err.Error()))
	b.WriteString("\n\n")

	// Helpful hint.
	b.WriteString(MetaStyle.Render("Check the URL, network connection, and try again with Ctrl+R"))

	// Add a visual separator and common troubleshooting tips.
	b.WriteString("\n\n")
	b.WriteString(MetaStyle.Render("── Common causes ──"))
	b.WriteString("\n")
	b.WriteString(MetaStyle.Render("•  URL is unreachable or the hostname is wrong"))
	b.WriteString("\n")
	b.WriteString(MetaStyle.Render("•  Connection refused (no server listening on that port)"))
	b.WriteString("\n")
	b.WriteString(MetaStyle.Render("•  Request timed out (server didn't respond in time)"))
	b.WriteString("\n")
	b.WriteString(MetaStyle.Render("•  TLS/SSL certificate verification failed"))

	return b.String()
}
