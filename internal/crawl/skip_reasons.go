package crawl

// SkipReason is the machine-readable reason a URL was skipped.
type SkipReason string

const (
	SkipReasonOffDomain          SkipReason = "off_domain"
	SkipReasonAlreadySeen        SkipReason = "already_seen"
	SkipReasonCapReached         SkipReason = "cap_reached"
	SkipReasonNonHTMLContentType SkipReason = "non_html_content_type"
	SkipReasonInvalidURL         SkipReason = "invalid_url"
	SkipReasonLowSignalPage      SkipReason = "low_signal_page"
	SkipReasonFetchError         SkipReason = "fetch_error"
)
