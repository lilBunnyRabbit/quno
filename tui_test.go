package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(t *testing.T, content string) *todoModel {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "todo.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := newTodoModel(&config{docs: dir, projects: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

func fileOf(t *testing.T, m *todoModel) string {
	t.Helper()
	b, err := os.ReadFile(m.path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func key(m *todoModel, keys ...string) {
	for _, k := range keys {
		switch k {
		case "enter":
			m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		case "esc":
			m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		case " ":
			m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
		default:
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
		}
	}
}

func click(m *todoModel, y int) {
	m.Update(tea.MouseMsg{X: 4, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
}

func TestClickTogglesRow(t *testing.T) {
	m := newTestModel(t, "- [ ] a\n- [ ] b\n")
	click(m, tuiHeaderRows+1)
	if got, want := fileOf(t, m), "- [ ] a\n- [x] b\n"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if m.cursor != 1 {
		t.Fatalf("cursor %d", m.cursor)
	}
	click(m, tuiHeaderRows+1)
	if got := fileOf(t, m); got != "- [ ] a\n- [ ] b\n" {
		t.Fatalf("second click did not untoggle: %q", got)
	}
}

func TestClickOutsideListIsNoop(t *testing.T) {
	m := newTestModel(t, "- [ ] a\n")
	click(m, 0)
	click(m, tuiHeaderRows+5)
	if got := fileOf(t, m); got != "- [ ] a\n" {
		t.Fatalf("file changed: %q", got)
	}
}

func TestKeysMoveToggleDeleteClear(t *testing.T) {
	m := newTestModel(t, "# Todo\n\n- [ ] a\nnote line\n- [ ] b\n")
	key(m, "j", " ")
	if got := fileOf(t, m); got != "# Todo\n\n- [ ] a\nnote line\n- [x] b\n" {
		t.Fatalf("toggle b: %q", got)
	}
	key(m, "k", "d")
	if got := fileOf(t, m); got != "# Todo\n\nnote line\n- [x] b\n" {
		t.Fatalf("delete a: %q", got)
	}
	key(m, "c")
	if got := fileOf(t, m); got != "# Todo\n\nnote line\n" {
		t.Fatalf("clear done: %q", got)
	}
	if len(m.items) != 0 || m.cursor != 0 {
		t.Fatalf("items %d cursor %d", len(m.items), m.cursor)
	}
}

func TestAddAppendsAndSelects(t *testing.T) {
	m := newTestModel(t, "- [x] old\n")
	key(m, "n")
	if !m.adding {
		t.Fatal("not in adding mode")
	}
	key(m, "b", "u", "y", " ", "m", "i", "l", "k", "enter")
	if got := fileOf(t, m); got != "- [x] old\n- [ ] buy milk\n" {
		t.Fatalf("add: %q", got)
	}
	if m.adding || m.cursor != 1 {
		t.Fatalf("adding=%v cursor=%d", m.adding, m.cursor)
	}
	key(m, "n", "z", "esc")
	if got := fileOf(t, m); got != "- [x] old\n- [ ] buy milk\n" {
		t.Fatalf("esc should not add: %q", got)
	}
}

func TestScrollKeepsCursorVisible(t *testing.T) {
	var content string
	for i := 0; i < 50; i++ {
		content += "- [ ] item\n"
	}
	m := newTestModel(t, content)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	key(m, "G")
	rows := m.visibleRows()
	if m.cursor != 49 || m.top != 49-rows+1 {
		t.Fatalf("cursor %d top %d rows %d", m.cursor, m.top, rows)
	}
	click(m, tuiHeaderRows)
	if got := m.top; got != 49-rows+1 {
		t.Fatalf("click scrolled unexpectedly: top %d", got)
	}
	if !isCheckedLine(m.lines[m.items[m.top]]) {
		t.Fatal("click on first visible row did not toggle the item under it")
	}
}

func isCheckedLine(l string) bool { return len(l) >= len(checked) && l[:len(checked)] == checked }

func TestQuitKeys(t *testing.T) {
	m := newTestModel(t, "")
	for _, k := range []string{"q"} {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
		if cmd == nil {
			t.Fatalf("%s did not quit", k)
		}
	}
}
