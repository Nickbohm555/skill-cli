package content

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkoukk/tiktoken-go"
	"github.com/tmc/langchaingo/textsplitter"
)

const (
	DefaultChunkSizeTokens    = 500
	DefaultChunkOverlapTokens = 80
	defaultChunkEncodingName  = "cl100k_base"
	chunkIDHashLength         = 12
	pageIDPrefixLength        = 12
)

// ChunkConfig controls semantic-first chunking with token guardrails.
type ChunkConfig struct {
	ChunkSizeTokens    int
	ChunkOverlapTokens int
	EncodingName       string
}

// Chunk is the stable content unit emitted from a normalized page.
type Chunk struct {
	ID         string
	PageID     string
	Order      int
	Text       string
	TokenCount int
	Checksum   string
}

// DefaultChunkConfig returns the Phase 2 default chunking policy.
func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		ChunkSizeTokens:    DefaultChunkSizeTokens,
		ChunkOverlapTokens: DefaultChunkOverlapTokens,
		EncodingName:       defaultChunkEncodingName,
	}
}

// BuildChunks converts one normalized page into deterministic, reviewable
// chunks using semantic-first markdown splitting followed by token enforcement.
func BuildChunks(page NormalizedPage) ([]Chunk, error) {
	return BuildChunksWithConfig(page, DefaultChunkConfig())
}

// BuildChunksWithConfig applies a caller-supplied chunking configuration.
func BuildChunksWithConfig(page NormalizedPage, cfg ChunkConfig) ([]Chunk, error) {
	cfg = normalizeChunkConfig(cfg)
	if err := validateChunkConfig(cfg); err != nil {
		return nil, err
	}

	content := strings.TrimSpace(page.Markdown)
	if content == "" {
		content = strings.TrimSpace(page.PlainText)
	}
	if content == "" {
		return nil, nil
	}

	encoder, err := tiktoken.GetEncoding(cfg.EncodingName)
	if err != nil {
		return nil, fmt.Errorf("load token encoding %q: %w", cfg.EncodingName, err)
	}

	tokenLen := func(input string) int {
		return len(encoder.Encode(input, nil, nil))
	}

	tokenSplitter := textsplitter.NewTokenSplitter(
		textsplitter.WithChunkSize(cfg.ChunkSizeTokens),
		textsplitter.WithChunkOverlap(cfg.ChunkOverlapTokens),
		textsplitter.WithEncodingName(cfg.EncodingName),
	)

	markdownSplitter := textsplitter.NewMarkdownTextSplitter(
		textsplitter.WithChunkSize(cfg.ChunkSizeTokens),
		textsplitter.WithChunkOverlap(cfg.ChunkOverlapTokens),
		textsplitter.WithLenFunc(tokenLen),
		textsplitter.WithCodeBlocks(true),
		textsplitter.WithJoinTableRows(true),
		textsplitter.WithSecondSplitter(tokenSplitter),
	)

	rawChunks, err := markdownSplitter.SplitText(content)
	if err != nil {
		return nil, fmt.Errorf("split markdown into chunks: %w", err)
	}

	pageID := stableChunkPageID(page)
	chunks := make([]Chunk, 0, len(rawChunks))
	for _, rawChunk := range rawChunks {
		normalizedChunk := strings.TrimSpace(rawChunk)
		if normalizedChunk == "" {
			continue
		}

		if tokenLen(normalizedChunk) > cfg.ChunkSizeTokens {
			fallbackChunks, err := tokenSplitter.SplitText(normalizedChunk)
			if err != nil {
				return nil, fmt.Errorf("split oversized chunk: %w", err)
			}

			for _, fallbackChunk := range fallbackChunks {
				chunkText := strings.TrimSpace(fallbackChunk)
				if chunkText == "" {
					continue
				}

				chunks = append(chunks, newChunk(pageID, len(chunks), chunkText, tokenLen(chunkText)))
			}

			continue
		}

		chunks = append(chunks, newChunk(pageID, len(chunks), normalizedChunk, tokenLen(normalizedChunk)))
	}

	return chunks, nil
}

func validateChunkConfig(cfg ChunkConfig) error {
	if cfg.ChunkSizeTokens <= 0 {
		return fmt.Errorf("chunk size must be positive")
	}
	if cfg.ChunkOverlapTokens < 0 {
		return fmt.Errorf("chunk overlap must be non-negative")
	}
	if cfg.ChunkOverlapTokens >= cfg.ChunkSizeTokens {
		return fmt.Errorf("chunk overlap must be smaller than chunk size")
	}

	return nil
}

func normalizeChunkConfig(cfg ChunkConfig) ChunkConfig {
	defaults := DefaultChunkConfig()
	isZeroValue := cfg.ChunkSizeTokens == 0 && cfg.ChunkOverlapTokens == 0 && strings.TrimSpace(cfg.EncodingName) == ""

	if cfg.ChunkSizeTokens <= 0 {
		cfg.ChunkSizeTokens = defaults.ChunkSizeTokens
	}
	if cfg.ChunkOverlapTokens < 0 {
		cfg.ChunkOverlapTokens = defaults.ChunkOverlapTokens
	}
	if cfg.ChunkOverlapTokens == 0 && isZeroValue {
		cfg.ChunkOverlapTokens = defaults.ChunkOverlapTokens
	}
	if strings.TrimSpace(cfg.EncodingName) == "" {
		cfg.EncodingName = defaults.EncodingName
	}

	return cfg
}

func stableChunkPageID(page NormalizedPage) string {
	if pageID := strings.TrimSpace(page.ID); pageID != "" {
		return pageID
	}
	if canonicalURL := strings.TrimSpace(page.CanonicalURL); canonicalURL != "" {
		return stablePageID(canonicalURL)
	}
	if sourceURL := strings.TrimSpace(page.SourceURL); sourceURL != "" {
		return stablePageID(sourceURL)
	}

	return checksum(strings.TrimSpace(page.Title))
}

func newChunk(pageID string, order int, text string, tokenCount int) Chunk {
	contentChecksum := checksum(text)

	return Chunk{
		ID:         stableChunkID(pageID, order, contentChecksum),
		PageID:     pageID,
		Order:      order,
		Text:       text,
		TokenCount: tokenCount,
		Checksum:   contentChecksum,
	}
}

func stableChunkID(pageID string, order int, contentChecksum string) string {
	pagePrefix := pageID
	if len(pagePrefix) > pageIDPrefixLength {
		pagePrefix = pagePrefix[:pageIDPrefixLength]
	}

	hashPart := contentChecksum
	if len(hashPart) > chunkIDHashLength {
		hashPart = hashPart[:chunkIDHashLength]
	}

	return pagePrefix + "-" + strconv.Itoa(order+1) + "-" + hashPart
}
