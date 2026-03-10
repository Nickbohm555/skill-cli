---
phase: 04-validation-quality-gates
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/validation/model.go
  - internal/validation/parse_skill.go
  - internal/validation/report.go
  - internal/validation/schema_validate.go
  - internal/validation/semantic_validate.go
  - internal/validation/skill.schema.json
  - internal/validation/validation_test.go
autonomous: true
must_haves:
  truths:
    - "A generated skill is rejected immediately when required sections or schema constraints are missing or malformed."
    - "Validation output is deterministic and machine-readable with stable rule IDs and severity labels."
    - "In-scope and out-of-scope boundaries are required and semantically checked, not just string-present."
  artifacts:
    - path: "internal/validation/model.go"
      provides: "Normalized skill candidate model used by all validation passes."
    - path: "internal/validation/parse_skill.go"
      provides: "Frontmatter plus markdown section parsing into normalized model."
    - path: "internal/validation/schema_validate.go"
      provides: "Structural/schema validation with blocking error mapping."
    - path: "internal/validation/semantic_validate.go"
      provides: "Boundary-focused semantic validation rules for scope quality."
    - path: "internal/validation/report.go"
      provides: "Ordered issue report model with stable rule IDs and severity."
    - path: "internal/validation/skill.schema.json"
      provides: "Runtime schema source of truth for strict required-field checks."
  key_links:
    - from: "internal/validation/parse_skill.go"
      to: "internal/validation/model.go"
      via: "parser output normalized into candidate struct used by validators"
      pattern: "ParseSkill|CandidateSkill"
    - from: "internal/validation/schema_validate.go"
      to: "internal/validation/report.go"
      via: "schema failures converted into stable Error issues"
      pattern: "RuleID|SeverityError|AddIssue"
    - from: "internal/validation/semantic_validate.go"
      to: "internal/validation/report.go"
      via: "semantic scope checks emit deterministic prioritized issues"
      pattern: "VAL\\.SCOPE|SortIssues"
---

<objective>
Implement the deterministic validation core that enforces strict structural gates first and semantic scope quality second.

Purpose: Fulfill VAL-01 and VAL-03 by establishing fail-closed acceptance rules and explicit boundary requirements before any downstream progression.
Output: Validation model/parser, schema contract, semantic boundary checks, and deterministic report/test coverage.
</objective>

<execution_context>
@~/.cursor/get-shit-done/workflows/execute-plan.md
@~/.cursor/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/STATE.md
@.planning/phases/04-validation-quality-gates/04-CONTEXT.md
@.planning/phases/04-validation-quality-gates/04-RESEARCH.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create normalized skill model, markdown parser, and issue report contract</name>
  <files>internal/validation/model.go, internal/validation/parse_skill.go, internal/validation/report.go</files>
  <action>Define `CandidateSkill` and section-level model types that represent all required fields for Codex skill acceptability. Implement parser logic that reads `SKILL.md` frontmatter and required headings into that model (avoid regex-only parsing; parse by structured markdown traversal). Add `ValidationIssue`/`ValidationReport` with stable `RuleID`, `Severity` (`Error`, `Warning`), deterministic ordering, and helper methods (`HasBlockingIssues`, `NextBlockingIssue`).</action>
  <verify>`go test ./...` compiles and parser/report tests validate deterministic issue ordering for repeated runs on the same fixture.</verify>
  <done>A single normalized input contract and deterministic issue-report contract exist for all validation passes, with stable IDs suitable for prompt targeting.</done>
</task>

<task type="auto">
  <name>Task 2: Implement strict structural/schema validation pass</name>
  <files>internal/validation/schema_validate.go, internal/validation/skill.schema.json, internal/validation/validation_test.go</files>
  <action>Create JSON Schema that enforces required sections/fields and strict shape constraints for candidate skill structure. Implement structural validation that compiles/applies the schema and maps violations into stable blocking `Error` issues (no best-effort bypasses). Ensure required-section absence and malformed values produce explicit rule IDs/messages that can be consumed by the fix loop in the next plan.</action>
  <verify>`go test ./internal/validation -run Structural -v` passes with fixtures that prove missing required sections fail closed.</verify>
  <done>Structural validation is deterministic and blocks progression on any missing/malformed required content as required by VAL-01.</done>
</task>

<task type="auto">
  <name>Task 3: Implement semantic boundary validation for in-scope/out-of-scope quality</name>
  <files>internal/validation/semantic_validate.go, internal/validation/validation_test.go</files>
  <action>Add semantic checks for explicit scope boundaries: both in-scope and out-of-scope sections must exist, contain non-trivial entries, and avoid vague catch-all phrasing. Emit stable `VAL.SCOPE.*` rule IDs with deterministic priority so the first blocking issue is predictable. Keep semantic validation as required pass two after structural validation, not optional linting.</action>
  <verify>`go test ./internal/validation -run Semantic -v` passes and includes positive/negative fixtures for boundary specificity.</verify>
  <done>Validator enforces explicit and semantically acceptable scope boundaries, satisfying VAL-03 and producing deterministic rule-level output.</done>
</task>

</tasks>

<verification>
1. Run `go test ./internal/validation -v` and confirm structural and semantic suites both pass.
2. Validate deterministic ordering by executing the same fixture test multiple times and confirming the same first blocking issue is reported.
3. Confirm `HasBlockingIssues()` only reacts to `Error` severity and does not block on warning-only reports.
</verification>

<success_criteria>
- Any missing required section or schema violation yields at least one blocking `Error` and fails closed.
- Semantic boundary checks require explicit in-scope and out-of-scope definitions with rule-level traceability.
- Validation report is machine-readable and deterministic, enabling targeted follow-up prompts in the next wave.
</success_criteria>

<output>
After completion, create `.planning/phases/04-validation-quality-gates/04-01-SUMMARY.md`
</output>

## Section 1: Core validator contracts - strict structural and semantic checks

**Single goal:** Establish a deterministic validation core that blocks invalid skill output and emits stable issue data for guided remediation.

**Details:**
- Parse `SKILL.md` into normalized typed structures before applying validation rules.
- Run validation in two passes: structural/schema first, semantic boundary rules second.
- Enforce fail-closed behavior on any blocking `Error`.
- Use stable rule IDs and deterministic issue ordering to support one-issue-at-a-time follow-up.

**Tech stack and dependencies**
- Libraries/packages (Go modules): `github.com/yuin/goldmark`, `go.abhg.dev/goldmark/frontmatter`, `github.com/santhosh-tekuri/jsonschema/v6`, and `github.com/go-playground/validator/v10` as documented in `04-RESEARCH.md`.
- Tooling/runtime: no Docker or runtime changes; validation executes within existing Go CLI test/build workflow.

**Files and purpose**

| File | Purpose |
|------|---------|
| internal/validation/model.go | Defines normalized candidate skill structures shared by parser and validators. |
| internal/validation/parse_skill.go | Parses markdown/frontmatter into the normalized model. |
| internal/validation/report.go | Defines issue severities, rule IDs, sorting, and blocking helpers. |
| internal/validation/schema_validate.go | Runs strict schema validation and maps violations to blocking issues. |
| internal/validation/semantic_validate.go | Enforces explicit in-scope/out-of-scope semantic rules. |
| internal/validation/skill.schema.json | Declares structural source-of-truth constraints for candidate skills. |
| internal/validation/validation_test.go | Covers structural fail-closed behavior, semantic boundaries, and deterministic issue ordering. |

**How to test:** Run `go test ./internal/validation -v` and verify failing fixtures return explicit blocking errors while valid fixtures pass both validation passes.
