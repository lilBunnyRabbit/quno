# quno

Quick notes, v2: a dump with an agent attached.

- **Vault** `~/dev/docs` — `todo.md`, `quests/`, `knowledge/`. One kind of note: a quest is idea, brief, investigation, ADR and findings in one file; status is the phase (`raw ready investigating proposed adr in-progress done dropped`); a Proposal section is the team-facing pitch, `proposed` means waiting on their answer. Schema: `meta/Conventions.md`. Flows: `meta/Cheatsheet.md`.
- **Skills** — this repo is a Claude Code plugin: `/quno:quest` (file what is under discussion at whatever stage it is), `/quno:parse` (triage raw), `/quno:start` (clarify, ask the route, then investigate / propose to the team / record ADR / implement inside the quest file), `/quno:lint` (re-read the whole vault: dead links, duplicates, contradictions, stale lines). `/learn` stays global.
- **CLI** `quno` — Go. `q`, `t`, `ls`, `cat`, `start`, `resume`, `drop`, `rm`, `open`, `parse`, `lint`, `sync`, `git`, `snapshot`, `project`, `path`. Quests carry `id:` (7 hex, git style) written at capture and backfilled on load; `<quest>` is an id or unique prefix (3+), a slug, or a title substring. `ls` is one line per quest, id first, content-sized columns; `ls -l` adds the slug, `-s <status>` filters, `-a` includes done and dropped. `drop` sets `status: dropped`, `rm` deletes after `y/N` (`-f` skips, piped stdin refuses). `open` fires `obsidian://open?path=` for the vault, `todo`, `home`, or a quest. `cat <quest> [section]` prints the file or one `## section` (`quno cat <id> proposal | pbcopy` is how a proposal leaves the vault). `start` with several words that match no quest captures them as a new quest first, so `quno start new onboarding checklist` goes from idea to session in one command; `/quno:start <idea>` does the same inside a session. `start`, `resume`, `parse` exec `claude` with the prompt first and `--add-dir <docs>` after (the flag is variadic and would swallow the prompt). Bare `quno` (or `quno ui [todo|quests]`) opens the TUI: Bubble Tea, alt screen, mouse on, `tab` switches the two tabs. Todo: click or `space` toggles, wheel or `j`/`k` moves, `n` adds, `d` deletes, `c` clears checked, `q` quits. Quests: same list as `ls` with a detail block (slug, repo, sections present, first Idea lines); `enter` runs claude in the quest's repo and returns to the list when the session ends, `r` resumes its session, `o` opens it in Obsidian, `/` filters on id, status, project, slug, title (typing a closed status reveals it), `a` shows done and dropped. No delete and no drop from the TUI by design. In tmux you need `set -g mouse on`. Color is ANSI, off when piped or with `NO_COLOR`, forced with `CLICOLOR_FORCE=1`; `--porcelain` is tab-separated and never colored.

```
mise install                                  # go 1.24 from mise.toml
go build -o ~/.local/bin/quno .
quno -h
```

**History and backup.** The vault is not a git repo and stays that way: `quno sync` commits it into a bare git dir kept outside it (`~/.local/share/quno/history/<vault>.git`, passed as `--git-dir` with the vault as `--work-tree`), so nothing inside the vault ever sees a repository. `quno git <args>` is git against that history: `quno git log --oneline`, `quno git diff HEAD~1`, `quno git checkout <sha> -- quests/x.md` to pull one file back. `.obsidian/workspace*.json`, `.DS_Store` and `.trash/` are excluded.

`quno snapshot` syncs, bundles the whole history (`git bundle --all`), encrypts the bundle with [age](https://age-encryption.org) to the public key under `[vault]` and keeps the newest 14. Only the public key is on disk, so it runs unattended and the output is safe to put on a synced or third-party drive; the provider sees ciphertext. `quno snapshot keygen` makes the pair once: public key into `quno.toml`, secret printed once for the password manager. `quno snapshot restore <file> <dir>` decrypts (secret from `-i <file>`, `$QUNO_AGE_KEY` or a prompt) and clones with full history; the file is standard age, so `age -d -i key.txt x.bundle.age > v.bundle && git clone v.bundle` works without quno.

```toml
[vault]
recipient = "age1..."                     # written by quno snapshot keygen
snapshots = "~/Library/Mobile Documents/com~apple~CloudDocs/quno"   # default ~/.local/share/quno/snapshots/<vault>
keep = 14
# history = "~/.local/share/quno/history/docs-xxxxxx.git"
```

Daily, `~/Library/LaunchAgents/dev.quno.snapshot.plist`, then `launchctl load` it:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>dev.quno.snapshot</string>
  <key>ProgramArguments</key><array><string>/Users/<you>/.local/bin/quno</string><string>snapshot</string></array>
  <key>StartCalendarInterval</key><dict><key>Hour</key><integer>13</integer><key>Minute</key><integer>0</integer></dict>
  <key>StandardErrorPath</key><string>/tmp/quno-snapshot.log</string>
</dict></plist>
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

Spec: `~/dev/docs/quests/quno-capture-system.md`.
