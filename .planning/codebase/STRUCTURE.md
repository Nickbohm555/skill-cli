# Codebase Structure

**Analysis Date:** 2026-03-10

## Directory Layout

```text
cli-skill/
├── cmd/                     # CLI binary entrypoint package
│   └── cli-skill/           # main package for executable startup
├── internal/                # Private application/domain packages
│   ├── app/generate/        # Generation pipeline orchestration stages and gates
│   ├── cli/                 # Command handlers and interactive prompt adapters
│   ├── content/             # Extraction, normalization, chunking, summarization, review view
│   ├── crawl/               # Same-domain crawl engine and URL classification
│   ├── install/             # Preflight, approval, transaction, activation verification
│   ├── overlap/             # Installed-skill indexing and overlap/conflict resolution
│   ├── refinement/          # Interactive field state machine and clarity/revision flow
│   └── validation/          # Skill parser, schema/semantic validators, validation model
├── assets/                  # Static SVG docs assets
├── .planning/               # Planning artifacts and generated codebase analysis docs
├── go.mod                   # Go module and dependency definitions
├── go.sum                   # Go dependency checksums
└── README.md                # Project overview and usage
```

## Directory Purposes

**`cmd/cli-skill/`:**
- Purpose: Own executable startup and top-level command execution.
- Contains: `main.go`.
- Key files: `cmd/cli-skill/main.go`.

**`internal/cli/command/`:**
- Purpose: Define Cobra root/subcommands and command-specific orchestration/rendering.
- Contains: `crawl.go`, `process.go`, `refine.go` (+ tests).
- Key files: `internal/cli/command/crawl.go`, `internal/cli/command/process.go`, `internal/cli/command/refine.go`.

**`internal/cli/prompts/`:**
- Purpose: Convert refinement fields into UI prompt plans and render review sections.
- Contains: `refinement_form.go`, `review_renderer.go` (+ tests).
- Key files: `internal/cli/prompts/refinement_form.go`, `internal/cli/prompts/review_renderer.go`.

**`internal/crawl/`:**
- Purpose: Canonicalize URLs, derive docs root, run bounded crawler, and classify skips.
- Contains: `engine.go`, `classify.go`, `normalize.go`, `docs_root.go`, `types.go`.
- Key files: `internal/crawl/engine.go`, `internal/crawl/classify.go`, `internal/crawl/types.go`.

**`internal/content/`:**
- Purpose: Process fetched pages into attributed chunks and summary-first review output.
- Contains: `extract.go`, `normalize.go`, `dedupe.go`, `chunk.go`, `pipeline.go`, `summarize.go`, `review_view.go`, `attribution.go`, `types.go`.
- Key files: `internal/content/pipeline.go`, `internal/content/summarize.go`, `internal/content/review_view.go`.

**`internal/refinement/`:**
- Purpose: Keep deterministic field state and run adaptive deepening/revision logic.
- Contains: `session.go`, `flow.go`, `clarity.go`, `validator.go`, `field_graph.go`, `revise.go`.
- Key files: `internal/refinement/session.go`, `internal/refinement/flow.go`, `internal/refinement/field_graph.go`.

**`internal/validation/`:**
- Purpose: Parse markdown skill docs and produce sorted structural/semantic validation issues.
- Contains: `parse_skill.go`, `schema_validate.go`, `semantic_validate.go`, `report.go`, `model.go`, `followup_prompt.go`, `skill.schema.json`.
- Key files: `internal/validation/parse_skill.go`, `internal/validation/schema_validate.go`, `internal/validation/semantic_validate.go`.

**`internal/overlap/`:**
- Purpose: Build installed skill index and compute conflict-resolution decisions from overlap signals.
- Contains: `index_installed.go`, `detect.go`, `classify.go`, `decision_flow.go`, `report.go`, `resolution_summary.go`, `model.go`.
- Key files: `internal/overlap/index_installed.go`, `internal/overlap/detect.go`, `internal/overlap/decision_flow.go`.

**`internal/install/`:**
- Purpose: Enforce preflight/approval gates and perform atomic install + activation verification.
- Contains: `preflight_gates.go`, `approval_prompt.go`, `preview_diff.go`, `transaction.go`, `activate_verify.go`, `errors.go`, `model.go`.
- Key files: `internal/install/preflight_gates.go`, `internal/install/transaction.go`, `internal/install/activate_verify.go`.

**`internal/app/generate/`:**
- Purpose: Compose validation, overlap, and install layers into reusable phase stage APIs.
- Contains: `gate.go`, `fix_loop.go`, `overlap_stage.go`, `install_stage.go`.
- Key files: `internal/app/generate/fix_loop.go`, `internal/app/generate/overlap_stage.go`, `internal/app/generate/install_stage.go`.

## Key File Locations

