package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// No --add-dir <docs>: the system prompt lists added directories, which would point the session at
// the vault. Read outside cwd prompts in -p and a prompt is a denial, so the bare Read rule is the access.
const (
	evalAllowedTools = "Read,Grep,Glob,Bash(cat:*),Bash(sed:*),Bash(grep:*),Bash(rg:*),Bash(head:*),Bash(tail:*),Bash(ls:*),Bash(find:*)"
	evalDeniedTools  = "Edit,Write,NotebookEdit"
	evalCaseTimeout  = 15 * time.Minute
)

type evalCase struct {
	ID      string `json:"id"`
	Prompt  string `json:"prompt"`
	File    string `json:"file"`
	Pattern string `json:"pattern"`
	re      *regexp.Regexp
}

type evalResult struct {
	id      string
	read    bool
	applied bool
	turns   int
	tokens  int
	cost    float64
	session string
	status  string
}

type evalOpts struct {
	model    string
	maxTurns int
	budget   float64
	save     bool
	only     []string
}

func (c *config) evalDir() string { return filepath.Join(c.docs, "meta", "eval") }

func cmdEval(cfg *config, args []string) error {
	opts := evalOpts{model: "opus", maxTurns: 12, budget: 1, save: true}
	list := false
	value := func(i *int, flag string) (string, error) {
		if *i+1 == len(args) {
			return "", fmt.Errorf("%s needs a value", flag)
		}
		*i++
		return args[*i], nil
	}
	for i := 0; i < len(args); i++ {
		var v string
		var err error
		switch args[i] {
		case "--list":
			list = true
		case "--no-save":
			opts.save = false
		case "--case", "-c":
			if v, err = value(&i, args[i]); err == nil {
				opts.only = append(opts.only, v)
			}
		case "--model", "-m":
			if v, err = value(&i, args[i]); err == nil {
				opts.model = v
			}
		case "--max-turns":
			if v, err = value(&i, args[i]); err == nil {
				opts.maxTurns, err = strconv.Atoi(v)
			}
		case "--budget":
			if v, err = value(&i, args[i]); err == nil {
				opts.budget, err = strconv.ParseFloat(v, 64)
			}
		default:
			return fmt.Errorf("unknown flag %q", args[i])
		}
		if err != nil {
			return err
		}
	}
	cases, err := loadEvalCases(filepath.Join(cfg.evalDir(), "retrieval.jsonl"))
	if err != nil {
		return err
	}
	if cases, err = selectEvalCases(cases, opts.only); err != nil {
		return err
	}
	idw, fw := 4, 0
	for _, c := range cases {
		idw, fw = max(idw, len(c.ID)), max(fw, len(c.File))
	}
	if list {
		for _, c := range cases {
			fmt.Println(out.yellow(pad(c.ID, idw)) + "  " + out.dim(pad(c.File, fw)) + "  " + firstLine(c.Prompt, 70))
		}
		return nil
	}
	scratch := filepath.Join(stateDir(), "eval")
	if err := os.MkdirAll(scratch, 0o755); err != nil {
		return err
	}
	runsDir := filepath.Join(cfg.evalDir(), "runs")
	prevName, prev := previousRun(runsDir)
	fmt.Println(out.dim(fmt.Sprintf("model %s · %d cases · max-turns %d · budget $%.2f · scratch %s", opts.model, len(cases), opts.maxTurns, opts.budget, contract(scratch))))
	fmt.Println(out.bold(pad("case", idw)+"  read  applied  turns  tokens  cost   status") + out.dim("     session   change"))
	var results []evalResult
	for _, c := range cases {
		r := runEvalCase(scratch, c, opts)
		results = append(results, r)
		fmt.Println(evalRow(r, idw, prev))
	}
	reads, applieds, cost := tally(results)
	summary := fmt.Sprintf("read %d/%d · applied %d/%d · $%.2f", reads, len(results), applieds, len(results), cost)
	if prevName != "" {
		pr, pa := 0, 0
		for _, r := range results {
			if p, ok := prev[r.id]; ok {
				if p.read {
					pr++
				}
				if p.applied {
					pa++
				}
			}
		}
		summary += fmt.Sprintf("   previous %s: read %d, applied %d", prevName, pr, pa)
	}
	fmt.Println(out.bold(summary))
	if !opts.save {
		return nil
	}
	path := filepath.Join(runsDir, time.Now().Format("2006-01-02-1504")+".md")
	if err := writeRun(path, opts, results, prevName); err != nil {
		return err
	}
	fmt.Println(out.dim("saved " + contract(path)))
	return nil
}

