# Summary: 01-01 Crawl Contracts and Normalization Foundation

## Completed

- Added shared crawl accounting models in `internal/crawl/types.go`.
- Kept a stable machine-readable skip taxonomy in `internal/crawl/skip_reasons.go`.
- Implemented deterministic URL normalization, canonical key generation, and same-domain helpers in `internal/crawl/normalize.go`.
- Added table-driven coverage in `internal/crawl/normalize_test.go` for invalid URLs, fragment stripping, tracking query removal, canonical query ordering, path cleaning, relative resolution, and same-domain boundaries.

## Verification

- `go test ./internal/crawl -v` passed.
- Confirmed skip-reason literals are defined only in `internal/crawl/skip_reasons.go` and reused through the shared `SkipReason` type rather than duplicated ad hoc strings.

## Notes

- No blockers came up during this plan.
- Plan `01-01` is complete and the next implementation target is `01-02 / Task 1`.
