# External Integrations

**Analysis Date:** 2026-03-10

## APIs & External Services

**LLM Provider:**
- OpenAI Responses API - summarizing attributed chunks during content processing.
  - SDK/Client: `github.com/openai/openai-go/v3` in `internal/content/summarize.go`.
  - Auth: `OPENAI_API_KEY` read in `internal/content/summarize.go`.

**Web Documentation Fetching:**
- Arbitrary user-provided docs hosts over HTTP/HTTPS - crawl and fetch source pages from `--url`.
  - SDK/Client: `github.com/gocolly/colly/v2` in `internal/crawl/engine.go` and stdlib `net/http` in `internal/cli/command/process.go`.
  - Auth: Not detected (unauthenticated fetch behavior in `internal/cli/command/process.go` and `internal/crawl/engine.go`).

## Data Storage

**Databases:**
- Not detected.
  - Connection: Not applicable.
  - Client: Not applicable.

**File Storage:**
- Local filesystem only; generated/installed skills are written as `SKILL.md` under `CODEX_HOME/skills` (or `~/.codex/skills`) in `internal/install/transaction.go` and `internal/overlap/index_installed.go`.

**Caching:**
- None detected (no Redis/memcached/cache client dependencies in `go.mod`; no cache integration in `internal/**/*.go`).

## Authentication & Identity

**Auth Provider:**
- Custom environment-key auth for OpenAI only.
  - Implementation: API key from `OPENAI_API_KEY` gates provider selection in `internal/content/summarize.go`; no user/session auth layer is present in `cmd/cli-skill/main.go` or `internal/cli/command/*.go`.

## Monitoring & Observability

**Error Tracking:**
- None detected (no Sentry/Datadog/Bugsnag SDKs in `go.mod`).

**Logs:**
- CLI stdout/stderr rendering and error propagation only via `fmt.Fprintf` and command `RunE` returns in `cmd/cli-skill/main.go`, `internal/cli/command/crawl.go`, and `internal/cli/command/process.go`.

## CI/CD & Deployment

**Hosting:**
- Not applicable; distributed as a local CLI binary from `cmd/cli-skill/main.go`.

**CI Pipeline:**
- Not detected (no `.github/workflows/*.yml` in repo root and no other CI manifest detected).

## Environment Configuration

**Required env vars:**
- `CODEX_HOME` - optional override for installed skill root in `internal/overlap/index_installed.go`.
- `OPENAI_API_KEY` - required only for OpenAI summarization path in `internal/content/summarize.go`; fallback mode works without it.

**Secrets location:**
- Process environment variables only (`os.Getenv` usage in `internal/content/summarize.go` and `internal/overlap/index_installed.go`); no committed secret file integration detected.

## Webhooks & Callbacks

**Incoming:**
- None (no HTTP server endpoints or webhook handlers detected in `cmd/` or `internal/`).

**Outgoing:**
- HTTP GET crawl/fetch requests to target docs URLs in `internal/crawl/engine.go` and `internal/cli/command/process.go`.
- OpenAI API calls for summarization when configured in `internal/content/summarize.go`.

---

*Integration audit: 2026-03-10*
