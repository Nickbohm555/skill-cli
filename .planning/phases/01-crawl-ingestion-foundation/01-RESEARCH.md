# Phase 01: Crawl & Ingestion Foundation - Research

**Researched:** 2026-03-10
**Domain:** Same-domain documentation crawl foundation for Go CLI
**Confidence:** MEDIUM-HIGH

## Summary

Phase 01 should be planned as a deterministic, bounded crawl subsystem with explicit filtering and accounting, not as a generic scraper. The implementation should prioritize exact requirement coverage for CRAWL-01..04: same-domain enforcement, hard default cap at 50 processed pages, transparent skip reasons, and a final summary with discovered/processed/skipped counts.

Given current constraints and the existing Go-first architecture direction, the standard stack for this phase is `colly/v2` for crawl orchestration plus `net/url` canonicalization rules and strict HTML/docs-page filtering before enqueue and before process. The planning-critical decision left open in context (query + fragment handling) should be resolved now as a default canonicalization policy to avoid duplicate crawl explosion and inconsistent counts.

**Primary recommendation:** Implement a synchronous BFS-style crawl on top of `colly/v2` with canonical URL keys (`fragment stripped`, query normalized with tracking params removed) and mandatory per-URL skip reasons emitted into the final summary report.

## Standard Stack

The established libraries/tools for this domain:

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/gocolly/colly/v2` | `v2.3.0` | Same-domain crawling, link discovery, callback lifecycle | Provides `AllowedDomains`, `URLFilters`, depth controls, and request lifecycle hooks that directly map to CRAWL-01..04 without hand-rolled scheduler complexity. |
| Go `net/url` | Go stdlib (`go1.26.1` docs) | URL parsing, canonicalization, relative resolution | Canonical URL handling is the core correctness boundary for dedupe and same-domain checks. |
| Go `mime` (`ParseMediaType`) | Go stdlib (`go1.26.1` docs) | Robust content-type parsing for HTML gating | Prevents false positives from loose `Content-Type` matching and supports explicit unsupported-type skip reasons. |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/PuerkitoBio/goquery` | `v1.11.0` | Parse HTML and extract anchors/content signals | Use inside `OnResponse`/processing to classify docs-like pages and extract crawl candidates from HTML. |
| Go `path` | Go stdlib (`go1.26.1` docs) | URL path cleanup (`path.Clean`) | Use during canonicalization to collapse noisy path forms before dedupe keys are generated. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `colly/v2` | Hand-rolled `net/http` + queue + parser | More control, but reinvents crawl lifecycle, visit dedupe, and callback orchestration; higher bug risk for cap and accounting logic in Phase 01. |
| Query normalization | Drop all query params | Simpler, but can collapse distinct docs pages (e.g., language/version routes encoded in query) and hide content unexpectedly. |

**Installation:**
```bash
go get github.com/gocolly/colly/v2
go get github.com/PuerkitoBio/goquery
```

## Architecture Patterns

### Recommended Project Structure
```text
internal/crawl/
├── engine.go          # crawl orchestration + cap enforcement + summary accounting
├── normalize.go       # URL normalization + docs-root derivation + canonical key logic
├── classify.go        # docs-like and low-signal page classifiers
├── skip_reasons.go    # enum/constants for skip reason taxonomy
└── types.go           # CrawlResult, PageRecord, SummaryCounts
```

### Pattern 1: Deterministic Bounded Crawl Loop
**What:** A single orchestrator that owns all counters and enforces a hard processed-page cap (`50` default).
**When to use:** Always for Phase 01; asynchronous crawling is unnecessary for this cap and makes exact accounting harder.
**Example:**
```go
// Source: https://go-colly.org/docs/examples/max_depth/ (visit lifecycle pattern),
// adapted for processed-cap enforcement in this project.
processed := 0
capPages := 50

c.OnRequest(func(r *colly.Request) {
    if processed >= capPages {
        r.Abort()
        recordSkip(r.URL.String(), SkipCapReached)
        return
    }
})

c.OnResponse(func(r *colly.Response) {
    if !isDocsLikeHTML(r) {
        recordSkip(r.Request.URL.String(), SkipUnsupportedOrLowSignal)
        return
    }
    processed++
    recordProcessed(r.Request.URL.String())
})
```

