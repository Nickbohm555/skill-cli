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

Notes:
- Reused the existing `internal/crawl/types.go` and `internal/crawl/skip_reasons.go` implementation because the crawl result models and stable skip taxonomy were already present and correctly wired through `SkippedRecord.Reason`.
- Added `go.mod` with module path `github.com/Nickbohm555/skill-cli` so Go-native verification could run in this repository.
- Verification run output:
  - `go test ./...` -> `? github.com/Nickbohm555/skill-cli/internal/crawl [no test files]`
  - `go test ./internal/crawl -v` -> `? github.com/Nickbohm555/skill-cli/internal/crawl [no test files]`

## Section 2 — 01-crawl-ingestion-foundation — 01-01 — Task 1 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-01-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 1 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-01` / `task=1` / `status=verified`.

Notes:
- Re-ran the Task 1 verification commands without adding implementation scope.
- Confirmed the skip taxonomy strings are defined only in `internal/crawl/skip_reasons.go` and consumed through `SkippedRecord.Reason`, so there are no duplicated ad hoc skip-reason literals in the codebase.
- Verification run output:
  - `go test ./...` -> `? github.com/Nickbohm555/skill-cli/internal/crawl [no test files]`
  - `go test ./internal/crawl -v` -> `? github.com/Nickbohm555/skill-cli/internal/crawl [no test files]`

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

Notes:
- Added `internal/crawl/normalize.go` with shared helpers for entry URL normalization, canonical key generation, and same-domain checks so later crawl code can reuse one boundary policy.
- The normalization policy now resolves relative links against a base URL, strips fragments unconditionally, removes tracking params (`utm_*`, `gclid`, `fbclid`), preserves other query params in stable order, cleans paths, and normalizes hosts plus default ports.
- Verification run output:
  - `go fmt ./...` -> no output
  - `go test ./...` -> `? github.com/Nickbohm555/skill-cli/internal/crawl [no test files]`
