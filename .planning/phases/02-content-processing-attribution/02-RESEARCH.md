# Phase 2: Content Processing & Attribution - Research

**Researched:** 2026-03-10
**Domain:** Go content normalization, chunk summarization, and source attribution pipeline
**Confidence:** HIGH

## Summary

This research focused on how to convert Phase 1 crawl outputs into clean, inspectable, attributable context for downstream generation, while honoring locked decisions: minimal chrome stripping, light cleanup, high-confidence dedupe, medium chunk size, 1-2 line gist summaries, summary-first review, and per-chunk attribution.

The standard implementation pattern is a three-stage transform: (1) extract main readable content while preserving important structure, (2) normalize into markdown/text that retains tables/code/media hints, and (3) split into bounded chunks with metadata-first provenance (`source_url`, headings, checksums, token counts). Summaries should be generated as schema-validated records so every chunk has a stable 1-2 line gist and attribution fields for UI review.

The strongest recommendation is to rely on maintained parser/converter/splitter libraries and keep custom logic only for policy decisions (boilerplate thresholds, dedupe confidence gates, and attribution display). This avoids fragile hand-rolled HTML/chunking code and keeps Phase 2 planning task-oriented.

**Primary recommendation:** Use `go-readability v2` + `html-to-markdown/v2` + `langchaingo/textsplitter` + schema-validated summary output, with chunk metadata as the source of truth for attribution.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `codeberg.org/readeck/go-readability/v2` | `v2.1.1` | Main-content extraction with light chrome removal | Maintained Readability.js-aligned extractor; upstream `go-shiori/go-readability` is deprecated in favor of this line. |
| `github.com/JohannesKaufmann/html-to-markdown/v2` | `v2.5.0` | HTML to markdown normalization with structural fidelity | Supports plugin-based conversion, table plugin, and custom renderers for selective tag handling. |
| `github.com/tmc/langchaingo` (`textsplitter`) | `v0.1.14` | Recursive/token chunking with overlap and heading options | Provides tested splitter primitives including token splitter and metadata-preserving split document flow. |
| `github.com/openai/openai-go/v3` | `v3.26.0` | Structured chunk gist generation (1-2 lines) | Official SDK supports schema-constrained structured output workflows. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/pkoukk/tiktoken-go` | `v0.1.8` | Token counting and tokenizer alignment | Use when controlling chunk size by token budget and validating prompt footprint. |
| `github.com/microcosm-cc/bluemonday` | `v1.0.27` | Sanitization of untrusted HTML when needed | Use if rendered/previewed HTML is shown; keep conversion pipeline safe from script/style payloads. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `go-readability/v2` | `github.com/PuerkitoBio/goquery` only extraction | `goquery` is excellent for DOM selection, but not a full readability/content scoring engine by itself. |
| `html-to-markdown/v2` | direct `.Text()` extraction from DOM | Faster but loses structure (tables/code/image context) needed by your phase decisions. |
| `langchaingo textsplitter` | custom splitter implementation | Custom splitter increases edge-case burden (boundaries, overlap, tables, headings, multibyte length). |

**Installation:**
```bash
go get codeberg.org/readeck/go-readability/v2
go get github.com/JohannesKaufmann/html-to-markdown/v2
go get github.com/tmc/langchaingo@v0.1.14
go get github.com/openai/openai-go/v3@v3.26.0
go get github.com/pkoukk/tiktoken-go
go get github.com/microcosm-cc/bluemonday
```

## Architecture Patterns

### Recommended Project Structure
```text
internal/content/
├── extract/            # readability extraction + minimal chrome policy
├── normalize/          # html->markdown, code/table/media preservation
├── dedupe/             # exact and high-confidence near-duplicate policy
├── chunk/              # splitter config + chunk metadata population
├── summarize/          # 1-2 line gist generation via structured output
└── present/            # summary-first view model + raw expansion references
```

### Pattern 1: Two-Layer Content Model
**What:** Keep both `normalized` and `review` representations, not one lossy blob.
**When to use:** Always in this phase, because users need summary-first + raw expansion.
**Example:**
```go
type NormalizedChunk struct {
    ID          string
    SourceURL   string
    PageTitle   string
    HeadingPath []string
    RawMarkdown string
    RawText     string
    TokenCount  int
    Checksum    string
}

