package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/zaidejjo/ax/internal/client"
	"github.com/zaidejjo/ax/internal/history"
)

// ─── Pane Index ──────────────────────────────────────────────────────────────

type paneIndex int

const (
	paneSidebar  paneIndex = iota // 0 — history list
	paneRequest                   // 1 — request builder
	paneResponse                  // 2 — response viewer
)

const numPanes = 3

// ─── Messages ────────────────────────────────────────────────────────────────

type focusMsg struct {
	direction int
}

type executeRequestMsg struct{}

type responseReceivedMsg struct {
	resp *client.Response
	err  error
}

// ─── History Messages ────────────────────────────────────────────────────────

// loadHistoryMsg triggers an async load of history entries from the DB.
type loadHistoryMsg struct{}

// historyLoadedMsg carries the result of a load operation.
type historyLoadedMsg struct {
	entries []history.Entry
	err     error
}

// deleteEntryMsg triggers an async delete of a specific history entry.
type deleteEntryMsg struct {
	id int64
}

// entryDeletedMsg carries the result of a delete operation.
type entryDeletedMsg struct {
	err error
}

// historySavedMsg is sent after a history entry is saved (fire-and-forget).
type historySavedMsg struct{}

// ─── Toast / Clipboard Messages ──────────────────────────────────────────────

// toastMsg carries a message to display in the footer for a few seconds.
type toastMsg struct {
	text string
}

// clearToastMsg clears the current toast text.
type clearToastMsg struct{}

// copyResponseMsg triggers a clipboard copy of the response body.
type copyResponseMsg struct {
	body string
}

// ─── Root Model ──────────────────────────────────────────────────────────────

type model struct {
	// Terminal dimensions (set on WindowSizeMsg).
	width  int
	height int
	ready  bool

	// Which pane has keyboard focus.
	activePane paneIndex

	// Sub-panes.
	sidebar  sidebarPane
	request  requestPane
	response responsePane

	// HTTP client for executing requests.
	client *client.Client

	// loading is true while an HTTP request is in-flight.
	loading bool

	// spinner animates while loading is true.
	spinner spinner.Model

	// history is the SQLite-backed persistent store.
	history *history.Store

	// historyPath tracks the DB path for error messages.
	historyPath string

	// The version string set at build time (injected by main).
	Version string

	// help is the keybinding reference overlay.
	help helpOverlay

	// toastText is an ephemeral message shown in the footer.
	toastText string

	// toastExpire is when the toast should be cleared.
	toastExpire time.Time

	// sidebarHidden toggles the history sidebar on/off with Ctrl+H.
	// When hidden, the request and response panes expand to full width.
	sidebarHidden bool
}

// New creates a fully initialized root model, including opening the SQLite
// history database.
func New(version string) *model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	m := &model{
		sidebar:    newSidebarPane(),
		request:    newRequestPane(),
		response:   newResponsePane(),
		client:     client.New(),
		spinner:    s,
		activePane: paneRequest,
		Version:    version,
		help:       newHelpOverlay(),
	}

	// Focus the initial active pane so the textinput accepts key input
	// immediately on startup.
	m.request, _ = m.request.Focus()

	// Open history store (non-fatal on error — history is optional).
	if path, err := historyDBPath(); err == nil {
		store, err := history.Open(path)
		if err == nil {
			m.history = store
			m.historyPath = path
		}
	}

	return m
}

