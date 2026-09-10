package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type quest struct {
	*note
	slug string
}

type gitInfo struct{ top, branch string }

func gitInfoFor(dir string) gitInfo {
	out := func(args ...string) string {
		c := exec.Command("git", args...)
		c.Dir = dir
		b, err := c.Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}
	return gitInfo{top: out("rev-parse", "--show-toplevel"), branch: out("branch", "--show-current")}
}

func cmdQuest(cfg *config, args []string) error {
	text := strings.TrimSpace(strings.Join(args, " "))
	if text == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		text = strings.TrimSpace(string(b))
	}
	if text == "" {
		return fmt.Errorf("nothing to capture")
	}
	now := time.Now()
	dir := cwd()
	g := gitInfoFor(dir)
	qdir := filepath.Join(cfg.docs, "quests")
	if err := os.MkdirAll(qdir, 0o755); err != nil {
		return err
	}
	path := uniquePath(qdir, now.Format("2006-01-02")+"-"+slugify(text, 5))
	capture := "_" + now.Format("2006-01-02 15:04")
	if g.top != "" {
		capture += " · " + filepath.Base(g.top)
	}
	if g.branch != "" {
		capture += " · " + g.branch
	}
	capture += "_"
	heading, _, _ := strings.Cut(text, "\n")
	existing, err := loadQuests(cfg)
	if err != nil {
		return err
	}
	id := newID(existing)
	content := fmt.Sprintf(`---
id: %s
project: %s
status: raw
created: %s
repo: %s
session:
related: []
---

# %s
%s

> Hub: [[Home]]

## Idea
%s
`, id, cfg.projectFor(dir), now.Format("2006-01-02"), contract(g.top), heading, capture, text)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	if out.enabled {
		fmt.Println(out.green("✓"), out.yellow(id), contract(path))
	} else {
		fmt.Println(contract(path))
	}
	return nil
}

func slugify(text string, maxWords int) string {
	var words []string
	for _, w := range strings.Fields(strings.ToLower(text)) {
		var b strings.Builder
		for _, r := range w {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				b.WriteRune(r)
			}
		}
		if b.Len() > 0 {
			words = append(words, b.String())
		}
		if len(words) == maxWords {
			break
		}
	}
	if len(words) == 0 {
		return "quest"
	}
	return strings.Join(words, "-")
}

func uniquePath(dir, name string) string {
	path := filepath.Join(dir, name+".md")
	for i := 2; ; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
		path = filepath.Join(dir, fmt.Sprintf("%s-%d.md", name, i))
	}
}

func loadQuests(cfg *config) ([]quest, error) {
	paths, err := filepath.Glob(filepath.Join(cfg.docs, "quests", "*.md"))
	if err != nil {
		return nil, err
	}
	var qs []quest
	for _, p := range paths {
		n, err := readNote(p)
		if err != nil {
			return nil, err
		}
		qs = append(qs, quest{note: n, slug: strings.TrimSuffix(filepath.Base(p), ".md")})
	}
	for _, q := range qs {
		if q.front["id"] != "" {
			continue
		}
		id := newID(qs)
		if err := setFront(q.note, "id", id); err != nil {
			return nil, err
		}
		q.front["id"] = id
	}
	return qs, nil
}

// newID is 7 hex chars like a git short hash; files written by skills get one on the next load.
func newID(taken []quest) string {
	for {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			panic(err)
		}
		id := hex.EncodeToString(b)[:7]
		clash := false
		for _, q := range taken {
			if q.front["id"] == id {
				clash = true
				break
			}
		}
		if !clash {
			return id
		}
	}
}

var statusRank = map[string]int{"in-progress": 0, "ready": 1, "raw": 2, "done": 3, "dropped": 4}

// sortedQuests is the one ordering every view and every numeric reference uses; done and dropped
// sort last so the numbers shown by a default ls are the same ones -a shows.
func sortedQuests(cfg *config) ([]quest, error) {
	qs, err := loadQuests(cfg)
	if err != nil {
		return nil, err
	}
	sort.Slice(qs, func(i, j int) bool {
		a, b := qs[i], qs[j]
		if ra, rb := statusRank[a.front["status"]], statusRank[b.front["status"]]; ra != rb {
			return ra < rb
		}
		if a.front["created"] != b.front["created"] {
			return a.front["created"] < b.front["created"]
		}
		return a.slug < b.slug
	})
	return qs, nil
}

