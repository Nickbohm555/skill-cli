# Skill Weaver

## What This Is

Skill Weaver is a Go CLI tool that turns a documentation page URL into a Codex skill scaffold through an interactive, adaptive question flow. It is designed for your personal Codex workflow so skills are generated in the right format, scoped atomically, and installed where Codex can use them. The tool only installs a generated skill after explicit user approval.

## Core Value

Generate a skill that is actually usable in Codex, with clear scope and correct installation, in one guided flow.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] User can run `skill-generator` and enter a single documentation URL as input.
- [ ] The CLI runs an adaptive interactive questioning loop until required skill fields are complete.
- [ ] The generator enforces strict skill format validation and blocks output when required sections are weak or missing.
- [ ] On validation failure, the CLI asks targeted follow-up questions and retries generation.
- [ ] Generated skill content defines a narrow, single capability with explicit boundaries.
- [ ] The tool checks overlap against existing installed skills and surfaces potential conflicts.
- [ ] If overlap is detected, the CLI offers merge/update behavior instead of silent overwrite.
- [ ] The skill is only written to `$CODEX_HOME/skills/<skill-name>` after explicit user approval.
- [ ] The final installed skill is immediately usable by Codex.

### Out of Scope

- Multi-page crawling or domain-wide ingestion in v1 — intentionally constrained to one page for quality and predictability.
- Team/shared governance workflows in v1 — first release is optimized for a single-user workflow.
- Auto-install without confirmation — conflicts with explicit approval requirement.

## Context

This project starts greenfield in an empty repository. The desired user experience is not a one-shot converter; it is a guided conversation that keeps refining output until it is valid and useful. Atomicity and uniqueness are first-class concerns: each generated skill should do one thing well and avoid overlapping responsibilities with existing skills.

## Constraints

- **Tech stack**: Go CLI implementation — selected for maintainability and strong CLI distribution characteristics.
- **Input scope**: Single URL only in v1 — keeps extraction deterministic and reduces complexity.
- **Validation policy**: Strict block-and-re-ask — output is not accepted when quality gates fail.
- **Installation target**: `$CODEX_HOME/skills` — generated artifacts must land where Codex can load them.
- **Approval gate**: Explicit user approval before write/install — protects against unintended changes.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Implement v1 in Go | Prioritize maintainability and robust CLI packaging | — Pending |
| Limit ingestion to one page URL | Minimize ambiguity and control extraction quality in first version | — Pending |
| Use adaptive questioning with strict validation | Ensure required skill schema and quality before installation | — Pending |
| Detect overlap and offer merge/update | Prevent duplicate or conflicting skills while preserving user intent | — Pending |
| Install only after explicit user approval | Keep user control over local Codex skill registry | — Pending |

---
*Last updated: 2026-03-10 after initialization*
