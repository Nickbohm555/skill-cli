# Technology Stack

**Analysis Date:** 2026-03-10

## Languages

**Primary:**
- Go 1.25.8 - CLI runtime and all production code in `cmd/cli-skill/main.go` and `internal/**/*.go` (declared in `go.mod`).

**Secondary:**
- Markdown (version not applicable) - skill payload format and docs rendered/parsed by `internal/install/transaction.go`, `internal/validation/parse_skill.go`, and `README.md`.

## Runtime

**Environment:**
- Go toolchain 1.25.x with module mode (`go 1.25.8` in `go.mod`).

**Package Manager:**
- Go Modules (managed by `go.mod` and `go.sum`).
- Lockfile: present (`go.sum`).

## Frameworks

**Core:**
- `github.com/spf13/cobra` v1.10.2 - CLI command tree and argument parsing in `internal/cli/command/crawl.go`, `internal/cli/command/process.go`, and `internal/cli/command/refine.go`.
- `github.com/gocolly/colly/v2` v2.3.0 - bounded same-domain crawl engine in `internal/crawl/engine.go`.

**Testing:**
- Go standard `testing` package (version tied to Go 1.25.x) - unit/integration tests in `internal/**/*_test.go`.

**Build/Dev:**
- Go compiler/tooling (`go build`, `go test`, `go vet`, `go fmt`) referenced in `README.md` and driven by module metadata in `go.mod`.

## Key Dependencies

**Critical:**
- `github.com/openai/openai-go/v3` v3.26.0 - structured chunk summarization provider in `internal/content/summarize.go`.
- `github.com/gocolly/colly/v2` v2.3.0 - crawl orchestration and HTML link discovery in `internal/crawl/engine.go`.
- `github.com/spf13/cobra` v1.10.2 - top-level CLI command surface in `cmd/cli-skill/main.go` and `internal/cli/command/crawl.go`.

**Infrastructure:**
- `github.com/JohannesKaufmann/html-to-markdown/v2` v2.5.0 and `golang.org/x/net/html` v0.47.0 - content normalization/parsing in `internal/content/normalize.go` and `internal/content/extract.go`.
- `github.com/pkoukk/tiktoken-go` v0.1.8 and `github.com/tmc/langchaingo` v0.1.14 - chunk sizing/splitting in `internal/content/chunk.go`.
- `github.com/yuin/goldmark` v1.7.16 and `go.abhg.dev/goldmark/frontmatter` v0.3.0 - SKILL.md parsing and validation in `internal/validation/parse_skill.go`.
- `charm.land/huh/v2` v2.0.3 - interactive prompt dependencies for refinement flows (declared in `go.mod`, integrated through prompt abstractions in `internal/cli/prompts/*.go`).

## Configuration

**Environment:**
- Set `OPENAI_API_KEY` to enable OpenAI-backed summaries; without it, deterministic local fallback summaries are used in `internal/content/summarize.go`.
- Set `CODEX_HOME` to control installed skill location; default fallback is `~/.codex/skills` in `internal/overlap/index_installed.go`.

**Build:**
- Module and dependency config: `go.mod`, `go.sum`.
- CLI entrypoint compile target: `cmd/cli-skill/main.go`.
- Planning/runtime metadata used by GSD workflow: `.planning/config.json`.

## Platform Requirements

**Development:**
- Local machine with Go 1.25.x and filesystem write access to workspace + target skill directory (`README.md`, `internal/install/transaction.go`).
- Network egress to docs URLs for crawl/fetch via `internal/crawl/engine.go` and `internal/cli/command/process.go`.

**Production:**
- Not detected as a hosted service; deployment target is a local CLI binary (`go install ./cmd/cli-skill` in `README.md`) executed in user environments.

---

*Stack analysis: 2026-03-10*
