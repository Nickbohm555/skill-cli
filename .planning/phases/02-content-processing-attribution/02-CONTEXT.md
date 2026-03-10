# Phase 2: Content Processing & Attribution - Context

**Gathered:** 2026-03-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Transform Phase 1 crawl outputs into clean, normalized, inspectable content with source attribution suitable for downstream generation. This phase covers processing, chunking, summarization, and attribution visibility; it does not add new end-user capabilities beyond these boundaries.

</domain>

<decisions>
## Implementation Decisions

### Normalization Rules
- Use minimal chrome stripping; keep most page text unless clearly irrelevant.
- Apply light cleanup to preserve near-original formatting.
- Deduplicate similar content globally across pages when confidence is high.

### Chunk Structure and Size
- Use medium chunk sizes as the default.
- Chunk summaries should be brief gists (1-2 lines).
- Default review experience should be summary-first with optional raw chunk expansion.

### Attribution Visibility
- Require per-chunk source attribution.

### Edge Content Handling
- Preserve key code snippets and truncate very large code blocks.
- Preserve table structure in normalized output.
- Keep alt text and captions from media as context hints.

### Claude's Discretion
- Boilerplate/disclaimer noise handling strategy.
- Chunk boundary strategy (semantic-first, fixed windows, or hybrid guardrails).
- Attribution display default presentation.
- Multi-source attribution display format.
- Source-review depth beyond basic per-chunk URL visibility.
- Long-list filtering strategy (navigation/reference list handling).

</decisions>

<specifics>
## Specific Ideas

No external product references were specified. Preferences emphasized preserving source fidelity while keeping outputs inspectable and concise.

</specifics>

<deferred>
## Deferred Ideas

None - discussion stayed within phase scope.

</deferred>

---

*Phase: 02-content-processing-attribution*
*Context gathered: 2026-03-10*
