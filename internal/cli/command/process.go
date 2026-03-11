package command

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Nickbohm555/skill-cli/internal/content"
	"github.com/Nickbohm555/skill-cli/internal/crawl"
	"github.com/spf13/cobra"
)

const (
	processFetchTimeout = 20 * time.Second
	processRunTimeout   = 2 * time.Minute
	rawExcerptRunes     = 400
)

type processReport struct {
	Crawl      crawl.CrawlResult
	Review     content.ReviewView
	Warnings   []string
	PageCount  int
	Deduped    int
	ChunkCount int
}

func newProcessCommand() *cobra.Command {
	var entryURL string
	var includeRaw bool

	cmd := &cobra.Command{
		Use:   "process",
		Short: "Process crawled documentation into summarized review chunks",
		RunE: func(cmd *cobra.Command, args []string) error {
			entryURL = strings.TrimSpace(entryURL)
			if entryURL == "" {
				return fmt.Errorf("missing required --url value")
			}

			baseCtx := cmd.Context()
			if baseCtx == nil {
				baseCtx = context.Background()
			}

			ctx, cancel := context.WithTimeout(baseCtx, processRunTimeout)
			defer cancel()

			report, err := runProcess(ctx, entryURL)
			if err != nil {
				return fmt.Errorf("process failed: %w", err)
			}

			renderProcessReport(cmd.OutOrStdout(), report, includeRaw)
			renderProcessWarnings(cmd.ErrOrStderr(), report.Warnings)
			return nil
		},
	}

	cmd.Flags().StringVar(&entryURL, "url", "", "Documentation entry URL to process")
	cmd.Flags().BoolVar(&includeRaw, "include-raw", false, "Include raw chunk excerpts in the review output")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

func runProcess(ctx context.Context, entryURL string) (processReport, error) {
	crawlResult, err := crawl.ExecuteCrawl(entryURL)
	if err != nil {
		return processReport{}, err
	}
	if len(crawlResult.Processed) == 0 {
		return processReport{}, fmt.Errorf("crawl produced no processable pages")
	}

	pages, warnings, err := fetchProcessedPages(ctx, crawlResult.Processed)
	if err != nil {
		return processReport{}, err
	}

	normalizedPages := make([]content.NormalizedPage, 0, len(pages))
	for _, page := range pages {
		extracted, err := content.ExtractReadable(page)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: extract failed: %v", page.URL, err))
			continue
		}

		normalized, err := content.NormalizeContent(extracted)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: normalize failed: %v", extracted.SourceURL, err))
			continue
		}

		normalizedPages = append(normalizedPages, normalized)
	}

	if len(normalizedPages) == 0 {
		return processReport{}, fmt.Errorf("content processing produced no normalized pages")
	}

	dedupedPages := content.ApplyConservativeDedupe(normalizedPages)
	dedupedCount := 0
	for _, page := range dedupedPages {
		if page.Deduped {
			dedupedCount++
		}
	}

	chunks, err := content.ProcessToChunks(dedupedPages)
	if err != nil {
		return processReport{}, err
	}
	if len(chunks) == 0 {
		return processReport{}, fmt.Errorf("content processing produced no attributed chunks")
	}

	summaries, err := content.SummarizeChunks(ctx, chunks)
	if err != nil {
		return processReport{}, err
	}

	review, err := content.BuildReviewView(summaries, chunks)
	if err != nil {
		return processReport{}, err
	}

	return processReport{
		Crawl:      crawlResult,
		Review:     review,
		Warnings:   warnings,
		PageCount:  len(normalizedPages),
		Deduped:    dedupedCount,
		ChunkCount: len(chunks),
	}, nil
}

