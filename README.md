# quno

Quick notes, v2: a dump with an agent attached.

- **Vault** `~/dev/docs` — `todo.md`, `quests/`, `investigations/`, `decisions/`, `knowledge/`. Schema: `meta/Conventions.md`.
- **Skills** — this repo is a Claude Code plugin: `/quno:quest`, `/quno:start`, `/quno:parse`, `/quno:investigation`, `/quno:adr`. `/learn` stays global.
- **CLI** `quno` — Go, stdlib only. `q`, `t`, `ls`, `start`, `resume`, `drop`, `rm`, `open`, `parse`, `project`, `path`. Quests carry `id:` (7 hex, git style) written at capture and backfilled on load; `<quest>` is an id or unique prefix (3+), a slug, or a title substring. `rm` asks `y/N` on a terminal, `-f` skips, piped stdin without `-f` refuses. `ls` is one line per quest, id first, content-sized columns, no slug, so it stays readable at any width; `ls -l` adds the slug on a dim second line. `drop` sets `status: dropped`, `rm` deletes the file. `open` fires `obsidian://open?path=` for the vault, `todo`, `home`, or a quest. Local commands are file I/O; `start`, `resume`, `parse` exec `claude` with `--add-dir <docs>`. Bare `quno` (or `quno ui`) opens the todo TUI: Bubble Tea, alt screen, mouse on. Click or `space` toggles, wheel or `j`/`k` moves, `n` adds, `d` deletes, `c` clears checked, `q` quits. Every action writes `todo.md`; Obsidian sees it live. In tmux you need `set -g mouse on` for clicks. Quests have no TUI yet; `quno ls --porcelain` is the fzf-ready shape. Color is ANSI, on for a terminal, off when piped or with `NO_COLOR`, forced with `CLICOLOR_FORCE=1`; `--porcelain` is never colored.

```
mise install                                  # go 1.23 from mise.toml
go build -o ~/.local/bin/quno .
quno help
```

Install the skills: `ln -s "$(pwd)" ~/.claude/skills/quno` — loads next session as `quno@skills-dir`.

The skills write to `~/dev/docs` from whatever cwd you are in. Without allow rules that prompts on every file (and is silently denied in `claude -p`). User-level `~/.claude/settings.json`:

```json
"permissions": {
  "allow": [
    "Read(~/dev/**)",
    "Edit(~/dev/docs/**)",
    "Bash(mv /Users/<you>/dev/docs/*)",
    "Bash(echo $CLAUDE_CODE_SESSION_ID)",
    "Bash(git rev-parse --show-toplevel)",
    "Bash(git branch --show-current)"
  ]
}
```

Optional, lets `/quno:parse` delete merged or demoted quests unattended: `"Bash(rm /Users/<you>/dev/docs/quests/*)"`. Claude cannot add `rm`/`mv` rules for you (auto-mode classifier blocks edits to permission rules); add them by hand. Without them parse writes content in place, marks merged sources `status: dropped`, and prints the `mv`/`rm` lines for you to run. The `:*` prefix form did not match a two-argument `mv` in testing; use the glob form shown.

Spec: `~/dev/docs/investigations/quno-capture-system.md`. Flows: `~/dev/docs/meta/Cheatsheet.md`.
