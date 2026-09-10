---
name: start
description: Pick up a quest from the quno vault in this session — read its brief, record this session id on the quest, set it in-progress, do the work (implement, or investigate and record a decision), write findings. Use when the user says /quno:start <slug>, "start quest X", "pick up X", or when spawned by `quno start`. No argument lists the quests that are ready.
argument-hint: <quest slug or title fragment>
---

# quno:start

Adopt a quest in this session. The quest file is the ticket; its Brief is the prompt.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Find.** `$ARGUMENTS` empty → list quests with `status: ready` or `in-progress` (title, project, created), stop. Else match `id:` exactly or by unique prefix, then `<docs>/quests/<slug>.md` exactly, then by filename or `# title` substring. Several → list them, stop. None → say so, stop.
2. **Raw?** `status: raw` → apply the `/quno:parse` steps to this one quest first (brief, retitle, rename), continue with the new path.
3. **cwd.** `repo` set and cwd not inside it → stop; tell the user to `cd <repo>` and rerun, or use `quno start <slug>`. Project context (CLAUDE.md, memory, relative paths) is cwd-based; never work from the wrong directory.
4. **Claim.** `echo $CLAUDE_CODE_SESSION_ID` → `session:`; `status: in-progress`. State in two lines what the Brief asks for, then go.
5. **Work.** Follow `## Brief`.
   - Implement route: do it. Repo rules (CLAUDE.md, style) apply. No commit unless the user asks.
   - Investigate route (Brief opens with `Investigate first`, or the work turns out to need a decision): research, then `/quno:investigation` to archive it; a decision → `/quno:adr`. Both find this quest through its `session:` and link both ways.
6. **Findings.** Append `## Findings`: outcome, PR or branch, decisions, links. Finished → `status: done`. Stopping mid-way → Findings holds the state so far and the exact next step; status stays `in-progress`.
7. **Report** one line: quest path, status, what landed.

## Rules

- Never rewrite `## Idea`. A wrong Brief may be corrected; note the correction in Findings.
- Wikilinks only to files verified in `<docs>`.
- A lesson that generalizes beyond this repo goes through `/learn`, not into Findings.
