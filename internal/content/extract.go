package content

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
	"golang.org/x/net/html"
)

// ExtractReadable reduces a crawled page to readable HTML/text while
// preserving stable identifiers and audit metadata for downstream stages.
func ExtractReadable(page CrawledPage) (ExtractedPage, error) {
	pageURL, canonicalURL, err := normalizePageURLs(page)
	if err != nil {
		return ExtractedPage{}, err
	}

	htmlInput, err := validateHTML(page.HTML)
	if err != nil {
		return ExtractedPage{}, err
	}

	article, err := readability.FromReader(strings.NewReader(htmlInput), pageURL)
	if err != nil {
		return ExtractedPage{}, fmt.Errorf("%w: %v", ErrUnreadableContent, err)
	}

	readableHTML, err := renderArticleHTML(article)
	if err != nil {
		return ExtractedPage{}, fmt.Errorf("%w: %v", ErrUnreadableContent, err)
	}

	readableText, err := renderArticleText(article)
	if err != nil {
		return ExtractedPage{}, fmt.Errorf("%w: %v", ErrUnreadableContent, err)
	}

	readableHTML = strings.TrimSpace(readableHTML)
	readableText = strings.TrimSpace(readableText)
	if readableHTML == "" && readableText == "" {
		return ExtractedPage{}, fmt.Errorf("%w: readable extraction returned empty output", ErrUnreadableContent)
	}

	title := strings.TrimSpace(article.Title())
	if title == "" {
		title = strings.TrimSpace(page.Title)
	}

	canonical := canonicalURL.String()
	return ExtractedPage{
		ID:           stablePageID(canonical),
		SourceURL:    pageURL.String(),
		CanonicalURL: canonical,
		Title:        title,
		ReadableHTML: readableHTML,
		ReadableText: readableText,
		Stats: NormalizationStats{
			SourceHTMLBytes:   len(htmlInput),
			ReadableHTMLBytes: len(readableHTML),
			ReadableTextBytes: len(readableText),
			PlainTextBytes:    len(readableText),
			WordCount:         len(strings.Fields(readableText)),
		},
		Metadata: ProcessingMetadata{
			SiteName:         strings.TrimSpace(article.SiteName()),
			Byline:           strings.TrimSpace(article.Byline()),
			Excerpt:          strings.TrimSpace(article.Excerpt()),
			Language:         strings.TrimSpace(article.Language()),
			SourceChecksum:   checksum(htmlInput),
			ReadableChecksum: checksum(readableHTML + "\n" + readableText),
		},
	}, nil
}

func normalizePageURLs(page CrawledPage) (*url.URL, *url.URL, error) {
	rawURL := strings.TrimSpace(page.URL)
	if rawURL == "" {
		return nil, nil, fmt.Errorf("%w: missing source url", ErrInvalidPageURL)
	}

	pageURL, err := url.Parse(rawURL)
	if err != nil || pageURL.Scheme == "" || pageURL.Host == "" {
		if err == nil {
			err = errors.New("url must include scheme and host")
		}
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidPageURL, err)
	}

	canonicalInput := strings.TrimSpace(page.CanonicalURL)
	if canonicalInput == "" {
		canonicalInput = rawURL
	}

	canonicalURL, err := url.Parse(canonicalInput)
	if err != nil || canonicalURL.Scheme == "" || canonicalURL.Host == "" {
		if err == nil {
			err = errors.New("canonical url must include scheme and host")
		}
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidPageURL, err)
	}

	return pageURL, canonicalURL, nil
}

func validateHTML(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("%w: empty html", ErrInvalidPageHTML)
	}

	doc, err := html.Parse(strings.NewReader(trimmed))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPageHTML, err)
	}
	if !containsMeaningfulHTML(doc) {
		return "", fmt.Errorf("%w: no meaningful html nodes found", ErrInvalidPageHTML)
	}

	return trimmed, nil
}

func containsMeaningfulHTML(node *html.Node) bool {
	if node == nil {
		return false
	}

	if node.Type == html.ElementNode {
		switch node.Data {
		case "article", "main", "section", "div", "p", "pre", "code", "table", "ul", "ol":
			return true
		}
	}

	if node.Type == html.TextNode && strings.TrimSpace(node.Data) != "" {
		return true
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if containsMeaningfulHTML(child) {
			return true
		}
	}

	return false
}

func renderArticleHTML(article readability.Article) (string, error) {
	var buf bytes.Buffer
	if err := article.RenderHTML(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func renderArticleText(article readability.Article) (string, error) {
	var buf bytes.Buffer
	if err := article.RenderText(&buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func stablePageID(canonicalURL string) string {
	return checksum(canonicalURL)
}

func checksum(input string) string {
	sum := sha256.Sum256([]byte(input))
	return hex.EncodeToString(sum[:])
}
