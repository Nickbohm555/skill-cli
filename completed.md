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

## Section 4 — 01-crawl-ingestion-foundation — 01-01 — Task 2 (Verification)
Inputs:
- Plan file: `.planning/phases/01-crawl-ingestion-foundation/01-01-PLAN.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-CONTEXT.md`
- Reference: `.planning/phases/01-crawl-ingestion-foundation/01-RESEARCH.md`
Steps:
1. Re-run verification for Task 2 (or broader checks if required).
2. If fixes required, implement and rerun verification until clean.
3. Update `.planning/STATE.md` with `phase=01-crawl-ingestion-foundation` / `plan=01-01` / `task=2` / `status=verified`.

Notes:
- Re-ran the Task 2 verification commands without expanding implementation scope because this run is verification-only.
- Confirmed skip-reason literals exist only in `internal/crawl/skip_reasons.go` and are consumed via `SkippedRecord.Reason`, with no duplicated ad hoc string usage elsewhere in the repository.
- Verified the current normalization helpers still compile cleanly before Task 3 adds explicit table-driven coverage.
- Verification run output:
  - `go test ./...` -> `? github.com/Nickbohm555/skill-cli/internal/crawl [no test files]`
  - `go test ./internal/crawl -v` -> `? github.com/Nickbohm555/skill-cli/internal/crawl [no test files]`

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

Notes:
- Added `internal/crawl/normalize_test.go` with table-driven coverage for invalid URL handling, fragment stripping, tracking query removal, stable query ordering, path cleaning, relative URL resolution, default-port normalization, and same-domain true/false cases.
- The test fixtures use docs-like URL shapes (`/docs/...`, relative links from a docs base, mixed host casing, tracking params) so boundary behavior stays aligned with the phase research notes.
- No blockers came up; the existing normalization helpers already matched the expected boundary policy, so this run stayed test-focused.
- Verification run output:
  - `go test ./internal/crawl -run Normalize -v` -> `ok  	github.com/Nickbohm555/skill-cli/internal/crawl	0.483s`

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

Notes:
- Re-ran the plan-level verification command for Task 3 and all normalization and same-domain tests passed without code changes.
- Confirmed the skip-reason taxonomy remains centralized in `internal/crawl/skip_reasons.go`, with no duplicated ad hoc reason literals elsewhere in the repository.
- Created `.planning/phases/01-crawl-ingestion-foundation/01-01-SUMMARY.md` and advanced `.planning/STATE.md` to Plan `01-02` / Task `1` as the next execution target.
- Verification run output:
  - `go test ./internal/crawl -v` -> `ok  	github.com/Nickbohm555/skill-cli/internal/crawl	0.514s`
  - `rg -n 'off_domain|already_seen|cap_reached|non_html_content_type|invalid_url|low_signal_page|fetch_error' .` -> matches only in `internal/crawl/skip_reasons.go`

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

Notes:
- Added `internal/crawl/docs_root.go` with a deterministic `DeriveDocsRoot` helper that normalizes the entry URL first, then applies explicit path-segment precedence: `docs`, `documentation`, `guide`, `guides`, with fallback to site root.
- Added `internal/crawl/docs_root_test.go` with table-driven fixtures for top-level docs, nested documentation roots, guide roots, precedence behavior, site-root fallback, and invalid URL handling.
- No blockers came up; the existing normalization helpers were reused directly so the docs-root policy stays aligned with the canonical URL boundary.
- Verification run output:
  - `go test ./internal/crawl -run DocsRoot -v` -> `ok  	github.com/Nickbohm555/skill-cli/internal/crawl	0.486s`
