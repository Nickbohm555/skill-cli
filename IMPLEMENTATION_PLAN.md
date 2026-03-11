Tasks are in **required implementation order** (1...n). Each section = one context window. Complete one section at a time.

Goal: mimic `/gsd-execute-phase` sequentially across phases 1–6. Each plan is executed task-by-task. Each task has a follow-up verification session in the next run.

Current section to work on: section 86. (move +1 after each turn)

---

## Global Inputs (every section reads these)
- `.planning/ROADMAP.md`
- `.planning/STATE.md`
- `.planning/config.json` (if present)

## Phase Discovery (every new plan)
- Find phase dir: `.planning/phases/01-*` or `.planning/phases/1-*` for phase 1; same pattern for later phases.
- Plan files: `*-PLAN.md` inside phase dir.
- Skip plans with existing `*-SUMMARY.md` unless explicitly re-running.

---

## Execution Protocol (applies to all sections)
- **Execution Session**: implement exactly one task from a plan.
- **Verification Session**: verify the previous task and update state.
- Always update `.planning/STATE.md` after each task and after each verification.
- If a task introduces a checkpoint, stop and create a new session for the checkpoint resolution.

## Run Scope (must be explicit)
- **One run = one task implementation OR one task verification.**
- Runs alternate: Execution → Verification → Execution → Verification.
- Do not implement and verify in the same run.

---

## Section 1 — 01-crawl-ingestion-foundation — 01-01 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-01-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Create crawl contracts and skip taxonomy).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-01` / `task=1` / `status=implemented`.

## Section 2 — 01-crawl-ingestion-foundation — 01-01 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-01-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-01` / `task=1` / `status=verified`.

## Section 3 — 01-crawl-ingestion-foundation — 01-01 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-01-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Implement URL normalization and same-domain boundary helpers).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-01` / `task=2` / `status=implemented`.

## Section 4 — 01-crawl-ingestion-foundation — 01-01 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-01-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-01` / `task=2` / `status=verified`.

## Section 5 — 01-crawl-ingestion-foundation — 01-01 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-01-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Add table-driven normalization tests for boundary correctness).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-01` / `task=3` / `status=implemented`.

## Section 6 — 01-crawl-ingestion-foundation — 01-01 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-01-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-01` / `task=3` / `status=verified`.
4. Create `01-01-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 7 — 01-crawl-ingestion-foundation — 01-02 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-02-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Implement docs-root derivation policy).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-02` / `task=1` / `status=implemented`.

## Section 8 — 01-crawl-ingestion-foundation — 01-02 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-02-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-02` / `task=1` / `status=verified`.

## Section 9 — 01-crawl-ingestion-foundation — 01-02 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-02-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Implement docs-like and low-signal classifiers).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-02` / `task=2` / `status=implemented`.

## Section 10 — 01-crawl-ingestion-foundation — 01-02 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-02-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-02` / `task=2` / `status=verified`.

## Section 11 — 01-crawl-ingestion-foundation — 01-02 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-02-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Add table-driven tests for classifier edge cases).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-02` / `task=3` / `status=implemented`.

## Section 12 — 01-crawl-ingestion-foundation — 01-02 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-02-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-02` / `task=3` / `status=verified`.
4. Create `01-02-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 13 — 01-crawl-ingestion-foundation — 01-03 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-03-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Build bounded crawl engine with strict accounting).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-03` / `task=1` / `status=implemented`.

## Section 14 — 01-crawl-ingestion-foundation — 01-03 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-03-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-03` / `task=1` / `status=verified`.

## Section 15 — 01-crawl-ingestion-foundation — 01-03 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-03-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Add engine behavior tests for CRAWL-01..04).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-03` / `task=2` / `status=implemented`.

## Section 16 — 01-crawl-ingestion-foundation — 01-03 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-03-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-03` / `task=2` / `status=verified`.

## Section 17 — 01-crawl-ingestion-foundation — 01-03 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-03-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Wire crawl command and render final user summary).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-03` / `task=3` / `status=implemented`.

## Section 18 — 01-crawl-ingestion-foundation — 01-03 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-03-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-03` / `task=3` / `status=verified`.
4. Create `01-03-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.
5. Update `.planning/ROADMAP.md` and `.planning/STATE.md` to mark the phase complete.

## Section 19 — 02-content-processing-attribution — 02-01 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-01-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Establish content contracts and extraction dependencies).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-01` / `task=1` / `status=implemented`.

## Section 20 — 02-content-processing-attribution — 02-01 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-01-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-01` / `task=1` / `status=verified`.

## Section 21 — 02-content-processing-attribution — 02-01 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-01-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Implement structure-preserving normalization and conservative dedupe).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-01` / `task=2` / `status=implemented`.