func loadEvalCases(path string) ([]evalCase, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	if lines == nil {
		return nil, fmt.Errorf("no cases at %s", contract(path))
	}
	var cases []evalCase
	seen := map[string]bool{}
	for n, line := range lines {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "//") {
			continue
		}
		var c evalCase
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", contract(path), n+1, err)
		}
		if c.ID == "" || c.Prompt == "" || c.File == "" || c.Pattern == "" {
			return nil, fmt.Errorf("%s:%d: id, prompt, file and pattern are all required", contract(path), n+1)
		}
		if seen[c.ID] {
			return nil, fmt.Errorf("%s:%d: duplicate id %q", contract(path), n+1, c.ID)
		}
		seen[c.ID] = true
		if c.re, err = regexp.Compile("(?is)" + c.Pattern); err != nil {
			return nil, fmt.Errorf("%s:%d: pattern: %w", contract(path), n+1, err)
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func selectEvalCases(cases []evalCase, only []string) ([]evalCase, error) {
	if len(only) == 0 {
		return cases, nil
	}
	byID := map[string]evalCase{}
	for _, c := range cases {
		byID[c.ID] = c
	}
	var picked []evalCase
	for _, id := range only {
		c, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("no case %q (quno eval --list)", id)
		}
		picked = append(picked, c)
	}
	return picked, nil
}

func evalArgv(c evalCase, opts evalOpts) []string {
	return []string{
		"-p", c.Prompt,
		"--model", opts.model,
		"--output-format", "stream-json", "--verbose",
		"--allowedTools", evalAllowedTools,
		"--disallowedTools", evalDeniedTools,
		"--max-turns", strconv.Itoa(opts.maxTurns),
		"--max-budget-usd", strconv.FormatFloat(opts.budget, 'f', -1, 64),
	}
}

func runEvalCase(scratch string, c evalCase, opts evalOpts) evalResult {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return evalResult{id: c.ID, status: err.Error()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), evalCaseTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, evalArgv(c, opts)...)
	cmd.Dir = scratch
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return evalResult{id: c.ID, status: err.Error()}
	}
	if err := cmd.Start(); err != nil {
		return evalResult{id: c.ID, status: err.Error()}
	}
	r, scoreErr := scoreStream(stdout, c)
	waitErr := cmd.Wait()
	switch {
	case ctx.Err() != nil:
		r.status = "timeout"
	case scoreErr != nil:
		r.status = scoreErr.Error()
	case r.status == "" && waitErr != nil:
		r.status = firstLine(strings.TrimSpace(stderr.String()), 60)
		if r.status == "" {
			r.status = waitErr.Error()
		}
	case r.status == "":
		r.status = "no result"
	}
	return r
}

type streamEvent struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Message struct {
		Content []struct {
			Type  string `json:"type"`
			Name  string `json:"name"`
			Input struct {
				FilePath string `json:"file_path"`
				Path     string `json:"path"`
				Command  string `json:"command"`
			} `json:"input"`
		} `json:"content"`
	} `json:"message"`
	Result    string  `json:"result"`
	SessionID string  `json:"session_id"`
	NumTurns  int     `json:"num_turns"`
	Cost      float64 `json:"total_cost_usd"`
	Usage     struct {
		Input         int `json:"input_tokens"`
		Output        int `json:"output_tokens"`
		CacheCreation int `json:"cache_creation_input_tokens"`
		CacheRead     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

// Subagent tool_uses arrive in the parent's stream, so a delegated read still counts.
func scoreStream(stream io.Reader, c evalCase) (evalResult, error) {
	r := evalResult{id: c.ID}
	sc := bufio.NewScanner(stream)
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var ev streamEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		switch ev.Type {
		case "assistant":
			for _, b := range ev.Message.Content {
				if b.Type != "tool_use" {
					continue
				}
				for _, s := range []string{b.Input.FilePath, b.Input.Path, b.Input.Command} {
					if s != "" && strings.Contains(s, c.File) {
						r.read = true
					}
				}
			}
		case "result":
			r.applied = c.re.MatchString(ev.Result)
			r.turns = ev.NumTurns
			r.tokens = ev.Usage.Input + ev.Usage.Output + ev.Usage.CacheCreation + ev.Usage.CacheRead
			r.cost = ev.Cost
			r.session = ev.SessionID
			r.status = strings.TrimPrefix(ev.Subtype, "error_")
			if r.status == "success" {
				r.status = "ok"
			}
		}
	}
	return r, sc.Err()
}

