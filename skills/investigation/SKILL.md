---
name: investigation
description: Archive an investigation, audit, or trade-off analysis already done in this conversation as a decision-ready note in the quno vault (~/dev/docs/investigations/) — frontmatter per Conventions, findings with file:line evidence, bottom line, wikilinks, README index line, and a link back to the quest being worked if one is active. Use when the user says /quno:investigation, "log this investigation", "archive this analysis", or after researching "should we do X / which to convert / is this sound" and the conclusion should outlive the chat.
argument-hint: [topic / title] [--update <slug>] [--quest <slug>]
allowed-tools: Read, Write, Edit, Grep, Glob, Bash(git:*), Bash(date:*), Bash(mkdir:*), Bash(gh:*), Bash(echo:*), Agent, AskUserQuestion
---

# quno:investigation

Archive, not investigate. The thinking has happened in this conversation; capture it so a reader can act without re-reading the chat. Every load-bearing claim lands with its evidence (`file:line`, command, the verdict that overrode a first guess); every recommendation carries its hazard; the doc opens with the question and closes with the bottom line.

**Arguments**: `$ARGUMENTS`

- Empty → capture the investigation just completed; infer the title.
- A title → doc title and slug.
- `--update <slug>` → amend the existing doc. A doc with the resolved slug already existing → update regardless.
- `--quest <slug>` → link to that quest. Without it, detect the active quest: the file in `<docs>/quests/` whose `session:` equals `$CLAUDE_CODE_SESSION_ID`.

## Resolve

- `<docs>`: `$QUNO_DOCS` if set, else the `docs` value in `~/dev/docs/meta/quno.toml`, else `~/dev/docs`.
- `project`: the `[projects]` entry in `quno.toml` whose path (with `~` expanded) is the longest prefix of the cwd. No match → empty.
- Read `<docs>/meta/Conventions.md` once. It is the schema; this file does not restate it.

## Steps

1. **Gather** from this conversation: the question, what was checked (with `file:line` or command evidence), the findings, adversarial corrections or overrides, recommendations with their hazards, open questions. Nothing substantive to capture → say so and stop; no empty templates. Load-bearing but fuzzy → re-check with Read or Grep, never hedge.
2. **Locate.** `<docs>/investigations/<topic-slug>.md`, kebab-case, no date prefix (living docs). Exists → update mode.
3. **Write** from `<docs>/meta/templates/investigation.md`. Frontmatter: `project`, `status: open`, `created` (today), `aliases` (short search names), `pr` when there is one, `related` (only wikilinks whose target exists in `<docs>`: sibling investigations, decisions, the quest). Body: `# Investigation — <topic>`, `> Hub: [[Home]]`, run line `_Run <date> · scope: … · method: … · related: …_`, one framing paragraph, findings shaped to the type, `## Bottom line`.
   - Recommendation audit: ranked table (item · location · verdict · fit · effort · risk), then tiers with a one-line why and the biggest hazard.
   - Root-cause or soundness study: diagnosis with evidence → why → options (minimal / proper / structural) → recommendation.
   - Survey: categorized inventory with counts and representative examples.

   Updating: amend the relevant sections, keep the original run date as `Run <orig>, updated <new>`, add a dated line on what changed. Do not rewrite untouched sections.
4. **Index.** `<docs>/investigations/README.md`: add or refresh `- [[<slug>|<topic>]] — <one-line takeaway> · <date>`.
5. **Quest link.** Active or given quest: add `[[<slug>]]` to the quest's `related`, add `[[<quest-slug>]]` to this doc's `related`, append one line under the quest's `## Findings`: `- investigation: [[<slug>]] — <takeaway>`. Never touch the quest's `## Idea`.
6. **Report** one line: path, created or updated, one-line summary, quest linked or not.

## Rules

- Faithful to the conversation. No upgrading tentative findings, no dropping the overrides that make the record trustworthy.
- Evidence over assertion; hypotheses marked as such.
- Tables over prose for comparisons; bold the verdict words; no padding. Read a neighboring file in the folder first to match voice.
- Wikilinks only to files verified in `<docs>`. PRs, issues, Slack: markdown links.
- Notes, not commits. No git commit or push unless asked.