// ─── Init ────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		func() tea.Msg {
			return m.spinner.Tick()
		},
	}

	// Load history on startup if the store is available.
	if m.history != nil {
		cmds = append(cmds, loadHistoryCmd(m.history))
	}

	return tea.Batch(cmds...)
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// ── Terminal resize ────────────────────────────────────────────────
	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height

		var cmd tea.Cmd
		m.sidebar, _ = m.sidebar.Update(msg)
		m.request, _ = m.request.Update(msg)
		m.response, cmd = m.response.Update(msg)
		return m, cmd

	// ── Spinner tick (animation frame) ────────────────────────────────
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.loading {
			m.response.SetSpinnerText(m.spinner.View())
		}
		return m, cmd

	// ── Key presses ───────────────────────────────────────────────────
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	// ── Focus shift (programmatic) ────────────────────────────────────
	case focusMsg:
		return m.handleFocusShift(msg)

	// ── Execute request (Ctrl+R) ──────────────────────────────────────
	case executeRequestMsg:
		return m.handleExecuteRequest()

	// ── Response received ─────────────────────────────────────────────
	case responseReceivedMsg:
		return m.handleResponseReceived(msg)

	// ── History loaded ────────────────────────────────────────────────
	case historyLoadedMsg:
		return m.handleHistoryLoaded(msg)

	// ── Entry deleted ─────────────────────────────────────────────────
	case entryDeletedMsg:
		return m.handleEntryDeleted(msg)

	// ── History saved (silent) ─────────────────────────────────────────
	case historySavedMsg:
		return m, nil

	// ── Copy response to clipboard ─────────────────────────────────────
	case copyResponseMsg:
		return m.handleCopyResponse(msg)

	// ── Toast message ──────────────────────────────────────────────────
	case toastMsg:
		m.toastText = msg.text
		m.toastExpire = time.Now().Add(4 * time.Second)
		return m, tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
			return clearToastMsg{}
		})

	// ── Clear toast ────────────────────────────────────────────────────
	case clearToastMsg:
		m.toastText = ""
		return m, nil

	// ── Paste text into URL input ─────────────────────────────────────
	case pasteTextMsg:
		m.request.SetValue(msg.text)
		m.toastText = "✓ Pasted from clipboard"
		m.toastExpire = time.Now().Add(2 * time.Second)
		return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return clearToastMsg{}
		})
	}

	return m, nil
}

// ─── Key Dispatch ────────────────────────────────────────────────────────────

func (m model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// ── Help overlay consumes all keys when visible ────────────────────
	if m.help.Visible() {
		m.help.Hide()
		return m, nil
	}

	switch msg.String() {
	// ── Quit ───────────────────────────────────────────────────────────
	case "q", "ctrl+c":
		return m, tea.Quit

	// ── Help overlay ─────────────────────────────────────────────────—
	case "?":
		m.help.Toggle()
		return m, nil

	// ── Focus switching ───────────────────────────────────────────────
	case "tab":
		return m.handleFocusShift(focusMsg{direction: +1})
	case "shift+tab":
		return m.handleFocusShift(focusMsg{direction: -1})

	// ── Execute request ───────────────────────────────────────────────
	case "ctrl+r":
		if m.loading {
			return m, nil
		}
		return m.handleExecuteRequest()

	// ── Cycle HTTP method (only when not typing in URL input) ─────────
	case "m":
		if m.activePane != paneRequest {
			m.request.cycleMethod()
			return m, nil
		}
		// Fall through to textinput when request pane is active.

	// ── Toggle sidebar visibility ─────────────────────────────────────
	case "ctrl+h":
		m.sidebarHidden = !m.sidebarHidden
		return m, nil

	// ── Paste from clipboard into URL input ───────────────────────────
	case "ctrl+p", "ctrl+v", "ctrl+shift+v":
		if m.activePane == paneRequest {
			return m, pasteURLFromClipboardCmd()
		}
		return m, nil

	// ── Copy response body to clipboard ───────────────────────────────
	case "ctrl+y":
		if !m.response.HasBody() {
			return m, nil
		}
		return m, func() tea.Msg {
			return copyResponseMsg{body: m.response.BodyText()}
		}

	// ── Enter: sidebar select or execute request ──────────────────────
	case "enter":
		switch m.activePane {
		case paneSidebar:
			return m.handleSidebarSelect()
		case paneRequest:
			if m.loading {
				return m, nil
			}
			return m.handleExecuteRequest()
		}
		// Fall through to delegate to active pane for other panes.

	// ── Sidebar: delete entry (Ctrl+D) ────────────────────────────────
	case "ctrl+d":
		if m.activePane == paneSidebar {
			return m.handleSidebarDelete()
		}
		// Fall through to delegate to active pane.
	}

	// Delegate to the active pane.
	var cmd tea.Cmd
	switch m.activePane {
	case paneSidebar:
		m.sidebar, cmd = m.sidebar.Update(msg)
	case paneRequest:
		m.request, cmd = m.request.Update(msg)
	case paneResponse:
		m.response, cmd = m.response.Update(msg)
	}
	return m, cmd
}

