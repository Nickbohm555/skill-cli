package crawl

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

const DefaultProcessedPageCap = 50

// EngineOptions configures bounded crawl execution.
type EngineOptions struct {
	ProcessedPageCap int
	HTTPClient       *http.Client
}

type crawlCandidate struct {
	Raw          string
	CanonicalURL string
	Depth        int
}

type crawlState struct {
	result      CrawlResult
	root        string
	rootFailure error

	processedCap int
	queue        []crawlCandidate
	seen         map[string]struct{}
	finalized    map[string]struct{}
	processable  map[string]struct{}
}

// ExecuteCrawl runs a same-domain bounded crawl using default engine options.
func ExecuteCrawl(entryURL string) (CrawlResult, error) {
	return ExecuteCrawlWithOptions(entryURL, EngineOptions{})
}

// ExecuteCrawlWithOptions runs a same-domain bounded crawl using synchronous
// collector behavior and centralized summary accounting.
func ExecuteCrawlWithOptions(entryURL string, options EngineOptions) (CrawlResult, error) {
	entry, err := NormalizeEntryURL(entryURL)
	if err != nil {
		return CrawlResult{}, fmt.Errorf("normalize entry url: %w", err)
	}

	root, err := DeriveDocsRoot(entry.String())
	if err != nil {
		return CrawlResult{}, fmt.Errorf("derive docs root: %w", err)
	}

	state := newCrawlState(entry.String(), root.String(), options.ProcessedPageCap)
	collector := colly.NewCollector(
		colly.Async(false),
		colly.AllowedDomains(root.Hostname()),
	)
	if options.HTTPClient != nil {
		collector.SetClient(options.HTTPClient)
	}

	collector.OnRequest(func(r *colly.Request) {
		candidate := candidateFromContext(r.Ctx)
		if candidate.CanonicalURL == "" {
			candidate = crawlCandidate{
				Raw:          r.URL.String(),
				CanonicalURL: r.URL.String(),
			}
		}

		sameDomain, sameDomainErr := SameDomain(r.URL.String(), root)
		switch {
		case sameDomainErr != nil:
			state.recordRequestSkip(candidate, SkipReasonInvalidURL, sameDomainErr.Error())
			r.Abort()
		case !sameDomain:
			state.recordRequestSkip(candidate, SkipReasonOffDomain, "candidate resolved outside crawl root domain")
			r.Abort()
		case state.processedLimitReached():
			state.recordRequestSkip(candidate, SkipReasonCapReached, fmt.Sprintf("processed page cap %d reached", state.processedCap))
			r.Abort()
		}
	})

	collector.OnResponse(func(r *colly.Response) {
		candidate := candidateFromContext(r.Ctx)
		if state.isFinalized(candidate.CanonicalURL) {
			return
		}

		outcome, classifyErr := ClassifyCandidate(r.Request.URL.String(), r.Headers.Get("Content-Type"), root)
		if classifyErr != nil {
			state.recordRequestSkip(candidate, SkipReasonInvalidURL, classifyErr.Error())
			return
		}
		if !outcome.DocsLike {
			state.recordRequestSkip(candidate, outcome.SkipReason, fmt.Sprintf("content type %q rejected", r.Headers.Get("Content-Type")))
			return
		}

		state.markProcessable(candidate.CanonicalURL)
		state.recordProcessed(candidate)
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		parent := candidateFromContext(e.Request.Ctx)
		if !state.isProcessable(parent.CanonicalURL) {
			return
		}

		state.enqueueCandidate(e.Attr("href"), e.Request.URL.String(), parent.Depth+1)
	})

	collector.OnError(func(r *colly.Response, err error) {
		if err == nil {
			return
		}

		candidate := crawlCandidate{}
		if r != nil && r.Request != nil {
			candidate = candidateFromContext(r.Request.Ctx)
			if candidate.CanonicalURL == "" && r.Request.URL != nil {
				candidate = crawlCandidate{
					Raw:          r.Request.URL.String(),
					CanonicalURL: r.Request.URL.String(),
				}
			}
		}

		if candidate.CanonicalURL == "" || state.isFinalized(candidate.CanonicalURL) {
			return
		}

		state.recordRequestSkip(candidate, SkipReasonFetchError, err.Error())
	})

	state.enqueueCandidate(root.String(), root.String(), 0)

	for len(state.queue) > 0 {
		candidate := state.popCandidate()
		if state.processedLimitReached() {
			state.recordRequestSkip(candidate, SkipReasonCapReached, fmt.Sprintf("processed page cap %d reached", state.processedCap))
			continue
		}

		ctx := colly.NewContext()
		ctx.Put("raw", candidate.Raw)
		ctx.Put("canonical", candidate.CanonicalURL)
		ctx.Put("depth", strconv.Itoa(candidate.Depth))

		if err := collector.Request(http.MethodGet, candidate.CanonicalURL, nil, ctx, nil); err != nil {
			if state.isFinalized(candidate.CanonicalURL) {
				continue
			}
			state.recordRequestSkip(candidate, SkipReasonFetchError, err.Error())
		}
	}

	if state.rootFailure != nil {
		return state.result, state.rootFailure
	}

	return state.result, nil
}

func newCrawlState(entryURL string, rootURL string, capValue int) *crawlState {
	if capValue <= 0 {
		capValue = DefaultProcessedPageCap
	}

	return &crawlState{
		result: CrawlResult{
			EntryURL: entryURL,
			RootURL:  rootURL,
		},
		root:         rootURL,
		processedCap: capValue,
		seen:         make(map[string]struct{}),
		finalized:    make(map[string]struct{}),
		processable:  make(map[string]struct{}),
	}
}

