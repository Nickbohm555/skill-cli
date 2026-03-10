package command

import (
	"fmt"
	"io"
	"strings"

	"github.com/Nickbohm555/skill-cli/internal/crawl"
	"github.com/spf13/cobra"
)

// NewRootCommand builds the top-level CLI command tree.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "cli-skill",
		Short:         "Generate Codex skills from documentation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newCrawlCommand())
	return rootCmd
}

func newCrawlCommand() *cobra.Command {
	var entryURL string

	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Crawl same-domain documentation pages from an entry URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			entryURL = strings.TrimSpace(entryURL)
			if entryURL == "" {
				return fmt.Errorf("missing required --url value")
			}

			result, err := crawl.ExecuteCrawl(entryURL)
			if err != nil {
				return fmt.Errorf("crawl failed: %w", err)
			}

			renderCrawlReport(cmd.OutOrStdout(), result)
			return nil
		},
	}

	cmd.Flags().StringVar(&entryURL, "url", "", "Documentation entry URL to crawl")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

func renderCrawlReport(w io.Writer, result crawl.CrawlResult) {
	fmt.Fprintf(w, "Entry URL: %s\n", result.EntryURL)
	fmt.Fprintf(w, "Docs root: %s\n", result.RootURL)
	fmt.Fprintf(w, "Processed pages (%d):\n", len(result.Processed))
	for _, page := range result.Processed {
		fmt.Fprintf(w, "- [%d] %s\n", page.Depth, page.CanonicalURL)
	}

	fmt.Fprintf(w, "Skipped pages (%d):\n", len(result.Skipped))
	for _, skipped := range result.Skipped {
		fmt.Fprintf(
			w,
			"- [%d] %s (%s)",
			skipped.Depth,
			displayURL(skipped.CanonicalURL, skipped.URL),
			skipped.Reason,
		)
		if detail := strings.TrimSpace(skipped.Detail); detail != "" {
			fmt.Fprintf(w, " - %s", detail)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w, "Summary:")
	fmt.Fprintf(w, "- Discovered: %d\n", result.Summary.Discovered)
	fmt.Fprintf(w, "- Processed: %d\n", result.Summary.Processed)
	fmt.Fprintf(w, "- Skipped: %d\n", result.Summary.Skipped)
}

func displayURL(canonical string, raw string) string {
	canonical = strings.TrimSpace(canonical)
	if canonical != "" {
		return canonical
	}

	raw = strings.TrimSpace(raw)
	if raw != "" {
		return raw
	}

	return "<unknown>"
}