// ─── Focus Management ────────────────────────────────────────────────────────

func (m model) handleFocusShift(msg focusMsg) (tea.Model, tea.Cmd) {
	m.blurPane()

	next := int(m.activePane) + msg.direction
	if next < 0 {
		next = numPanes - 1
	} else if next >= numPanes {
		next = 0
	}
	m.activePane = paneIndex(next)

	cmd := m.focusPane()
	return m, cmd
}

func (m *model) blurPane() {
	switch m.activePane {
	case paneSidebar:
	case paneRequest:
		m.request = m.request.Blur()
	case paneResponse:
		m.response = m.response.Blur()
	}
}

func (m model) focusPane() tea.Cmd {
	switch m.activePane {
	case paneSidebar:
		return nil
	case paneRequest:
		var cmd tea.Cmd
		m.request, cmd = m.request.Focus()
		return cmd
	case paneResponse:
		var cmd tea.Cmd
		m.response, cmd = m.response.Focus()
		return cmd
	}
	return nil
}

// ─── Request Execution ───────────────────────────────────────────────────────

func (m model) handleExecuteRequest() (tea.Model, tea.Cmd) {
	input := m.request.Value()
	if input == "" {
		return m, nil
	}

	// Parse the input using the xh/httpie-style shorthand parser.
	parsed, err := client.Parse(input)
	if err != nil {
		m.response.SetError(fmt.Errorf("parse error: %w", err))
		m.activePane = paneResponse
		var cmd tea.Cmd
		m.response, cmd = m.response.Focus()
		return m, cmd
	}

	// Update the method badge to reflect what the parser detected.
	m.request.method = parsed.Method

	m.loading = true
	m.response.SetLoading(m.spinner.View())

	return m, func() tea.Msg {
		resp, err := m.client.Do(parsed)
		return responseReceivedMsg{resp: resp, err: err}
	}
}

// ─── Response Handling ───────────────────────────────────────────────────────

func (m model) handleResponseReceived(msg responseReceivedMsg) (tea.Model, tea.Cmd) {
	m.loading = false

	if msg.err != nil {
		m.response.SetError(msg.err)
		// Still switch to response pane so user sees the error.
		m.activePane = paneResponse
		var cmd tea.Cmd
		m.response, cmd = m.response.Focus()
		return m, cmd
	}

	m.response.SetResponse(msg.resp)

	// Auto-save to history.
	var saveCmd tea.Cmd
	if m.history != nil {
		entry := m.buildHistoryEntry(msg.resp)
		saveCmd = saveHistoryCmd(m.history, entry)
	}

	// Refresh sidebar list to include the new entry.
	var reloadCmd tea.Cmd
	if m.history != nil {
		reloadCmd = loadHistoryCmd(m.history)
	}

	// Switch focus to response pane so the user can scroll.
	m.activePane = paneResponse
	var focusCmd tea.Cmd
	m.response, focusCmd = m.response.Focus()

	return m, tea.Batch(focusCmd, saveCmd, reloadCmd)
}

