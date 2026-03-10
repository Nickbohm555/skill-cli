package crawl

import (
	"net/url"
	"strings"
)

var docsRootSegmentPrecedence = []string{
	"docs",
	"documentation",
	"guide",
	"guides",
}

// DeriveDocsRoot normalizes an entry URL down to the nearest deterministic
// docs root using explicit path-segment precedence before falling back to "/".
func DeriveDocsRoot(raw string) (*url.URL, error) {
	entry, err := NormalizeEntryURL(raw)
	if err != nil {
		return nil, err
	}

	root := &url.URL{
		Scheme: entry.Scheme,
		Host:   entry.Host,
		Path:   deriveDocsRootPath(entry.Path),
	}

	return root, nil
}

func deriveDocsRootPath(rawPath string) string {
	trimmed := strings.Trim(rawPath, "/")
	if trimmed == "" {
		return "/"
	}

	segments := strings.Split(trimmed, "/")
	for _, marker := range docsRootSegmentPrecedence {
		for idx, segment := range segments {
			if segment == marker {
				return "/" + strings.Join(segments[:idx+1], "/")
			}
		}
	}

	return "/"
}
