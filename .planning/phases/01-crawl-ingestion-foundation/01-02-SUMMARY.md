# Summary: 01-02 Docs-Root and Classification Heuristics

## Completed

- Added deterministic docs-root derivation in `internal/crawl/docs_root.go`.
- Added conservative docs-like and low-signal classification helpers in `internal/crawl/classify.go`.
- Expanded `internal/crawl/classify_test.go` with table-driven edge cases for mixed-case HTML media types, malformed or missing headers, obvious low-signal assets, and valid docs pages that must remain processable.

## Verification

- `go test ./internal/crawl -v` passed.
- Confirmed `ClassifyCandidate` exposes explicit outcomes only: `DocsLike`, `non_html_content_type`, `low_signal_page`, or an error that callers can map to `invalid_url`, with no silent drop path.

## Notes

- No blockers came up during this plan.
- Plan `01-02` is complete and the next implementation target is `01-03 / Task 1`.
