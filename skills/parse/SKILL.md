---
name: parse
description: Triage raw quests in the quno vault — for every quest with status raw (or one named quest), dedupe, decide todo vs quest vs investigate-first, write a Brief against the real repo, retitle, rename, set status ready. Use when the user says /quno:parse, "parse quests", "triage the inbox", or when spawned by `quno parse`.
argument-hint: [id or slug | --project <name>]
---

# quno:parse

Batch triage. Classification alone moves a one-liner to a different state; the Brief is what makes a quest pickup-able weeks later.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Collect.** `<docs>/quests/*.md` with `status: raw`. `$ARGUMENTS` names one (id, prefix or slug) → only that one; `--project <name>` → filter. Nothing raw → say so, stop.
2. **Per quest**, oldest `created` first:
   1. **Read** `## Idea` and `repo`. Repo missing on disk → brief without entry points and flag it.
   2. **Dedupe** against the other quests. Duplicate → merge its Idea into the survivor as `_<date> addendum:_`, delete the file (shape in step 4), note it. Related → `related`.
   3. **Classify.**
      - todo: no brief needed, minutes, or not code → append `- [ ] <idea> #<project>` to `<docs>/todo.md`, delete the quest file.
      - quest: write the Brief against `repo`, read-only: scope, entry points as `path:line` you opened, constraints, done-when.
      - investigate first: a real brief needs research → Brief opens with `Investigate first:` and the concrete questions.
      - propose first: needs the team's decision, buy-in or hands before any work → Brief opens with `Propose first:` and what to ask them; start writes the Proposal.
      - drop candidate: noise → leave untouched, collect for the question at the end. Never drop silently.
   4. **Retitle, rename.** `# <title>` ≤8 words. One plain Bash call, absolute paths, no quotes, no `~`: `mv /Users/<you>/dev/docs/quests/<old>.md /Users/<you>/dev/docs/quests/<slug>.md`; deleting: `rm /Users/<you>/dev/docs/quests/<old>.md`. These exact shapes match the permission allow rules. Denied (non-interactive, no rule) → set the source to `status: dropped`, put survivors in its `related`, report it so the user deletes by hand. `id` stays with the file; a merged duplicate's id dies with it. Collision → `-2`.
   5. **Status** `ready`. `## Idea` and the capture line stay verbatim.
3. **Report** a table: id · old name → new path · project · route (implement / investigate / propose / todo / dup) · one-line brief. Then list drop candidates as a question.

## Rules

- Never invent a `path:line`; open the file at that line before citing it.
- Wikilinks only to files verified in `<docs>`.
- One Brief per quest. Split only when the Idea holds clearly separate items with separate done-whens; then sibling files, cross-linked.
