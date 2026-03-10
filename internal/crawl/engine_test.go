package crawl

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestEngineSameDomainSkipReasonsAndCanonicalDedupe(t *testing.T) {
	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, "<html><body><main>external</main></body></html>")
	}))
	defer external.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/docs":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprintf(
				w,
				`<html><body>
					<a href="/docs/page?a=1&b=2">page canonical</a>
					<a href="/docs/page?b=2&a=1&utm_source=test#overview">page duplicate variant</a>
					<a href="%s/docs/outside">outside</a>
					<a href="/assets/manual.pdf">pdf asset</a>
					<a href="/api/schema">json endpoint</a>
					<a href="http://%%">broken</a>
				</body></html>`,
				external.URL,
			)
		case "/docs/page":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprint(w, "<html><body><main>doc page</main></body></html>")
		case "/api/schema":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"openapi":"3.1.0"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	result, err := ExecuteCrawl(server.URL + "/docs")
	if err != nil {
		t.Fatalf("ExecuteCrawl() error = %v", err)
	}

	assertSummaryIntegrity(t, result)

	if result.Summary.Processed != 2 {
		t.Fatalf("Processed = %d, want 2", result.Summary.Processed)
	}

	for _, page := range result.Processed {
		if !strings.HasPrefix(page.CanonicalURL, server.URL) {
			t.Fatalf("processed off-domain page: %+v", page)
		}
	}

	pageCanonical, err := CanonicalKey("/docs/page?a=1&b=2", mustNormalizeURL(t, server.URL+"/docs"))
	if err != nil {
		t.Fatalf("CanonicalKey() error = %v", err)
	}

	if processedCountForCanonical(result, pageCanonical) != 1 {
		t.Fatalf("processed count for canonical %q = %d, want 1", pageCanonical, processedCountForCanonical(result, pageCanonical))
	}

	wantReasons := map[SkipReason]int{
		SkipReasonAlreadySeen:        1,
		SkipReasonInvalidURL:         1,
		SkipReasonLowSignalPage:      1,
		SkipReasonNonHTMLContentType: 1,
		SkipReasonOffDomain:          1,
	}

	for reason, wantCount := range wantReasons {
		if got := countSkippedReason(result, reason); got != wantCount {
			t.Fatalf("skip count for %q = %d, want %d", reason, got, wantCount)
		}
	}
}

func TestEngineRespectsDefaultProcessedCap(t *testing.T) {
	const extraPages = 55

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if r.URL.Path == "/docs" {
			var body strings.Builder
			body.WriteString("<html><body>")
			for i := range extraPages {
				body.WriteString(fmt.Sprintf(`<a href="/docs/page-%d">page-%d</a>`, i, i))
			}
			body.WriteString("</body></html>")
			_, _ = fmt.Fprint(w, body.String())
			return
		}

		if strings.HasPrefix(r.URL.Path, "/docs/page-") {
			_, _ = fmt.Fprintf(w, "<html><body><main>%s</main></body></html>", r.URL.Path)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	result, err := ExecuteCrawl(server.URL + "/docs")
	if err != nil {
		t.Fatalf("ExecuteCrawl() error = %v", err)
	}

	assertSummaryIntegrity(t, result)

	if result.Summary.Processed != DefaultProcessedPageCap {
		t.Fatalf("Processed = %d, want %d", result.Summary.Processed, DefaultProcessedPageCap)
	}

	wantCapSkips := extraPages + 1 - DefaultProcessedPageCap
	if got := countSkippedReason(result, SkipReasonCapReached); got != wantCapSkips {
		t.Fatalf("cap_reached skips = %d, want %d", got, wantCapSkips)
	}
}

func assertSummaryIntegrity(t *testing.T, result CrawlResult) {
	t.Helper()

	if result.Summary.Discovered != result.Summary.Processed+result.Summary.Skipped {
		t.Fatalf(
			"Discovered = %d, want processed + skipped = %d",
			result.Summary.Discovered,
			result.Summary.Processed+result.Summary.Skipped,
		)
	}
}

func processedCountForCanonical(result CrawlResult, canonical string) int {
	count := 0
	for _, page := range result.Processed {
		if page.CanonicalURL == canonical {
			count++
		}
	}
	return count
}

func countSkippedReason(result CrawlResult, reason SkipReason) int {
	count := 0
	for _, skipped := range result.Skipped {
		if skipped.Reason == reason {
			count++
		}
	}
	return count
}

func mustNormalizeURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	normalized, err := NormalizeEntryURL(raw)
	if err != nil {
		t.Fatalf("NormalizeEntryURL() error = %v", err)
	}

	return normalized
}