### Pattern 2: Canonical URL Keying Before Deduplication
**What:** Normalize URL before enqueue and before counting discovered/processed/skipped.
**When to use:** For every candidate URL (entry URL and extracted links).
**Example:**
```go
// Source: https://pkg.go.dev/net/url (Parse, ResolveReference, Values.Encode sorted by key)
// plus project decision for fragment/query policy.
func canonicalKey(raw string, base *url.URL) (string, error) {
    u, err := url.Parse(raw)
    if err != nil {
        return "", err
    }
    if base != nil {
        u = base.ResolveReference(u)
    }
    u.Fragment = "" // browsers strip fragment from HTTP requests; use one canonical key

    q := u.Query()
    for _, k := range []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "gclid", "fbclid"} {
        q.Del(k)
    }
    u.RawQuery = q.Encode() // sorted by key per net/url docs
    u.Path = path.Clean(u.Path)
    return u.String(), nil
}
```

### Pattern 3: Skip Taxonomy as First-Class Output
**What:** Every non-processed candidate gets a machine-readable skip reason.
**When to use:** Required for CRAWL-03 transparency and CRAWL-04 summary consistency.
**Example skip reasons:**
- `off_domain`
- `already_seen`
- `cap_reached`
- `non_html_content_type`
- `invalid_url`
- `low_signal_page`
- `fetch_error`

### Anti-Patterns to Avoid
- **Counter updates in multiple callbacks:** Splits ownership of discovered/processed/skipped and causes inconsistent final summary.
- **Dedupe on raw URL strings:** Causes duplicates from fragments, path variants, and query-order differences.
- **Silent filtering:** Any skip without explicit reason breaks CRAWL-03 and makes debugging impossible.
- **Default async crawl in Phase 01:** Makes exact 50-page cap and deterministic counts harder than necessary.

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Crawl lifecycle + request scheduling | Custom recursive fetch/queue engine | `colly/v2` collector callbacks | Colly already provides robust request hooks, domain filters, and crawl controls. |
| URL parsing/relative resolution | String concatenation for links | `net/url` parse + `ResolveReference` | Correct RFC-compliant URL handling is subtle and error-prone when hand-built. |
| Content-Type parsing | `strings.Contains(contentType, "html")` only | `mime.ParseMediaType` | Header values include params/casing; parsing avoids false decisions. |

**Key insight:** Phase 01 quality is mostly correctness of boundaries (domain, cap, skip semantics). Existing libraries solve boundary mechanics better than custom code.

## Common Pitfalls

### Pitfall 1: Cap Overshoot Under Concurrency
**What goes wrong:** More than 50 pages get processed due to in-flight requests.
**Why it happens:** Cap checked too late or async scheduling allows bursts.
**How to avoid:** Use synchronous crawl mode for Phase 01, check cap in `OnRequest`, and only increment processed count after docs-like HTML validation.
**Warning signs:** `processed > 50` in summary or non-deterministic counts between runs.

### Pitfall 2: Broken Same-Domain Logic
**What goes wrong:** Off-domain URLs are crawled or valid same-host URLs are skipped.
**Why it happens:** Comparing raw host strings without normalization or resolving relative URLs incorrectly.
**How to avoid:** Resolve against base URL, compare normalized hostnames, and apply `AllowedDomains` plus explicit guard checks.
**Warning signs:** Processed URLs include external hosts or many false `off_domain` skips for relative links.

### Pitfall 3: Low-Signal Heuristics Too Aggressive
**What goes wrong:** Valid docs pages get filtered out.
**Why it happens:** Overly broad extension/path denylists or simplistic text-length checks.
**How to avoid:** Start with conservative skip rules (non-HTML + obvious binary/media paths) and keep `low_signal_page` criteria explicit and testable.
**Warning signs:** Very low processed count from clearly rich docs sites.

### Pitfall 4: Query/Fragment Policy Drift
**What goes wrong:** Duplicate inflation or page loss depending on ad hoc URL handling.
**Why it happens:** Query and fragment rules are undocumented or inconsistently applied in different call sites.
**How to avoid:** Centralize canonicalization in one function and document default policy in code and phase docs.
**Warning signs:** Same page appears multiple times with minor URL differences.

