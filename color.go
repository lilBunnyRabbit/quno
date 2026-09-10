package main

import (
	"os"
	"strings"
	"unicode/utf8"
)

type palette struct{ enabled bool }

var (
	out    = newPalette(os.Stdout)
	errOut = newPalette(os.Stderr)
)

func newPalette(f *os.File) palette {
	if v := os.Getenv("CLICOLOR_FORCE"); v != "" && v != "0" {
		return palette{enabled: true}
	}
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return palette{}
	}
	return palette{enabled: isTerminal(f)}
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func (p palette) wrap(code, s string) string {
	if !p.enabled || s == "" {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (p palette) bold(s string) string    { return p.wrap("1", s) }
func (p palette) dim(s string) string     { return p.wrap("2", s) }
func (p palette) strike(s string) string  { return p.wrap("9", s) }
func (p palette) red(s string) string     { return p.wrap("31", s) }
func (p palette) green(s string) string   { return p.wrap("32", s) }
func (p palette) yellow(s string) string  { return p.wrap("33", s) }
func (p palette) cyan(s string) string    { return p.wrap("36", s) }
func (p palette) magenta(s string) string { return p.wrap("35", s) }

// status colors display by the meaning of value; display is usually value already padded.
func (p palette) status(value, display string) string {
	switch value {
	case "in-progress":
		return p.yellow(p.bold(display))
	case "ready":
		return p.green(display)
	case "investigating":
		return p.yellow(display)
	case "adr":
		return p.magenta(display)
	case "raw":
		return p.dim(display)
	case "done":
		return p.dim(p.green(display))
	case "dropped":
		return p.dim(p.red(display))
	}
	return display
}

// tags colors every #word in a line; runs before any other wrapping so widths stay plain text.
func (p palette) tags(s string) string {
	if !p.enabled {
		return s
	}
	words := strings.Split(s, " ")
	for i, w := range words {
		if strings.HasPrefix(w, "#") && len(w) > 1 {
			words[i] = p.cyan(w)
		}
	}
	return strings.Join(words, " ")
}

// pad right-pads to n runes before color is applied, since escape codes break %-Ns widths.
func pad(s string, n int) string {
	if w := utf8.RuneCountInString(s); w < n {
		return s + strings.Repeat(" ", n-w)
	}
	return s
}