// buildHistoryEntry converts an HTTP response into a history.Entry for saving.
func (m model) buildHistoryEntry(resp *client.Response) history.Entry {
	headers := make(map[string]string, len(resp.Headers))
	for key, values := range resp.Headers {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return history.Entry{
		Method:   m.request.Method(),
		URL:      m.request.Value(),
		Headers:  headers,
		Body:     string(resp.Body),
		Status:   resp.StatusCode,
		BodySize: resp.BodySize,
		Duration: resp.Duration,
	}
}

// ─── Sidebar Actions ─────────────────────────────────────────────────────────

// handleSidebarSelect loads the selected history entry into the request pane
// and switches focus to the request pane so the user can modify or re-run it.
func (m model) handleSidebarSelect() (tea.Model, tea.Cmd) {
	entry, ok := m.sidebar.SelectedEntry()
	if !ok {
		return m, nil
	}

	// Load into request pane.
	m.request.method = entry.Method
	m.request.SetValue(entry.URL)

	// Switch focus to request pane.
	m.activePane = paneRequest
	m.blurPane()
	cmd := m.focusPane()
	return m, cmd
}

// handleSidebarDelete deletes the selected history entry from the DB and
// refreshes the sidebar list.
func (m model) handleSidebarDelete() (tea.Model, tea.Cmd) {
	entry, ok := m.sidebar.SelectedEntry()
	if !ok || m.history == nil {
		return m, nil
	}

	return m, func() tea.Msg {
		err := m.history.Delete(entry.ID)
		return entryDeletedMsg{err: err}
	}
}

// ─── History Commands ────────────────────────────────────────────────────────

// loadHistoryCmd returns a Cmd that fetches history entries from the DB.
func loadHistoryCmd(s *history.Store) tea.Cmd {
	return func() tea.Msg {
		entries, err := s.List(100, 0)
		if err != nil {
			return historyLoadedMsg{err: err}
		}
		return historyLoadedMsg{entries: entries}
	}
}

// saveHistoryCmd returns a fire-and-forget Cmd that persists an entry.
func saveHistoryCmd(s *history.Store, entry history.Entry) tea.Cmd {
	return func() tea.Msg {
		_, err := s.Insert(entry)
		if err != nil {
			// Silently fail — history persistence should never interrupt
			// the user experience.
			return nil
		}
		return historySavedMsg{}
	}
}

// ─── Clipboard ───────────────────────────────────────────────────────────────

// handleCopyResponse copies the response body to the system clipboard and
// shows a toast confirming success or reporting an error.
func (m model) handleCopyResponse(msg copyResponseMsg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		if err := copyToClipboard(msg.body); err != nil {
			return toastMsg{text: fmt.Sprintf("✗ Clipboard error: %v", err)}
		}
		return toastMsg{text: "✓ Response body copied to clipboard"}
	}
}

// ─── Paste Command ───────────────────────────────────────────────────────────

// pasteURLFromClipboardCmd returns a Cmd that reads the system clipboard and
// sends a toastMsg with the result. The paste happens in the request pane's
// handlePasteClipboardMsg handler which is set up via the normal update flow.
//
// We use a two-step approach: the Cmd reads clipboard (async), then the
// toastMsg is handled by the root model which just displays the result.
// The actual paste into the URL input is done inside the Cmd since we have
// access to the text string synchronously after reading the clipboard.
func pasteURLFromClipboardCmd() tea.Cmd {
	return func() tea.Msg {
		text, err := pasteFromClipboard()
		if err != nil {
			return toastMsg{text: fmt.Sprintf("✗ Paste error: %v", err)}
		}
		if text == "" {
			return toastMsg{text: "Clipboard is empty"}
		}
		// Return the text so the root model can apply it.
		return pasteTextMsg{text: text}
	}
}

// pasteTextMsg carries clipboard text to be inserted into the URL input.
type pasteTextMsg struct {
	text string
}

// ─── History Message Handlers ────────────────────────────────────────────────

func (m model) handleHistoryLoaded(msg historyLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		// Non-fatal: history is unavailable; sidebar stays empty.
		return m, nil
	}
	m.sidebar.SetEntries(msg.entries)
	return m, nil
}

