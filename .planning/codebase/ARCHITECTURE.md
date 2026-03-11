# Architecture

**Analysis Date:** 2026-03-10

## Pattern Overview

**Overall:** Layered CLI pipeline with package-level domain modules

**Key Characteristics:**
- CLI entrypoint and command routing are isolated in `cmd/cli-skill/main.go` and `internal/cli/command/`.
- Domain logic is split by phase-aligned packages in `internal/crawl/`, `internal/content/`, `internal/refinement/`, `internal/validation/`, `internal/overlap/`, and `internal/install/`.
- Orchestration stages for downstream generation/install gates are centralized in `internal/app/generate/` as composable stage structs and loop controllers.

## Layers

**CLI Entry Layer:**
- Purpose: Parse commands/flags, run command handlers, and render user-facing output.
- Location: `cmd/cli-skill/main.go`, `internal/cli/command/crawl.go`, `internal/cli/command/process.go`, `internal/cli/command/refine.go`
- Contains: Cobra root/subcommands, command flag parsing, stdout/stderr report rendering, refinement REPL loop wiring.
- Depends on: `internal/crawl/`, `internal/content/`, `internal/refinement/`, `internal/cli/prompts/`, `github.com/spf13/cobra`.
- Used by: End users invoking the `cli-skill` binary.

**Prompt/Presentation Layer:**
- Purpose: Adapt refinement field definitions into interactive prompt plans and review output text.
- Location: `internal/cli/prompts/refinement_form.go`, `internal/cli/prompts/review_renderer.go`
- Contains: Prompt plan adapters, control metadata (`input` vs `select`), review rendering helpers.
- Depends on: `internal/refinement/`.
- Used by: `internal/cli/command/refine.go` through `prompts.RefinementFormAdapter`.

**Crawl Domain Layer:**
- Purpose: Execute bounded same-domain URL discovery and classify processed/skipped pages with reasons.
- Location: `internal/crawl/engine.go`, `internal/crawl/classify.go`, `internal/crawl/normalize.go`, `internal/crawl/docs_root.go`, `internal/crawl/types.go`
- Contains: Crawl queue/state machine, canonicalization, docs-root derivation, skip taxonomy, summary counters.
- Depends on: `github.com/gocolly/colly/v2`, Go `net/url`, `net/http`.
- Used by: `internal/cli/command/crawl.go`, `internal/cli/command/process.go`.

**Content Processing Layer:**
- Purpose: Transform crawled pages into normalized content, attributed chunks, and summary-first review projections.
- Location: `internal/content/extract.go`, `internal/content/normalize.go`, `internal/content/dedupe.go`, `internal/content/chunk.go`, `internal/content/pipeline.go`, `internal/content/summarize.go`, `internal/content/review_view.go`
- Contains: Readability extraction, markdown/plaintext normalization, conservative dedupe, chunk attribution, summarization with deterministic fallback, expansion mapping.
- Depends on: `codeberg.org/readeck/go-readability/v2`, `github.com/JohannesKaufmann/html-to-markdown/v2`, `github.com/openai/openai-go/v3`, `github.com/pkoukk/tiktoken-go`.
- Used by: `internal/cli/command/process.go`.

**Refinement State Machine Layer:**
- Purpose: Maintain required refinement fields, run adaptive question/deepening flow, and enforce commit readiness.
- Location: `internal/refinement/session.go`, `internal/refinement/flow.go`, `internal/refinement/clarity.go`, `internal/refinement/validator.go`, `internal/refinement/field_graph.go`, `internal/refinement/revise.go`
- Contains: Session field registry/state, flow events, deepening policy, revise dependency graph, validation report for review gate.
- Depends on: Go stdlib only.
- Used by: `internal/cli/command/refine.go`, `internal/cli/prompts/`.

