package content

import "errors"

var (
	ErrInvalidPageURL    = errors.New("invalid page url")
	ErrInvalidPageHTML   = errors.New("invalid page html")
	ErrUnreadableContent = errors.New("unreadable content")
)

// CrawledPage is the raw Phase 1 page payload handed to content processing.
type CrawledPage struct {
	URL          string
	CanonicalURL string
	Title        string
	ContentType  string
	HTML         string
	Depth        int
}

// ProcessingMetadata captures extractor and dedupe metadata that downstream
// stages can preserve for attribution and auditability.
type ProcessingMetadata struct {
	SiteName         string
	Byline           string
	Excerpt          string
	Language         string
	SourceChecksum   string
	ReadableChecksum string
}

// NormalizationStats tracks stable byte and word counts across processing
// stages so later chunking and attribution can reason about page shape.
type NormalizationStats struct {
	SourceHTMLBytes   int
	ReadableHTMLBytes int
	ReadableTextBytes int
	MarkdownBytes     int
	PlainTextBytes    int
	WordCount         int
}

// ExtractedPage is the deterministic output of readable-content extraction and
// the input to later normalization stages.
type ExtractedPage struct {
	ID           string
	SourceURL    string
	CanonicalURL string
	Title        string
	ReadableHTML string
	ReadableText string
	Stats        NormalizationStats
	Metadata     ProcessingMetadata
}

// NormalizedPage is the shared Phase 2 record shape used by normalization,
// dedupe, chunking, and summarization.
type NormalizedPage struct {
	ID               string
	SourceURL        string
	CanonicalURL     string
	Title            string
	ReadableHTML     string
	Markdown         string
	PlainText        string
	Stats            NormalizationStats
	Metadata         ProcessingMetadata
	Deduped          bool
	DuplicateOf      string
	DuplicateReason  string
	NormalizationErr string
}
