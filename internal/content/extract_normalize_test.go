package content

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestExtractReadable(t *testing.T) {
	t.Parallel()

	longBody := strings.Repeat("Detailed installation guidance. ", 80)

	tests := []struct {
		name        string
		page        CrawledPage
		wantErr     error
		assertions  []string
		notContains []string
	}{
		{
			name: "extracts readable article content",
			page: CrawledPage{
				URL:          "https://docs.example.com/guides/install",
				CanonicalURL: "https://docs.example.com/guides/install",
				Title:        "Install Guide",
				HTML: `<!doctype html>
<html>
  <head><title>Install Guide</title></head>
  <body>
    <header>Docs navigation</header>
    <nav>Sidebar links</nav>
    <article>
      <h1>Install Guide</h1>
      <p>This guide keeps the important setup details.</p>
      <p>` + longBody + `</p>
      <p>npm install cli-skill</p>
    </article>
    <footer>Copyright footer</footer>
  </body>
</html>`,
			},
			assertions: []string{
				"This guide keeps the important setup details.",
				"npm install cli-skill",
			},
			notContains: []string{
				"Sidebar links",
			},
		},
		{
			name: "keeps short content and does not over-strip",
			page: CrawledPage{
				URL:          "https://docs.example.com/reference/env",
				CanonicalURL: "https://docs.example.com/reference/env",
				Title:        "Env Vars",
				HTML: `<!doctype html>
<html>
  <body>
    <main>
      <p>CLI_SKILL_API_KEY is required for authenticated calls.</p>
    </main>
  </body>
</html>`,
			},
			assertions: []string{
				"CLI_SKILL_API_KEY is required for authenticated calls.",
			},
		},
		{
			name: "rejects invalid page url",
			page: CrawledPage{
				URL:          "docs.example.com/guides/install",
				CanonicalURL: "https://docs.example.com/guides/install",
				HTML:         "<article><p>body</p></article>",
			},
			wantErr: ErrInvalidPageURL,
		},
		{
			name: "rejects empty html",
			page: CrawledPage{
				URL:          "https://docs.example.com/guides/install",
				CanonicalURL: "https://docs.example.com/guides/install",
				HTML:         "   ",
			},
			wantErr: ErrInvalidPageHTML,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ExtractReadable(tc.page)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ExtractReadable() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ExtractReadable() error = %v", err)
			}

			if got.ID == "" {
				t.Fatalf("ExtractReadable() returned empty stable ID")
			}
			if got.Title != tc.page.Title {
				t.Fatalf("ExtractReadable() title = %q, want %q", got.Title, tc.page.Title)
			}
			if got.Metadata.SourceChecksum == "" {
				t.Fatalf("ExtractReadable() returned empty source checksum")
			}
			if got.Metadata.ReadableChecksum == "" {
				t.Fatalf("ExtractReadable() returned empty readable checksum")
			}
			if got.Stats.SourceHTMLBytes == 0 || got.Stats.ReadableHTMLBytes == 0 || got.Stats.PlainTextBytes == 0 {
				t.Fatalf("ExtractReadable() returned incomplete stats: %+v", got.Stats)
			}

			combined := got.ReadableHTML + "\n" + got.ReadableText
			for _, want := range tc.assertions {
				if !strings.Contains(combined, want) {
					t.Fatalf("ExtractReadable() output missing %q\noutput=%s", want, combined)
				}
			}
			for _, unwanted := range tc.notContains {
				if strings.Contains(combined, unwanted) {
					t.Fatalf("ExtractReadable() output unexpectedly contained %q\noutput=%s", unwanted, combined)
				}
			}
		})
	}
}

func TestNormalizeContentPreservesStructure(t *testing.T) {
	t.Parallel()

	extracted := ExtractedPage{
		ID:           "page-1",
		SourceURL:    "https://docs.example.com/reference/tables",
		CanonicalURL: "https://docs.example.com/reference/tables",
		Title:        "Reference",
		ReadableHTML: `<article>
<h1>Reference</h1>
<p>Use the table and code sample below.</p>
<table>
  <thead><tr><th>Flag</th><th>Description</th></tr></thead>
  <tbody><tr><td>--help</td><td>Show help</td></tr></tbody>
</table>
<pre><code>` + strings.Repeat("x", maxCodeBlockRunes+50) + `</code></pre>
<figure>
  <img src="/images/arch.png" alt="Architecture diagram" title="System architecture">
  <figcaption>Rendered pipeline overview.</figcaption>
</figure>
<video src="/media/demo.mp4" title="CLI walkthrough" data-caption="Demo clip"></video>
</article>`,
		ReadableText: "Reference\nUse the table and code sample below.\nFlag Description\n--help Show help\nArchitecture diagram\nRendered pipeline overview.",
		Stats: NormalizationStats{
			SourceHTMLBytes:   1200,
			ReadableHTMLBytes: 900,
			ReadableTextBytes: 140,
			PlainTextBytes:    140,
			WordCount:         16,
		},
		Metadata: ProcessingMetadata{
			SourceChecksum:   "source-checksum",
			ReadableChecksum: "readable-checksum",
		},
	}

	normalized, err := NormalizeContent(extracted)
	if err != nil {
		t.Fatalf("NormalizeContent() error = %v", err)
	}

	if normalized.ID != extracted.ID {
		t.Fatalf("NormalizeContent() ID = %q, want %q", normalized.ID, extracted.ID)
	}
	if normalized.Metadata.SourceChecksum != extracted.Metadata.SourceChecksum {
		t.Fatalf("NormalizeContent() source checksum = %q, want %q", normalized.Metadata.SourceChecksum, extracted.Metadata.SourceChecksum)
	}
	if normalized.Markdown == "" || normalized.PlainText == "" {
		t.Fatalf("NormalizeContent() returned empty normalized content: %+v", normalized)
	}
	if normalized.Stats.MarkdownBytes == 0 || normalized.Stats.PlainTextBytes == 0 {
		t.Fatalf("NormalizeContent() returned incomplete stats: %+v", normalized.Stats)
	}

	for _, want := range []string{
		"| Flag",
		"| --help | Show help",
		"Architecture diagram",
		"Rendered pipeline overview.",
		"Embedded media | label: CLI walkthrough | source: /media/demo.mp4",
		fmt.Sprintf(codeBlockTruncationMessage, maxCodeBlockRunes),
	} {
		if !strings.Contains(normalized.Markdown, want) {
			t.Fatalf("NormalizeContent() markdown missing %q\nmarkdown=%s", want, normalized.Markdown)
		}
	}

	if strings.Contains(normalized.Markdown, strings.Repeat("x", maxCodeBlockRunes+10)) {
		t.Fatalf("NormalizeContent() did not truncate oversized code block")
	}
}

