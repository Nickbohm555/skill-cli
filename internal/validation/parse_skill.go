package validation

import (
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	gmtext "github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

type rawFrontmatter struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Metadata    map[string]interface{} `yaml:"metadata"`
}

func ParseSkill(input []byte) (CandidateSkill, error) {
	candidate := newCandidateSkill()

	md := goldmark.New(goldmark.WithExtensions(&frontmatter.Extender{}))
	ctx := parser.NewContext()
	if err := md.Convert(input, io.Discard, parser.WithContext(ctx)); err != nil {
		return CandidateSkill{}, fmt.Errorf("parse markdown: %w", err)
	}
	doc := md.Parser().Parse(gmtext.NewReader(input), parser.WithContext(ctx))

	if fm := frontmatter.Get(ctx); fm != nil {
		var raw rawFrontmatter
		if err := fm.Decode(&raw); err != nil {
			return CandidateSkill{}, fmt.Errorf("decode frontmatter: %w", err)
		}
		candidate.Metadata = SkillMetadata{
			Name:        strings.TrimSpace(raw.Name),
			Description: strings.TrimSpace(raw.Description),
			Extra:       flattenMetadata(raw.Metadata),
		}
	}

	var current *parsedSection
	for node := doc.FirstChild(); node != nil; node = node.NextSibling() {
		heading, ok := node.(*ast.Heading)
		if !ok {
			if current != nil {
				current.blocks = append(current.blocks, node)
			}
			continue
		}

		headingText := strings.TrimSpace(extractText(input, heading))
		switch heading.Level {
		case 1:
			applySection(&candidate, input, current)
			if candidate.Title == "" {
				candidate.Title = headingText
			}
			current = nil
		case 2:
			applySection(&candidate, input, current)
			slug := slugHeading(headingText)
			current = &parsedSection{
				heading: headingText,
				slug:    slug,
				blocks:  make([]ast.Node, 0),
			}
		default:
			if current != nil {
				current.blocks = append(current.blocks, node)
			}
		}
	}
	applySection(&candidate, input, current)

	return candidate, nil
}

type parsedSection struct {
	heading string
	slug    string
	blocks  []ast.Node
}

func applySection(candidate *CandidateSkill, source []byte, section *parsedSection) {
	if section == nil {
		return
	}

	binding, ok := candidate.bindSection(section.slug)
	if !ok {
		return
	}

	body, items := collectSectionContent(source, section.blocks)
	if binding.kind == sectionKindText && binding.text != nil {
		binding.text.Heading = section.heading
		binding.text.Body = body
		return
	}
	if binding.kind == sectionKindList && binding.list != nil {
		binding.list.Heading = section.heading
		binding.list.Intro = body
		binding.list.Items = items
	}
}

func collectSectionContent(source []byte, blocks []ast.Node) (string, []string) {
	paragraphs := make([]string, 0)
	items := make([]string, 0)

	for _, block := range blocks {
		switch node := block.(type) {
		case *ast.List:
			items = append(items, collectListItems(source, node)...)
		case *ast.Heading:
			text := strings.TrimSpace(extractText(source, node))
			if text != "" {
				paragraphs = append(paragraphs, text)
			}
		default:
			text := strings.TrimSpace(extractText(source, node))
			if text != "" {
				paragraphs = append(paragraphs, text)
			}
		}
	}

	return strings.Join(normalizeLines(paragraphs), "\n\n"), normalizeLines(items)
}

func collectListItems(source []byte, list *ast.List) []string {
	items := make([]string, 0, list.ChildCount())
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		text := strings.TrimSpace(extractText(source, item))
		if text != "" {
			items = append(items, text)
		}
	}
	return items
}

func extractText(source []byte, node ast.Node) string {
	var parts []string
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			switch n.Kind() {
			case ast.KindParagraph, ast.KindHeading, ast.KindListItem, ast.KindBlockquote:
				parts = append(parts, "\n")
			}
			return ast.WalkContinue, nil
		}

		switch typed := n.(type) {
		case *ast.Text:
			parts = append(parts, string(typed.Segment.Value(source)))
			if typed.HardLineBreak() || typed.SoftLineBreak() {
				parts = append(parts, "\n")
			}
		case *ast.CodeSpan:
			parts = append(parts, strings.TrimSpace(string(typed.Text(source))))
		case *ast.FencedCodeBlock:
			for i := 0; i < typed.Lines().Len(); i++ {
				line := typed.Lines().At(i)
				parts = append(parts, string(line.Value(source)))
			}
			parts = append(parts, "\n")
		case *ast.CodeBlock:
			for i := 0; i < typed.Lines().Len(); i++ {
				line := typed.Lines().At(i)
				parts = append(parts, string(line.Value(source)))
			}
			parts = append(parts, "\n")
		}

		return ast.WalkContinue, nil
	})

	text := strings.Join(parts, "")
	text = strings.ReplaceAll(text, "\u00a0", " ")
	lines := strings.FieldsFunc(text, func(r rune) bool { return r == '\n' || r == '\r' })
	return strings.Join(normalizeLines(lines), "\n")
}

func flattenMetadata(metadata map[string]interface{}) map[string]string {
	if len(metadata) == 0 {
		return nil
	}

	flattened := make(map[string]string, len(metadata))
	for key, value := range metadata {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey == "" || value == nil {
			continue
		}
		flattened[trimmedKey] = strings.TrimSpace(fmt.Sprint(value))
	}
	if len(flattened) == 0 {
		return nil
	}
	return flattened
}

func slugHeading(heading string) string {
	var builder strings.Builder
	lastHyphen := false

	for _, r := range strings.ToLower(strings.TrimSpace(heading)) {
		if isSlugRune(r) {
			builder.WriteRune(r)
			lastHyphen = false
			continue
		}
		if lastHyphen {
			continue
		}
		builder.WriteByte('-')
		lastHyphen = true
	}

	return strings.Trim(builder.String(), "-")
}

func isSlugRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