func (m model) handleEntryDeleted(msg entryDeletedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil || m.history == nil {
		return m, nil
	}
	// Refresh the sidebar list after deletion.
	return m, loadHistoryCmd(m.history)
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m model) View() tea.View {
	if !m.ready {
		return tea.NewView("ax — TUI API Client (loading...)\n")
	}

	// ── Compute layout dimensions ──────────────────────────────────────
	bodyHeight := m.height - 2
	if bodyHeight < 3 {
		bodyHeight = 3
	}

	const borderSize = 2

	var sidebarWidth, mainWidth int
	if m.sidebarHidden {
		sidebarWidth = 0
		mainWidth = m.width
	} else {
		sidebarWidth = int(float64(m.width) * 0.3)
		mainWidth = m.width - sidebarWidth
		if mainWidth < 20 {
			mainWidth = 20
		}
		if sidebarWidth < 10 {
			sidebarWidth = 10
		}
	}

	mainInnerW := mainWidth - borderSize

	reqInnerH := (bodyHeight-borderSize*2)*40/100 - borderSize
	respInnerH := (bodyHeight - borderSize*2) - reqInnerH - borderSize*2
	if reqInnerH < 3 {
		reqInnerH = 3
	}
	if respInnerH < 3 {
		respInnerH = 3
	}

	// ── Render each pane ───────────────────────────────────────────────
	reqContent := m.request.View(mainInnerW, reqInnerH)
	respContent := m.response.View(mainInnerW, respInnerH)
	requestView := PaneBorder(m.activePane == paneRequest, mainWidth).
		Render(reqContent)
	responseView := PaneBorder(m.activePane == paneResponse, mainWidth).
		Render(respContent)

	// ── Assemble layout ────────────────────────────────────────────────
	header := HeaderStyle.
		Width(m.width).
		Render(fmt.Sprintf(" ax — TUI API Client  v%s", m.Version))

	var body string
	if m.sidebarHidden {
		body = lipgloss.JoinVertical(lipgloss.Top, requestView, responseView)
	} else {
		sideInnerW := sidebarWidth - borderSize
		sideInnerH := bodyHeight - borderSize
		sideContent := m.sidebar.View(sideInnerW, sideInnerH)
		sidebarView := PaneBorder(m.activePane == paneSidebar, sidebarWidth).
			Render(sideContent)
		body = lipgloss.JoinHorizontal(
			lipgloss.Top,
			sidebarView,
			lipgloss.JoinVertical(lipgloss.Top, requestView, responseView),
		)
	}

	footer := FooterStyle.
		Width(m.width).
		Render(m.renderFooter())

	// ── Assemble main view ────────────────────────────────────────────
	mainView := lipgloss.JoinVertical(lipgloss.Top, header, body, footer)

	// ── Overlay help on top if visible ─────────────────────────────────
	if m.help.Visible() {
		helpView := m.help.View(m.width, m.height)
		if helpView != "" {
			mainView = helpView
		}
	}

	v := tea.NewView(mainView)
	v.AltScreen = true
	return v
}

// renderFooter builds the status bar / help text.
func (m model) renderFooter() string {
	var left, right string

	// ── Toast takes priority if active and not expired ─────────────────
	if m.toastText != "" && time.Now().Before(m.toastExpire) {
		left = fmt.Sprintf(" %s", m.toastText)
	} else if m.loading {
		left = fmt.Sprintf(" %s  Sending request...", m.spinner.View())
	} else if m.activePane == paneSidebar && m.sidebar.Len() > 0 {
		left = " [Enter] Load  [Ctrl+D] Delete  [Ctrl+H] Sidebar  [?] Help  [q] Quit"
	} else {
		left = " [Tab] Focus  [Ctrl+R] Send  [m] Method  [Ctrl+P] Paste  [Ctrl+Y] Copy  [Ctrl+H] Sidebar  [?] Help  [q] Quit"
	}

	methodInfo := fmt.Sprintf(" %s ", MethodColor(m.request.Method()).Render(m.request.Method()))
	right = methodInfo

	filler := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if filler < 1 {
		filler = 1
	}
	return left + stringsRepeat(" ", filler) + right
}

// ─── Config Path ─────────────────────────────────────────────────────────────

// historyDBPath returns the platform-specific path to the SQLite history file.
func historyDBPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}

	appDir := filepath.Join(configDir, "ax")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create config directory %s: %w", appDir, err)
	}

	return filepath.Join(appDir, "history.db"), nil
}

// ─── Import-level string helpers ─────────────────────────────────────────────

func stringsCount(s, substr string) int {
	n := 0
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			n++
			i += len(substr) - 1
		}
	}
	return n
}

func stringsRepeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}