## Section 22 — 02-content-processing-attribution — 02-01 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-01-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-01` / `task=2` / `status=verified`.

## Section 23 — 02-content-processing-attribution — 02-01 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-01-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Add extraction/normalization regression tests).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-01` / `task=3` / `status=implemented`.

## Section 24 — 02-content-processing-attribution — 02-01 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-01-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-01` / `task=3` / `status=verified`.
4. Create `02-01-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 25 — 02-content-processing-attribution — 02-02 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-02-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Implement chunking strategy with semantic-first token guardrails).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-02` / `task=1` / `status=implemented`.

## Section 26 — 02-content-processing-attribution — 02-02 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-02-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-02` / `task=1` / `status=verified`.

## Section 27 — 02-content-processing-attribution — 02-02 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-02-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Attach attribution at chunk creation and wire pipeline orchestration).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-02` / `task=2` / `status=implemented`.

## Section 28 — 02-content-processing-attribution — 02-02 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-02-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-02` / `task=2` / `status=verified`.

## Section 29 — 02-content-processing-attribution — 02-02 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-02-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Add chunking and attribution persistence tests).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-02` / `task=3` / `status=implemented`.

## Section 30 — 02-content-processing-attribution — 02-02 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-02-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-02` / `task=3` / `status=verified`.
4. Create `02-02-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 31 — 02-content-processing-attribution — 02-03 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-03-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Implement schema-validated chunk summarization).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-03` / `task=1` / `status=implemented`.

## Section 32 — 02-content-processing-attribution — 02-03 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-03-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-03` / `task=1` / `status=verified`.

## Section 33 — 02-content-processing-attribution — 02-03 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-03-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Build summary-first review model with raw expansion references).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-03` / `task=2` / `status=implemented`.

## Section 34 — 02-content-processing-attribution — 02-03 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-03-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-03` / `task=2` / `status=verified`.

## Section 35 — 02-content-processing-attribution — 02-03 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-03-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Wire process command output and add summarization regression tests).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-03` / `task=3` / `status=implemented`.

## Section 36 — 02-content-processing-attribution — 02-03 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/02-content-processing-attribution/02-03-PLAN.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-CONTEXT.md`
- Reference: `.planning/phases/02-content-processing-attribution/02-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=02-content-processing-attribution` / `plan=02-03` / `task=3` / `status=verified`.
4. Create `02-03-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.
5. Update `.planning/ROADMAP.md` and `.planning/STATE.md` to mark the phase complete.

## Section 37 — 03-interactive-refinement-loop — 03-01 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-01-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Define session and field dependency contracts).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-01` / `task=1` / `status=implemented`.

## Section 38 — 03-interactive-refinement-loop — 03-01 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-01-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-01` / `task=1` / `status=verified`.

## Section 39 — 03-interactive-refinement-loop — 03-01 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-01-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Implement clarity scoring and deepening safeguards).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-01` / `task=2` / `status=implemented`.

## Section 40 — 03-interactive-refinement-loop — 03-01 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-01-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-01` / `task=2` / `status=verified`.

## Section 41 — 03-interactive-refinement-loop — 03-01 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-01-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Build readiness validator and test commit gate behavior).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-01` / `task=3` / `status=implemented`.

## Section 42 — 03-interactive-refinement-loop — 03-01 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-01-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-01` / `task=3` / `status=verified`.
4. Create `03-01-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 43 — 03-interactive-refinement-loop — 03-02 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-02-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Build huh-based adapters for primary and deepening questions).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-02` / `task=1` / `status=implemented`.

## Section 44 — 03-interactive-refinement-loop — 03-02 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-02-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-02` / `task=1` / `status=verified`.

## Section 45 — 03-interactive-refinement-loop — 03-02 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-02-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Render sectioned final review with readiness indicators).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-02` / `task=2` / `status=implemented`.

## Section 46 — 03-interactive-refinement-loop — 03-02 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-02-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-02` / `task=2` / `status=verified`.

## Section 47 — 03-interactive-refinement-loop — 03-02 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-02-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Test deepening fallback behavior and deterministic option routing).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-02` / `task=3` / `status=implemented`.

## Section 48 — 03-interactive-refinement-loop — 03-02 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-02-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-02` / `task=3` / `status=verified`.
4. Create `03-02-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 49 — 03-interactive-refinement-loop — 03-03 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-03-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Implement deterministic refinement loop orchestration).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-03` / `task=1` / `status=implemented`.

## Section 50 — 03-interactive-refinement-loop — 03-03 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-03-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-03` / `task=1` / `status=verified`.