**Validation Layer:**
- Purpose: Parse skill markdown and enforce structural + semantic quality rules.
- Location: `internal/validation/parse_skill.go`, `internal/validation/schema_validate.go`, `internal/validation/semantic_validate.go`, `internal/validation/report.go`, `internal/validation/model.go`, `internal/validation/skill.schema.json`
- Contains: Candidate model binding, JSON schema compiler/mapper, semantic scope rules, issue prioritization, follow-up prompt text.
- Depends on: `github.com/santhosh-tekuri/jsonschema/v6`, `github.com/yuin/goldmark`, `go.abhg.dev/goldmark/frontmatter`.
- Used by: `internal/app/generate/fix_loop.go`, `internal/install/transaction.go`, `internal/install/activate_verify.go`, `internal/overlap/index_installed.go`.

**Overlap & Install Layer:**
- Purpose: Compare candidate skill against installed inventory, collect explicit resolution decision, and perform approval-gated atomic install.
- Location: `internal/overlap/detect.go`, `internal/overlap/decision_flow.go`, `internal/overlap/index_installed.go`, `internal/install/preflight_gates.go`, `internal/install/approval_prompt.go`, `internal/install/transaction.go`, `internal/install/activate_verify.go`
- Contains: Installed skill indexing, weighted overlap classification, decision prompt flow, preflight blockers, preview/diff rendering, staged rename transaction, post-install verification.
- Depends on: `charm.land/huh/v2`, filesystem APIs, `internal/validation/`.
- Used by: `internal/app/generate/overlap_stage.go`, `internal/app/generate/install_stage.go`.

**Application Orchestration Layer:**
- Purpose: Compose validation/overlap/install gates into reusable stage APIs for full generation pipeline control.
- Location: `internal/app/generate/gate.go`, `internal/app/generate/fix_loop.go`, `internal/app/generate/overlap_stage.go`, `internal/app/generate/install_stage.go`
- Contains: Gate decisions, iterative fix loop, overlap handoff shaping, install stage dependency injection points.
- Depends on: `internal/validation/`, `internal/overlap/`, `internal/install/`.
- Used by: Package tests currently (`internal/app/generate/*_test.go`); not yet wired into `internal/cli/command/`.

## Data Flow

**Crawl Command Flow (`crawl`):**

1. `cmd/cli-skill/main.go` builds root command via `internal/cli/command.NewRootCommand()`.
2. `internal/cli/command/crawl.go` validates `--url`, calls `crawl.ExecuteCrawl(entryURL)`.
3. `internal/crawl/engine.go` canonicalizes URLs, enforces same-domain + cap, classifies responses, and records `CrawlResult`.
4. `internal/cli/command/crawl.go` renders processed/skipped pages and summary counts.

**Process Command Flow (`process`):**

1. `internal/cli/command/process.go` runs `crawl.ExecuteCrawl`, then fetches processed page HTML.
2. `internal/content/extract.go` + `internal/content/normalize.go` convert HTML into normalized page content.
3. `internal/content/dedupe.go` marks duplicates; `internal/content/pipeline.go` produces attributed chunks.
4. `internal/content/summarize.go` summarizes chunks via OpenAI provider or deterministic fallback.
5. `internal/content/review_view.go` builds summary-first review records + expansion map.
6. `internal/cli/command/process.go` prints review blocks and warnings.

**Refine Command Flow (`refine`):**

1. `internal/cli/command/refine.go` creates `refinement.SessionState` and `refinement.Flow`.
2. `internal/refinement/flow.go` asks primaries/deepening through the command-layer console interface.
3. `internal/cli/prompts/refinement_form.go` produces prompt plans from field state.
4. `internal/refinement/validator.go` returns readiness report; flow enters review when all required fields are complete.
5. `internal/cli/command/refine.go` handles `revise <field>` and `commit`, then emits JSON payload.

**Validation to Install Pipeline Flow (orchestration package):**