type prevScore struct{ read, applied bool }

// previousRun returns the newest run file's name and its per-case scores; names sort by time.
func previousRun(dir string) (string, map[string]prevScore) {
	names, _ := filepath.Glob(filepath.Join(dir, "*.md"))
	if len(names) == 0 {
		return "", nil
	}
	sort.Strings(names)
	path := names[len(names)-1]
	lines, err := readLines(path)
	if err != nil {
		return "", nil
	}
	scores := map[string]prevScore{}
	for _, line := range lines {
		cells := strings.Split(line, "|")
		if len(cells) < 4 || !strings.HasPrefix(line, "|") {
			continue
		}
		id := strings.TrimSpace(cells[1])
		if id == "" || id == "case" || strings.HasPrefix(id, "-") {
			continue
		}
		scores[id] = prevScore{read: strings.TrimSpace(cells[2]) == "y", applied: strings.TrimSpace(cells[3]) == "y"}
	}
	return strings.TrimSuffix(filepath.Base(path), ".md"), scores
}

func yn(b bool) string {
	if b {
		return "y"
	}
	return "n"
}

func evalRow(r evalResult, idw int, prev map[string]prevScore) string {
	mark := func(b bool) string {
		if b {
			return out.green("y")
		}
		return out.red("n")
	}
	change := ""
	if p, ok := prev[r.id]; ok {
		var parts []string
		if p.read != r.read {
			parts = append(parts, "read "+yn(p.read)+">"+yn(r.read))
		}
		if p.applied != r.applied {
			parts = append(parts, "applied "+yn(p.applied)+">"+yn(r.applied))
		}
		change = strings.Join(parts, ", ")
	}
	status := r.status
	if status != "ok" {
		status = out.yellow(status)
	}
	session := r.session
	if len(session) > 8 {
		session = session[:8]
	}
	return pad(r.id, idw) + "  " + mark(r.read) + "     " + mark(r.applied) + "        " + pad(strconv.Itoa(r.turns), 5) + "  " + pad(kilo(r.tokens), 6) + "  " + pad(fmt.Sprintf("%.2f", r.cost), 5) + "  " + pad(status, 9) + "  " + out.dim(pad(session, 8)) + "  " + out.dim(change)
}

func kilo(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}
	return fmt.Sprintf("%dk", (n+500)/1000)
}

func tally(results []evalResult) (reads, applieds int, cost float64) {
	for _, r := range results {
		if r.read {
			reads++
		}
		if r.applied {
			applieds++
		}
		cost += r.cost
	}
	return reads, applieds, cost
}

func writeRun(path string, opts evalOpts, results []evalResult, prevName string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	reads, applieds, cost := tally(results)
	var b strings.Builder
	fmt.Fprintf(&b, "# Retrieval eval %s\n\n", strings.TrimSuffix(filepath.Base(path), ".md"))
	fmt.Fprintf(&b, "model %s · %d cases · read %d/%d · applied %d/%d · $%.2f · max-turns %d · budget $%.2f", opts.model, len(results), reads, len(results), applieds, len(results), cost, opts.maxTurns, opts.budget)
	if prevName != "" {
		fmt.Fprintf(&b, " · previous [[%s]]", prevName)
	}
	b.WriteString("\n\n| case | read | applied | turns | tokens | cost | status | session |\n|---|---|---|---|---|---|---|---|\n")
	for _, r := range results {
		fmt.Fprintf(&b, "| %s | %s | %s | %d | %d | %.2f | %s | %s |\n", r.id, yn(r.read), yn(r.applied), r.turns, r.tokens, r.cost, r.status, r.session)
	}
	b.WriteString("\nInspect a case: `claude --resume <session>` from any cwd.\n")
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func firstLine(s string, n int) string {
	s, _, _ = strings.Cut(s, "\n")
	if len(s) > n {
		return s[:n-1] + "…"
	}
	return s
}