type ReviewChunk struct {
    ChunkID      string
    SourceURL    string
    GistSummary  string // 1-2 lines
    PreviewText  string
    ExpandTarget string // key to full raw chunk
}
```

### Pattern 2: Structure-First Chunking, Token Guardrails Second
**What:** Split by semantic boundaries first (headings/paragraphs/tables/code blocks), then enforce medium token caps with overlap.
**When to use:** Default mode for docs/reference pages.
**Example:**
```go
// Source: https://raw.githubusercontent.com/tmc/langchaingo/main/textsplitter/token_splitter.go
splitter := textsplitter.NewTokenSplitter(
    textsplitter.WithChunkSize(500),
    textsplitter.WithChunkOverlap(80),
)
chunks, err := splitter.SplitText(markdownText)
```

### Pattern 3: Attribution as Chunk Metadata, Never UI-Derived
**What:** Store provenance directly on each chunk record (`source_url`, heading path, offsets/checksum), then render from metadata.
**When to use:** Always; required by CONT-03.
**Example:**
```go
// Source: https://raw.githubusercontent.com/tmc/langchaingo/main/schema/documents.go
doc := schema.Document{
    PageContent: chunkText,
    Metadata: map[string]any{
        "source_url": sourceURL,
        "page_title": pageTitle,
        "chunk_id":   chunkID,
        "checksum":   checksum,
    },
}
```

### Anti-Patterns to Avoid
- **Single-pass plaintext extraction:** Loses table/code/media context and weakens downstream quality.
- **Attribution only at page-level:** Fails per-chunk provenance requirement and makes review ambiguous.
- **Aggressive dedupe thresholding:** Drops distinct but similar sections; keep dedupe to high-confidence overlaps only.
- **Summary-only storage:** Prevents raw expansion and creates irreversible information loss.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Main-content extraction | Custom "remove nav/footer regex" pipeline | `go-readability/v2` | Readability heuristics are battle-tested and handle varied DOM patterns better. |
| HTML to markdown fidelity | Ad-hoc tag-by-tag converter | `html-to-markdown/v2` with plugins | Tables/code/link/image handling and extension hooks already exist. |
| Chunk splitting mechanics | Custom overlap/merge splitter | `langchaingo/textsplitter` | Handles recursive boundaries and token-based splitting with configurable overlap. |
| Structured gist outputs | Prompt-only freeform summaries | JSON-schema-constrained model output | Ensures each chunk summary is parseable and consistently shaped. |

**Key insight:** Hand-rolling extraction/conversion/splitting causes silent quality regressions and high maintenance burden; custom code should define policy thresholds, not core text-processing mechanics.

## Common Pitfalls

### Pitfall 1: Deprecated extractor dependency
**What goes wrong:** Team uses deprecated `github.com/go-shiori/go-readability` as primary.
**Why it happens:** Older package still appears in search results and examples.
**How to avoid:** Use `codeberg.org/readeck/go-readability/v2` as default extractor.
**Warning signs:** New extraction bugs fixed upstream but not available in your pinned package.

### Pitfall 2: Losing code/table/media fidelity during normalization
**What goes wrong:** Conversion drops table alignment, mangles code fences, or removes alt/caption context.
**Why it happens:** Plain text extraction shortcuts bypass structure-preserving conversion.
**How to avoid:** Normalize through markdown conversion with table support; preserve image/link context fields.
**Warning signs:** Chunk previews have flattened rows, missing snippet context, or image references vanish.

### Pitfall 3: Chunking by bytes/characters only
**What goes wrong:** Medium chunks become inconsistent by model token limits, causing prompt inefficiency.
**Why it happens:** Character-length chunking alone does not track tokenizer boundaries.
**How to avoid:** Use token-aware splitter or token count validation (`tiktoken-go`) after semantic split.
**Warning signs:** Frequent oversized prompt payloads and variable chunk utility.

### Pitfall 4: Dedupe overreach
**What goes wrong:** Similar but distinct sections are removed.
**Why it happens:** Broad near-duplicate rules applied without confidence gating.
**How to avoid:** Restrict dedupe to exact hash and very high-confidence normalized overlap only; log suppressions.
**Warning signs:** Missing sections users expected to review; unexplained chunk count drops.

### Pitfall 5: Attribution attached after summarization
**What goes wrong:** Summaries cannot be reliably traced back to source chunks.
**Why it happens:** Provenance is added in presentation layer instead of chunk creation.
**How to avoid:** Stamp source metadata at chunk creation and carry unchanged through summarization and UI.
**Warning signs:** Summary rows without stable `chunk_id`/`source_url`.

## Code Examples

Verified patterns from official sources:

### Main-content extraction
```go
// Source: https://pkg.go.dev/codeberg.org/readeck/go-readability/v2
article, err := readability.FromReader(srcReader, baseURL)
if err != nil { return err }
contentHTML := article.Content()
textContent := article.TextContent()
```

### HTML->Markdown conversion with extension points
```go
// Source: https://pkg.go.dev/github.com/JohannesKaufmann/html-to-markdown/v2
markdown, err := htmltomarkdown.ConvertString(inputHTML)
if err != nil { return err }
```

### Token chunking with overlap
```go
// Source: https://raw.githubusercontent.com/tmc/langchaingo/main/textsplitter/token_splitter.go
splitter := textsplitter.NewTokenSplitter(
    textsplitter.WithChunkSize(500),
    textsplitter.WithChunkOverlap(80),
)
parts, err := splitter.SplitText(markdown)
if err != nil { return err }
```

### Structured summary output call pattern
```go
// Source: https://raw.githubusercontent.com/openai/openai-go/main/README.md
resp, err := client.Responses.New(ctx, responses.ResponseNewParams{
    Model: openai.ChatModelGPT5_2,
    Input: responses.ResponseNewParamsInputUnion{OfString: openai.String(prompt)},
    Text: responses.ResponseTextConfigParam{
        Format: responses.ResponseFormatTextConfigParamOfJSONSchema(
            "chunk_summary",
            chunkSummarySchema,
        ),
    },
})
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `github.com/go-shiori/go-readability` primary | `codeberg.org/readeck/go-readability/v2` | 2025-2026 | Better maintained Readability-compatible extraction path; avoid deprecated package line. |
| Freeform summary strings | Structured output with JSON schema | GPT-4o+ era onward | Predictable machine-consumable chunk summaries and safer downstream parsing. |
| Character-only chunk sizing | Token-aware chunking guardrails | Modern LLM pipelines | More stable prompt budgeting and better retrieval/generation efficiency. |

