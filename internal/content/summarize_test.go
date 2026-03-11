package content

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSummarizeChunksUsesStructuredProviderOutput(t *testing.T) {
	t.Parallel()

	chunk := summaryTestChunk()
	summaries, err := SummarizeChunksWithConfig(context.Background(), []AttributedChunk{chunk}, SummaryConfig{
		Provider: stubSummaryProvider{
			record: SummaryRecord{
				ChunkID:    chunk.Attribution.ChunkID,
				SourceURL:  chunk.Attribution.SourceURL,
				Summary:    "Install the CLI before authenticating.\nKeep the generated token available for later commands.",
				Confidence: summaryConfidenceHigh,
				Notes:      "provider result",
			},
		},
	})
	if err != nil {
		t.Fatalf("SummarizeChunksWithConfig() error = %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("SummarizeChunksWithConfig() returned %d summaries, want 1", len(summaries))
	}

	got := summaries[0]
	if got.ChunkID != chunk.Attribution.ChunkID {
		t.Fatalf("summary chunk_id = %q, want %q", got.ChunkID, chunk.Attribution.ChunkID)
	}
	if got.SourceURL != chunk.Attribution.SourceURL {
		t.Fatalf("summary source_url = %q, want %q", got.SourceURL, chunk.Attribution.SourceURL)
	}
	if got.Summary != "Install the CLI before authenticating.\nKeep the generated token available for later commands." {
		t.Fatalf("summary text = %q", got.Summary)
	}
	if got.Confidence != summaryConfidenceHigh {
		t.Fatalf("summary confidence = %q, want %q", got.Confidence, summaryConfidenceHigh)
	}
	if got.Notes != "provider result" {
		t.Fatalf("summary notes = %q, want provider result", got.Notes)
	}
	if got.UsedFallback {
		t.Fatal("summary unexpectedly used fallback")
	}
	if got.FallbackReason != "" {
		t.Fatalf("fallback reason = %q, want empty", got.FallbackReason)
	}
	if got.Attribution.ChunkID != chunk.Attribution.ChunkID || got.Attribution.SourceURL != chunk.Attribution.SourceURL {
		t.Fatalf("summary attribution = %+v, want %+v", got.Attribution, chunk.Attribution)
	}
	if countSummaryLines(got.Summary) != 2 {
		t.Fatalf("summary line count = %d, want 2", countSummaryLines(got.Summary))
	}
}

func TestSummarizeChunksFallsBackWhenProviderUnavailable(t *testing.T) {
	t.Parallel()

	chunk := summaryTestChunk()
	summaries, err := SummarizeChunksWithConfig(context.Background(), []AttributedChunk{chunk}, SummaryConfig{})
	if err != nil {
		t.Fatalf("SummarizeChunksWithConfig() error = %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("SummarizeChunksWithConfig() returned %d summaries, want 1", len(summaries))
	}

	got := summaries[0]
	if !got.UsedFallback {
		t.Fatal("summary did not use fallback")
	}
	if got.FallbackReason != "provider unavailable" {
		t.Fatalf("fallback reason = %q, want provider unavailable", got.FallbackReason)
	}
	if got.Confidence != summaryConfidenceLow {
		t.Fatalf("summary confidence = %q, want %q", got.Confidence, summaryConfidenceLow)
	}
	if got.ChunkID != chunk.Attribution.ChunkID || got.SourceURL != chunk.Attribution.SourceURL {
		t.Fatalf("summary attribution identifiers were not preserved: %+v", got)
	}
	if countSummaryLines(got.Summary) == 0 || countSummaryLines(got.Summary) > 2 {
		t.Fatalf("fallback summary line count = %d, want 1 or 2", countSummaryLines(got.Summary))
	}
	if !strings.Contains(got.Notes, fallbackSummaryNote) {
		t.Fatalf("fallback notes = %q, want %q", got.Notes, fallbackSummaryNote)
	}
}

func TestSummarizeChunksBoundsProviderSummaryToTwoLines(t *testing.T) {
	t.Parallel()

	chunk := summaryTestChunk()
	summaries, err := SummarizeChunksWithConfig(context.Background(), []AttributedChunk{chunk}, SummaryConfig{
		Provider: stubSummaryProvider{
			record: SummaryRecord{
				ChunkID:   chunk.Attribution.ChunkID,
				SourceURL: chunk.Attribution.SourceURL,
				Summary: " Install the CLI before authenticating. \n\n" +
					"Keep the generated token available for later commands.\n" +
					"Ignore this third line.",
			},
		},
	})
	if err != nil {
		t.Fatalf("SummarizeChunksWithConfig() error = %v", err)
	}

	got := summaries[0]
	if got.UsedFallback {
		t.Fatal("summary unexpectedly used fallback for line normalization")
	}
	if countSummaryLines(got.Summary) != 2 {
		t.Fatalf("summary line count = %d, want 2", countSummaryLines(got.Summary))
	}
	if strings.Contains(got.Summary, "Ignore this third line.") {
		t.Fatalf("summary retained third line: %q", got.Summary)
	}
}

func TestSummarizeChunksFallsBackWhenProviderReturnsInvalidRecord(t *testing.T) {
	t.Parallel()

	chunk := summaryTestChunk()
	summaries, err := SummarizeChunksWithConfig(context.Background(), []AttributedChunk{chunk}, SummaryConfig{
		Provider: stubSummaryProvider{
			record: SummaryRecord{
				ChunkID:   "wrong-chunk-id",
				SourceURL: chunk.Attribution.SourceURL,
				Summary:   "line one\nline two\nline three",
			},
		},
	})
	if err != nil {
		t.Fatalf("SummarizeChunksWithConfig() error = %v", err)
	}

	got := summaries[0]
	if !got.UsedFallback {
		t.Fatal("summary did not use fallback for invalid provider record")
	}
	if got.FallbackReason != "summary validation failed" {
		t.Fatalf("fallback reason = %q, want summary validation failed", got.FallbackReason)
	}
	if got.ChunkID != chunk.Attribution.ChunkID {
		t.Fatalf("summary chunk_id = %q, want %q", got.ChunkID, chunk.Attribution.ChunkID)
	}
	if got.SourceURL != chunk.Attribution.SourceURL {
		t.Fatalf("summary source_url = %q, want %q", got.SourceURL, chunk.Attribution.SourceURL)
	}
	if countSummaryLines(got.Summary) == 0 || countSummaryLines(got.Summary) > 2 {
		t.Fatalf("fallback summary line count = %d, want 1 or 2", countSummaryLines(got.Summary))
	}
	if !strings.Contains(got.Notes, fallbackSummaryNote) {
		t.Fatalf("fallback notes = %q, want fallback note", got.Notes)
	}
}

func TestSummarizeChunksFallsBackWhenProviderOmitsSourceURL(t *testing.T) {
	t.Parallel()

	chunk := summaryTestChunk()
	summaries, err := SummarizeChunksWithConfig(context.Background(), []AttributedChunk{chunk}, SummaryConfig{
		Provider: stubSummaryProvider{
			record: SummaryRecord{
				ChunkID: chunk.Attribution.ChunkID,
				Summary: "Install the CLI before authenticating.",
			},
		},
	})
	if err != nil {
		t.Fatalf("SummarizeChunksWithConfig() error = %v", err)
	}

	got := summaries[0]
	if !got.UsedFallback {
		t.Fatal("summary did not use fallback for schema validation error")
	}
	if got.FallbackReason != "summary validation failed" {
		t.Fatalf("fallback reason = %q, want summary validation failed", got.FallbackReason)
	}
	if got.SourceURL != chunk.Attribution.SourceURL {
		t.Fatalf("summary source_url = %q, want %q", got.SourceURL, chunk.Attribution.SourceURL)
	}
	if countSummaryLines(got.Summary) == 0 || countSummaryLines(got.Summary) > 2 {
		t.Fatalf("fallback summary line count = %d, want 1 or 2", countSummaryLines(got.Summary))
	}
}

func TestSummarizeChunksFallsBackWhenProviderErrors(t *testing.T) {
	t.Parallel()

	chunk := summaryTestChunk()
	summaries, err := SummarizeChunksWithConfig(context.Background(), []AttributedChunk{chunk}, SummaryConfig{
		Provider: stubSummaryProvider{err: errors.New("provider boom")},
	})
	if err != nil {
		t.Fatalf("SummarizeChunksWithConfig() error = %v", err)
	}

	got := summaries[0]
	if !got.UsedFallback {
		t.Fatal("summary did not use fallback for provider error")
	}
	if got.FallbackReason != "provider boom" {
		t.Fatalf("fallback reason = %q, want provider boom", got.FallbackReason)
	}
	if !strings.Contains(got.Notes, "provider boom") {
		t.Fatalf("fallback notes = %q, want provider error detail", got.Notes)
	}
}

type stubSummaryProvider struct {
	record SummaryRecord
	err    error
}

func (p stubSummaryProvider) Summarize(_ context.Context, _ SummaryInput) (SummaryRecord, error) {
	if p.err != nil {
		return SummaryRecord{}, p.err
	}
	return p.record, nil
}

func summaryTestChunk() AttributedChunk {
	page := NormalizedPage{
		ID:           "page-summary",
		SourceURL:    "https://docs.example.com/guides/install",
		CanonicalURL: "https://docs.example.com/guides/install",
		Title:        "Install Guide",
	}
	chunk := Chunk{
		ID:       "page-summary:chunk-000",
		PageID:   page.ID,
		Order:    0,
		Text:     "# Install Guide\n\nInstall the CLI before authenticating. Then create a token and export it for later commands.",
		Checksum: "abc123",
	}

	return AttributedChunk{
		Chunk:       chunk,
		Attribution: NewChunkAttribution(page, chunk),
	}
}