func cmdList(cfg *config, args []string) error {
	all, long, porcelain, status := false, false, false, ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all", "-a":
			all = true
		case "--long", "-l":
			long = true
		case "--porcelain":
			porcelain = true
		case "--status", "-s":
			if i+1 == len(args) {
				return fmt.Errorf("--status needs a value")
			}
			i++
			status = args[i]
		default:
			return fmt.Errorf("unknown flag %q", args[i])
		}
	}
	qs, err := sortedQuests(cfg)
	if err != nil {
		return err
	}
	var rows []quest
	for _, q := range qs {
		st := q.front["status"]
		if status != "" && st != status {
			continue
		}
		if status == "" && !all && (st == "done" || st == "dropped") {
			continue
		}
		rows = append(rows, q)
	}
	if len(rows) == 0 {
		if !porcelain {
			fmt.Println(out.dim("no quests"))
		}
		return nil
	}
	if porcelain {
		for _, q := range rows {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", q.front["id"], q.slug, q.front["status"], q.front["project"], q.front["created"], q.title())
		}
		return nil
	}
	sw, pw := 0, 0
	for _, q := range rows {
		sw = max(sw, utf8.RuneCountInString(q.front["status"]))
		pw = max(pw, utf8.RuneCountInString(q.front["project"]))
	}
	indent := strings.Repeat(" ", sw+pw+25)
	for _, q := range rows {
		st := q.front["status"]
		line := out.yellow(pad(q.front["id"], 7)) + "  " + out.status(st, pad(st, sw)) + "  " + out.cyan(pad(q.front["project"], pw)) + "  " + out.dim(pad(q.front["created"], 10)) + "  " + q.title()
		fmt.Println(strings.TrimRight(line, " "))
		if long {
			fmt.Println(indent + out.dim(q.slug))
		}
	}
	return nil
}

// findQuest resolves an id or unique id prefix (3+ chars), an exact slug, or a unique slug / title substring.
func findQuest(cfg *config, needle string) (quest, error) {
	qs, err := sortedQuests(cfg)
	if err != nil {
		return quest{}, err
	}
	needle = strings.TrimSuffix(needle, ".md")
	lower := strings.ToLower(needle)
	var byPrefix []quest
	for _, q := range qs {
		if q.front["id"] == lower || q.slug == needle {
			return q, nil
		}
		if len(lower) >= 3 && strings.HasPrefix(q.front["id"], lower) {
			byPrefix = append(byPrefix, q)
		}
	}
	if len(byPrefix) == 1 {
		return byPrefix[0], nil
	}
	if len(byPrefix) > 1 {
		return quest{}, ambiguous(needle, byPrefix)
	}
	var matches []quest
	for _, q := range qs {
		if strings.Contains(strings.ToLower(q.slug), lower) || strings.Contains(strings.ToLower(q.title()), lower) {
			matches = append(matches, q)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return quest{}, fmt.Errorf("no quest matches %q", needle)
	}
	return quest{}, ambiguous(needle, matches)
}

func ambiguous(needle string, qs []quest) error {
	lines := make([]string, 0, len(qs))
	for _, q := range qs {
		lines = append(lines, q.front["id"]+"  "+q.title())
	}
	return fmt.Errorf("%q matches several quests:\n  %s", needle, strings.Join(lines, "\n  "))
}

func cmdStart(cfg *config, args []string) error {
	if len(args) == 0 {
		return execClaude(cfg, "/quno:start")
	}
	q, err := findQuest(cfg, args[0])
	if err != nil {
		return err
	}
	if err := enterRepo(q); err != nil {
		return err
	}
	return execClaude(cfg, "/quno:start "+q.slug)
}

func cmdResume(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: quno resume <slug>")
	}
	q, err := findQuest(cfg, args[0])
	if err != nil {
		return err
	}
	session := q.front["session"]
	if session == "" {
		return fmt.Errorf("%s has no session recorded; use quno start %s", q.slug, q.slug)
	}
	if err := enterRepo(q); err != nil {
		return err
	}
	return execClaude(cfg, "", "--resume", session)
}

func enterRepo(q quest) error {
	repo := expand(q.front["repo"])
	if repo == "" {
		return nil
	}
	if err := os.Chdir(repo); err != nil {
		return fmt.Errorf("repo %s: %w", repo, err)
	}
	return nil
}
