package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const questDetailRows = 5

var (
	styleID      = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleProject = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleSelect  = lipgloss.NewStyle().Bold(true)
	statusStyles = map[string]lipgloss.Style{
		"in-progress":   lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true),
		"adr":           lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		"proposed":      lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		"investigating": lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		"ready":         lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		"raw":           lipgloss.NewStyle().Faint(true),
		"done":          lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Faint(true),
		"dropped":       lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Faint(true),
	}
)

type questModel struct {
	cfg       *config
	all       []quest
	rows      []quest
	cursor    int
	top       int
	width     int
	height    int
	showAll   bool
	filtering bool
	filter    textinput.Model
	notice    string
}

type claudeDoneMsg struct{ err error }

func newQuestModel(cfg *config) (*questModel, error) {
	ti := textinput.New()
	ti.Prompt = "  / "
	ti.Placeholder = "filter by id, status, project, title; esc clears"
	ti.CharLimit = 80
	m := &questModel{cfg: cfg, filter: ti, width: 80, height: 24}
	return m, m.reload()
}

func (m *questModel) editing() bool { return m.filtering }

func (m *questModel) reload() error {
	qs, err := sortedQuests(m.cfg)
	if err != nil {
		return err
	}
	m.all = qs
	m.apply()
	return nil
}

// apply rebuilds the visible rows from showAll and the filter text, keeping the cursor on the same quest.
func (m *questModel) apply() {
	current := ""
	if m.cursor < len(m.rows) {
		current = m.rows[m.cursor].front["id"]
	}
	needle := strings.ToLower(strings.TrimSpace(m.filter.Value()))
	m.rows = m.rows[:0]
	for _, q := range m.all {
		st := q.front["status"]
		closed := st == "done" || st == "dropped"
		named := needle != "" && strings.HasPrefix(st, needle)
		if closed && !m.showAll && !named {
			continue
		}
		if needle != "" && !questMatches(q, needle) {
			continue
		}
		m.rows = append(m.rows, q)
	}
	m.cursor = 0
	for i, q := range m.rows {
		if q.front["id"] == current {
			m.cursor = i
		}
	}
	m.clamp()
}

func questMatches(q quest, needle string) bool {
	for _, s := range []string{q.front["id"], q.front["status"], q.front["project"], q.slug, q.title()} {
		if strings.Contains(strings.ToLower(s), needle) {
			return true
		}
	}
	return false
}

func (m *questModel) visibleRows() int {
	rows := m.height - tuiHeaderRows - tuiFooterRows - questDetailRows
	if m.filtering {
		rows--
	}
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m *questModel) clamp() {
	if m.cursor > len(m.rows)-1 {
		m.cursor = len(m.rows) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	rows := m.visibleRows()
	if m.cursor < m.top {
		m.top = m.cursor
	}
	if m.cursor >= m.top+rows {
		m.top = m.cursor - rows + 1
	}
	if m.top < 0 {
		m.top = 0
	}
}

func (m *questModel) move(delta int) {
	m.cursor += delta
	m.clamp()
}

func (m *questModel) selected() (quest, bool) {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return quest{}, false
	}
	return m.rows[m.cursor], true
}

func (m *questModel) fail(err error) {
	if err != nil {
		m.notice = styleError.Render(" " + err.Error())
	}
}

func (m *questModel) Init() tea.Cmd { return nil }

func (m *questModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.filter.Width = msg.Width - 6
		m.clamp()
	case reloadMsg:
		m.fail(m.reload())
	case claudeDoneMsg:
		m.fail(msg.err)
		if err := m.reload(); err != nil {
			m.fail(err)
		}
	case tea.MouseMsg:
		m.notice = ""
		return m.handleMouse(msg)
	case tea.KeyMsg:
		m.notice = ""
		if m.filtering {
			return m.handleFilterKey(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *questModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch {
	case msg.Button == tea.MouseButtonWheelUp:
		m.move(-1)
	case msg.Button == tea.MouseButtonWheelDown:
		m.move(1)
	case msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft:
		row := msg.Y - tuiHeaderRows
		if row < 0 || row >= m.visibleRows() {
			return m, nil
		}
		if i := m.top + row; i < len(m.rows) {
			m.cursor = i
			m.clamp()
		}
	}
	return m, nil
}

func (m *questModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		m.move(1)
	case "k", "up":
		m.move(-1)
	case "g", "home":
		m.cursor = 0
		m.clamp()
	case "G", "end":
		m.cursor = len(m.rows) - 1
		m.clamp()
	case "enter", "s":
		return m, m.start()
	case "r":
		return m, m.resume()
	case "o":
		if q, ok := m.selected(); ok {
			m.fail(openURL(obsidianURL(q.path)))
		}
	case "a":
		m.showAll = !m.showAll
		m.fail(m.reload())
	case "/":
		m.filtering = true
		m.clamp()
		return m, m.filter.Focus()
	}
	return m, nil
}

func (m *questModel) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.filter.SetValue("")
		m.filtering = false
		m.filter.Blur()
		m.apply()
		return m, nil
	case "enter":
		m.filtering = false
		m.filter.Blur()
		m.apply()
		return m, nil
	}
	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)
	m.apply()
	return m, cmd
}

