package crawl

// CrawlResult captures the full outcome of a bounded crawl run.
type CrawlResult struct {
	EntryURL  string
	RootURL   string
	Summary   SummaryCounts
	Processed []PageRecord
	Skipped   []SkippedRecord
}

// SummaryCounts provides stable aggregate counters for final user reporting.
type SummaryCounts struct {
	Discovered int
	Processed  int
	Skipped    int
}

// PageRecord tracks a page that was accepted and processed by the crawler.
type PageRecord struct {
	URL          string
	CanonicalURL string
	Title        string
	Depth        int
}

// SkippedRecord tracks a candidate URL that was not processed and why.
type SkippedRecord struct {
	URL          string
	CanonicalURL string
	Reason       SkipReason
	Detail       string
	Depth        int
}
