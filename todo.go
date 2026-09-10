package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	unchecked = "- [ ] "
	checked   = "- [x] "
)

func readLines(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	text := strings.TrimRight(string(b), "\n")
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

func writeLines(path string, lines []string) error {
	out := strings.Join(lines, "\n")
	if out != "" {
		out += "\n"
	}
	return os.WriteFile(path, []byte(out), 0o644)
}

func isTodo(line string) bool {
	return strings.HasPrefix(line, unchecked) || strings.HasPrefix(line, checked)
}

func cmdTodo(cfg *config, args []string) error {
	path := filepath.Join(cfg.docs, "todo.md")
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		n := 0
		for _, l := range lines {
			if isTodo(l) {
				n++
				fmt.Printf("%s %s\n", out.dim(fmt.Sprintf("%2d", n)), todoLine(l))
			}
		}
		if n == 0 {
			fmt.Println(out.dim("no todos"))
		}
		return nil
	}
	switch args[0] {
	case "done":
		if len(args) < 2 {
			return fmt.Errorf("usage: quno t done <n>")
		}
		want, err := strconv.Atoi(args[1])
		if err != nil || want < 1 {
			return fmt.Errorf("bad index %q", args[1])
		}
		n := 0
		for i, l := range lines {
			if !isTodo(l) {
				continue
			}
			n++
			if n != want {
				continue
			}
			lines[i] = checked + strings.TrimPrefix(strings.TrimPrefix(l, unchecked), checked)
			if err := writeLines(path, lines); err != nil {
				return err
			}
			fmt.Println(todoLine(lines[i]))
			return nil
		}
		return fmt.Errorf("no todo %d", want)
	case "clear":
		all := len(args) > 1 && args[1] == "--all"
		var keep []string
		removed := 0
		for _, l := range lines {
			if strings.HasPrefix(l, checked) || (all && isTodo(l)) {
				removed++
				continue
			}
			keep = append(keep, l)
		}
		if err := writeLines(path, keep); err != nil {
			return err
		}
		fmt.Println(out.dim(fmt.Sprintf("cleared %d", removed)))
		return nil
	}
	text := strings.TrimSpace(strings.Join(args, " "))
	if project := cfg.projectFor(cwd()); project != "" && !strings.Contains(text, "#"+project) {
		text += " #" + project
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	lines = append(lines, unchecked+text)
	if err := writeLines(path, lines); err != nil {
		return err
	}
	fmt.Println(todoLine(unchecked + text))
	return nil
}

func todoLine(l string) string {
	if strings.HasPrefix(l, checked) {
		return out.dim(out.strike("[x] " + strings.TrimPrefix(l, checked)))
	}
	return out.green("[ ]") + " " + out.tags(strings.TrimPrefix(l, unchecked))
}
