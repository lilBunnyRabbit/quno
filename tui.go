package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	tuiHeaderRows = 2
	tuiFooterRows = 2
)

var (
	styleTitle  = lipgloss.NewStyle().Bold(true)
	styleDim    = lipgloss.NewStyle().Faint(true)
	styleCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleOpen   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleDone   = lipgloss.NewStyle().Faint(true).Strikethrough(true)
	styleTag    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleError  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

type todoModel struct {
	cfg    *config
	path   string
	lines  []string
	items  []int
	cursor int
	top    int
	width  int
	height int
	adding bool
	input  textinput.Model
	notice string
}

type tabModel interface {
	tea.Model
	editing() bool
}

// rootModel holds both tabs and only routes: tab switches, everything else goes to the active one.
type rootModel struct {
	tabs   []tabModel
	active int
}

func runTUI(cfg *config, tab string) error {
	todo, err := newTodoModel(cfg)
	if err != nil {
		return err
	}
	quests, err := newQuestModel(cfg)
	if err != nil {
		return err
	}
	m := &rootModel{tabs: []tabModel{todo, quests}}
	switch tab {
	case "", "todo":
	case "quests", "quest", "q":
		m.active = 1
	default:
		return fmt.Errorf("unknown tab %q (todo or quests)", tab)
	}
	_, err = tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	return err
}

func (m *rootModel) Init() tea.Cmd { return nil }

func (m *rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for _, t := range m.tabs {
			t.Update(msg)
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "tab" && !m.tabs[m.active].editing() {
			m.active = (m.active + 1) % len(m.tabs)
			_, cmd := m.tabs[m.active].Update(reloadMsg{})
			return m, cmd
		}
	}
	_, cmd := m.tabs[m.active].Update(msg)
	return m, cmd
}

func (m *rootModel) View() string { return m.tabs[m.active].View() }

type reloadMsg struct{}

var tabNames = []string{"todo", "quests"}

// tabBar is the first header row of every tab: names left, a right-aligned status string.
func tabBar(active, right string, width int) string {
	var names []string
	for _, n := range tabNames {
		if n == active {
			names = append(names, styleTitle.Render(n))
		} else {
			names = append(names, styleDim.Render(n))
		}
	}
	left := " " + strings.Join(names, styleDim.Render("  ·  "))
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func newTodoModel(cfg *config) (*todoModel, error) {
	ti := textinput.New()
	ti.Prompt = "  + "
	ti.Placeholder = "new todo, enter to save, esc to cancel"
	ti.CharLimit = 200
	m := &todoModel{cfg: cfg, path: filepath.Join(cfg.docs, "todo.md"), input: ti, width: 80, height: 24}
	return m, m.reload()
}

func (m *todoModel) reload() error {
	lines, err := readLines(m.path)
	if err != nil {
		return err
	}
	m.lines = lines
	m.items = m.items[:0]
	for i, l := range lines {
		if isTodo(l) {
			m.items = append(m.items, i)
		}
	}
	m.clamp()
	return nil
}

func (m *todoModel) visibleRows() int {
	rows := m.height - tuiHeaderRows - tuiFooterRows
	if m.adding {
		rows--
	}
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m *todoModel) clamp() {
	if m.cursor > len(m.items)-1 {
		m.cursor = len(m.items) - 1
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

func (m *todoModel) move(delta int) {
	m.cursor += delta
	m.clamp()
}

func (m *todoModel) save() error {
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}
	if err := writeLines(m.path, m.lines); err != nil {
		return err
	}
	return m.reload()
}

func (m *todoModel) toggle(i int) error {
	if i < 0 || i >= len(m.items) {
		return nil
	}
	li := m.items[i]
	if l := m.lines[li]; strings.HasPrefix(l, checked) {
		m.lines[li] = unchecked + strings.TrimPrefix(l, checked)
	} else {
		m.lines[li] = checked + strings.TrimPrefix(l, unchecked)
	}
	return m.save()
}

func (m *todoModel) remove(i int) error {
	if i < 0 || i >= len(m.items) {
		return nil
	}
	li := m.items[i]
	m.lines = append(m.lines[:li:li], m.lines[li+1:]...)
	return m.save()
}

func (m *todoModel) clearDone() (int, error) {
	keep := make([]string, 0, len(m.lines))
	removed := 0
	for _, l := range m.lines {
		if strings.HasPrefix(l, checked) {
			removed++
			continue
		}
		keep = append(keep, l)
	}
	m.lines = keep
	return removed, m.save()
}

func (m *todoModel) add(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if p := m.cfg.projectFor(cwd()); p != "" && !strings.Contains(text, "#"+p) {
		text += " #" + p
	}
	m.lines = append(m.lines, unchecked+text)
	if err := m.save(); err != nil {
		return err
	}
	m.cursor = len(m.items) - 1
	m.clamp()
	return nil
}

func (m *todoModel) fail(err error) {
	if err != nil {
		m.notice = styleError.Render(" " + err.Error())
	}
}

func (m *todoModel) Init() tea.Cmd { return nil }

func (m *todoModel) editing() bool { return m.adding }

func (m *todoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.input.Width = msg.Width - 6
		m.clamp()
	case reloadMsg:
		m.fail(m.reload())
	case tea.MouseMsg:
		m.notice = ""
		return m.handleMouse(msg)
	case tea.KeyMsg:
		m.notice = ""
		if m.adding {
			return m.handleAddKey(msg)
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *todoModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
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
		i := m.top + row
		if i >= len(m.items) {
			return m, nil
		}
		m.cursor = i
		m.fail(m.toggle(i))
	}
	return m, nil
}

func (m *todoModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.cursor = len(m.items) - 1
		m.clamp()
	case " ", "enter", "x":
		m.fail(m.toggle(m.cursor))
	case "d", "delete", "backspace":
		m.fail(m.remove(m.cursor))
	case "c":
		n, err := m.clearDone()
		m.fail(err)
		if err == nil {
			m.notice = styleDim.Render(fmt.Sprintf(" cleared %d", n))
		}
	case "n", "a", "+":
		m.adding = true
		m.input.SetValue("")
		m.clamp()
		return m, m.input.Focus()
	case "r":
		m.fail(m.reload())
	}
	return m, nil
}

func (m *todoModel) handleAddKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.adding = false
		m.input.Blur()
		m.clamp()
		return m, nil
	case "enter":
		text := m.input.Value()
		m.adding = false
		m.input.Blur()
		m.fail(m.add(text))
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *todoModel) View() string {
	open, done := 0, 0
	for _, li := range m.items {
		if strings.HasPrefix(m.lines[li], checked) {
			done++
		} else {
			open++
		}
	}
	var b strings.Builder
	b.WriteString(tabBar("todo", styleDim.Render(fmt.Sprintf("%d open · %d done ", open, done)), m.width) + "\n\n")
	rows := m.visibleRows()
	for r := 0; r < rows; r++ {
		i := m.top + r
		switch {
		case i < len(m.items):
			b.WriteString(m.renderItem(i, i == m.cursor))
		case r == 0 && len(m.items) == 0 && !m.adding:
			b.WriteString(styleDim.Render("  no todos · n to add one"))
		}
		b.WriteString("\n")
	}
	if m.adding {
		b.WriteString(m.input.View() + "\n")
	}
	footer := m.notice
	if footer == "" {
		footer = styleDim.Render(" j/k move · space toggle · n new · d delete · c clear done · tab quests · q quit")
	}
	b.WriteString("\n" + footer)
	return b.String()
}

func (m *todoModel) renderItem(i int, selected bool) string {
	l := m.lines[m.items[i]]
	done := strings.HasPrefix(l, checked)
	text := strings.TrimPrefix(strings.TrimPrefix(l, checked), unchecked)
	avail := m.width - 8
	if avail < 10 {
		avail = 10
	}
	if utf8.RuneCountInString(text) > avail {
		text = string([]rune(text)[:avail-1]) + "…"
	}
	cursor := "  "
	if selected {
		cursor = styleCursor.Render("› ")
	}
	if done {
		return cursor + styleDone.Render("[x] "+text)
	}
	return cursor + styleOpen.Render("[ ]") + " " + colorTags(text)
}

func colorTags(s string) string {
	words := strings.Split(s, " ")
	for i, w := range words {
		if strings.HasPrefix(w, "#") && len(w) > 1 {
			words[i] = styleTag.Render(w)
		}
	}
	return strings.Join(words, " ")
}