## Section 51 — 03-interactive-refinement-loop — 03-03 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-03-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Add `revise &lt;field&gt;` handling with impact-aware re-ask).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-03` / `task=2` / `status=implemented`.

## Section 52 — 03-interactive-refinement-loop — 03-03 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-03-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-03` / `task=2` / `status=verified`.

## Section 53 — 03-interactive-refinement-loop — 03-03 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-03-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Wire refine CLI command and enforce final commit gate).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-03` / `task=3` / `status=implemented`.

## Section 54 — 03-interactive-refinement-loop — 03-03 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/03-interactive-refinement-loop/03-03-PLAN.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-CONTEXT.md`
- Reference: `.planning/phases/03-interactive-refinement-loop/03-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=03-interactive-refinement-loop` / `plan=03-03` / `task=3` / `status=verified`.
4. Create `03-03-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.
5. Update `.planning/ROADMAP.md` and `.planning/STATE.md` to mark the phase complete.

## Section 55 — 04-validation-quality-gates — 04-01 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-01-core-validator-contracts-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Create normalized skill model, markdown parser, and issue report contract).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-01` / `task=1` / `status=implemented`.

## Section 56 — 04-validation-quality-gates — 04-01 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-01-core-validator-contracts-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-01` / `task=1` / `status=verified`.

## Section 57 — 04-validation-quality-gates — 04-01 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-01-core-validator-contracts-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Implement strict structural/schema validation pass).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-01` / `task=2` / `status=implemented`.

## Section 58 — 04-validation-quality-gates — 04-01 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-01-core-validator-contracts-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-01` / `task=2` / `status=verified`.

## Section 59 — 04-validation-quality-gates — 04-01 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-01-core-validator-contracts-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Implement semantic boundary validation for in-scope/out-of-scope quality).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-01` / `task=3` / `status=implemented`.

## Section 60 — 04-validation-quality-gates — 04-01 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-01-core-validator-contracts-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-01` / `task=3` / `status=verified`.
4. Create `04-01-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 61 — 04-validation-quality-gates — 04-02 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-02-guided-fix-loop-gating-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Implement stable rule-to-follow-up prompt mapping).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-02` / `task=1` / `status=implemented`.

## Section 62 — 04-validation-quality-gates — 04-02 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-02-guided-fix-loop-gating-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-02` / `task=1` / `status=verified`.

## Section 63 — 04-validation-quality-gates — 04-02 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-02-guided-fix-loop-gating-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Add one-issue-at-a-time fix loop with immediate revalidation).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-02` / `task=2` / `status=implemented`.

## Section 64 — 04-validation-quality-gates — 04-02 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-02-guided-fix-loop-gating-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-02` / `task=2` / `status=verified`.

## Section 65 — 04-validation-quality-gates — 04-02 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-02-guided-fix-loop-gating-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Centralize progression gate and enforce fail-closed policy).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-02` / `task=3` / `status=implemented`.

## Section 66 — 04-validation-quality-gates — 04-02 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/04-validation-quality-gates/04-02-guided-fix-loop-gating-PLAN.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-CONTEXT.md`
- Reference: `.planning/phases/04-validation-quality-gates/04-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=04-validation-quality-gates` / `plan=04-02` / `task=3` / `status=verified`.
4. Create `04-02-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.
5. Update `.planning/ROADMAP.md` and `.planning/STATE.md` to mark the phase complete.

## Section 67 — 05-overlap-conflict-resolution — 05-01 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-01-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Define overlap domain contracts and report schema).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-01` / `task=1` / `status=implemented`.

## Section 68 — 05-overlap-conflict-resolution — 05-01 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-01-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-01` / `task=1` / `status=verified`.

## Section 69 — 05-overlap-conflict-resolution — 05-01 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-01-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Build installed-skill indexer with normalized profile extraction).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-01` / `task=2` / `status=implemented`.

## Section 70 — 05-overlap-conflict-resolution — 05-01 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-01-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-01` / `task=2` / `status=verified`.

## Section 71 — 05-overlap-conflict-resolution — 05-01 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-01-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Implement layered overlap detection and severity classification).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-01` / `task=3` / `status=implemented`.

## Section 72 — 05-overlap-conflict-resolution — 05-01 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-01-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-01` / `task=3` / `status=verified`.
4. Create `05-01-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 73 — 05-overlap-conflict-resolution — 05-02 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-02-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Implement explicit decision flow with safe interruption behavior).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-02` / `task=1` / `status=implemented`.

## Section 74 — 05-overlap-conflict-resolution — 05-02 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-02-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-02` / `task=1` / `status=verified`.

## Section 75 — 05-overlap-conflict-resolution — 05-02 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-02-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Add resolution summary artifact and pre-install display contract).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-02` / `task=2` / `status=implemented`.

