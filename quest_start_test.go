package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestStartTargetCapturesSeveralWordsThatMatchNothing(t *testing.T) {
	dir := t.TempDir()
	qdir := filepath.Join(dir, "quests")
	if err := os.MkdirAll(qdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(qdir, "alpha.md"), []byte(questFile("aaaaaaa", "ready", "Alpha thing")), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := &config{docs: dir, projects: map[string]string{}}

	q, err := startTarget(cfg, []string{"alpha"})
	if err != nil || q.slug != "alpha" {
		t.Fatalf("existing slug: got %q, %v", q.slug, err)
	}
	q, err = startTarget(cfg, []string{"Alpha", "thing"})
	if err != nil || q.slug != "alpha" {
		t.Fatalf("title substring with a space: got %q, %v", q.slug, err)
	}

	var missing noMatch
	if _, err := startTarget(cfg, []string{"zzz"}); !errors.As(err, &missing) {
		t.Fatalf("one unknown word: want noMatch, got %v", err)
	}

	q, err = startTarget(cfg, []string{"new", "onboarding", "checklist"})
	if err != nil {
		t.Fatal(err)
	}
	if q.front["status"] != "raw" || q.title() != "new onboarding checklist" {
		t.Fatalf("captured quest: status %q title %q", q.front["status"], q.title())
	}
	if _, err := os.Stat(filepath.Join(qdir, q.slug+".md")); err != nil {
		t.Fatal(err)
	}
	if again, err := startTarget(cfg, []string{"new", "onboarding", "checklist"}); err != nil || again.front["id"] != q.front["id"] {
		t.Fatalf("second start must find the captured quest, got %q, %v", again.front["id"], err)
	}
}
