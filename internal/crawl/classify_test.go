package crawl

import "testing"

func TestClassifyDocsLikeHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "html with charset stays processable",
			contentType: "text/html; charset=utf-8",
			want:        true,
		},
		{
			name:        "xhtml stays processable",
			contentType: "application/xhtml+xml",
			want:        true,
		},
		{
			name:        "json is not docs-like html",
			contentType: "application/json",
			want:        false,
		},
		{
			name:        "missing header is not docs-like html",
			contentType: "",
			want:        false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := IsDocsLikeHTML(tc.contentType); got != tc.want {
				t.Fatalf("IsDocsLikeHTML(%q) = %t, want %t", tc.contentType, got, tc.want)
			}
		})
	}
}

func TestClassifyLowSignalPage(t *testing.T) {
	t.Parallel()

	base, err := NormalizeEntryURL("https://docs.example.com/docs/")
	if err != nil {
		t.Fatalf("NormalizeEntryURL(base) error = %v", err)
	}

	tests := []struct {
		name    string
		raw     string
		want    bool
		wantErr bool
	}{
		{
			name: "obvious image asset is low signal",
			raw:  "https://docs.example.com/assets/logo.svg",
			want: true,
		},
		{
			name: "relative javascript asset is low signal",
			raw:  "../static/app.js",
			want: true,
		},
		{
			name: "docs article path remains processable",
			raw:  "https://docs.example.com/docs/getting-started/install",
			want: false,
		},
		{
			name: "docs article with query remains processable",
			raw:  "../reference/auth?lang=en",
			want: false,
		},
		{
			name:    "invalid candidate returns error",
			raw:     "http://[::1",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := IsLowSignalPage(tc.raw, base)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("IsLowSignalPage(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("IsLowSignalPage(%q) error = %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("IsLowSignalPage(%q) = %t, want %t", tc.raw, got, tc.want)
			}
		})
	}
}

func TestClassifyCandidate(t *testing.T) {
	t.Parallel()

	base, err := NormalizeEntryURL("https://docs.example.com/docs/")
	if err != nil {
		t.Fatalf("NormalizeEntryURL(base) error = %v", err)
	}

	tests := []struct {
		name        string
		raw         string
		contentType string
		want        ClassificationOutcome
		wantErr     bool
	}{
		{
			name:        "non html content maps to explicit skip reason",
			raw:         "https://docs.example.com/docs/openapi.json",
			contentType: "application/json",
			want:        ClassificationOutcome{SkipReason: SkipReasonNonHTMLContentType},
		},
		{
			name:        "low signal asset maps to explicit skip reason",
			raw:         "https://docs.example.com/assets/logo.svg",
			contentType: "text/html; charset=utf-8",
			want:        ClassificationOutcome{SkipReason: SkipReasonLowSignalPage},
		},
		{
			name:        "docs article is accepted",
			raw:         "../guides/install",
			contentType: "text/html; charset=utf-8",
			want:        ClassificationOutcome{DocsLike: true},
		},
		{
			name:        "invalid low signal candidate bubbles error",
			raw:         "http://[::1",
			contentType: "text/html",
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ClassifyCandidate(tc.raw, tc.contentType, base)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ClassifyCandidate(%q, %q) error = nil, want error", tc.raw, tc.contentType)
				}
				return
			}
			if err != nil {
				t.Fatalf("ClassifyCandidate(%q, %q) error = %v", tc.raw, tc.contentType, err)
			}
			if got != tc.want {
				t.Fatalf("ClassifyCandidate(%q, %q) = %+v, want %+v", tc.raw, tc.contentType, got, tc.want)
			}
		})
	}
}