**Deprecated/outdated:**
- `github.com/go-shiori/go-readability` as primary extractor: deprecated by maintainers in favor of readeck module path.

## Open Questions

1. **Near-duplicate algorithm beyond exact-match dedupe**
   - What we know: high-confidence overlap removal is required; exact and normalized hash checks are safe.
   - What's unclear: whether to adopt a dedicated near-dup library now or defer until real corpus metrics exist.
   - Recommendation: ship Phase 2 with exact + strict normalized dedupe only; instrument duplicate rates and revisit in later phase.

2. **Default medium chunk target by model family**
   - What we know: token-aware chunking is available and preferred.
   - What's unclear: exact default (`~400`, `~500`, `~700`) for best balance with later Phase 3 prompts.
   - Recommendation: start at 500 tokens with 80 overlap, then calibrate using real prompt-size telemetry.

## Sources

### Primary (HIGH confidence)
- `codeberg.org/readeck/go-readability/v2` docs: https://pkg.go.dev/codeberg.org/readeck/go-readability/v2
- Readeck fork docs and compatibility notes: https://codeberg.org/readeck/go-readability
- `html-to-markdown/v2` docs: https://pkg.go.dev/github.com/JohannesKaufmann/html-to-markdown/v2
- `goquery` docs (fallback/DOM extraction constraints): https://pkg.go.dev/github.com/PuerkitoBio/goquery
- LangChainGo text splitter docs: https://tmc.github.io/langchaingo/docs/modules/data_connection/text_splitters/
- LangChainGo splitter source (`token_splitter`, options, split docs): https://raw.githubusercontent.com/tmc/langchaingo/main/textsplitter/token_splitter.go
- OpenAI structured outputs guide: https://platform.openai.com/docs/guides/structured-outputs
- OpenAI Go SDK README (structured outputs example): https://raw.githubusercontent.com/openai/openai-go/main/README.md
- `tiktoken-go` docs: https://pkg.go.dev/github.com/pkoukk/tiktoken-go
- `bluemonday` docs: https://pkg.go.dev/github.com/microcosm-cc/bluemonday

### Secondary (MEDIUM confidence)
- Ecosystem pattern references for chunking and attribution (validated directionally against official docs): web search results on 2026 chunking best practices.

### Tertiary (LOW confidence)
- Web-search-only near-duplicate recommendations (MinHash/LSH blog posts) not backed by an actively maintained Go-first official standard package for this project yet.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - based on current official package docs, version metadata, and maintainer deprecation guidance.
- Architecture: HIGH - directly aligned with package capabilities and locked phase decisions.
- Pitfalls: MEDIUM-HIGH - mostly verified by official docs; dedupe-specific tuning remains context-dependent.

**Research date:** 2026-03-10
**Valid until:** 2026-04-09
