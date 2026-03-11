# Codebase Concerns

**Analysis Date:** 2026-03-10

## Tech Debt

**Duplicate target-path resolution logic across install flow:**
- Issue: Install target path derivation is duplicated in multiple layers with parallel logic (`filepath.Join(rootDir, target.SkillID)`), which increases drift risk when path policy changes.
- Files: `internal/install/transaction.go`, `internal/install/preview_diff.go`, `internal/app/generate/install_stage.go`
- Impact: Inconsistent install location, preview/report mismatch, and harder safe-path policy rollout.
- Fix approach: Centralize target resolution in one package-level function and reuse it from transaction, preview, and stage orchestration.

**Stringly-typed interactive command protocol in refinement loop:**
- Issue: Review commands are parsed with ad-hoc string checks and regex/token matching instead of a typed command model.
- Files: `internal/cli/command/refine.go`, `internal/refinement/revise.go`
- Impact: Behavioral regressions are likely when adding commands or aliases; user-facing errors are easier to introduce.
- Fix approach: Introduce a typed command parser/dispatcher with explicit command structs and unit tests per command variant.

## Known Bugs

**Potential install path escape via unsanitized skill ID:**
- Symptoms: A crafted `SkillID` containing path traversal segments can resolve outside the intended root.
- Files: `internal/install/transaction.go`, `internal/install/preview_diff.go`, `internal/app/generate/install_stage.go`
- Trigger: Provide `InstallTarget{RootDir: "...", SkillID: "../outside"}` and run install path resolution.
- Workaround: Only pass trusted slug-like IDs and reject any `SkillID` containing separators or traversal segments before install.

## Security Considerations

**Filesystem write boundary is not enforced during install path resolution:**
- Risk: Path traversal can redirect writes/renames to unintended filesystem locations.
- Files: `internal/install/transaction.go`, `internal/install/model.go`
- Current mitigation: Preflight/approval gates exist in `internal/install/preflight_gates.go`, but they validate readiness/conflicts rather than path safety.
- Recommendations: Validate `SkillID` against a strict slug regex, normalize+clean target path, then assert target remains within canonicalized `RootDir`.

**Unbounded HTTP response body reads in processing pipeline:**
- Risk: Large or malicious pages can cause high memory usage or process instability.
- Files: `internal/cli/command/process.go`
- Current mitigation: Per-request timeout exists (`processFetchTimeout`), but no response-size cap.
- Recommendations: Replace direct `io.ReadAll` with `io.LimitReader` (or max-bytes guard), and skip/truncate oversized bodies with warning telemetry.

## Performance Bottlenecks

**Sequential fetch/normalize/chunk/summarize pipeline:**
- Problem: End-to-end processing is serialized per page/chunk and scales linearly with network/model latency.
- Files: `internal/cli/command/process.go`, `internal/content/pipeline.go`, `internal/content/summarize.go`
- Cause: Single-threaded loops in `fetchProcessedPages` and `SummarizeChunksWithConfig`.
- Improvement path: Add bounded worker pools for fetch/normalize/summarize stages with deterministic ordering at output boundary.

**Double network fetch for processed pages:**
- Problem: Crawl discovers pages, then process stage fetches them again, increasing latency and external load.
- Files: `internal/crawl/engine.go`, `internal/cli/command/process.go`
- Cause: `ExecuteCrawl` only records metadata; `fetchProcessedPages` performs a second GET pass.
- Improvement path: Carry response body/content-type from crawl stage (or optional cache layer) to avoid redundant downloads.

## Fragile Areas

**Refinement graph construction can panic on configuration errors:**
- Files: `internal/refinement/field_graph.go`
- Why fragile: `DefaultFieldGraph` panics on `NewFieldGraph` error, converting config mistakes into hard process termination.
- Safe modification: Keep registry/edge updates atomic and validated in tests before wiring into defaults.
- Test coverage: `internal/refinement/session_test.go` and `internal/refinement/flow_test.go` cover flow behavior, but no direct panic-safety regression test for `DefaultFieldGraph`.