## Code Examples

Verified patterns from official sources:

### Same-domain and URL filtering with Colly
```go
// Source: https://go-colly.org/docs/introduction/configuration
// and https://go-colly.org/docs/examples/url_filter/
c := colly.NewCollector(
    colly.AllowedDomains("docs.example.com"),
    colly.URLFilters(regexp.MustCompile(`^https://docs\.example\.com/`)),
    colly.MaxDepth(4),
)
```

### Parse Content-Type safely
```go
// Source: https://pkg.go.dev/mime#ParseMediaType
mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
if err != nil || mediaType != "text/html" {
    recordSkip(url, SkipNonHTMLContentType)
    return
}
```

### Canonicalize query ordering for stable dedupe keys
```go
// Source: https://pkg.go.dev/net/url (Values.Encode sorts by key)
q := u.Query()
q.Del("utm_source")
q.Del("fbclid")
u.RawQuery = q.Encode()
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `github.com/gocolly/colly` v1 import path | `github.com/gocolly/colly/v2` module path | v2 module adoption | Use v2 path for maintained releases and current APIs. |
| Loose HTML parsing assumptions | Explicit UTF-8 expectation in goquery usage | Ongoing; documented in goquery README | Plan to treat malformed/non-UTF-8 pages as skipped/unsupported in Phase 01 boundaries. |
| Ad hoc URL keying | Canonicalization with stdlib URL APIs | Standard modern crawler practice | Improves determinism of discovered/processed/skipped counts. |

**Deprecated/outdated:**
- `colly` v1-style import path for new code: replace with `/v2` module path.

## Open Questions

1. **Query parameter policy strictness**
   - What we know: fragment stripping should be unconditional; query handling was intentionally left open.
   - What's unclear: whether to keep all non-tracking query params or use an allowlist (`lang`, `version`, etc.) from day one.
   - Recommendation: start with "keep non-tracking params" + canonical sort; add optional allowlist policy in a later requirement (aligns with CRAWL-05 advanced policies).

2. **Docs-root normalization heuristic**
   - What we know: decision says normalize entry URL to nearest docs root before crawl.
   - What's unclear: exact heuristic order (`/docs`, `/documentation`, breadcrumb/nav markers, sitemap hints).
   - Recommendation: document deterministic heuristic precedence in plan tasks and include fixture-based tests for at least 4 URL shapes.

3. **Potential scope conflict with earlier single-page guidance**
   - What we know: older stack research favored single-page ingestion for v1, while current roadmap/requirements include bounded same-domain crawl.
   - What's unclear: whether older docs should be updated before implementation to avoid planning drift.
   - Recommendation: treat CRAWL-01..04 as source of truth and schedule a doc-alignment task in planning.

## Sources

### Primary (HIGH confidence)
- https://go-colly.org/docs/introduction/configuration - collector options (`AllowedDomains`, max depth env var, transport tuning)
- https://go-colly.org/docs/examples/max_depth/ - bounded traversal example pattern
- https://go-colly.org/docs/examples/url_filter/ - URL filter pattern
- https://pkg.go.dev/github.com/gocolly/colly/v2 - current module path/version metadata and feature set
- https://pkg.go.dev/net/url - parse/resolve/canonicalization APIs and `Values.Encode` sort behavior
- https://pkg.go.dev/mime#ParseMediaType - robust content-type parsing
- https://pkg.go.dev/path#Clean - path normalization behavior
- https://pkg.go.dev/github.com/PuerkitoBio/goquery - HTML parsing constraints and current version metadata

### Secondary (MEDIUM confidence)
- https://go-colly.org/docs/best_practices/crawling - operational best practices for crawl jobs

### Tertiary (LOW confidence)
- Web ecosystem discovery searches for "Go crawler best practices 2026" used only for direction; all critical claims above were verified against official docs.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - official package/docs coverage is strong for selected libraries.
- Architecture: MEDIUM-HIGH - patterns are prescriptive and aligned to requirements, but exact project code constraints are still greenfield.
- Pitfalls: MEDIUM - grounded in common crawler failure modes and doc evidence, but some are implementation-behavior dependent.

**Research date:** 2026-03-10
**Valid until:** 2026-04-09
