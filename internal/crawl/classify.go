package crawl

import (
	"mime"
	"net/url"
	"path"
	"strings"
)

var docsLikeMediaTypes = map[string]struct{}{
	"application/xhtml+xml": {},
	"text/html":             {},
}

var lowSignalExtensions = map[string]struct{}{
	".7z":    {},
	".avi":   {},
	".bmp":   {},
	".css":   {},
	".csv":   {},
	".eot":   {},
	".gif":   {},
	".gz":    {},
	".ico":   {},
	".jpeg":  {},
	".jpg":   {},
	".js":    {},
	".json":  {},
	".map":   {},
	".mov":   {},
	".mp3":   {},
	".mp4":   {},
	".pdf":   {},
	".png":   {},
	".svg":   {},
	".tar":   {},
	".tgz":   {},
	".ttf":   {},
	".txt":   {},
	".wav":   {},
	".webm":  {},
	".webp":  {},
	".woff":  {},
	".woff2": {},
	".xml":   {},
	".zip":   {},
}

var lowSignalBaseNames = map[string]struct{}{
	"favicon.ico":  {},
	"humans.txt":   {},
	"robots.txt":   {},
	"security.txt": {},
	"sitemap.xml":  {},
}

// ClassificationOutcome makes skip semantics explicit for crawl orchestration.
type ClassificationOutcome struct {
	DocsLike   bool
	SkipReason SkipReason
}

// ClassifyCandidate combines content-type and URL signals into an explicit
// outcome the engine can map directly into skipped accounting.
func ClassifyCandidate(raw string, contentType string, base *url.URL) (ClassificationOutcome, error) {
	if !IsDocsLikeHTML(contentType) {
		return ClassificationOutcome{SkipReason: SkipReasonNonHTMLContentType}, nil
	}

	lowSignal, err := IsLowSignalPage(raw, base)
	if err != nil {
		return ClassificationOutcome{}, err
	}
	if lowSignal {
		return ClassificationOutcome{SkipReason: SkipReasonLowSignalPage}, nil
	}

	return ClassificationOutcome{DocsLike: true}, nil
}

// IsDocsLikeHTML reports whether a parsed response content type is suitable for
// documentation page processing.
func IsDocsLikeHTML(contentType string) bool {
	if strings.TrimSpace(contentType) == "" {
		return false
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}

	_, ok := docsLikeMediaTypes[strings.ToLower(mediaType)]
	return ok
}

// IsLowSignalPage reports whether a candidate URL is an obvious non-docs asset
// even if it is encountered in otherwise docs-shaped navigation.
func IsLowSignalPage(raw string, base *url.URL) (bool, error) {
	normalized, err := normalizeURL(raw, base)
	if err != nil {
		return false, err
	}

	baseName := strings.ToLower(path.Base(normalized.Path))
	if _, ok := lowSignalBaseNames[baseName]; ok {
		return true, nil
	}

	ext := strings.ToLower(path.Ext(baseName))
	if ext == "" {
		return false, nil
	}

	_, ok := lowSignalExtensions[ext]
	return ok, nil
}
