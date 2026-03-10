---
phase: 04-validation-quality-gates
plan: 02
type: execute
wave: 2
depends_on:
  - 04-01
files_modified:
  - internal/validation/followup_prompt.go
  - internal/app/generate/fix_loop.go
  - internal/app/generate/gate.go
  - internal/app/generate/fix_loop_test.go
  - internal/app/generate/gate_test.go
autonomous: true
must_haves:
  truths:
    - "When validation fails, user receives a targeted follow-up prompt tied to the first blocking rule."
    - "The system requests one focused edit at a time, then revalidates immediately."
    - "Progression remains blocked until the candidate has zero blocking errors."
  artifacts:
    - path: "internal/validation/followup_prompt.go"
      provides: "Rule-ID to targeted follow-up prompt mapping."
    - path: "internal/app/generate/fix_loop.go"
      provides: "Single-issue guided remediation loop with revalidation after each edit."
    - path: "internal/app/generate/gate.go"
      provides: "Central fail-closed progression gate based on validation severity."
    - path: "internal/app/generate/fix_loop_test.go"
      provides: "Behavior tests for one-issue-at-a-time prompt and retry loop."
    - path: "internal/app/generate/gate_test.go"
      provides: "Guarantees warning-only reports do not block while any Error always blocks."
  key_links:
    - from: "internal/app/generate/fix_loop.go"
      to: "internal/validation/followup_prompt.go"
      via: "next blocking issue mapped to one targeted prompt per iteration"
      pattern: "NextBlockingIssue|PromptForRule"
    - from: "internal/app/generate/fix_loop.go"
      to: "internal/validation/report.go"
      via: "loop exit/continue based on blocking issue state"
      pattern: "HasBlockingIssues|NextBlockingIssue"
    - from: "internal/app/generate/gate.go"
      to: "internal/validation/schema_validate.go"
      via: "post-validation gate blocks downstream progression on Errors"
      pattern: "CanProceed|ValidateCandidate"
---

<objective>
Wire targeted follow-up prompting and fail-closed progression gating into generation flow so users can only continue after acceptable output.

Purpose: Fulfill VAL-02 and complete Phase 04 gating behavior by coupling validation issues to guided remediation and strict progression control.
Output: Rule-targeted prompt mapping, single-edit retry orchestration, and centralized progression gate with test coverage.
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
@.planning/phases/04-validation-quality-gates/04-01-SUMMARY.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Implement stable rule-to-follow-up prompt mapping</name>
  <files>internal/validation/followup_prompt.go</files>
  <action>Create a deterministic mapping from validator `RuleID` to targeted question templates that request only the missing or malformed content needed to resolve that rule. Keep prompts concise, specific, and scoped to one correction target (for example, missing in-scope list, weak out-of-scope boundary, malformed metadata). Include fallback behavior for unknown rule IDs that still preserves fail-closed handling.</action>
  <verify>`go test ./...` compiles and mapping tests confirm every blocking `VAL.*` rule used in Phase 04 has a non-empty targeted prompt.</verify>
  <done>Every blocking validation rule can produce a deterministic, targeted follow-up question usable by remediation loop logic.</done>
</task>

<task type="auto">
  <name>Task 2: Add one-issue-at-a-time fix loop with immediate revalidation</name>
  <files>internal/app/generate/fix_loop.go, internal/app/generate/fix_loop_test.go</files>
  <action>Implement remediation loop orchestration that (1) runs validation, (2) selects first blocking issue, (3) asks exactly one mapped follow-up question, (4) applies one focused edit to candidate output, and (5) revalidates immediately. Preserve unlimited retries and avoid multi-issue prompt batching. Ensure loop exits only when no blocking errors remain or user explicitly cancels.</action>
  <verify>`go test ./internal/app/generate -run FixLoop -v` passes with fixtures proving single-issue prompting and immediate revalidation each cycle.</verify>
  <done>Validation failures trigger guided, targeted, one-edit remediation cycles until candidate passes or user exits.</done>
</task>

<task type="auto">
  <name>Task 3: Centralize progression gate and enforce fail-closed policy</name>
  <files>internal/app/generate/gate.go, internal/app/generate/gate_test.go</files>
  <action>Create a dedicated progression gate (`CanProceed` or equivalent) that is the only authority for allowing downstream workflow continuation. Gate must deny progression on any blocking `Error`, allow progression when only warnings remain, and return explicit reason payloads for blocked states. Replace any ad hoc issue checks in generation flow with this centralized gate call.</action>
  <verify>`go test ./internal/app/generate -run Gate -v` passes and shows block-on-error / allow-on-warning-only behavior.</verify>
  <done>Downstream progression is fail-closed and deterministic, fully aligned with Phase 04 quality-gate contract.</done>
</task>

</tasks>

<verification>
1. Run `go test ./internal/validation ./internal/app/generate -v` and confirm prompt mapping, fix-loop, and gate suites pass together.
2. Simulate candidate with two blocking issues and confirm exactly one issue is prompted per iteration until cleared.
3. Confirm progression remains blocked for any report containing `Error`, and becomes allowed only after blocking issues are resolved.
</verification>

<success_criteria>
- Validation failure always produces targeted next-step guidance rather than generic or silent failure behavior.
- Guided loop resolves one blocking issue at a time with immediate revalidation and unlimited retries.
- Progression/install-prep path cannot continue while blocking validation errors remain.
</success_criteria>

<output>
After completion, create `.planning/phases/04-validation-quality-gates/04-02-SUMMARY.md`
</output>

## Section 2: Guided remediation and progression gate wiring

**Single goal:** Ensure validation findings drive deterministic one-at-a-time correction prompts and hard-stop progression until output is acceptable.

**Details:**
- Map every blocking rule to a targeted follow-up prompt template.
- Orchestrate a single-edit retry loop that revalidates after each correction.
- Centralize proceed/block decision into one fail-closed gate used by generation flow.
- Keep warning-level issues non-blocking while preserving visibility in reports.

**Tech stack and dependencies**
- Libraries/packages (Go modules): use existing validation stack from Wave 1 plus `charm.land/huh/v2` for one-question prompt interactions where CLI prompting is required.
- Tooling/runtime: no container changes; integrates into existing CLI flow and Go test commands.

**Files and purpose**

| File | Purpose |
|------|---------|
| internal/validation/followup_prompt.go | Maps blocking rule IDs to focused follow-up questions. |
| internal/app/generate/fix_loop.go | Implements one-issue-at-a-time remediation orchestration with revalidation. |
| internal/app/generate/gate.go | Provides centralized fail-closed progression decision logic. |
| internal/app/generate/fix_loop_test.go | Verifies single-issue prompt behavior and loop control semantics. |
| internal/app/generate/gate_test.go | Verifies progression blocks on errors and allows warning-only reports. |

**How to test:** Run `go test ./internal/app/generate -v` and manually validate that failing candidates trigger one targeted prompt per loop iteration until blocking errors are cleared.
