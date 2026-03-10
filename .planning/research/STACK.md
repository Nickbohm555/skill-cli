# Stack Research

**Domain:** Go CLI tool that generates Codex skills from one docs URL
**Researched:** 2026-03-10
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.25.x | Primary language/runtime | Current stable Go line for 2025+ CLIs; strong stdlib for filesystem/path/process ops needed for `$CODEX_HOME/skills` install gates. **Confidence: HIGH** |
| `github.com/spf13/cobra` | v1.10.2 | CLI command model and UX | De-facto Go CLI standard for subcommands, help/completion, and predictable UX. **Confidence: HIGH** |
| `charm.land/huh/v2` | v2.0.3 | Adaptive interactive questioning flow | Modern maintained prompt/form system; strong fit for iterative ask/validate/re-ask loops. **Confidence: HIGH** |
| `github.com/openai/openai-go/v3` | v3.26.0 | LLM-powered draft generation + targeted follow-up prompts | Official Go SDK; fastest path to robust generation + constrained retries from validation feedback. **Confidence: HIGH** |
| `github.com/spf13/viper` | v1.21.0 | Config and env management | Standard pairing with Cobra for API keys, model defaults, and runtime behavior toggles. **Confidence: HIGH** |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/PuerkitoBio/goquery` | v1.11.0 | Parse fetched HTML and extract canonical doc text | Use for single-page URL ingestion in v1; deterministic extraction and simpler failure handling than crawlers. **Confidence: HIGH** |
| `github.com/hashicorp/go-retryablehttp` | v0.7.8 | Retry/backoff around URL fetch and API calls | Use for flaky docs hosts and transient API/network failures. **Confidence: HIGH** |
| `github.com/santhosh-tekuri/jsonschema/v6` | v6.0.2 | Strict schema validation of generated skill structure | Use as hard gate before approval/install; fail fast with structured errors for targeted re-asks. **Confidence: HIGH** |
| `github.com/go-playground/validator/v10` | v10.30.1 | Semantic/field validation beyond schema | Use for quality checks schema cannot express well (scope narrowness, boundary rules, required examples). **Confidence: HIGH** |
| `github.com/BurntSushi/toml` | v1.6.0 | Parse skill metadata and existing skill manifests | Use when scanning installed skills and validating template metadata blocks. **Confidence: HIGH** |
| `github.com/sahilm/fuzzy` | v0.1.1 | Fast candidate overlap detection by name/keywords | Use as first-pass overlap candidate filter before deeper similarity checks. **Confidence: MEDIUM** |
| `dario.cat/mergo` | v1.0.2 | Merge/update behavior when overlap is approved | Use for controlled map/struct merge semantics in update paths. **Confidence: HIGH** |
| `github.com/spf13/afero` | v1.15.0 | Filesystem abstraction for install + tests | Use to unit test install/approval paths without touching real `$CODEX_HOME`. **Confidence: HIGH** |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| `golangci-lint` | Lint and static checks | Gate PRs; enable `gosec`, `errcheck`, `staticcheck`, `govet` for CLI reliability. |
| `goreleaser` | Cross-platform binary packaging | Produce reproducible single-binary releases for macOS/Linux/Windows. |
| `gotestsum` + `testing` | Test execution/reporting | Keep approval-gate, validation-retry, and install behavior under deterministic tests. |

## Installation

```bash
# Core
go get github.com/spf13/cobra@v1.10.2
go get charm.land/huh/v2@v2.0.3
go get github.com/openai/openai-go/v3@v3.26.0
go get github.com/spf13/viper@v1.21.0

# Supporting
go get github.com/PuerkitoBio/goquery@v1.11.0
go get github.com/hashicorp/go-retryablehttp@v0.7.8
go get github.com/santhosh-tekuri/jsonschema/v6@v6.0.2
go get github.com/go-playground/validator/v10@v10.30.1
go get github.com/BurntSushi/toml@v1.6.0
go get github.com/sahilm/fuzzy@v0.1.1
go get dario.cat/mergo@v1.0.2
go get github.com/spf13/afero@v1.15.0