func fetchProcessedPages(ctx context.Context, processed []crawl.PageRecord) ([]content.CrawledPage, []string, error) {
	client := &http.Client{Timeout: processFetchTimeout}
	pages := make([]content.CrawledPage, 0, len(processed))
	warnings := make([]string, 0)

	for _, record := range processed {
		targetURL := strings.TrimSpace(record.CanonicalURL)
		if targetURL == "" {
			targetURL = strings.TrimSpace(record.URL)
		}
		if targetURL == "" {
			warnings = append(warnings, "skipped processed page with empty URL")
			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: build request failed: %v", targetURL, err))
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: fetch failed: %v", targetURL, err))
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: read body failed: %v", targetURL, readErr))
			continue
		}
		if closeErr != nil {
			warnings = append(warnings, fmt.Sprintf("%s: close body failed: %v", targetURL, closeErr))
		}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			warnings = append(warnings, fmt.Sprintf("%s: unexpected status %s", targetURL, resp.Status))
			continue
		}

		pages = append(pages, content.CrawledPage{
			URL:          targetURL,
			CanonicalURL: targetURL,
			Title:        strings.TrimSpace(record.Title),
			ContentType:  strings.TrimSpace(resp.Header.Get("Content-Type")),
			HTML:         string(body),
			Depth:        record.Depth,
		})
	}

	if len(pages) == 0 {
		return nil, warnings, fmt.Errorf("could not fetch any processed pages")
	}

	return pages, warnings, nil
}

func renderProcessReport(w io.Writer, report processReport, includeRaw bool) {
	fmt.Fprintf(w, "Entry URL: %s\n", report.Crawl.EntryURL)
	fmt.Fprintf(w, "Docs root: %s\n", report.Crawl.RootURL)
	fmt.Fprintf(w, "Crawl pages: %d processed / %d skipped\n", len(report.Crawl.Processed), len(report.Crawl.Skipped))
	fmt.Fprintf(w, "Content pages: %d normalized / %d deduped\n", report.PageCount, report.Deduped)
	fmt.Fprintf(w, "Review chunks: %d summaries / %d attributed chunks\n", len(report.Review.Chunks), report.ChunkCount)
	fmt.Fprintln(w, "Summary-first review:")

	for index, chunk := range report.Review.Chunks {
		fmt.Fprintf(w, "[%d] %s\n", index+1, chunk.ChunkID)
		fmt.Fprintf(w, "  source_url: %s\n", chunk.SourceURL)
		fmt.Fprintf(w, "  summary: %s\n", formatIndentedMultiline(chunk.Summary, "           "))
		if confidence := strings.TrimSpace(chunk.Confidence); confidence != "" {
			fmt.Fprintf(w, "  confidence: %s\n", confidence)
		}
		if notes := strings.TrimSpace(chunk.Notes); notes != "" {
			fmt.Fprintf(w, "  notes: %s\n", formatIndentedMultiline(notes, "         "))
		}
		fmt.Fprintf(w, "  expand_target: %s\n", chunk.ExpandTarget.Key)
		fmt.Fprintf(w, "  reference: %s\n", chunk.Attribution.Reference)

		if includeRaw {
			raw := report.Review.Expansions[chunk.ExpandTarget.Key]
			fmt.Fprintf(w, "  raw_excerpt: %s\n", formatIndentedMultiline(excerptRunes(raw.Text, rawExcerptRunes), "               "))
		}
	}
}

func renderProcessWarnings(w io.Writer, warnings []string) {
	if len(warnings) == 0 {
		return
	}

	fmt.Fprintln(w, "Warnings:")
	for _, warning := range warnings {
		fmt.Fprintf(w, "- %s\n", warning)
	}
}

func formatIndentedMultiline(input string, indent string) string {
	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(input), "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, line := range lines {
		if i > 0 {
			builder.WriteString("\n")
			builder.WriteString(indent)
		}
		builder.WriteString(strings.TrimSpace(line))
	}

	return builder.String()
}

func excerptRunes(input string, limit int) string {
	input = strings.TrimSpace(input)
	if input == "" || limit <= 0 {
		return input
	}
	if utf8.RuneCountInString(input) <= limit {
		return input
	}

	runes := []rune(input)
	if limit == 1 {
		return string(runes[:1])
	}

	return strings.TrimSpace(string(runes[:limit-1])) + "…"
}
