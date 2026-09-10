package main

import (
	"os"
	"path/filepath"
	"strings"
)

type note struct {
	path  string
	front map[string]string
	body  string
}

func readNote(path string) (*note, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	n := &note{path: path, front: map[string]string{}, body: string(raw)}
	if !strings.HasPrefix(n.body, "---\n") {
		return n, nil
	}
	rest := n.body[4:]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return n, nil
	}
	for _, line := range strings.Split(rest[:end], "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(line, " ") {
			continue
		}
		n.front[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	n.body = rest[end+5:]
	return n, nil
}

func (n *note) title() string {
	for _, line := range strings.Split(n.body, "\n") {
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(line[2:])
		}
	}
	return strings.TrimSuffix(filepath.Base(n.path), ".md")
}
