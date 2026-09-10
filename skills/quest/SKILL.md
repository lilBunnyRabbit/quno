---
name: quest
description: Capture an idea as a quest in the quno vault from inside a session — distill a title, resolve project and repo from cwd, dedupe against existing quests and investigations, write a brief against the real repo, file it as status ready. Use when the user says /quno:quest, "quest this", "make this a quest", or wants an idea saved for later work. With no arguments, capture the idea currently under discussion.
argument-hint: <idea text, or empty to capture the idea under discussion>
---

# quno:quest

Capture one idea as a quest file. The user pays one line; you write the title and the Brief.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- `project`: the `[projects]` entry in `quno.toml` whose path (with `~` expanded) is the longest prefix of the cwd. No match → empty.
- `repo`: `git rev-parse --show-toplevel` in cwd, written `~`-relative; `branch`: `git branch --show-current`. Outside git → both empty.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Idea.** `$ARGUMENTS` verbatim. Empty → distill the idea under discussion and quote it the way the user framed it. Nothing concrete under discussion → say so, stop.
2. **Dedupe.** Grep `<docs>/quests/` and `<docs>/investigations/` for the idea's key terms. Same idea exists → append `_<YYYY-MM-DD> addendum:_ <text>` under its `## Idea`, report that path, stop. Related but different → continue and put it in `related`.
3. **Title, slug.** Title ≤8 words. Slug kebab-case. File `<docs>/quests/<slug>.md`; collision → `-2`.
4. **Brief.** Against the repo at `repo`, read-only, absolute paths. Scope (what changes, what explicitly does not), entry points as `path:line` you have actually opened, constraints, done-when. A real brief needs research first → open the Brief with `Investigate first:` and the concrete questions. Never invent a `path:line`.
5. **Todo instead?** No brief needed, minutes of work, or not code → append `- [ ] <idea> #<project>` to `<docs>/todo.md`, say so, stop.
6. **Write.** Keys per Conventions: `id` (leave empty; the CLI assigns a 7-hex id on its next run), `project`, `status: ready`, `created` (today), `repo`, `session` (empty), `related` (only wikilinks whose target file exists in `<docs>`). Body:

   ```markdown
   # <title>
   _<YYYY-MM-DD HH:mm> · <repo basename> · <branch>_

   > Hub: [[Home]]

   ## Idea
   <verbatim>

   ## Brief
   <scope · entry points · constraints · done-when>
   ```

7. **Report** one line: `quest: <path> · ready · <project>` plus any dedupe note.

## Rules

- Never rewrite `## Idea`.
- Read the repo, never edit it here.
- Wikilinks only to files verified in `<docs>`. Code, PRs, Slack stay plain text.
