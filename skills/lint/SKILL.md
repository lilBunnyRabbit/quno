---
name: lint
description: Health pass over the quno vault — knowledge/ first (dead wikilinks, orphan notes, duplicate or contradicting lines, date-bound and upstream-bug lines to re-verify, overlong index lines that belong in notes, hot-facts budget, lines sitting at the wrong level), then quests (unresolved related targets, stale in-progress, raw backlog). Applies mechanical fixes, asks about judgment calls, writes a brief. Use when the user says /quno:lint, "lint the vault", "check the knowledge vault", or when spawned by `quno lint`.
argument-hint: [knowledge | quests | --report]
---

# quno:lint

`/learn` compiles one lesson at a time against whatever index is in context. Nothing re-reads the whole vault. This does, rarely: monthly, or when hot facts pass ~10 lines.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- Read `<docs>/meta/Conventions.md` and `~/.claude/skills/learn/SKILL.md` once. Together they are the schema; this file does not restate it.
- `$ARGUMENTS`: `knowledge` or `quests` limits scope; empty = both. `--report` = flag everything, change nothing.

## Steps

1. **Read all of it.** Every `.md` under `<docs>/knowledge/` in one batch, plus `ls -l <docs>/knowledge/snippets/bin`; for quests only frontmatter plus the first line under each heading.
2. **Knowledge checks**, in this order:
   1. **Links.** Every `[[target]]` resolves to a file in `<docs>`; a bare `[[INDEX]]` inside a stack file means the sibling INDEX. Every root domain pointer path exists. Every stack file is named in its domain INDEX and in the root pointer. Every note file has at least one inbound link. Every script in `snippets/bin/` is executable, passes `bash -n`, and is named in `snippets/INDEX.md`; every `bin/<name>` that index mentions exists.
   2. **Duplicates.** Same lesson in two places (hot fact restating a stack-file line, INDEX line restating a note) → the more specific place keeps the text, the other keeps one clause plus `[[link]]`.
   3. **Contradictions.** Two lines, same trigger, opposite conclusion → question, both quoted verbatim, never resolved by guessing.
   4. **Stale.** Lines carrying a date, a version pin, an "upstream bug", a `workaround`, or a "snapshot of" note → list with the claim to re-verify. Verify only when one command or one changelog read settles it; otherwise leave the line, append ` — unverified since <YYYY-MM>` to it, and list it. A later pass that settles the claim removes the marker; one that cannot leaves it, never moves the date. The marker is the one lint edit that touches a lesson: it tells the next session the line is a hypothesis, not a fact, and it survives the brief.
   5. **Symptom first.** An index line or a root domain pointer has to be findable before the cause is known: it leads with what was observed (error text, the thing that broke, the tool involved), the conclusion follows. A line that opens with the cause (`Narrow sideEffects glob …`) or a pointer that names a category (`tool gotchas`) gets reordered or given the symptoms it covers; substance untouched.
   6. **Shape.** Index line past ~60 words or holding repro steps → note file plus a one-line pointer. Hot facts past ~10 → name the coldest to demote. Stack-neutral line in a stack file → domain INDEX; API-bound line in an INDEX → stack file. `snippets/INDEX.md` is exempt from the word cap: a line there is trigger + invocation + gotcha, and code lives in `snippets/` by design.
3. **Quest checks.** `related:` targets resolve. `status: in-progress` untouched (mtime) for 14+ days → list. `status: raw` present → say how many, point at `/quno:parse`. Every quest carries `> Hub: [[Home]]`.
4. **Apply** mechanical fixes unless `--report`: dead link repaired or removed, missing pointer or header line added, duplicate collapsed to a link, overlong line moved to a note, cause-first line reordered to symptom first, `unverified since` marker added or removed. Everything under Contradictions stays a question; under Stale only the marker is applied, the claim itself stays a question.
5. **Brief**, one screen: changed · flagged (questions) · re-verify list · demote candidates. Table for flagged lines: file · line · check · action asked.

## Rules

- Never rewrite a lesson's substance without evidence from a command, a changelog, or the user.
- Never delete a line to resolve a contradiction; ask.
- Do not open repos to re-verify unless one file read settles it.
- Reading `<docs>/knowledge/astra/` counts; it is project reference and still gets link and stale checks.