func TestNormalizeContentFallsBackToReadableText(t *testing.T) {
	t.Parallel()

	page := ExtractedPage{
		ID:           "page-2",
		SourceURL:    "https://docs.example.com/reference/plain",
		CanonicalURL: "https://docs.example.com/reference/plain",
		Title:        "Plain",
		ReadableText: "Single line fallback text",
		Metadata: ProcessingMetadata{
			SourceChecksum:   "source-checksum-2",
			ReadableChecksum: "readable-checksum-2",
		},
	}

	normalized, err := NormalizeContent(page)
	if err != nil {
		t.Fatalf("NormalizeContent() error = %v", err)
	}

	if !strings.Contains(normalized.Markdown, "Single line fallback text") {
		t.Fatalf("NormalizeContent() markdown missing readable text fallback: %q", normalized.Markdown)
	}
	if normalized.PlainText != "Single line fallback text" {
		t.Fatalf("NormalizeContent() plain text = %q, want %q", normalized.PlainText, "Single line fallback text")
	}
}

func TestApplyConservativeDedupe(t *testing.T) {
	t.Parallel()

	base := []NormalizedPage{
		{
			ID:        "page-a",
			SourceURL: "https://docs.example.com/a",
			Metadata: ProcessingMetadata{
				SourceChecksum: "same-source",
			},
			Markdown:  "# Install\nUse `cli-skill init`.",
			PlainText: "Install Use cli-skill init.",
		},
		{
			ID:        "page-b",
			SourceURL: "https://docs.example.com/b",
			Metadata: ProcessingMetadata{
				SourceChecksum: "same-source",
			},
			Markdown:  "# Install\nUse `cli-skill init`.",
			PlainText: "Install Use cli-skill init.",
		},
		{
			ID:        "page-c",
			SourceURL: "https://docs.example.com/c",
			Metadata:  ProcessingMetadata{SourceChecksum: "source-c"},
			Markdown:  "# Install\nStep one\nStep two",
			PlainText: "Install Step one Step two",
		},
		{
			ID:        "page-d",
			SourceURL: "https://docs.example.com/d",
			Metadata:  ProcessingMetadata{SourceChecksum: "source-d"},
			Markdown:  "# Install\nStep one\nStep three",
			PlainText: "Install Step one Step three",
		},
	}

	got := ApplyConservativeDedupe(base)
	if len(got) != len(base) {
		t.Fatalf("ApplyConservativeDedupe() len = %d, want %d", len(got), len(base))
	}

	if got[1].Deduped != true || got[1].DuplicateOf != "page-a" || got[1].DuplicateReason != DuplicateReasonSourceChecksumMatch {
		t.Fatalf("ApplyConservativeDedupe() duplicate by source checksum not recorded correctly: %+v", got[1])
	}
	if got[2].Deduped || got[3].Deduped {
		t.Fatalf("ApplyConservativeDedupe() over-deduped similar but distinct pages: page-c=%+v page-d=%+v", got[2], got[3])
	}

	normalizedMatch := ApplyConservativeDedupe([]NormalizedPage{
		{
			ID:        "page-e",
			SourceURL: "https://docs.example.com/e",
			Metadata:  ProcessingMetadata{SourceChecksum: "source-e"},
			Markdown:  "# API\nValue",
			PlainText: "API Value",
		},
		{
			ID:        "page-f",
			SourceURL: "https://docs.example.com/f",
			Metadata:  ProcessingMetadata{SourceChecksum: "source-f"},
			Markdown:  "# API\nValue",
			PlainText: "API Value",
		},
	})

	if !normalizedMatch[1].Deduped || normalizedMatch[1].DuplicateOf != "page-e" || normalizedMatch[1].DuplicateReason != DuplicateReasonNormalizedFormMatch {
		t.Fatalf("ApplyConservativeDedupe() normalized-form duplicate not recorded correctly: %+v", normalizedMatch[1])
	}
}
