package content

import "testing"

func TestBuildReviewViewIncludesSummaryAttributionAndExpansion(t *testing.T) {
	t.Parallel()

	chunk := summaryTestChunk()
	summary := ChunkSummary{
		ChunkID:     chunk.Attribution.ChunkID,
		SourceURL:   chunk.Attribution.SourceURL,
		Summary:     "Install the CLI before authenticating.\nStore the token for later commands.",
		Confidence:  summaryConfidenceHigh,
		Notes:       "provider result",
		Attribution: cloneAttribution(chunk.Attribution),
	}

	view, err := BuildReviewView([]ChunkSummary{summary}, []AttributedChunk{chunk})
	if err != nil {
		t.Fatalf("BuildReviewView() error = %v", err)
	}
	if len(view.Chunks) != 1 {
		t.Fatalf("BuildReviewView() chunks = %d, want 1", len(view.Chunks))
	}
	if len(view.Expansions) != 1 {
		t.Fatalf("BuildReviewView() expansions = %d, want 1", len(view.Expansions))
	}

	got := view.Chunks[0]
	if got.ChunkID != chunk.Attribution.ChunkID {
		t.Fatalf("review chunk_id = %q, want %q", got.ChunkID, chunk.Attribution.ChunkID)
	}
	if got.SourceURL != chunk.Attribution.SourceURL {
		t.Fatalf("review source_url = %q, want %q", got.SourceURL, chunk.Attribution.SourceURL)
	}
	if got.Summary != summary.Summary {
		t.Fatalf("review summary = %q, want %q", got.Summary, summary.Summary)
	}
	if got.Attribution.Reference != chunk.Attribution.Reference {
		t.Fatalf("review attribution reference = %q, want %q", got.Attribution.Reference, chunk.Attribution.Reference)
	}

	raw, ok := view.Expansions[got.ExpandTarget.Key]
	if !ok {
		t.Fatalf("missing expansion for key %q", got.ExpandTarget.Key)
	}
	if raw.Text != chunk.Chunk.Text {
		t.Fatalf("raw expansion text = %q, want %q", raw.Text, chunk.Chunk.Text)
	}
	if raw.Reference != chunk.Attribution.Reference {
		t.Fatalf("raw expansion reference = %q, want %q", raw.Reference, chunk.Attribution.Reference)
	}
	if len(raw.HeadingPath) == 0 {
		t.Fatal("raw expansion heading path was empty")
	}
}

func TestBuildReviewViewSupportsMultipleSourcesWithoutCollapsingProvenance(t *testing.T) {
	t.Parallel()

	first := summaryTestChunk()
	second := summaryTestChunk()
	second.Chunk.ID = "page-summary:chunk-001"
	second.Chunk.Text = "# Authenticate\n\nCreate a token before calling the API."
	second.Attribution = NewChunkAttribution(NormalizedPage{
		ID:           "page-auth",
		SourceURL:    "https://docs.example.com/guides/auth",
		CanonicalURL: "https://docs.example.com/guides/auth",
		Title:        "Auth Guide",
	}, second.Chunk)

	summaries := []ChunkSummary{
		{
			ChunkID:     first.Attribution.ChunkID,
			SourceURL:   first.Attribution.SourceURL,
			Summary:     "Install the CLI before authenticating.",
			Attribution: cloneAttribution(first.Attribution),
		},
		{
			ChunkID:     second.Attribution.ChunkID,
			SourceURL:   second.Attribution.SourceURL,
			Summary:     "Create a token before calling the API.",
			Attribution: cloneAttribution(second.Attribution),
		},
	}

	view, err := BuildReviewView(summaries, []AttributedChunk{first, second})
	if err != nil {
		t.Fatalf("BuildReviewView() error = %v", err)
	}
	if len(view.Chunks) != 2 {
		t.Fatalf("BuildReviewView() chunks = %d, want 2", len(view.Chunks))
	}
	if len(view.Expansions) != 2 {
		t.Fatalf("BuildReviewView() expansions = %d, want 2", len(view.Expansions))
	}

	for _, reviewChunk := range view.Chunks {
		raw, ok := view.Expansions[reviewChunk.ExpandTarget.Key]
		if !ok {
			t.Fatalf("missing expansion for %q", reviewChunk.ExpandTarget.Key)
		}
		if raw.SourceURL != reviewChunk.SourceURL {
			t.Fatalf("raw expansion source_url = %q, want %q", raw.SourceURL, reviewChunk.SourceURL)
		}
		if raw.ChunkID != reviewChunk.ChunkID {
			t.Fatalf("raw expansion chunk_id = %q, want %q", raw.ChunkID, reviewChunk.ChunkID)
		}
	}
}

func TestBuildReviewViewErrorsWhenRawExpansionIsMissing(t *testing.T) {
	t.Parallel()

	chunk := summaryTestChunk()
	_, err := BuildReviewView([]ChunkSummary{
		{
			ChunkID:     chunk.Attribution.ChunkID,
			SourceURL:   chunk.Attribution.SourceURL,
			Summary:     "Install the CLI before authenticating.",
			Attribution: cloneAttribution(chunk.Attribution),
		},
	}, nil)
	if err == nil {
		t.Fatal("BuildReviewView() error = nil, want missing raw chunk error")
	}
}
