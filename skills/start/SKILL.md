---
name: start
description: Pick up a quest from the quno vault in this session — read it, ask the user whatever is still unclear until the idea is solid, ask whether to implement or investigate, record this session id on the quest, set it in-progress, do the work, write findings. Use when the user says /quno:start <id or slug>, "start quest X", "pick up X", or when spawned by `quno start`. No argument lists the quests that are ready.
argument-hint: <quest id, slug or title fragment>
---

# quno:start

Adopt a quest in this session. The quest file is the ticket. Before any work the idea has to be solid enough that the user and you mean the same thing, and the user has to pick the route.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Find.** `$ARGUMENTS` empty → list quests with `status: ready` or `in-progress` (id, title, project, created), stop. Else match `id:` exactly or by unique prefix, then `<docs>/quests/<slug>.md` exactly, then by filename or `# title` substring. Several → list them, stop. None → say so, stop.
2. **Raw?** `status: raw` → apply the `/quno:parse` steps to this one quest first (brief, retitle, rename), continue with the new path.
3. **cwd.** `repo` set and cwd not inside it → stop; tell the user to `cd <repo>` and rerun, or use `quno start <id>`. Project context (CLAUDE.md, memory, relative paths) is cwd-based; never work from the wrong directory.
4. **Understand.** Read `## Idea`, `## Brief`, any `## Findings`, and open the Brief's entry points in the repo. Then restate the quest in three lines: what changes, where, done-when. Everything that is still open after reading the repo goes into one AskUserQuestion, batched, each question with concrete options plus the free-text fallback: which surface or flow, scope edges, behaviour on the edge cases you found, what done looks like, contradictions between Idea and Brief, anything the Idea assumes that the code does not support. Do not ask what the repo answers; do not ask if the Brief already answers it. Repeat once if the answers open new gaps. Write the outcome into the quest under `## Brief` as a `**Clarified <YYYY-MM-DD>:**` block (decisions, edge cases, done-when as agreed); fix scope or entry points in the Brief if they were wrong. Never touch `## Idea`.
5. **Route.** Ask, one AskUserQuestion, options in this order: Implement now · Investigate first (research, `/quno:investigation`, `/quno:adr` if a decision falls out, then decide again) · Brief only (stop here, quest stays `ready`) · Drop (`status: dropped`, one line why under Findings). A Brief that opens with `Investigate first:` puts that option first and says why. In-progress quests being resumed skip this step unless the Findings say the route was undecided.
6. **Claim.** Only for Implement or Investigate: `echo $CLAUDE_CODE_SESSION_ID` → `session:`; `status: in-progress`. Say in one line what you are about to do, then go.
7. **Work.**
   - Implement: do it. Repo rules (CLAUDE.md, style) apply. No commit unless the user asks.
   - Investigate: research, then `/quno:investigation` to archive it; a decision → `/quno:adr`. Both find this quest through its `session:` and link both ways. When the investigation lands, ask Implement now or Brief only, and continue.
8. **Findings.** Append or extend `## Findings`: outcome, PR or branch, decisions, links. Finished → `status: done`. Stopping mid-way → Findings holds the state so far and the exact next step; status stays `in-progress`.
9. **Report** one line: quest id, status, what landed.

## Rules

- Questions before work. Solid means the user could hand the quest to someone else and get the same result; if the restatement in step 4 needed no questions, say so and move to the route.
- Never rewrite `## Idea`. A wrong Brief may be corrected; note the correction under Clarified.
- Wikilinks only to files verified in `<docs>`.
- A lesson that generalizes beyond this repo goes through `/learn`, not into Findings.
