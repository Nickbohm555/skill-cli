package content

import "strings"

// ChunkAttribution carries the source metadata that must remain attached to a
// chunk through downstream summarization and review flows.
type ChunkAttribution struct {
	SourceURL   string
	PageTitle   string
	HeadingPath []string
	ChunkID     string
	Checksum    string
	Reference   string
}

// AttributedChunk is the summarization-ready record emitted by the content
// pipeline after chunking and attribution stamping.
type AttributedChunk struct {
	Chunk       Chunk
	Attribution ChunkAttribution
}

func cloneAttribution(in ChunkAttribution) ChunkAttribution {
	out := in
	if in.HeadingPath != nil {
		out.HeadingPath = append([]string(nil), in.HeadingPath...)
	}
	return out
}

// NewChunkAttribution stamps metadata directly from the normalized page and
// chunk so attribution does not need to be reconstructed later.
func NewChunkAttribution(page NormalizedPage, chunk Chunk) ChunkAttribution {
	sourceURL := firstNonEmpty(
		strings.TrimSpace(page.SourceURL),
		strings.TrimSpace(page.CanonicalURL),
	)
	pageTitle := firstNonEmpty(
		strings.TrimSpace(page.Title),
		sourceURL,
		stableChunkPageID(page),
	)
	headingPath := deriveHeadingPath(page, chunk)

	return ChunkAttribution{
		SourceURL:   sourceURL,
		PageTitle:   pageTitle,
		HeadingPath: headingPath,
		ChunkID:     strings.TrimSpace(chunk.ID),
		Checksum:    firstNonEmpty(strings.TrimSpace(chunk.Checksum), checksum(strings.TrimSpace(chunk.Text))),
		Reference:   sourceURL + "#" + strings.TrimSpace(chunk.ID),
	}
}

// HasRequiredFields reports whether the attribution satisfies the minimum
// metadata contract required by the Phase 2 pipeline.
func (a ChunkAttribution) HasRequiredFields() bool {
	return strings.TrimSpace(a.SourceURL) != "" &&
		strings.TrimSpace(a.PageTitle) != "" &&
		len(a.HeadingPath) > 0 &&
		strings.TrimSpace(a.ChunkID) != "" &&
		strings.TrimSpace(a.Checksum) != "" &&
		strings.TrimSpace(a.Reference) != ""
}

func deriveHeadingPath(page NormalizedPage, chunk Chunk) []string {
	headings := extractChunkHeadings(chunk.Text)
	if len(headings) > 0 {
		return headings
	}

	fallback := firstNonEmpty(
		strings.TrimSpace(page.Title),
		strings.TrimSpace(page.CanonicalURL),
		strings.TrimSpace(page.SourceURL),
		stableChunkPageID(page),
	)
	if fallback == "" {
		return []string{"untitled"}
	}

	return []string{fallback}
}

func extractChunkHeadings(text string) []string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	stack := make([]string, 0, 6)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}

		level := 0
		for level < len(trimmed) && trimmed[level] == '#' {
			level++
		}
		if level == 0 || level >= len(trimmed) || trimmed[level] != ' ' {
			continue
		}

		heading := strings.TrimSpace(trimmed[level+1:])
		if heading == "" {
			continue
		}

		if level-1 < len(stack) {
			stack = stack[:level-1]
		}
		stack = append(stack, heading)
	}

	if len(stack) == 0 {
		return nil
	}

	path := make([]string, len(stack))
	copy(path, stack)
	return path
}
