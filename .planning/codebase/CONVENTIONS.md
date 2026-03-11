# Coding Conventions

**Analysis Date:** 2026-03-10

## Naming Patterns

**Files:**
- Use `snake_case.go` for source and `snake_case_test.go` for tests in the same package directory, for example `internal/crawl/engine.go` and `internal/crawl/engine_test.go`.
- Keep package names short and domain-aligned (`crawl`, `content`, `refinement`, `validation`, `overlap`, `install`) under `internal/`.

**Functions:**
- Use `PascalCase` for exported constructors/APIs such as `NewRootCommand` in `internal/cli/command/crawl.go`, `ExecuteCrawlWithOptions` in `internal/crawl/engine.go`, and `InstallTransaction` in `internal/install/transaction.go`.
- Use `camelCase` for package-private helpers such as `renderProcessReport` and `excerptRunes` in `internal/cli/command/process.go`.
- Prefix constructor-like functions with `New` and guard required dependencies early, as shown in `NewFlow` in `internal/refinement/flow.go`.

**Variables:**
- Use descriptive `camelCase` names (`entryURL`, `crawlResult`, `normalizedPages`, `deepeningAttempts`) rather than abbreviations in `internal/cli/command/process.go` and `internal/refinement/flow.go`.
- Keep receiver names short single letters (`s`, `f`, `e`) for methods in `internal/crawl/engine.go`, `internal/refinement/flow.go`, and `internal/install/transaction.go`.

**Types:**
- Use `PascalCase` for exported structs/enums (`FlowState`, `InstallError`, `OverlapFinding`) in `internal/refinement/flow.go`, `internal/install/errors.go`, and `internal/overlap/model.go`.
- Model enum-like values as typed strings with explicit const sets (`FlowEventType`, `ResolutionMode`, `ErrorCode`) in `internal/refinement/flow.go`, `internal/overlap/model.go`, and `internal/install/errors.go`.

## Code Style

**Formatting:**
- Use Go standard formatting (`go fmt ./...`) as the source of truth (documented in `AGENTS.md` and `README.md`).
- Keep line wrapping pragmatic; long `fmt.Errorf` and struct literals are split vertically, as in `internal/install/transaction.go`.

**Linting:**
- Use `go vet ./...` as the baseline static check (documented in `AGENTS.md` and `README.md`).
- No repository lint config (`.golangci*`, `.editorconfig`, `.eslintrc*`, `.prettierrc*`) is detected at repo root; treat idiomatic Go + vet-clean code as the enforced standard.

## Import Organization

**Order:**
1. Standard library imports first (`fmt`, `strings`, `net/http`) as in `internal/cli/command/process.go`.
2. Internal module imports (`github.com/Nickbohm555/skill-cli/internal/...`) next, as in `internal/cli/command/crawl.go`.
3. External third-party imports last (`github.com/spf13/cobra`, `github.com/yuin/goldmark`) as in `internal/cli/command/crawl.go` and `internal/validation/parse_skill.go`.

**Path Aliases:**
- Use fully qualified import paths; no alias/path-shortcut scheme is used.
- One explicit alias pattern is used only for collision/readability (`gmtext` in `internal/validation/parse_skill.go`).

## Error Handling

**Patterns:**
- Return errors (not panics) with `%w` wrapping and operation context, for example `fmt.Errorf("process failed: %w", err)` in `internal/cli/command/process.go` and `fmt.Errorf("create stage dir: %w", err)` in `internal/install/transaction.go`.
- Validate inputs early and fail fast (`missing required --url value`, nil dependency checks) in `internal/cli/command/crawl.go` and `internal/refinement/flow.go`.
- Use sentinel/custom typed errors for policy gates and classification (`ErrInstallApprovalRequired`, `InstallError.Is`, `ErrorCodeOf`) in `internal/install/errors.go`.
- Use `errors.Is` and `errors.As` compatibility patterns, not string-only matching, in `internal/install/errors.go` and `internal/content/types.go`.

## Logging

**Framework:** None (structured logger not detected).

**Patterns:**
- Use explicit CLI output rendering (`fmt.Fprintf`) to provided writers (`cmd.OutOrStdout`, `cmd.ErrOrStderr`) instead of logger side effects in `internal/cli/command/crawl.go` and `internal/cli/command/process.go`.
- Keep domain internals pure and return typed data/errors; presentation is concentrated in command/prompt layers (`internal/cli/command/*`, `internal/cli/prompts/*`).

## Comments

**When to Comment:**
- Add doc comments to exported constructors/types when behavior is non-trivial, for example `NewRootCommand` in `internal/cli/command/crawl.go` and type comments in `internal/content/types.go`.
- Prefer self-documenting names for private helpers; short private functions typically omit comments.

**JSDoc/TSDoc:**
- Not applicable for this Go codebase.
- Go doc comments are used where needed; there is no strict “every export must be commented” gate detected.

## Function Design

**Size:** Prefer medium-sized functions with explicit stages and helper extraction (`runProcess`, `fetchProcessedPages`, `renderProcessReport`) in `internal/cli/command/process.go`.

**Parameters:** Pass concrete domain structs/interfaces over unstructured maps (`InstallRequest`, `FlowState`, `transactionFS`) in `internal/install/transaction.go` and `internal/refinement/flow.go`.

**Return Values:**
- Return `(value, error)` consistently; zero-value structs on failure are standard (`processReport{}, err`, `FlowResult{}, err`).
- Use pure helper returns for derived values (`displayURL`, `ErrorCodeOf`, `slugHeading`) in `internal/cli/command/crawl.go`, `internal/install/errors.go`, and `internal/validation/parse_skill.go`.

## Module Design

**Exports:**
- Keep package APIs narrow and explicit via exported constructors/functions plus internal helpers in the same file, as in `internal/crawl/engine.go` and `internal/install/transaction.go`.
- Keep package-local seams via unexported interfaces for testability (`transactionFS`) in `internal/install/transaction.go`.

**Barrel Files:**
- Not used; Go package exports are consumed directly by package path (for example `internal/refinement`, `internal/content`, `internal/install`).

---

*Convention analysis: 2026-03-10*
