package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dtrugman/qory/lib/message"
	"github.com/dtrugman/qory/lib/session"
)

type historyProvider interface {
	HistoryAll(limit int) ([]session.SessionPreview, error)
	HistorySession(id string) (session.Session, error)
	HistoryDelete(id string) error
}

const (
	dateFormat = "Jan 02 2006 15:04"

	// historyLength is the maximum number of sessions loaded into the browser.
	historyLength = 10

	// Fixed lines of overhead: 1 header + up to 3 nav lines + 2 separators + 1 help line.
	viewOverheadLines = 7
)

var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	previewBodyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	moreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// roleStyle maps each message role to its header colour; unknown roles fall back to normalStyle.
	roleStyle = map[message.Role]lipgloss.Style{
		message.RoleUser:      lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true),
		message.RoleAssistant: lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true),
		message.RoleSystem:    lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Bold(true),
	}
)

type listState struct {
	cursor        int
	showPreview   bool
	previewOffset int
}

type sessionModel struct {
	provider historyProvider
	previews []session.SessionPreview

	selected  string
	quitting  bool
	statusMsg string

	width  int
	height int

	list listState
}

func newSessionModel(provider historyProvider) (sessionModel, error) {
	previews, err := provider.HistoryAll(historyLength)
	if err != nil {
		return sessionModel{}, err
	}
	return sessionModel{
		provider: provider,
		previews: previews,
		list:     listState{showPreview: true},
		width:    80,
		height:   24,
	}, nil
}

func (m sessionModel) Init() tea.Cmd {
	return nil
}