**Entry Points:**
- `cmd/cli-skill/main.go`: Binary startup and command execution.
- `internal/cli/command/crawl.go`: Root command factory (`NewRootCommand`) plus `crawl` command.
- `internal/cli/command/process.go`: `process` command with crawl + content orchestration.
- `internal/cli/command/refine.go`: `refine` command and review/commit loop.

**Configuration:**
- `go.mod`: Module path, Go version, and pinned dependencies.
- `README.md`: Usage contract and high-level phase map.
- `internal/validation/skill.schema.json`: Structural schema contract for candidate skill validation.

**Core Logic:**
- `internal/crawl/engine.go`: Crawl queue/state machine and domain filtering.
- `internal/content/pipeline.go`: Normalized page -> attributed chunk pipeline.
- `internal/content/summarize.go`: Provider-backed summary with deterministic fallback.
- `internal/refinement/flow.go`: Adaptive question/deepening/revise flow state machine.
- `internal/overlap/detect.go`: Overlap scoring/classification.
- `internal/install/transaction.go`: Atomic staged install transaction.

**Testing:**
- `internal/crawl/*_test.go`: Crawl normalization/classification/engine tests.
- `internal/content/*_test.go`: Extraction/chunk/summarize/review projection tests.
- `internal/refinement/*_test.go`: Session, clarity, flow, and validator behavior tests.
- `internal/validation/validation_test.go`: Structural + semantic validation tests.
- `internal/overlap/*_test.go`: Indexing, detection, and decision flow tests.
- `internal/install/*_test.go`: Preflight, approval, preview/diff, transaction, activation tests.
- `internal/app/generate/*_test.go`: Gate/fix-loop/overlap/install stage orchestration tests.

## Naming Conventions

**Files:**
- Domain source files use lowercase snake_case names by behavior, e.g. `internal/install/preflight_gates.go`, `internal/overlap/decision_flow.go`.
- Unit tests are colocated with `_test.go` suffix, e.g. `internal/refinement/flow_test.go`.

**Directories:**
- Runtime code is under `internal/<domain>/` and command entrypoint code is under `cmd/cli-skill/`.
- Orchestration logic is nested under `internal/app/generate/` to separate multi-package pipeline coordination from pure domain packages.

## Where to Add New Code

**New CLI Subcommand:**
- Primary code: `internal/cli/command/<name>.go` with registration in `internal/cli/command/crawl.go` (`NewRootCommand`).
- Tests: `internal/cli/command/<name>_test.go`.

**New Crawl Capability:**
- Primary code: `internal/crawl/`.
- Tests: add/extend `internal/crawl/*_test.go`.

**New Content Processing Step:**
- Primary code: `internal/content/` (`pipeline.go` for orchestration boundaries, new focused file for pure transform).
- Tests: `internal/content/*_test.go`.

**New Refinement Rule or Field Dependency:**
- Primary code: `internal/refinement/clarity.go`, `internal/refinement/validator.go`, or `internal/refinement/field_graph.go`.
- Tests: `internal/refinement/clarity_test.go`, `internal/refinement/validator_test.go`, `internal/refinement/flow_test.go`.

**New Validation Rule:**
- Structural schema constraints: `internal/validation/skill.schema.json` + mapping in `internal/validation/schema_validate.go`.
- Semantic rule checks: `internal/validation/semantic_validate.go`.
- Tests: `internal/validation/validation_test.go`.

**New Overlap Heuristic:**
- Primary code: `internal/overlap/detect.go` and optionally `internal/overlap/classify.go`.
- Tests: `internal/overlap/detect_test.go` and `internal/overlap/model_report_test.go`.

**New Install Safety Gate or Verification:**
- Primary code: `internal/install/preflight_gates.go`, `internal/install/transaction.go`, or `internal/install/activate_verify.go`.
- Tests: matching files in `internal/install/*_test.go`.

**New Full-Pipeline Orchestration Rule:**
- Primary code: `internal/app/generate/` stage files.
- Tests: `internal/app/generate/*_test.go`.

**Shared Utilities:**
- Keep utilities inside the owning domain package (for example `internal/content/` helper in that package) instead of creating a global utility directory; no cross-domain `pkg/` utility package exists in current structure.

## Special Directories

**`.planning/`:**
- Purpose: Roadmap, phase summaries, and generated codebase docs.
- Generated: Yes (planning workflows generate/update many docs).
- Committed: Yes.

**`.planning/codebase/`:**
- Purpose: Machine-consumable architecture/stack/convention/concern documentation for planner/executor flows.
- Generated: Yes.
- Committed: Yes.

**`assets/`:**
- Purpose: Documentation visuals used by `README.md`.
- Generated: No.
- Committed: Yes.

---

*Structure analysis: 2026-03-10*
