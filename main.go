package main

import (
	"fmt"
	"os"
	"strings"
)

const usageText = `quno — quests (idea, investigation, ADR, implementation in one file), todo, and the agent that works them

  quno q <text>              new raw quest; project from cwd, repo and branch recorded
  quno q -                   same, text from stdin
  quno t <text>              add a todo line, #project from cwd
  quno t                     list todos, numbered
  quno t done <n>            check todo n
  quno t clear [--all]       delete checked todos; --all deletes every todo
  quno ls [-a | -s <status>] [-l] [--porcelain]
                             quests with ids; hides done and dropped, -l adds the slug line
                             status: raw ready investigating proposed adr in-progress done dropped
  quno cat <quest> [section]
                             print the file, or one ## section (quno cat <id> proposal | pbcopy)
  quno start [<quest>]       cd repo, claude "/quno:start <slug>"; no quest lists ready ones
  quno resume <quest>        cd repo, claude --resume <recorded session>
  quno drop <quest>          set status: dropped
  quno rm [-f] <quest>       delete the quest file; asks first unless -f
  quno open [<quest>|todo|home]
                             open the vault, or one note, in Obsidian
  quno parse [args]          claude "/quno:parse args"
  quno lint [args]           claude "/quno:lint args"
  quno project               project resolved from cwd
  quno path                  docs root
  quno [ui [todo|quests]]    TUI; tab switches todo and quests. quests: enter starts, r resumes,
                             o opens in Obsidian, / filters, a shows closed. No delete, no drop.

<quest> is an id or any unique prefix of it (3+ chars), a slug, or a title substring.
Docs root: $QUNO_DOCS, else ~/dev/docs. Routing: <docs>/meta/quno.toml.`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, errOut.red("quno:"), err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		if isTerminal(os.Stdout) {
			return runTUI(cfg, "")
		}
		printUsage()
		return nil
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "q", "quest":
		return cmdQuest(cfg, rest)
	case "t", "todo":
		return cmdTodo(cfg, rest)
	case "ls", "list":
		return cmdList(cfg, rest)
	case "start":
		return cmdStart(cfg, rest)
	case "resume":
		return cmdResume(cfg, rest)
	case "drop":
		return cmdDrop(cfg, rest)
	case "rm", "remove":
		return cmdRemove(cfg, rest)
	case "o", "open":
		return cmdOpen(cfg, rest)
	case "cat":
		return cmdCat(cfg, rest)
	case "parse":
		return execClaude(cfg, strings.TrimSpace("/quno:parse "+strings.Join(rest, " ")))
	case "lint":
		return execClaude(cfg, strings.TrimSpace("/quno:lint "+strings.Join(rest, " ")))
	case "project":
		fmt.Println(cfg.projectFor(cwd()))
		return nil
	case "path":
		fmt.Println(cfg.docs)
		return nil
	case "ui", "tui":
		return runTUI(cfg, strings.Join(rest, ""))
	case "-h", "--help":
		printUsage()
		return nil
	}
	return fmt.Errorf("unknown command %q (try quno -h)", cmd)
}

func printUsage() {
	for i, line := range strings.Split(usageText, "\n") {
		switch {
		case i == 0:
			fmt.Println(out.bold(line))
		case strings.HasPrefix(line, "  quno"):
			cmd, rest, _ := strings.Cut(strings.TrimPrefix(line, "  "), "  ")
			fmt.Println(strings.TrimRight("  "+out.cyan(pad(cmd, 25))+"  "+strings.TrimLeft(rest, " "), " "))
		case strings.HasPrefix(line, "Docs root"), strings.HasPrefix(line, "<quest>"):
			fmt.Println(out.dim(line))
		default:
			fmt.Println(line)
		}
	}
}