1. `internal/app/generate/fix_loop.go` repeatedly validates candidate skill (`ValidateStructural` + `ValidateSemantic`) and applies one blocking fix at a time.
2. `internal/app/generate/overlap_stage.go` runs `overlap.Detect` and `overlap.DecisionFlow.Decide`, then emits install handoff only when explicitly resolved.
3. `internal/app/generate/install_stage.go` runs `install.Preflight`, preview/diff rendering, approval capture, staged transaction, and activation verification.
4. `internal/install/transaction.go` writes `SKILL.md` to stage dir, validates parseability, and atomically activates via rename.

**State Management:**
- Crawl state is mutable and internal to `internal/crawl/engine.go` (`crawlState` with queue/seen/finalized maps).
- Refinement state is explicit in `internal/refinement/session.go` (`SessionState` with revision counter and per-field status).
- Validation/overlap/install states are immutable-ish value objects passed between stages (`validation.ValidationReport`, `overlap.OverlapReport`, `install.InstallRequest`).

## Key Abstractions

**Command Tree:**
- Purpose: Provide binary contract and subcommand entrypoints.
- Examples: `internal/cli/command/crawl.go`, `internal/cli/command/process.go`, `internal/cli/command/refine.go`
- Pattern: Cobra command factory functions with local flag vars and `RunE`.

**Candidate Skill Model:**
- Purpose: Canonical structured representation used by validation, overlap indexing, and install rendering.
- Examples: `internal/validation/model.go`, `internal/install/model.go`, `internal/overlap/index_installed.go`
- Pattern: Shared struct-based DTOs with normalization and path/rule annotations.

**Stage/Gate Decision Objects:**
- Purpose: Make progression blockers explicit instead of implicit boolean returns.
- Examples: `internal/app/generate/gate.go`, `internal/app/generate/overlap_stage.go`, `internal/install/preflight_gates.go`
- Pattern: `Allowed + Reason + BlockingIssue` structs propagated to caller.

**Provider/Function Injection Seams:**
- Purpose: Allow deterministic tests without live API/IO for major flows.
- Examples: `internal/content/summarize.go` (`SummaryProvider`), `internal/app/generate/install_stage.go` function fields, `internal/install/transaction.go` filesystem interface
- Pattern: Struct fields for replaceable behavior with defaults applied at runtime.

## Entry Points

**Binary Main:**
- Location: `cmd/cli-skill/main.go`
- Triggers: User runs `cli-skill ...`.
- Responsibilities: Construct root command, route stderr, execute command tree, set exit code on failure.

**Root Command Factory:**
- Location: `internal/cli/command/crawl.go` (`NewRootCommand`)
- Triggers: Called by `main`.
- Responsibilities: Register `crawl`, `process`, and `refine` subcommands.

**Pipeline Stages (Library Entry Points):**
- Location: `internal/app/generate/fix_loop.go`, `internal/app/generate/overlap_stage.go`, `internal/app/generate/install_stage.go`
- Triggers: Called by tests and future generate command wiring.
- Responsibilities: Drive gating progression from validation through install activation.

## Error Handling

**Strategy:** Fail closed with wrapped context errors and explicit gate blockers.

**Patterns:**
- Command handlers return wrapped errors (`fmt.Errorf("...: %w", err)`) and `main` owns process exit in `cmd/cli-skill/main.go`.
- Domain packages return typed blocker states plus sentinel errors where needed (`internal/install/errors.go`, `internal/app/generate/fix_loop.go`).
- Partial failures that should not abort are accumulated as warnings (`internal/cli/command/process.go`, `internal/overlap/index_installed.go`).

## Cross-Cutting Concerns

**Logging:** No centralized logger; user-visible status is emitted directly via command/report rendering in `internal/cli/command/` and prompt renderers in `internal/cli/prompts/`.

**Validation:** Structural + semantic validation is centralized in `internal/validation/` and reused by fix loop plus install verification.

**Authentication:** External auth is limited to OpenAI API key discovery in `internal/content/summarize.go` (`OPENAI_API_KEY`); no internal user/session auth layer.

---

*Architecture analysis: 2026-03-10*
