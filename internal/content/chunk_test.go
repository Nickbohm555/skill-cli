package content

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildChunksDeterministicIDsAndOrder(t *testing.T) {
	t.Parallel()

	page := NormalizedPage{
		ID:           "page-deterministic",
		SourceURL:    "https://docs.example.com/guides/install",
		CanonicalURL: "https://docs.example.com/guides/install",
		Title:        "Install Guide",
		Markdown: strings.Join([]string{
			"# Install Guide",
			"",
			strings.Repeat("This paragraph documents the setup workflow in a stable way. ", 30),
			"",
			"## Configure",
			"",
			strings.Repeat("Configuration details should still produce deterministic chunk output. ", 30),
		}, "\n"),
	}

	cfg := ChunkConfig{
		ChunkSizeTokens:    80,
		ChunkOverlapTokens: 10,
		EncodingName:       defaultChunkEncodingName,
	}

	first, err := BuildChunksWithConfig(page, cfg)
	if err != nil {
		t.Fatalf("BuildChunksWithConfig() error = %v", err)
	}
	second, err := BuildChunksWithConfig(page, cfg)
	if err != nil {
		t.Fatalf("BuildChunksWithConfig() second call error = %v", err)
	}

	if len(first) < 2 {
		t.Fatalf("BuildChunksWithConfig() returned %d chunks, want at least 2", len(first))
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("BuildChunksWithConfig() output was not deterministic\nfirst=%+v\nsecond=%+v", first, second)
	}

	for i, chunk := range first {
		if chunk.Order != i {
			t.Fatalf("chunk[%d] order = %d, want %d", i, chunk.Order, i)
		}
		if chunk.ID == "" {
			t.Fatalf("chunk[%d] returned empty stable ID", i)
		}
		if chunk.PageID != page.ID {
			t.Fatalf("chunk[%d] page ID = %q, want %q", i, chunk.PageID, page.ID)
		}
		if chunk.Checksum == "" {
			t.Fatalf("chunk[%d] returned empty checksum", i)
		}
	}
}

func TestBuildChunksEnforcesTokenCapGuardrails(t *testing.T) {
	t.Parallel()

	page := NormalizedPage{
		ID:           "page-token-cap",
		SourceURL:    "https://docs.example.com/reference/flags",
		CanonicalURL: "https://docs.example.com/reference/flags",
		Title:        "Flags",
		Markdown: strings.Join([]string{
			"# Flags",
			"",
			strings.Repeat("The command supports predictable token aware chunk boundaries. ", 80),
		}, "\n"),
	}

	cfg := ChunkConfig{
		ChunkSizeTokens:    40,
		ChunkOverlapTokens: 8,
		EncodingName:       defaultChunkEncodingName,
	}

	chunks, err := BuildChunksWithConfig(page, cfg)
	if err != nil {
		t.Fatalf("BuildChunksWithConfig() error = %v", err)
	}

	if len(chunks) < 2 {
		t.Fatalf("BuildChunksWithConfig() returned %d chunks, want multiple chunks for guardrail coverage", len(chunks))
	}

	for i, chunk := range chunks {
		if chunk.TokenCount > cfg.ChunkSizeTokens {
			t.Fatalf("chunk[%d] token count = %d, exceeds cap %d", i, chunk.TokenCount, cfg.ChunkSizeTokens)
		}
		if strings.TrimSpace(chunk.Text) == "" {
			t.Fatalf("chunk[%d] text was empty after splitting", i)
		}
	}
}

