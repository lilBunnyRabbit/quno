package main

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func questFile(id, status, title string) string {
	return "---\nid: " + id + "\nproject: p\nstatus: " + status + "\ncreated: 2026-09-01\nrepo:\nsession:\nrelated: []\n---\n\n# " + title + "\n\n## Idea\nfirst line\nsecond line\n\n## Brief\nscope\n"
}

func newQuestTestModel(t *testing.T) *questModel {
	t.Helper()
	dir := t.TempDir()
	qdir := filepath.Join(dir, "quests")
	if err := os.MkdirAll(qdir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"alpha.md": questFile("aaaaaaa", "ready", "Alpha thing"),
		"beta.md":  questFile("bbbbbbb", "in-progress", "Beta thing"),
		"gamma.md": questFile("ccccccc", "done", "Gamma thing"),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(qdir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	m, err := newQuestModel(&config{docs: dir, projects: map[string]string{}})
	if err != nil {
		t.Fatal(err)
	}
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return m
}

func qkey(m *questModel, keys ...string) tea.Cmd {
	var cmd tea.Cmd
	for _, k := range keys {
		switch k {
		case "enter":
			_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		case "esc":
			_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
		default:
			_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
		}
	}
	return cmd
}

func ids(m *questModel) []string {
	var out []string
	for _, q := range m.rows {
		out = append(out, q.front["id"])
	}
	return out
}

func TestQuestRowsHideClosedAndSortByPhase(t *testing.T) {
	m := newQuestTestModel(t)
	if got := ids(m); len(got) != 2 || got[0] != "bbbbbbb" || got[1] != "aaaaaaa" {
		t.Fatalf("rows %v", got)
	}
	qkey(m, "a")
	if got := ids(m); len(got) != 3 || got[2] != "ccccccc" {
		t.Fatalf("all rows %v", got)
	}
}

func TestQuestFilterKeepsCursorAndRevealsNamedStatus(t *testing.T) {
	m := newQuestTestModel(t)
	qkey(m, "j")
	if m.rows[m.cursor].front["id"] != "aaaaaaa" {
		t.Fatalf("cursor on %s", m.rows[m.cursor].front["id"])
	}
	qkey(m, "/", "a", "l", "p", "enter")
	if got := ids(m); len(got) != 1 || got[0] != "aaaaaaa" || m.cursor != 0 || m.filtering {
		t.Fatalf("filtered %v cursor %d filtering %v", got, m.cursor, m.filtering)
	}
	qkey(m, "/", "esc")
	if got := ids(m); len(got) != 2 {
		t.Fatalf("esc did not clear filter: %v", got)
	}
	qkey(m, "/", "d", "o", "n", "e", "enter")
	if got := ids(m); len(got) != 1 || got[0] != "ccccccc" {
		t.Fatalf("status filter %v", got)
	}
}

func TestQuestEnterStartsClaudeAndResumeNeedsSession(t *testing.T) {
	bin := t.TempDir()
	stub := filepath.Join(bin, "claude")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	m := newQuestTestModel(t)
	if cmd := qkey(m, "enter"); cmd == nil {
		t.Fatalf("enter returned no exec cmd: %s", m.notice)
	}
	if cmd := qkey(m, "r"); cmd != nil || m.notice == "" {
		t.Fatalf("resume without session should only notice; cmd=%v notice=%q", cmd != nil, m.notice)
	}
}

func TestQuestDetailShowsIdeaAndSections(t *testing.T) {
	m := newQuestTestModel(t)
	view := m.View()
	for _, want := range []string{"beta", "Idea · Brief", "first line", "second line"} {
		if !contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestTabSwitchesAndSkipsWhileEditing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "quests"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := &config{docs: dir, projects: map[string]string{}}
	todo, err := newTodoModel(cfg)
	if err != nil {
		t.Fatal(err)
	}
	quests, err := newQuestModel(cfg)
	if err != nil {
		t.Fatal(err)
	}
	root := &rootModel{tabs: []tabModel{todo, quests}}
	tab := tea.KeyMsg{Type: tea.KeyTab}
	root.Update(tab)
	if root.active != 1 {
		t.Fatalf("active %d", root.active)
	}
	root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	root.Update(tab)
	if root.active != 1 || !quests.filtering {
		t.Fatalf("tab while filtering switched: active %d filtering %v", root.active, quests.filtering)
	}
	root.Update(tea.KeyMsg{Type: tea.KeyEsc})
	root.Update(tab)
	if root.active != 0 {
		t.Fatalf("active %d after esc+tab", root.active)
	}
}
