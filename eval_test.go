package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const evalStream = `{"type":"system","subtype":"init","cwd":"/tmp/x"}
{"type":"assistant","message":{"content":[{"type":"text","text":"looking"},{"type":"tool_use","name":"Bash","input":{"command":"ls ~/.claude"}}]}}
{"type":"assistant","parent_tool_use_id":"toolu_1","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/Users/x/.claude/knowledge/tooling/INDEX.md"}}]}}
{"type":"result","subtype":"success","result":"Push over SSH: git push git@github.com:o/r.git b","session_id":"abc","num_turns":3,"total_cost_usd":0.25,"usage":{"input_tokens":10,"output_tokens":20,"cache_creation_input_tokens":30,"cache_read_input_tokens":40}}
`

func TestScoreStreamReadAndApplied(t *testing.T) {
	cases, err := loadEvalCases(writeCases(t, `{"id":"a","file":"knowledge/tooling/INDEX.md","pattern":"git@github\\.com|\\bssh\\b","prompt":"p"}`))
	if err != nil {
		t.Fatal(err)
	}
	r, err := scoreStream(strings.NewReader(evalStream), cases[0])
	if err != nil {
		t.Fatal(err)
	}
	if !r.read || !r.applied || r.turns != 3 || r.tokens != 100 || r.cost != 0.25 || r.session != "abc" || r.status != "ok" {
		t.Fatalf("got %+v", r)
	}
}

func TestScoreStreamMissesOtherFileAndSubtype(t *testing.T) {
	cases, err := loadEvalCases(writeCases(t, `{"id":"b","file":"knowledge/frontend/INDEX.md","pattern":"referrer","prompt":"p"}`))
	if err != nil {
		t.Fatal(err)
	}
	stream := strings.Replace(evalStream, `"subtype":"success"`, `"subtype":"error_max_turns"`, 1)
	r, err := scoreStream(strings.NewReader(stream), cases[0])
	if err != nil {
		t.Fatal(err)
	}
	if r.read || r.applied || r.status != "max_turns" {
		t.Fatalf("got %+v", r)
	}
}

func TestLoadEvalCasesRejectsBadLines(t *testing.T) {
	for _, line := range []string{
		`{"id":"a","file":"f","pattern":"(","prompt":"p"}`,
		`{"id":"a","file":"f","prompt":"p"}`,
		`{"id":"a","file":"f","pattern":"x","prompt":"p"}` + "\n" + `{"id":"a","file":"g","pattern":"y","prompt":"q"}`,
	} {
		if _, err := loadEvalCases(writeCases(t, line)); err == nil {
			t.Fatalf("accepted %q", line)
		}
	}
}

func TestPreviousRunReadsNewestTable(t *testing.T) {
	dir := t.TempDir()
	old := "# old\n\n| case | read | applied |\n|---|---|---|\n| a | y | y |\n"
	newer := "# new\n\n| case | read | applied | turns |\n|---|---|---|---|\n| a | n | y | 2 |\n| b | y | n | 4 |\n"
	for name, body := range map[string]string{"2026-09-01-1000.md": old, "2026-09-02-0900.md": newer} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	name, scores := previousRun(dir)
	if name != "2026-09-02-0900" || scores["a"] != (prevScore{read: false, applied: true}) || scores["b"] != (prevScore{read: true, applied: false}) {
		t.Fatalf("got %q %+v", name, scores)
	}
	if name, scores := previousRun(filepath.Join(dir, "none")); name != "" || scores != nil {
		t.Fatalf("empty dir: got %q %+v", name, scores)
	}
}

func writeCases(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "retrieval.jsonl")
	if err := os.WriteFile(path, []byte(body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
