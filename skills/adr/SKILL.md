---
name: adr
description: Author a numbered MADR-style decision record in the quno vault (~/dev/docs/decisions/NNNN-<slug>.md) — context, drivers, at least two options with honest pros and cons, outcome, consequences — or re-status / supersede an existing one, keeping the trail. Links the decision to the active quest and the investigation that informed it. Use when the user says /quno:adr, "record this decision", "write an ADR", or when a quest's investigate route reaches a decision.
argument-hint: [decision title or question] [--supersede <id>] [--status <id> <proposed|accepted|deprecated|superseded|rejected>] [--repo] [--quest <slug>]
allowed-tools: Read, Write, Edit, Grep, Glob, Bash(date:*), Bash(mkdir:*), Bash(git:*), Bash(gh:*), Bash(echo:*), Bash(ls:*), AskUserQuestion
---

# quno:adr

Facilitator and scribe for one decision. A real ADR names at least two options it could have taken, says why the others lost, and states the consequences it accepts.

**Arguments**: `$ARGUMENTS`

- A title or question → new ADR. Empty → ask for it.
- `--supersede <id>` → new ADR replacing `<id>`; both linked, old status flipped.
- `--status <id> <status>` → status change only, no new file.
- `--repo` → a code-architecture decision that belongs with the code: target the repo's `docs/adr/`, follow its numbering and format, do not commit. Default is the vault. A clearly code-architecture decision → say so and offer `--repo`; never silently file it in the vault.
- `--quest <slug>` → link to that quest. Without it, detect the active quest: the file in `<docs>/quests/` whose `session:` equals `$CLAUDE_CODE_SESSION_ID`.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- `project`: the `[projects]` entry in `quno.toml` whose path (with `~` expanded) is the longest prefix of the cwd. No match → empty.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Mode.** `--status` → step 5. `--supersede` → read the old ADR, carry its context forward. Otherwise new: a decision already discussed in this conversation → draft from it and confirm the gaps; nothing discussed → interview (step 2).
2. **Gather**, asking only for what is missing: one batched AskUserQuestion for the forks (which option, status), free text for the rest. Context and problem; decision drivers; considered options (at least two, each with pros and cons against the drivers); outcome and justification, or `proposed` plus what would settle it; consequences good / bad / neutral; links (investigations, quests, PRs). An investigation in this conversation or in `<docs>/investigations/` that drove this → link it, do not re-derive.
3. **Number and slug.** `ls <docs>/decisions | grep -oE '^[0-9]{4}' | sort -n | tail -1`; next is max plus one, zero-padded; none → `0001`. File `NNNN-<kebab-slug>.md`. Ids are immutable and never reused. Read a neighboring ADR first to match voice.
4. **Write** from `<docs>/meta/templates/adr.md`. Frontmatter: `project`, `id: "NNNN"`, `status`, `created`, `aliases`, `supersedes` / `superseded_by` when set, `pr` when set, `related` (only verified targets: investigation, quest, prior ADR). Body: `# ADR NNNN — <title>` · `> Hub: [[Home]]` · `> **Status:** <status> · <date>` · Context and problem statement · Decision drivers · Considered options (numbered, one line each) · Decision outcome (`Chosen: **X**, because …`) · Consequences (🟢 good / 🔴 bad / ⚪ neutral) · Pros and cons of the options · More information (related notes, PR, external refs).
5. **Status and supersede.** `--status`: update `status:` and the `> **Status:**` line, append a dated line under More information with the reason; never rewrite the decision. `--supersede`: after writing the new one, set the old to `status: superseded`, `superseded_by: "[[NNNN-new]]"`, dated note; the new one gets `supersedes: "[[old]]"`. ADRs are never deleted.
6. **Quest link.** Active or given quest: `[[NNNN-slug]]` into its `related`, the quest slug into this ADR's `related`, one line under the quest's `## Findings`: `- decision: [[NNNN-slug]] — <outcome>`. Never touch `## Idea`.
7. **Report** one line: path, id, status, one-line outcome, quest linked or not. Offer "mark accepted once decided" or "supersede NNNN".

## Rules

- Facilitate, never fabricate. Undecided → `proposed` and what would settle it.
- At least two real options; say why the losers lost.
- Numbered, immutable, never deleted; supersede instead.
- Bases list decisions live from `<docs>/decisions/`; there is no hand index to maintain.
- Wikilinks only to files verified in `<docs>`. PRs, issues, Slack: markdown links.
- `--repo`: never commit or push unless asked.
- Ambiguous scope, contested drivers, undecided outcome → ask, do not guess.
