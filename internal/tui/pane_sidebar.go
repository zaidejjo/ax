package tui

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/zaidejjo/ax/internal/history"
)

// ─── Sidebar Item ────────────────────────────────────────────────────────────

// sidebarItem wraps a history.Entry so it can be used as a list.Item.
type sidebarItem struct {
	entry history.Entry
}

// FilterValue returns the searchable text for the list's built-in filter.
func (i sidebarItem) FilterValue() string {
	return fmt.Sprintf("%s %s", i.entry.Method, i.entry.URL)
}

// ─── Sidebar Delegate ────────────────────────────────────────────────────────

// sidebarDelegate renders each history entry as a single-line row with a
// colored method badge, truncated URL, and status code.
type sidebarDelegate struct {
	width int // set on each render to truncate URLs properly
}

func (d sidebarDelegate) Height() int                             { return 2 } // entry + divider
func (d sidebarDelegate) Spacing() int                            { return 0 }
func (d sidebarDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d sidebarDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry := item.(sidebarItem).entry

	// Colored method badge: " GET " / " POST " etc.
	methodBadge := MethodColor(entry.Method).Render(fmt.Sprintf(" %s ", entry.Method))

	// Truncated URL to fit sidebar width (account for badge, status, spacing).
	remaining := d.width - lipgloss.Width(methodBadge) - 6 // 6 = status + spacing
	url := entry.URL
	if remaining > 0 && len(url) > remaining {
		url = url[:remaining-3] + "..."
	} else if remaining <= 0 {
		url = ""
	}

	// Status code with color.
	statusStr := fmt.Sprintf("%d", entry.Status)
	styledStatus := StatusStyle(entry.Status).Render(statusStr)

	// Build the content line.
	var line string
	if index == m.Index() {
		// Selected item: show with a pointer and highlight.
		pointer := lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			Render("▸")
		line = fmt.Sprintf("%s %s %s %s", pointer, methodBadge, url, styledStatus)
	} else {
		line = fmt.Sprintf("  %s %s %s", methodBadge, url, styledStatus)
	}

	fmt.Fprint(w, line)

	// Thin separator line between entries (not after the last one).
	if index < len(m.Items())-1 {
		// Build a full-width divider line.
		fillWidth := d.width
		if fillWidth < 1 {
			fillWidth = 1
		}
		dividerLine := ""
		for i := 0; i < fillWidth; i++ {
			dividerLine += "─"
		}
		styledDivider := lipgloss.NewStyle().Foreground(borderInactive).Render(dividerLine)
		fmt.Fprint(w, "\n"+styledDivider)
	}
}

// ─── Sidebar Pane ────────────────────────────────────────────────────────────

// sidebarPane displays saved request history using bubbles/list with a custom
// delegate for colored method badges and status codes.
type sidebarPane struct {
	list     list.Model
	delegate sidebarDelegate
	items    []list.Item // cached items for rebuilding list when needed
}

// newSidebarPane creates the sidebar pane with an empty list.
func newSidebarPane() sidebarPane {
	d := sidebarDelegate{}
	emptyItems := []list.Item{}
	l := list.New(emptyItems, d, 0, 0)
	l.Title = "History"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.DisableQuitKeybindings()

	return sidebarPane{
		list:     l,
		delegate: d,
	}
}

func (p sidebarPane) Init() tea.Cmd {
	return nil
}

// Update delegates all messages to the embedded list component.
func (p sidebarPane) Update(msg tea.Msg) (sidebarPane, tea.Cmd) {
	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}

// View renders the sidebar within the given dimensions.
func (p sidebarPane) View(width, height int) string {
	p.list.SetWidth(width)
	p.list.SetHeight(height)
	p.delegate.width = width

	// If there are no items, show a placeholder instead of an empty list.
	if len(p.items) == 0 {
		return p.emptyView(width, height)
	}

	return p.list.View()
}

// emptyView renders a helpful message when the history is empty.
func (p sidebarPane) emptyView(width, height int) string {
	s := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				SidebarHelpStyle.Render("No history yet"),
				"",
				SidebarHelpStyle.Render("Press Ctrl+R to"),
				SidebarHelpStyle.Render("send your first request"),
			),
		)
	return s
}

// ─── Public API for Root Model ───────────────────────────────────────────────

// SetEntries replaces the sidebar items with the given history entries.
func (p *sidebarPane) SetEntries(entries []history.Entry) {
	items := make([]list.Item, len(entries))
	for i, e := range entries {
		items[i] = sidebarItem{entry: e}
	}
	p.items = items
	p.list.SetItems(items)
}

// SelectedEntry returns the currently selected history entry.
// The second return value is false if the list is empty.
func (p sidebarPane) SelectedEntry() (history.Entry, bool) {
	if len(p.items) == 0 {
		return history.Entry{}, false
	}
	idx := p.list.Index()
	if idx < 0 || idx >= len(p.items) {
		return history.Entry{}, false
	}
	item, ok := p.items[idx].(sidebarItem)
	if !ok {
		return history.Entry{}, false
	}
	return item.entry, true
}

// Len returns the number of items in the sidebar.
func (p sidebarPane) Len() int {
	return len(p.items)
}
