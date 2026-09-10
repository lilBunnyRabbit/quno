---
name: parse
description: Triage raw quests in the quno vault — for every quest with status raw (or one named slug), dedupe, decide todo vs quest vs investigate-first, write a brief against the real repo, retitle, rename, set status ready. Use when the user says /quno:parse, "parse quests", "triage the inbox", or when spawned by `quno parse`.
argument-hint: [slug | --project <name>]
---

# quno:parse

Batch triage. Classification alone moves a one-liner to a different state; the Brief is the work that makes a quest pickup-able weeks later.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Collect.** `<docs>/quests/*.md` with `status: raw`. `$ARGUMENTS` is a slug → only that one; `--project <name>` → filter. Nothing raw → say so, stop.
2. **Per quest**, oldest `created` first:
   1. **Read** `## Idea` and `repo`. Repo path missing on disk → brief without entry points and flag it.
   2. **Dedupe** against the other quests and `<docs>/investigations/`. Duplicate → merge its Idea into the survivor as `_<date> addendum:_`, `rm` the file (shape in step 4), note it. Related → `related`.
   3. **Classify.**
      - todo: no brief needed, minutes, or not code → append `- [ ] <idea> #<project>` to `<docs>/todo.md`, then `rm` the quest file (shape in step 4).
      - quest: write the Brief against `repo`, read-only: scope, entry points as `path:line` you opened, constraints, done-when.
      - investigate first: a real brief needs research → Brief opens with `Investigate first:` and the concrete questions.
      - drop candidate: noise → leave it untouched, collect it for the question at the end. Never drop silently.
   4. **Retitle, rename.** `# <title>` ≤8 words. Rename with one plain Bash call, absolute paths, no quotes, no `~`: `mv /Users/<you>/dev/docs/quests/<old>.md /Users/<you>/dev/docs/quests/<slug>.md`. Splitting or merging: write the new files, then delete the source the same way: `rm /Users/<you>/dev/docs/quests/<old>.md`. These exact shapes match the permission allow rules; anything else prompts and fails in `-p`. Deletion denied (non-interactive, no `rm` rule) → set the source to `status: dropped`, put the survivors in its `related`, and report it so the user deletes by hand. Raw quests have no inbound links, rename is safe; `id` stays with the file, a merged duplicate's id dies with it. Collision → `-2`.
   5. **Status** `ready`. `## Idea` and the capture line under the title stay verbatim.
3. **Report** a table: old name → new path · project · route (implement / investigate / todo / dup) · one-line brief summary. Then list drop candidates as a question.

## Rules

- Never invent a `path:line`; open the file at that line before citing it.
- Wikilinks only to files verified in `<docs>`.
- One Brief per quest. Split only when the Idea holds clearly separate items with separate done-whens; then sibling files, cross-linked.
