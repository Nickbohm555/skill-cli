package content

import "strings"

// PipelineConfig controls the normalized-page to attributed-chunk orchestration.
type PipelineConfig struct {
	ChunkConfig ChunkConfig
}

// DefaultPipelineConfig returns the Phase 2 default pipeline settings.
func DefaultPipelineConfig() PipelineConfig {
	return PipelineConfig{
		ChunkConfig: DefaultChunkConfig(),
	}
}

// ProcessToChunks runs normalized pages through chunking and stamps
// attribution at chunk creation time for downstream summarization.
func ProcessToChunks(pages []NormalizedPage) ([]AttributedChunk, error) {
	return ProcessToChunksWithConfig(pages, DefaultPipelineConfig())
}

// ProcessToChunksWithConfig applies caller-supplied chunking configuration
// while preserving deterministic page and chunk ordering.
func ProcessToChunksWithConfig(pages []NormalizedPage, cfg PipelineConfig) ([]AttributedChunk, error) {
	cfg = normalizePipelineConfig(cfg)

	if len(pages) == 0 {
		return nil, nil
	}

	attributed := make([]AttributedChunk, 0, len(pages))
	for _, page := range pages {
		if shouldSkipPageForChunking(page) {
			continue
		}

		chunks, err := BuildChunksWithConfig(page, cfg.ChunkConfig)
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			record := AttributedChunk{
				Chunk:       chunk,
				Attribution: NewChunkAttribution(page, chunk),
			}
			if !record.Attribution.HasRequiredFields() {
				continue
			}

			attributed = append(attributed, record)
		}
	}

	if len(attributed) == 0 {
		return nil, nil
	}

	return attributed, nil
}

func normalizePipelineConfig(cfg PipelineConfig) PipelineConfig {
	cfg.ChunkConfig = normalizeChunkConfig(cfg.ChunkConfig)
	return cfg
}

func shouldSkipPageForChunking(page NormalizedPage) bool {
	if page.Deduped {
		return true
	}
	if strings.TrimSpace(page.NormalizationErr) != "" {
		return true
	}

	content := strings.TrimSpace(page.Markdown)
	if content == "" {
		content = strings.TrimSpace(page.PlainText)
	}

	return content == ""
}