func TestBuildChunksPreservesTableAndCodeBoundaries(t *testing.T) {
	t.Parallel()

	page := NormalizedPage{
		ID:           "page-structure",
		SourceURL:    "https://docs.example.com/reference/structure",
		CanonicalURL: "https://docs.example.com/reference/structure",
		Title:        "Structure",
		Markdown: strings.Join([]string{
			"# Structure",
			"",
			"Use the preserved structures below during review.",
			"",
			"| Flag | Description |",
			"| --- | --- |",
			"| --help | Show help |",
			"| --version | Print version |",
			"",
			"```bash",
			"cli-skill --help",
			"cli-skill --version",
			"```",
		}, "\n"),
	}

	chunks, err := BuildChunks(page)
	if err != nil {
		t.Fatalf("BuildChunks() error = %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("BuildChunks() returned no chunks")
	}

	foundTable := false
	foundCode := false
	for i, chunk := range chunks {
		if strings.Contains(chunk.Text, "| --help | Show help |") {
			foundTable = true
			if !strings.Contains(chunk.Text, "| --version | Print version |") {
				t.Fatalf("chunk[%d] did not keep related table rows together\nchunk=%s", i, chunk.Text)
			}
		}

		if strings.Contains(chunk.Text, "```bash") {
			foundCode = true
			if strings.Count(chunk.Text, "```") != 2 {
				t.Fatalf("chunk[%d] contains an imbalanced fenced code block\nchunk=%s", i, chunk.Text)
			}
			if !strings.Contains(chunk.Text, "cli-skill --version") {
				t.Fatalf("chunk[%d] did not preserve code block body\nchunk=%s", i, chunk.Text)
			}
		}
	}

	if !foundTable {
		t.Fatal("BuildChunks() did not preserve any table chunk")
	}
	if !foundCode {
		t.Fatal("BuildChunks() did not preserve any fenced code chunk")
	}
}

func TestProcessToChunksRequiresAttributionForEveryChunk(t *testing.T) {
	t.Parallel()

	pages := []NormalizedPage{
		{
			ID:           "page-attribution",
			SourceURL:    "https://docs.example.com/reference/pipeline",
			CanonicalURL: "https://docs.example.com/reference/pipeline",
			Title:        "Pipeline",
			Markdown: strings.Join([]string{
				"# Pipeline",
				"",
				strings.Repeat("Every emitted chunk must carry source attribution for review. ", 40),
				"",
				"## Inputs",
				"",
				strings.Repeat("Chunk creation should stamp metadata before downstream usage. ", 35),
			}, "\n"),
		},
	}

	chunks, err := ProcessToChunksWithConfig(pages, PipelineConfig{
		ChunkConfig: ChunkConfig{
			ChunkSizeTokens:    60,
			ChunkOverlapTokens: 10,
			EncodingName:       defaultChunkEncodingName,
		},
	})
	if err != nil {
		t.Fatalf("ProcessToChunksWithConfig() error = %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("ProcessToChunksWithConfig() returned %d chunks, want multiple attributed chunks", len(chunks))
	}

	for i, chunk := range chunks {
		if !chunk.Attribution.HasRequiredFields() {
			t.Fatalf("chunk[%d] attribution missing required fields: %+v", i, chunk.Attribution)
		}
		if chunk.Attribution.SourceURL != pages[0].SourceURL {
			t.Fatalf("chunk[%d] source_url = %q, want %q", i, chunk.Attribution.SourceURL, pages[0].SourceURL)
		}
		if chunk.Attribution.ChunkID != chunk.Chunk.ID {
			t.Fatalf("chunk[%d] chunk_id = %q, want %q", i, chunk.Attribution.ChunkID, chunk.Chunk.ID)
		}
		if len(chunk.Attribution.HeadingPath) == 0 {
			t.Fatalf("chunk[%d] heading_path was empty", i)
		}
	}
}

func TestAttributionRemainsUnchangedForDownstreamSummaryInput(t *testing.T) {
	t.Parallel()

	page := NormalizedPage{
		ID:           "page-summary-input",
		SourceURL:    "https://docs.example.com/reference/downstream",
		CanonicalURL: "https://docs.example.com/reference/downstream",
		Title:        "Downstream",
		Markdown: strings.Join([]string{
			"# Downstream",
			"",
			"## Summary Input",
			"",
			strings.Repeat("Chunk text may be forwarded, but attribution must stay stable. ", 20),
		}, "\n"),
	}

	attributed, err := ProcessToChunks([]NormalizedPage{page})
	if err != nil {
		t.Fatalf("ProcessToChunks() error = %v", err)
	}
	if len(attributed) == 0 {
		t.Fatal("ProcessToChunks() returned no attributed chunks")
	}

	original := cloneAttribution(attributed[0].Attribution)
	input := newSummaryInput(attributed[0])

	if input.Text != attributed[0].Chunk.Text {
		t.Fatalf("newSummaryInput() text = %q, want %q", input.Text, attributed[0].Chunk.Text)
	}
	if !reflect.DeepEqual(input.Attribution, original) {
		t.Fatalf("newSummaryInput() attribution = %+v, want %+v", input.Attribution, original)
	}
	if !reflect.DeepEqual(attributed[0].Attribution, original) {
		t.Fatalf("attribution changed after building downstream input\ngot=%+v\nwant=%+v", attributed[0].Attribution, original)
	}
}

type summaryInput struct {
	Text        string
	Attribution ChunkAttribution
}

func newSummaryInput(chunk AttributedChunk) summaryInput {
	return summaryInput{
		Text:        chunk.Chunk.Text,
		Attribution: cloneAttribution(chunk.Attribution),
	}
}

func cloneAttribution(in ChunkAttribution) ChunkAttribution {
	out := in
	if in.HeadingPath != nil {
		out.HeadingPath = append([]string(nil), in.HeadingPath...)
	}
	return out
}
