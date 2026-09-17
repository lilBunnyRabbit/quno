package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type config struct {
	docs      string
	projects  map[string]string
	history   string
	snapshots string
	recipient string
	keep      int
}

func (c *config) tomlPath() string {
	return filepath.Join(c.docs, "meta", "quno.toml")
}

func loadConfig() (*config, error) {
	env := os.Getenv("QUNO_DOCS")
	docs := env
	if docs == "" {
		docs = filepath.Join(home(), "dev", "docs")
	}
	cfg := &config{docs: expand(docs), projects: map[string]string{}}
	raw, err := os.ReadFile(cfg.tomlPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	section := ""
	for _, line := range strings.Split(string(raw), "\n") {
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.Trim(line, "[]")
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = expand(strings.Trim(strings.TrimSpace(value), `"`))
		switch section {
		case "":
			if key == "docs" && env == "" {
				cfg.docs = value
			}
		case "projects":
			cfg.projects[key] = value
		case "vault":
			switch key {
			case "history":
				cfg.history = value
			case "snapshots":
				cfg.snapshots = value
			case "recipient":
				cfg.recipient = value
			case "keep":
				cfg.keep, _ = strconv.Atoi(value)
			}
		}
	}
	return cfg, nil
}

func (c *config) projectFor(dir string) string {
	best, name := -1, ""
	for n, p := range c.projects {
		if (dir == p || strings.HasPrefix(dir, p+string(filepath.Separator))) && len(p) > best {
			best, name = len(p), n
		}
	}
	return name
}

func home() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

func expand(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		return filepath.Join(home(), p[1:])
	}
	return p
}

func contract(p string) string {
	h := home()
	if p == h {
		return "~"
	}
	if h != "" && strings.HasPrefix(p, h+string(filepath.Separator)) {
		return "~" + p[len(h):]
	}
	return p
}

func cwd() string {
	d, err := os.Getwd()
	if err != nil {
		return ""
	}
	return d
}
