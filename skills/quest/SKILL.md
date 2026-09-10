---
name: quest
description: File the thing under discussion as a quest in the quno vault at whatever stage it is — a fresh idea gets a title and a Brief (status ready); an investigation or decision already worked out in this conversation gets Investigation / ADR sections (status investigating or adr). Dedupes against existing quests. Use when the user says /quno:quest, "quest this", "archive this analysis", "record this decision", or wants something saved for later.
argument-hint: <idea text, or empty to file what is under discussion>
---

# quno:quest

One quest file per thing. The user pays one line; you write the rest.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- `project`: the `[projects]` entry in `quno.toml` whose path (with `~` expanded) is the longest prefix of the cwd. No match → empty.
- `repo`: `git rev-parse --show-toplevel` in cwd, written `~`-relative; `branch`: `git branch --show-current`. Outside git → both empty.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Content.** `$ARGUMENTS` verbatim is the Idea. Empty → take what this conversation holds for one topic: an idea (quote it the way the user framed it), a finished investigation, a decision, or all three. Nothing concrete → say so, stop.
2. **Dedupe.** Grep `<docs>/quests/` for the key terms. Same topic exists → extend that file: `_<YYYY-MM-DD> addendum:_` under Idea, or add the missing Investigation / ADR section and move status forward; report the path, stop. Related but different → put it in `related`.
3. **Todo instead?** No brief needed, minutes of work, or not code → append `- [ ] <idea> #<project>` to `<docs>/todo.md`, say so, stop.
4. **Title, slug.** Title ≤8 words. Kebab slug. `<docs>/quests/<slug>.md`; collision → `-2`.
5. **Sections**, in Conventions order, only those with content:
   - Idea: verbatim.
   - Brief: against the repo, read-only: scope, entry points as `path:line` you opened, constraints, done-when. A real brief needs research → open with `Investigate first:` and the concrete questions.
   - Investigation: the question, evidence per claim (`path:line`, command, the override that beat a first guess), options, `### Bottom line`. Faithful to the conversation; tentative stays tentative.
   - ADR: context, drivers, at least two options with pros and cons, chosen and why the losers lost, consequences. Undecided → say what would settle it.
6. **Status.** ADR written → `adr`. Investigation without a decision → `investigating`. Brief only → `ready`.
7. **Write.** Seven keys per Conventions: `id` empty (the CLI assigns), `session` empty, `related` only verified targets. Body: `# <title>`, `_<YYYY-MM-DD HH:mm> · <repo basename> · <branch>_`, `> Hub: [[Home]]`, sections.
8. **Report** one line: `quest: <path> · <status> · <project>` plus any dedupe note.

## Rules

- Never rewrite Idea. Read the repo, never edit it here. Never invent a `path:line`.
- Wikilinks only to files verified in `<docs>`. Code, PRs, Slack stay plain text.
