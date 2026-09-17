package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"filippo.io/age"
)

func newVaultTestConfig(t *testing.T) *config {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(root, "gitconfig"))
	t.Setenv("GIT_AUTHOR_NAME", "t")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@t")
	t.Setenv("GIT_COMMITTER_NAME", "t")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@t")
	docs := filepath.Join(root, "docs")
	if err := os.MkdirAll(filepath.Join(docs, "quests"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(docs, "quests", "alpha.md"), "one\n")
	return &config{
		docs:      docs,
		projects:  map[string]string{},
		history:   filepath.Join(root, "state", "history.git"),
		snapshots: filepath.Join(root, "state", "snapshots"),
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSyncCommitsOutsideTheVault(t *testing.T) {
	cfg := newVaultTestConfig(t)

	sha, changed, err := syncVault(cfg, "")
	if err != nil || sha == "" || changed != 1 {
		t.Fatalf("first sync: sha %q changed %d err %v", sha, changed, err)
	}
	if _, changed, err = syncVault(cfg, ""); err != nil || changed != 0 {
		t.Fatalf("second sync must be a no-op: changed %d err %v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(cfg.docs, ".git")); !os.IsNotExist(err) {
		t.Fatalf("vault must not hold a .git, stat err %v", err)
	}
	inside := exec.Command("git", "rev-parse", "--show-toplevel")
	inside.Dir = cfg.docs
	inside.Env = append(os.Environ(), "GIT_CEILING_DIRECTORIES="+filepath.Dir(cfg.docs))
	if err := inside.Run(); err == nil {
		t.Fatal("git must not see a repository from inside the vault")
	}

	writeTestFile(t, filepath.Join(cfg.docs, "quests", "alpha.md"), "two\n")
	writeTestFile(t, filepath.Join(cfg.docs, ".DS_Store"), "junk")
	if _, changed, err = syncVault(cfg, "chore: named"); err != nil || changed != 1 {
		t.Fatalf("edit plus an excluded file: changed %d err %v", changed, err)
	}
	log, err := vaultGit(cfg, "log", "--format=%s").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Fields(strings.ReplaceAll(string(log), "chore: ", "")); len(got) < 2 || got[0] != "named" {
		t.Fatalf("log: %q", log)
	}
}

func TestSnapshotRoundTripAndPrune(t *testing.T) {
	cfg := newVaultTestConfig(t)
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := snapshot(cfg, time.Now()); err == nil {
		t.Fatal("snapshot without a recipient must fail")
	}
	cfg.recipient = identity.Recipient().String()
	cfg.keep = 2

	base := time.Date(2026, 9, 17, 10, 0, 0, 0, time.UTC)
	var last string
	for i := 0; i < 3; i++ {
		writeTestFile(t, filepath.Join(cfg.docs, "quests", "alpha.md"), strings.Repeat("x", i+1)+"\n")
		if last, err = snapshot(cfg, base.Add(time.Duration(i)*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	names := snapshotFiles(cfg)
	if len(names) != 2 || filepath.Join(cfg.snapshots, names[1]) != last || !strings.Contains(names[0], "110000") {
		t.Fatalf("prune must keep the two newest: %v", names)
	}
	raw, err := os.ReadFile(last)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "xxx") || !strings.HasPrefix(string(raw), "age-encryption.org/v1") {
		t.Fatal("snapshot is not age ciphertext")
	}

	target := filepath.Join(t.TempDir(), "restored")
	other, _ := age.GenerateX25519Identity()
	if err := restoreSnapshot(last, target, other.String()); err == nil {
		t.Fatal("a different key must not decrypt")
	}
	if err := restoreSnapshot(last, target, identity.String()); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(target, "quests", "alpha.md"))
	if err != nil || string(got) != "xxx\n" {
		t.Fatalf("restored content %q, %v", got, err)
	}
	count, err := exec.Command("git", "-C", target, "rev-list", "--count", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(count)) != "3" {
		t.Fatalf("restored history: %q, %v", count, err)
	}
	if err := restoreSnapshot(last, target, identity.String()); err == nil {
		t.Fatal("restore into a non-empty dir must refuse")
	}
}

func TestWriteRecipientKeepsExistingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meta", "quno.toml")
	if err := writeRecipient(path, "age1new"); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); string(got) != "[vault]\nrecipient = \"age1new\"\n" {
		t.Fatalf("fresh file: %q", got)
	}

	writeTestFile(t, path, "docs = \"~/d\"\n\n[vault]\nkeep = 5\n\n[projects]\na = \"~/a\"\n")
	if err := writeRecipient(path, "age1new"); err != nil {
		t.Fatal(err)
	}
	want := "docs = \"~/d\"\n\n[vault]\nrecipient = \"age1new\"\nkeep = 5\n\n[projects]\na = \"~/a\"\n"
	if got, _ := os.ReadFile(path); string(got) != want {
		t.Fatalf("existing section:\n%s", got)
	}

	writeTestFile(t, path, "docs = \"~/d\"\n\n[projects]\na = \"~/a\"\n")
	if err := writeRecipient(path, "age1new"); err != nil {
		t.Fatal(err)
	}
	want = "docs = \"~/d\"\n\n[projects]\na = \"~/a\"\n\n[vault]\nrecipient = \"age1new\"\n"
	if got, _ := os.ReadFile(path); string(got) != want {
		t.Fatalf("appended section:\n%s", got)
	}
}