// start runs claude in the quest's repo and comes back to the list when the session ends.
func (m *questModel) start() tea.Cmd {
	q, ok := m.selected()
	if !ok {
		return nil
	}
	return m.runClaude(q, "/quno:start "+q.slug)
}

func (m *questModel) resume() tea.Cmd {
	q, ok := m.selected()
	if !ok {
		return nil
	}
	session := q.front["session"]
	if session == "" {
		m.notice = styleError.Render(" no session recorded; enter starts it")
		return nil
	}
	return m.runClaude(q, "", "--resume", session)
}

func (m *questModel) runClaude(q quest, prompt string, extra ...string) tea.Cmd {
	dir, err := q.repoDir()
	if err != nil {
		m.fail(err)
		return nil
	}
	c, err := claudeCmd(m.cfg, dir, prompt, extra...)
	if err != nil {
		m.fail(err)
		return nil
	}
	return tea.ExecProcess(c, func(err error) tea.Msg { return claudeDoneMsg{err: err} })
}

func (m *questModel) View() string {
	right := styleDim.Render(fmt.Sprintf("%d quests ", len(m.rows)))
	if f := strings.TrimSpace(m.filter.Value()); f != "" && !m.filtering {
		right = styleDim.Render("/"+f+" · ") + right
	}
	if m.showAll {
		right = styleDim.Render("all · ") + right
	}
	var b strings.Builder
	b.WriteString(tabBar("quests", right, m.width) + "\n\n")
	sw, pw := 0, 0
	for _, q := range m.rows {
		sw = max(sw, utf8.RuneCountInString(q.front["status"]))
		pw = max(pw, utf8.RuneCountInString(q.front["project"]))
	}
	rows := m.visibleRows()
	for r := 0; r < rows; r++ {
		i := m.top + r
		switch {
		case i < len(m.rows):
			b.WriteString(m.renderRow(m.rows[i], sw, pw, i == m.cursor))
		case r == 0 && len(m.rows) == 0:
			b.WriteString(styleDim.Render("  no quests · quno q <text> captures one"))
		}
		b.WriteString("\n")
	}
	if m.filtering {
		b.WriteString(m.filter.View() + "\n")
	}
	b.WriteString(m.renderDetail())
	footer := m.notice
	if footer == "" {
		footer = styleDim.Render(" j/k move · enter start · r resume · o obsidian · / filter · a all · tab todo · q quit")
	}
	b.WriteString("\n" + footer)
	return b.String()
}

func (m *questModel) renderRow(q quest, sw, pw int, selected bool) string {
	cursor := "  "
	if selected {
		cursor = styleCursor.Render("› ")
	}
	st := q.front["status"]
	left := cursor + styleID.Render(q.front["id"]) + "  " + statusStyles[st].Render(pad(st, sw)) + "  "
	if pw > 0 {
		left += styleProject.Render(pad(q.front["project"], pw)) + "  "
	}
	left += styleDim.Render(q.front["created"]) + "  "
	title := truncate(q.title(), m.width-lipgloss.Width(left))
	if selected {
		title = styleSelect.Render(title)
	}
	return left + title
}

// renderDetail is a fixed block under the list: slug and repo, sections present, first lines of the Idea.
func (m *questModel) renderDetail() string {
	lines := make([]string, questDetailRows)
	if q, ok := m.selected(); ok {
		meta := q.slug
		if repo := q.front["repo"]; repo != "" {
			meta += " · " + repo
		}
		lines[0] = styleDim.Render(truncate(" "+meta, m.width))
		parts := q.sections()
		if q.front["session"] != "" {
			parts = append(parts, "session ✓")
		}
		lines[1] = styleDim.Render(truncate(" "+strings.Join(parts, " · "), m.width))
		for i, l := range ideaLines(q, questDetailRows-2) {
			lines[2+i] = truncate(" "+l, m.width)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func ideaLines(q quest, n int) []string {
	var out []string
	in := false
	for _, line := range strings.Split(q.body, "\n") {
		if strings.HasPrefix(line, "## ") {
			if in {
				break
			}
			in = strings.TrimSpace(line[3:]) == "Idea"
			continue
		}
		if line = strings.TrimSpace(line); in && line != "" {
			out = append(out, line)
			if len(out) == n {
				break
			}
		}
	}
	return out
}

func truncate(s string, width int) string {
	if width < 4 {
		width = 4
	}
	if utf8.RuneCountInString(s) <= width {
		return s
	}
	return string([]rune(s)[:width-1]) + "…"
}