func (m sessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.list.cursor > 0 {
				m.list.cursor--
				m.list.previewOffset = 0
			}
		case "down", "j":
			if m.list.cursor < len(m.previews)-1 {
				m.list.cursor++
				m.list.previewOffset = 0
			}
		case "ctrl+j", "ctrl+down":
			m.list.previewOffset++
		case "ctrl+k", "ctrl+up":
			if m.list.previewOffset > 0 {
				m.list.previewOffset--
			}
		case "p":
			m.list.showPreview = !m.list.showPreview
			m.list.previewOffset = 0
		case "d":
			if len(m.previews) > 0 {
				id := m.previews[m.list.cursor].Name
				if err := m.provider.HistoryDelete(id); err != nil {
					m.statusMsg = fmt.Sprintf("Error: %v", err)
				} else {
					m.statusMsg = ""
					m.previews = append(m.previews[:m.list.cursor], m.previews[m.list.cursor+1:]...)
					if m.list.cursor >= len(m.previews) && m.list.cursor > 0 {
						m.list.cursor--
					}
					m.list.previewOffset = 0
				}
			}
		case "enter":
			if len(m.previews) > 0 {
				m.selected = m.previews[m.list.cursor].Name
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m sessionModel) View() string {
	if m.quitting {
		return ""
	}
	if len(m.previews) == 0 {
		return "No sessions found.\n"
	}

	var sb strings.Builder

	sb.WriteString(m.renderHeader())
	navCount := m.navCount()
	items, cursorInView := m.visibleItems(navCount)
	sb.WriteString(m.renderNavigator(items, cursorInView))

	if m.list.showPreview {
		previewHeight := max(1, m.height-navCount-viewOverheadLines)
		sb.WriteString(m.renderSeparator())
		sb.WriteString(m.renderPreview(previewHeight))
	}

	sb.WriteString(m.renderSeparator())
	if m.statusMsg != "" {
		sb.WriteString(helpStyle.Render(m.statusMsg))
		sb.WriteByte('\n')
	}
	sb.WriteString(helpStyle.Render(m.helpLine()))
	sb.WriteByte('\n')

	return sb.String()
}

// navCount returns how many sessions the navigator should show.
// When the preview is visible it stays at 3; when hidden it fills the
// available terminal height.
func (m sessionModel) navCount() int {
	if m.list.showPreview {
		return 3
	}
	// Available lines = height − 1 header − 1 separator − 1 help.
	available := max(3, m.height-3)
	return min(available, len(m.previews))
}

// visibleItems returns the slice of previews to display for the given window
// size and the index within that slice that corresponds to the current cursor.
// The cursor is kept as centred as possible, clamping at list boundaries.
func (m sessionModel) visibleItems(count int) ([]session.SessionPreview, int) {
	n := len(m.previews)
	if n == 0 {
		return nil, 0
	}

	// Ideal start keeps the cursor in the middle of the window.
	start := max(0, min(m.list.cursor-count/2, n-count))
	end := min(start+count, n)
	return m.previews[start:end], m.list.cursor - start
}

func (m sessionModel) renderHeader() string {
	n := len(m.previews)
	return normalStyle.Render(fmt.Sprintf("Session history (%d sessions)", n)) + "\n"
}

func (m sessionModel) renderNavigator(items []session.SessionPreview, cursorInView int) string {
	// Width of the widest index number, for consistent column alignment.
	numWidth := len(fmt.Sprintf("%d", len(m.previews)))

	var sb strings.Builder
	for i, p := range items {
		// Absolute 1-based index of this item in the full preview list.
		absNum := m.list.cursor - cursorInView + i + 1
		num := fmt.Sprintf("%*d", numWidth, absNum)
		line := fmt.Sprintf("%s. %s (%s)", num, p.Name, p.UpdatedAt.Format(dateFormat))
		if i == cursorInView {
			sb.WriteString(selectedStyle.Render("> " + line))
		} else {
			sb.WriteString(normalStyle.Render("  " + line))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func (m sessionModel) renderSeparator() string {
	width := m.width
	if width <= 0 {
		width = 40
	}
	return separatorStyle.Render(strings.Repeat("─", width)) + "\n"
}

func (m sessionModel) buildPreviewLines(messages []message.Message) []string {
	var lines []string
	for _, msg := range messages {
		style, ok := roleStyle[msg.Role]
		if !ok {
			style = normalStyle
		}
		role := strings.ToUpper(string(msg.Role))
		lines = append(lines, style.Render("--- "+role+" ---"))
		for _, l := range strings.Split(wordWrap(msg.Content, m.width), "\n") {
			lines = append(lines, previewBodyStyle.Render(l))
		}
		lines = append(lines, "")
	}
	return lines
}

func (m sessionModel) renderPreview(maxLines int) string {
	id := m.previews[m.list.cursor].Name
	sess, err := m.provider.HistorySession(id)
	if err != nil {
		return fmt.Sprintf("Error loading session: %v\n", err)
	}

	all := m.buildPreviewLines(sess.Messages)
	total := len(all)

	// Clamp the offset so we never scroll past the last screenful.
	offset := min(m.list.previewOffset, max(0, total-maxLines))

	// When there is more content than fits, reserve the last line for the indicator.
	contentLines := maxLines
	if total > maxLines {
		contentLines = maxLines - 1
	}

	end := min(offset+contentLines, total)

	var sb strings.Builder
	sb.WriteString(strings.Join(all[offset:end], "\n"))
	sb.WriteByte('\n')

	if remaining := total - end; remaining > 0 {
		sb.WriteString(moreStyle.Render(fmt.Sprintf("  ↓ %d more lines", remaining)))
		sb.WriteByte('\n')
	}

	return sb.String()
}

func (m sessionModel) helpLine() string {
	if m.list.showPreview {
		return "↑/k up  ↓/j down  ctrl+↑/↓ scroll preview  p hide preview  d delete  enter select  q quit"
	}
	return "↑/k up  ↓/j down  p show preview  d delete  enter select  q quit"
}

// wordWrap inserts newlines so no line exceeds maxWidth runes.
func wordWrap(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	var out strings.Builder
	for i, line := range strings.Split(text, "\n") {
		if i > 0 {
			out.WriteByte('\n')
		}
		for len(line) > maxWidth {
			out.WriteString(line[:maxWidth])
			out.WriteByte('\n')
			line = line[maxWidth:]
		}
		out.WriteString(line)
	}
	return out.String()
}

// ShowHistoryMenu presents an interactive session browser.
// Returns the selected session ID, or an empty string if the user quits without selecting.
func ShowHistoryMenu(provider historyProvider) (string, error) {
	m, err := newSessionModel(newCachingProvider(provider))
	if err != nil {
		return "", err
	}
	if len(m.previews) == 0 {
		return "", nil
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return "", err
	}

	return final.(sessionModel).selected, nil
}
