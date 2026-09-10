package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func cmdOpen(cfg *config, args []string) error {
	target := cfg.docs
	if len(args) > 0 {
		switch args[0] {
		case "todo":
			target = filepath.Join(cfg.docs, "todo.md")
		case "home":
			target = filepath.Join(cfg.docs, "Home.md")
		default:
			q, err := findQuest(cfg, args[0])
			if err != nil {
				return err
			}
			target = q.path
		}
	}
	return openURL("obsidian://open?path=" + strings.ReplaceAll(url.QueryEscape(target), "+", "%20"))
}

func openURL(u string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", u)
	default:
		c = exec.Command("xdg-open", u)
	}
	c.Stderr = os.Stderr
	return c.Run()
}

func cmdRemove(cfg *config, args []string) error {
	force := false
	var rest []string
	for _, a := range args {
		if a == "-f" || a == "--force" {
			force = true
		} else {
			rest = append(rest, a)
		}
	}
	if len(rest) == 0 {
		return fmt.Errorf("usage: quno rm [-f] <quest>")
	}
	q, err := findQuest(cfg, rest[0])
	if err != nil {
		return err
	}
	if !force {
		if !isTerminal(os.Stdin) {
			return fmt.Errorf("refusing to delete %s without a terminal; pass -f", q.slug)
		}
		fmt.Printf("delete %s %s? [y/N] ", out.yellow(q.front["id"]), out.bold(q.title()))
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println(out.dim("kept"))
			return nil
		}
	}
	if err := os.Remove(q.path); err != nil {
		return err
	}
	if out.enabled {
		fmt.Println(out.red("✗"), out.dim("removed"), out.yellow(q.front["id"]), q.title())
	} else {
		fmt.Println(contract(q.path))
	}
	return nil
}

func cmdDrop(cfg *config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: quno drop <quest>")
	}
	q, err := findQuest(cfg, args[0])
	if err != nil {
		return err
	}
	if err := setFront(q.note, "status", "dropped"); err != nil {
		return err
	}
	if out.enabled {
		fmt.Println(out.status("dropped", "dropped"), out.yellow(q.front["id"]), q.title())
	} else {
		fmt.Println(contract(q.path))
	}
	return nil
}

// setFront replaces a frontmatter key in place or inserts it as the first key.
func setFront(n *note, key, value string) error {
	raw, err := os.ReadFile(n.path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return fmt.Errorf("%s has no frontmatter", contract(n.path))
	}
	entry := key + ": " + value
	for i := 1; i < len(lines) && lines[i] != "---"; i++ {
		if strings.HasPrefix(lines[i], key+":") {
			lines[i] = entry
			return os.WriteFile(n.path, []byte(strings.Join(lines, "\n")), 0o644)
		}
	}
	lines = append(lines[:1], append([]string{entry}, lines[1:]...)...)
	return os.WriteFile(n.path, []byte(strings.Join(lines, "\n")), 0o644)
}