# Dev dependencies / tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/goreleaser/goreleaser/v2@latest
go install gotest.tools/gotestsum@latest
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| `cobra` | `urfave/cli` (`v3.7.0` latest) | Use `urfave/cli` only if you want a flatter command model and minimal framework conventions. |
| `huh/v2` | Bubble Tea custom prompts only | Use pure Bubble Tea if you need full-screen bespoke TUI control beyond form workflows. |
| `jsonschema/v6` + `validator/v10` | Hand-rolled validation only | Use hand-rolled only for very small prototypes; not for strict block-and-re-ask production gates. |
| `goquery` single-page extraction | Full crawler stacks | Use crawler stacks only in post-v1 when multi-page/domain ingestion becomes a requirement. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `github.com/AlecAivazis/survey` | Archived/read-only; weak long-term maintenance signal for core UX dependency. | `charm.land/huh/v2` |
| Unvalidated direct template writes | Allows malformed skills to pass through and fail only at runtime in Codex. | `jsonschema/v6` + semantic validators + block-and-re-ask loop |
| Auto-install without explicit confirmation | Violates approval gate requirement and increases risk of unwanted local skill mutations. | Explicit yes/no install gate before writes |
| Multi-page crawling in v1 | Expands ambiguity and failure modes; conflicts with single-URL deterministic v1 scope. | Single-page URL fetch + extraction |

## Stack Patterns by Variant

**If running API-backed adaptive generation (recommended v1):**
- Use `openai-go/v3` + strict schema validation + targeted re-ask prompts.
- Because this delivers the highest success rate for "usable in Codex" output with explicit correction loops.

**If running deterministic/offline mode later:**
- Keep same CLI/validation/install stack, swap generation engine for local rules or local model adapter.
- Because architecture remains stable while generation backend changes independently.

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| `go@1.25.x` | `goquery@v1.11.0` | `goquery@v1.11.0` requires Go 1.24+, so Go 1.25 is safe. |
| `go@1.25.x` | `validator/v10@v10.30.1` | Published metadata shows Go 1.24.0 requirement; Go 1.25 is safe. |
| `huh/v2@v2.0.3` | `bubbletea@v2.0.2` + `lipgloss@v2.0.1` | Keep Charm ecosystem major versions aligned on v2 to reduce UI integration friction. |
| `openai-go/v3@v3.26.0` | `go@1.25.x` | SDK requires Go 1.22+, so Go 1.25 is safe. |
| `jsonschema/v6@v6.0.2` | `go@1.25.x` | Module metadata lists Go 1.21 requirement; Go 1.25 is safe. |

## Sources

- [Go release history](https://go.dev/doc/devel/release) - verified Go 1.25 line and patch cadence. (HIGH)
- [Cobra releases](https://github.com/spf13/cobra/releases) - verified `v1.10.2` latest. (HIGH)
- [Viper releases](https://github.com/spf13/viper/releases) - verified `v1.21.0` latest. (HIGH)
- [Huh releases](https://github.com/charmbracelet/huh/releases) - verified `v2.0.3` latest. (HIGH)
- [Bubble Tea releases](https://github.com/charmbracelet/bubbletea/releases) - verified `v2.0.2` latest. (HIGH)
- [Lip Gloss releases](https://github.com/charmbracelet/lipgloss/releases) - verified `v2.0.1` latest. (HIGH)
- [OpenAI Go releases](https://github.com/openai/openai-go/releases) and [OpenAI Go README](https://raw.githubusercontent.com/openai/openai-go/main/README.md) - verified `v3.26.0` and Go version requirement. (HIGH)
- [goquery releases](https://github.com/PuerkitoBio/goquery/releases) - verified `v1.11.0` and Go 1.24+ requirement. (HIGH)
- [go-retryablehttp on pkg.go.dev](https://pkg.go.dev/github.com/hashicorp/go-retryablehttp?tab=versions) - verified `v0.7.8` default release metadata. (HIGH)
- [jsonschema releases](https://github.com/santhosh-tekuri/jsonschema/releases) and [jsonschema v6 pkg page](https://pkg.go.dev/github.com/santhosh-tekuri/jsonschema/v6) - verified `v6.0.2` and Go requirement. (HIGH)
- [validator releases](https://github.com/go-playground/validator/releases) and [validator pkg page](https://pkg.go.dev/github.com/go-playground/validator/v10) - verified `v10.30.1` and Go requirement metadata. (HIGH)
- [BurntSushi/toml releases](https://github.com/BurntSushi/toml/releases) - verified `v1.6.0` latest. (HIGH)
- [sahilm/fuzzy releases](https://github.com/sahilm/fuzzy/releases) - verified `v0.1.1` latest (older cadence). (MEDIUM)
- [mergo releases](https://github.com/darccio/mergo/releases) - verified `v1.0.2` latest. (HIGH)
- [afero releases](https://github.com/spf13/afero/releases) - verified `v1.15.0` latest. (HIGH)
- [urfave/cli releases](https://github.com/urfave/cli/releases) - verified `v3.7.0` latest for alternative comparison. (HIGH)
- Web search corroboration for archived `survey` status. (MEDIUM)

---
*Stack research for: Skill Weaver (CLI skill generation)*
*Researched: 2026-03-10*