func (s *crawlState) enqueueCandidate(raw string, base string, depth int) {
	var baseURL *url.URL
	if strings.TrimSpace(base) != "" {
		var err error
		baseURL, err = NormalizeEntryURL(base)
		if err != nil {
			s.recordEncounterSkip(crawlCandidate{Raw: raw, Depth: depth}, SkipReasonInvalidURL, err.Error())
			return
		}
	}

	canonical, err := CanonicalKey(raw, baseURL)
	if err != nil {
		s.recordEncounterSkip(crawlCandidate{Raw: raw, Depth: depth}, SkipReasonInvalidURL, err.Error())
		return
	}

	candidate := crawlCandidate{
		Raw:          raw,
		CanonicalURL: canonical,
		Depth:        depth,
	}

	sameDomain, sameDomainErr := sameDomainWithBase(raw, baseURL, s.root)
	if sameDomainErr != nil {
		s.recordEncounterSkip(candidate, SkipReasonInvalidURL, sameDomainErr.Error())
		return
	}
	if !sameDomain {
		s.recordEncounterSkip(candidate, SkipReasonOffDomain, "candidate resolved outside crawl root domain")
		return
	}

	lowSignal, lowSignalErr := IsLowSignalPage(raw, baseURL)
	if lowSignalErr != nil {
		s.recordEncounterSkip(candidate, SkipReasonInvalidURL, lowSignalErr.Error())
		return
	}
	if lowSignal {
		s.recordEncounterSkip(candidate, SkipReasonLowSignalPage, "candidate matched low-signal path policy")
		return
	}

	if _, seen := s.seen[candidate.CanonicalURL]; seen {
		s.recordEncounterSkip(candidate, SkipReasonAlreadySeen, "canonical url already discovered")
		return
	}

	s.seen[candidate.CanonicalURL] = struct{}{}
	s.queue = append(s.queue, candidate)
}

func (s *crawlState) popCandidate() crawlCandidate {
	candidate := s.queue[0]
	s.queue = s.queue[1:]
	return candidate
}

func (s *crawlState) processedLimitReached() bool {
	return s.result.Summary.Processed >= s.processedCap
}

func (s *crawlState) isFinalized(canonical string) bool {
	_, ok := s.finalized[canonical]
	return ok
}

func (s *crawlState) markProcessable(canonical string) {
	if canonical == "" {
		return
	}
	s.processable[canonical] = struct{}{}
}

func (s *crawlState) isProcessable(canonical string) bool {
	_, ok := s.processable[canonical]
	return ok
}

func (s *crawlState) recordProcessed(candidate crawlCandidate) {
	if candidate.CanonicalURL == "" || s.isFinalized(candidate.CanonicalURL) {
		return
	}

	s.finalized[candidate.CanonicalURL] = struct{}{}
	s.result.Processed = append(s.result.Processed, PageRecord{
		URL:          candidate.Raw,
		CanonicalURL: candidate.CanonicalURL,
		Depth:        candidate.Depth,
	})
	s.result.Summary.Processed++
	s.result.Summary.Discovered++
}

func (s *crawlState) recordRequestSkip(candidate crawlCandidate, reason SkipReason, detail string) {
	if candidate.CanonicalURL == "" || s.isFinalized(candidate.CanonicalURL) {
		return
	}

	s.finalized[candidate.CanonicalURL] = struct{}{}
	s.appendSkip(candidate, reason, detail)

	if candidate.CanonicalURL == s.root && s.result.Summary.Processed == 0 && s.rootFailure == nil {
		s.rootFailure = rootFailure(reason, candidate.Raw, detail)
	}
}

func (s *crawlState) recordEncounterSkip(candidate crawlCandidate, reason SkipReason, detail string) {
	s.appendSkip(candidate, reason, detail)
}

func (s *crawlState) appendSkip(candidate crawlCandidate, reason SkipReason, detail string) {
	s.result.Skipped = append(s.result.Skipped, SkippedRecord{
		URL:          candidate.Raw,
		CanonicalURL: candidate.CanonicalURL,
		Reason:       reason,
		Detail:       detail,
		Depth:        candidate.Depth,
	})
	s.result.Summary.Skipped++
	s.result.Summary.Discovered++
}

func candidateFromContext(ctx *colly.Context) crawlCandidate {
	if ctx == nil {
		return crawlCandidate{}
	}

	depth, _ := strconv.Atoi(ctx.Get("depth"))
	return crawlCandidate{
		Raw:          ctx.Get("raw"),
		CanonicalURL: ctx.Get("canonical"),
		Depth:        depth,
	}
}

func rootFailure(reason SkipReason, raw string, detail string) error {
	var prefix string
	switch reason {
	case SkipReasonFetchError:
		prefix = "entry url could not be fetched"
	case SkipReasonNonHTMLContentType:
		prefix = "entry url did not return docs-like html"
	case SkipReasonLowSignalPage:
		prefix = "entry url resolved to a low-signal page"
	case SkipReasonInvalidURL:
		prefix = "entry url was invalid after normalization"
	default:
		prefix = "entry url could not be crawled"
	}

	detail = strings.TrimSpace(detail)
	if detail == "" {
		return errors.New(prefix)
	}

	return fmt.Errorf("%s: %s (%s)", prefix, raw, detail)
}

func sameDomainWithBase(raw string, base *url.URL, root string) (bool, error) {
	rootURL, err := NormalizeEntryURL(root)
	if err != nil {
		return false, err
	}

	if base != nil {
		canonical, err := normalizeURL(raw, base)
		if err != nil {
			return false, err
		}

		return SameDomain(canonical.String(), rootURL)
	}

	return SameDomain(raw, rootURL)
}
