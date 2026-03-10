package content

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"net/url"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

const (
	maxCodeBlockRunes          = 2000
	codeBlockTruncationMessage = "[code block truncated after %d runes]"
)

// NormalizeContent converts extracted readable HTML into stable markdown and
// plain text while preserving tables, code blocks, and media context.
func NormalizeContent(page ExtractedPage) (NormalizedPage, error) {
	normalized := NormalizedPage{
		ID:           page.ID,
		SourceURL:    page.SourceURL,
		CanonicalURL: page.CanonicalURL,
		Title:        page.Title,
		ReadableHTML: strings.TrimSpace(page.ReadableHTML),
		Stats:        page.Stats,
		Metadata:     page.Metadata,
	}

	baseURL, err := url.Parse(strings.TrimSpace(page.SourceURL))
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		if err == nil {
			err = fmt.Errorf("url must include scheme and host")
		}
		normalized.NormalizationErr = fmt.Sprintf("%v: %v", ErrInvalidPageURL, err)
		return normalized, fmt.Errorf("%w: %v", ErrInvalidPageURL, err)
	}

	normalized.ReadableHTML, err = prepareReadableHTML(page.ReadableHTML, page.ReadableText)
	if err != nil {
		normalized.NormalizationErr = err.Error()
		return normalized, err
	}

	markdown, err := convertReadableHTML(normalized.ReadableHTML, baseURL.String())
	if err != nil {
		normalized.NormalizationErr = err.Error()
		return normalized, err
	}

	normalized.Markdown = cleanupMarkdown(markdown)
	normalized.PlainText = derivePlainText(page.ReadableText, normalized.Markdown)
	normalized.Stats.MarkdownBytes = len(normalized.Markdown)
	normalized.Stats.PlainTextBytes = len(normalized.PlainText)
	normalized.Stats.WordCount = len(strings.Fields(normalized.PlainText))

	return normalized, nil
}

func prepareReadableHTML(readableHTML string, readableText string) (string, error) {
	input := strings.TrimSpace(readableHTML)
	if input == "" {
		fallbackText := strings.TrimSpace(readableText)
		if fallbackText == "" {
			return "", fmt.Errorf("%w: empty readable content", ErrUnreadableContent)
		}

		return "<p>" + stdhtml.EscapeString(fallbackText) + "</p>", nil
	}

	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPageHTML, err)
	}

	truncateOversizedCodeBlocks(doc)
	preserveEmbeddedMediaContext(doc)

	body := findElement(doc, "body")
	if body == nil {
		return renderHTMLNode(doc)
	}

	return renderHTMLChildren(body)
}

func convertReadableHTML(inputHTML string, baseURL string) (string, error) {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)

	markdown, err := conv.ConvertString(
		inputHTML,
		converter.WithDomain(baseURL),
	)
	if err != nil {
		return "", fmt.Errorf("%w: markdown conversion failed: %v", ErrUnreadableContent, err)
	}

	if strings.TrimSpace(markdown) == "" {
		return "", fmt.Errorf("%w: markdown conversion returned empty output", ErrUnreadableContent)
	}

	return markdown, nil
}

func truncateOversizedCodeBlocks(root *html.Node) {
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node == nil {
			return
		}

		if node.Type == html.ElementNode && node.Data == "pre" {
			target := firstChildElement(node, "code")
			if target == nil {
				target = node
			}

			text := strings.TrimRight(nodeInnerText(target), "\n")
			if utf8.RuneCountInString(text) > maxCodeBlockRunes {
				replacement := truncateRunes(text, maxCodeBlockRunes) + "\n" + fmt.Sprintf(codeBlockTruncationMessage, maxCodeBlockRunes)
				replaceNodeText(target, replacement)
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}

	visit(root)
}

func preserveEmbeddedMediaContext(root *html.Node) {
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node == nil {
			return
		}

		if node.Type == html.ElementNode {
			switch node.Data {
			case "iframe", "video", "audio":
				context := mediaContextLine(node)
				if context != "" && node.Parent != nil {
					paragraph := &html.Node{
						Type: html.ElementNode,
						Data: "p",
					}
					paragraph.AppendChild(&html.Node{
						Type: html.TextNode,
						Data: context,
					})
					node.Parent.InsertBefore(paragraph, node)
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}

	visit(root)
}

func mediaContextLine(node *html.Node) string {
	label := firstNonEmpty(
		attrValue(node, "title"),
		attrValue(node, "aria-label"),
		attrValue(node, "aria-describedby"),
		attrValue(node, "data-caption"),
	)
	src := firstNonEmpty(
		attrValue(node, "src"),
		attrValue(node, "data-src"),
		attrValue(node, "poster"),
	)

	parts := []string{"Embedded media"}
	if label != "" {
		parts = append(parts, "label: "+label)
	}
	if src != "" {
		parts = append(parts, "source: "+src)
	}
	if len(parts) == 1 {
		return ""
	}

	return strings.Join(parts, " | ")
}

func cleanupMarkdown(markdown string) string {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	cleaned := make([]string, 0, len(lines))
	blankCount := 0

	for _, line := range lines {
		trimmedRight := strings.TrimRight(line, " \t")
		if strings.TrimSpace(trimmedRight) == "" {
			blankCount++
			if blankCount > 1 {
				continue
			}
			cleaned = append(cleaned, "")
			continue
		}

		blankCount = 0
		cleaned = append(cleaned, trimmedRight)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func derivePlainText(readableText string, markdown string) string {
	text := strings.TrimSpace(readableText)
	if text == "" {
		text = strings.TrimSpace(markdown)
	}

	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	cleaned := make([]string, 0, len(lines))
	blankCount := 0
	for _, line := range lines {
		trimmed := strings.Join(strings.Fields(line), " ")
		if trimmed == "" {
			blankCount++
			if blankCount > 1 {
				continue
			}
			cleaned = append(cleaned, "")
			continue
		}

		blankCount = 0
		cleaned = append(cleaned, trimmed)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func renderHTMLNode(node *html.Node) (string, error) {
	var buf bytes.Buffer
	if err := html.Render(&buf, node); err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidPageHTML, err)
	}

	return strings.TrimSpace(buf.String()), nil
}

func renderHTMLChildren(node *html.Node) (string, error) {
	var buf bytes.Buffer
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := html.Render(&buf, child); err != nil {
			return "", fmt.Errorf("%w: %v", ErrInvalidPageHTML, err)
		}
	}

	return strings.TrimSpace(buf.String()), nil
}

func findElement(node *html.Node, tag string) *html.Node {
	if node == nil {
		return nil
	}
	if node.Type == html.ElementNode && node.Data == tag {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if found := findElement(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func firstChildElement(node *html.Node, tag string) *html.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == tag {
			return child
		}
	}
	return nil
}

func nodeInnerText(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Data
	}

	var b strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		b.WriteString(nodeInnerText(child))
	}

	return b.String()
}

func replaceNodeText(node *html.Node, content string) {
	node.FirstChild = nil
	node.LastChild = nil
	textNode := &html.Node{
		Type:   html.TextNode,
		Data:   content,
		Parent: node,
	}
	node.FirstChild = textNode
	node.LastChild = textNode
}

func truncateRunes(input string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(input) <= limit {
		return input
	}

	var b strings.Builder
	count := 0
	for _, r := range input {
		if count == limit {
			break
		}
		b.WriteRune(r)
		count++
	}

	return strings.TrimRight(b.String(), "\n")
}

func attrValue(node *html.Node, name string) string {
	for _, attr := range node.Attr {
		if attr.Key == name {
			return strings.TrimSpace(attr.Val)
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
