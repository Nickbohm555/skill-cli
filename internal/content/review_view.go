package content

import (
	"fmt"
	"sort"
	"strings"
)

// ReviewChunk is the summary-first record shown to users during inspection.
// It keeps concise summary fields at the top level while preserving explicit
// attribution and a raw expansion target for full-fidelity review.
type ReviewChunk struct {
	ChunkID      string
	SourceURL    string
	Summary      string
	Confidence   string
	Notes        string
	Attribution  ChunkAttribution
	ExpandTarget ExpandTarget
}

// ExpandTarget is the stable lookup key a caller can use to fetch the raw
// chunk content from a review projection.
type ExpandTarget struct {
	Key       string
	ChunkID   string
	SourceURL string
	Reference string
}

// RawChunkView is the full-fidelity raw chunk data indexed behind each
// ExpandTarget so the default review experience can stay summary-first.
type RawChunkView struct {
	ChunkID     string
	SourceURL   string
	Text        string
	PageTitle   string
	HeadingPath []string
	Checksum    string
	Reference   string
}

// ReviewView groups concise review rows with their raw expansion lookup table.
type ReviewView struct {
	Chunks     []ReviewChunk
	Expansions map[string]RawChunkView
}

// BuildReviewView projects structured summaries plus attributed raw chunks into
// a concise-first review surface with explicit raw expansion references.
func BuildReviewView(summaries []ChunkSummary, chunks []AttributedChunk) (ReviewView, error) {
	if len(summaries) == 0 {
		return ReviewView{
			Chunks:     nil,
			Expansions: map[string]RawChunkView{},
		}, nil
	}

	rawByChunkID := make(map[string]AttributedChunk, len(chunks))
	for _, chunk := range chunks {
		chunkID := strings.TrimSpace(chunk.Attribution.ChunkID)
		if chunkID == "" {
			chunkID = strings.TrimSpace(chunk.Chunk.ID)
		}
		if chunkID == "" {
			return ReviewView{}, fmt.Errorf("review chunk missing chunk identifier")
		}
		if !chunk.Attribution.HasRequiredFields() {
			return ReviewView{}, fmt.Errorf("review chunk %q missing required attribution", chunkID)
		}
		rawByChunkID[chunkID] = AttributedChunk{
			Chunk:       chunk.Chunk,
			Attribution: cloneAttribution(chunk.Attribution),
		}
	}

	view := ReviewView{
		Chunks:     make([]ReviewChunk, 0, len(summaries)),
		Expansions: make(map[string]RawChunkView, len(summaries)),
	}

	for _, summary := range summaries {
		chunkID := strings.TrimSpace(summary.ChunkID)
		sourceURL := strings.TrimSpace(summary.SourceURL)
		if chunkID == "" || sourceURL == "" {
			return ReviewView{}, fmt.Errorf("review summary missing chunk/source identifier")
		}
		if !summary.Attribution.HasRequiredFields() {
			return ReviewView{}, fmt.Errorf("review summary %q missing required attribution", chunkID)
		}

		rawChunk, ok := rawByChunkID[chunkID]
		if !ok {
			return ReviewView{}, fmt.Errorf("review summary %q missing raw chunk for expansion", chunkID)
		}
		if strings.TrimSpace(rawChunk.Attribution.SourceURL) != sourceURL {
			return ReviewView{}, fmt.Errorf("review summary %q source_url %q did not match raw chunk %q", chunkID, sourceURL, rawChunk.Attribution.SourceURL)
		}

		key := reviewExpandKey(chunkID, sourceURL)
		view.Expansions[key] = RawChunkView{
			ChunkID:     chunkID,
			SourceURL:   sourceURL,
			Text:        rawChunk.Chunk.Text,
			PageTitle:   rawChunk.Attribution.PageTitle,
			HeadingPath: append([]string(nil), rawChunk.Attribution.HeadingPath...),
			Checksum:    rawChunk.Attribution.Checksum,
			Reference:   rawChunk.Attribution.Reference,
		}
		view.Chunks = append(view.Chunks, ReviewChunk{
			ChunkID:     chunkID,
			SourceURL:   sourceURL,
			Summary:     strings.TrimSpace(summary.Summary),
			Confidence:  strings.TrimSpace(summary.Confidence),
			Notes:       strings.TrimSpace(summary.Notes),
			Attribution: cloneAttribution(summary.Attribution),
			ExpandTarget: ExpandTarget{
				Key:       key,
				ChunkID:   chunkID,
				SourceURL: sourceURL,
				Reference: summary.Attribution.Reference,
			},
		})
	}

	sort.SliceStable(view.Chunks, func(i, j int) bool {
		if view.Chunks[i].SourceURL == view.Chunks[j].SourceURL {
			return view.Chunks[i].ChunkID < view.Chunks[j].ChunkID
		}
		return view.Chunks[i].SourceURL < view.Chunks[j].SourceURL
	})

	return view, nil
}

func reviewExpandKey(chunkID string, sourceURL string) string {
	return strings.TrimSpace(sourceURL) + "#" + strings.TrimSpace(chunkID)
}