**Install transaction rename/rollback sequence is failure-sensitive:**
- Files: `internal/install/transaction.go`, `internal/install/transaction_test.go`
- Why fragile: Multi-step rename/restore/cleanup path has partial-failure branches where backup/stage cleanup correctness is critical.
- Safe modification: Preserve current transaction ordering, add explicit invariants around backup lifecycle, and extend failure-injection tests per branch.
- Test coverage: Good branch coverage in `internal/install/transaction_test.go`, but no explicit path-boundary safety cases.

## Scaling Limits

**Crawl throughput is intentionally capped at small volume:**
- Current capacity: `DefaultProcessedPageCap = 50`.
- Limit: Documentation sets beyond 50 processable pages are skipped (`cap_reached`) and never summarized.
- Files: `internal/crawl/engine.go`, `internal/cli/command/crawl.go`
- Scaling path: Make cap configurable via CLI/config with sane defaults and expose skipped-by-cap metrics to users.

**Summary generation scales with chunk count and external API latency:**
- Current capacity: One provider call per chunk in sequence.
- Limit: Large docs can create high end-to-end latency and long interactive waits.
- Files: `internal/content/summarize.go`, `internal/cli/command/process.go`
- Scaling path: Add bounded concurrent summarization plus batch-level cancellation and progress reporting.

## Dependencies at Risk

**OpenAI Responses API dependence for preferred summarization path:**
- Risk: Provider behavior/schema shifts can degrade summary quality or force fallback mode.
- Impact: `SummarizeChunksWithConfig` can return deterministic fallback summaries with lower semantic quality.
- Files: `internal/content/summarize.go`
- Migration plan: Keep provider abstraction in `internal/content/summarize.go`, add alternate provider implementation and feature flag selection.

**Heavy indirect dependency surface increases upgrade volatility:**
- Risk: Numerous indirect modules increase chance of transitive breakage/security advisories.
- Impact: Build/test instability and slower dependency maintenance.
- Files: `go.mod`, `go.sum`
- Migration plan: Periodically prune unused dependencies, run targeted upgrade windows, and pin/verify transitive changes via CI.

## Missing Critical Features

**No explicit network retry/backoff strategy for processing fetches:**
- Problem: Transient network failures immediately degrade output completeness.
- Files: `internal/cli/command/process.go`
- Blocks: Reliable processing of unstable docs hosts in real-world conditions.

**No install path policy enforcement primitive:**
- Problem: Approval/preflight do not validate filesystem safety constraints.
- Files: `internal/install/transaction.go`, `internal/install/preflight_gates.go`, `internal/install/model.go`
- Blocks: Safe automation for untrusted/generated target identifiers.

## Test Coverage Gaps

**Processing command has no direct command-level tests:**
- What's not tested: `process` command orchestration, warning rendering behavior, and timeout/error plumbing.
- Files: `internal/cli/command/process.go`, `internal/cli/command/refine_test.go`
- Risk: CLI regressions can ship in output formatting, fetch error handling, or stage integration.
- Priority: High

**Path traversal/boundary tests are missing in install flow:**
- What's not tested: Rejection of malicious `SkillID` values and enforcement that resolved target stays inside `RootDir`.
- Files: `internal/install/transaction.go`, `internal/install/preview_diff.go`, `internal/install/transaction_test.go`, `internal/install/preview_diff_test.go`
- Risk: Filesystem writes outside intended install root can occur unnoticed.
- Priority: High

**Markdown parser edge-case coverage is limited for malformed/hostile inputs:**
- What's not tested: Deeply nested markdown, oversized frontmatter values, and ambiguous heading/list structures.
- Files: `internal/validation/parse_skill.go`, `internal/validation/validation_test.go`
- Risk: Parsing/validation drift can silently alter extracted candidate fields.
- Priority: Medium

---

*Concerns audit: 2026-03-10*
