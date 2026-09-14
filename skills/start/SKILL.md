---
name: start
description: Pick up a quest from the quno vault in this session — read it, ask the user whatever is still unclear until the idea is solid, ask the route (implement, investigate, propose to the team after exploring and extending the idea with the user, record an ADR, brief only, drop), record this session id, do the work inside the same quest file, write findings. Works at any phase: raw, ready, investigating, proposed, adr, in-progress. Use when the user says /quno:start <id or slug>, "start quest X", "pick up X", "continue the investigation on X", or when spawned by `quno start`. No argument lists the quests that are ready.
argument-hint: <quest id, slug or title fragment>
---

# quno:start

Adopt a quest in this session. The quest file is the ticket and the record: investigation, decision and findings all go into it. Before any work the idea has to be solid and the user has to pick the route.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Find.** `$ARGUMENTS` empty → list quests not `done` or `dropped` (id, status, title, project), stop. Else match `id:` exactly or by unique prefix, then `<docs>/quests/<slug>.md` exactly, then by filename or `# title` substring. Several → list them, stop. None → say so, stop.
2. **Raw?** `status: raw` → apply the `/quno:parse` steps to this one quest first (brief, retitle, rename), continue with the new path.
3. **cwd.** `repo` set and cwd not inside it → stop; tell the user to `cd <repo>` and rerun, or use `quno start <id>`. Project context is cwd-based; never work from the wrong directory.
4. **Understand.** Read every section present and open the Brief's entry points in the repo. Restate the quest in three lines: what changes, where, done-when; for `investigating` or `adr` quests, the bottom line so far and what is still open. Everything the repo did not answer goes into one AskUserQuestion, batched, concrete options plus free text: which surface or flow, scope edges, behaviour on the edge cases you found, done-when, contradictions between sections. Do not ask what the repo or the Brief answers. Repeat once if answers open new gaps. Write the outcome under `## Brief` as a `**Clarified <YYYY-MM-DD>:**` block; fix scope or entry points if they were wrong. Never touch `## Idea`. If the restatement needed no questions, say so and move on.
5. **Route.** One AskUserQuestion, options fitted to the phase and in this order: Implement now · Investigate (or continue investigating) · Propose to team (needs their decision, buy-in or hands) · Record ADR (when the decision is clear, yours or the team's) · Brief only (stop, status unchanged) · Drop (`status: dropped`, one line why under Findings). A Brief opening with `Investigate first:` or `Propose first:` puts that route first and says why. A `proposed` quest offers: Record ADR (the team answered) · Revise proposal · Implement now · Drop. An `in-progress` quest being resumed skips this unless Findings say the route is undecided.
6. **Claim.** For Implement, Investigate, Propose, Record ADR: `echo $CLAUDE_CODE_SESSION_ID` → `session:`. Status: `in-progress`, `investigating`, `proposed`, or `adr` when the ADR is written. Say in one line what you are about to do, then go.
7. **Work**, inside the quest file, sections per Conventions:
   - Investigate: research, then write or extend `## Investigation`: the question, evidence per claim (`path:line`, command, the override that beat a first guess), options, `### Bottom line`. Extending → dated `_<date>:_` lines, untouched parts stay. Then ask: Record ADR · Implement now · Stop here.
   - Propose: explore before writing. Two or three rounds of AskUserQuestion that push the idea outward, every option a concrete suggestion you formed from the repo and the product, never a blank to fill: who else this serves and how; the smallest version and the ambitious one; the adjacent problem the same change would solve; what it replaces, breaks or makes obsolete; what success looks like in numbers or behaviour; what would make the team say no and how you would answer; extensions you would add yourself, named, keep or drop. Each round builds on the last answer; stop when a round adds nothing. Record the trail under `## Brief` as `**Explored <date>:**`, one line per direction taken or rejected with the reason. Then one question on audience: who reads this, what you want from them (a yes/no, feedback, priority, people), deadline, what they already know. Then write or rewrite `## Proposal` for that audience: Problem · Proposal · Why now · Alternatives · Cost and risk · Open questions · Ask. Self-contained, one screen, plain language; no `path:line` unless a reader needs it. Status `proposed`. Report `quno cat <id> proposal | pbcopy` as the way out. Stop here; the team's answer comes back through Record ADR.
   - Record ADR: write `## ADR`: context, drivers, at least two options with pros and cons, chosen and why the losers lost, consequences. A team decision names who decided and when (`Decided 2026-09-14 with the team:`); a rejection is still an ADR, then `status: dropped` with the reason under Findings. Supersedes another quest's decision → `Supersedes [[slug]]` here, `Superseded by [[this-slug]]` and `status: done` there. Status `adr`. Then ask: Implement now · Stop here.
   - Implement: status `in-progress`, do it. Repo rules (CLAUDE.md, style) apply. No commit unless the user asks.
8. **Findings.** Append or extend `## Findings`: outcome, PR or branch, next step if stopping mid-way (status then stays where it is). Finished → `status: done`.
9. **Report** one line: id, status, what landed.

## Rules

- Questions before work. Solid means the user could hand the quest to someone else and get the same result.
- Never rewrite `## Idea`. A wrong Brief may be corrected; note the correction under Clarified. Explored keeps rejected directions too; they are the Alternatives the team will ask about.
- Evidence over assertion in Investigation and ADR; tables over prose for comparisons; never invent a `path:line`.
- Wikilinks only to files verified in `<docs>`.
- A lesson that generalizes beyond this repo goes through `/learn`, not into Findings.