## Section 76 — 05-overlap-conflict-resolution — 05-02 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-02-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-02` / `task=2` / `status=verified`.

## Section 77 — 05-overlap-conflict-resolution — 05-02 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-02-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Enforce generate-pipeline gating on unresolved conflict states).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-02` / `task=3` / `status=implemented`.

## Section 78 — 05-overlap-conflict-resolution — 05-02 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/05-overlap-conflict-resolution/05-02-PLAN.md`
- Reference: `.planning/phases/05-overlap-conflict-resolution/05-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=05-overlap-conflict-resolution` / `plan=05-02` / `task=3` / `status=verified`.
4. Create `05-02-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.
5. Update `.planning/ROADMAP.md` and `.planning/STATE.md` to mark the phase complete.

## Section 79 — 06-approval-gated-install-activation — 06-01 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-01-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Create typed install contracts and fail-closed error taxonomy).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-01` / `task=1` / `status=implemented`.

## Section 80 — 06-approval-gated-install-activation — 06-01 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-01-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-01` / `task=1` / `status=verified`.

## Section 81 — 06-approval-gated-install-activation — 06-01 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-01-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Implement centralized preflight gate for validation and conflict prerequisites).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-01` / `task=2` / `status=implemented`.

## Section 82 — 06-approval-gated-install-activation — 06-01 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-01-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-01` / `task=2` / `status=verified`.

## Section 83 — 06-approval-gated-install-activation — 06-01 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-01-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Add explicit approval flow with deny-by-default non-interactive policy).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-01` / `task=3` / `status=implemented`.

## Section 84 — 06-approval-gated-install-activation — 06-01 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-01-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-01` / `task=3` / `status=verified`.
4. Create `06-01-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 85 — 06-approval-gated-install-activation — 06-02 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-02-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Build deterministic preview and diff rendering before approval).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-02` / `task=1` / `status=implemented`.

## Section 86 — 06-approval-gated-install-activation — 06-02 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-02-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-02` / `task=1` / `status=verified`.

## Section 87 — 06-approval-gated-install-activation — 06-02 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-02-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Implement approval-gated atomic install transaction).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-02` / `task=2` / `status=implemented`.

## Section 88 — 06-approval-gated-install-activation — 06-02 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-02-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-02` / `task=2` / `status=verified`.

## Section 89 — 06-approval-gated-install-activation — 06-02 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-02-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Prove sequence integrity and no-write-before-approval behavior).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-02` / `task=3` / `status=implemented`.

## Section 90 — 06-approval-gated-install-activation — 06-02 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-02-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-02` / `task=3` / `status=verified`.
4. Create `06-02-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.

## Section 91 — 06-approval-gated-install-activation — 06-03 — Task 1 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-03-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 1 (Task 1: Implement post-install activation verification contract).
2. Implement Task 1.
3. Run Task 1 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-03` / `task=1` / `status=implemented`.

## Section 92 — 06-approval-gated-install-activation — 06-03 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-03-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-03` / `task=1` / `status=verified`.

## Section 93 — 06-approval-gated-install-activation — 06-03 — Task 2 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-03-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 2 (Task 2: Wire strict install sequence into generation pipeline stage).
2. Implement Task 2.
3. Run Task 2 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-03` / `task=2` / `status=implemented`.

## Section 94 — 06-approval-gated-install-activation — 06-03 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-03-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-03` / `task=2` / `status=verified`.

## Section 95 — 06-approval-gated-install-activation — 06-03 — Task 3 (Execution)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-03-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Read plan frontmatter + Task 3 (Task 3: Add end-to-end tests for INST-01..INST-04 acceptance paths).
2. Implement Task 3.
3. Run Task 3 verification steps from the plan.
4. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-03` / `task=3` / `status=implemented`.

## Section 96 — 06-approval-gated-install-activation — 06-03 — Task 3 (Verification)
Inputs:
- Plan file: `.planning/phases/06-approval-gated-install-activation/06-03-PLAN.md`
- Reference: `.planning/phases/06-approval-gated-install-activation/06-RESEARCH.md`
Steps:
1. Re-run verification for Task 3 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=06-approval-gated-install-activation` / `plan=06-03` / `task=3` / `status=verified`.
4. Create `06-03-SUMMARY.md` in the plan directory and update `.planning/STATE.md` to the next plan.
5. Update `.planning/ROADMAP.md` and `.planning/STATE.md` to mark the phase complete.

---

## Phase Completion Rules
- Summary creation happens after the final verification of each plan.
- Roadmap/state phase completion happens after the final verification of the last plan in the phase.
