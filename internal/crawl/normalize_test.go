package crawl

import "testing"

func TestNormalizeEntryURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{
			name:    "invalid url parse failure",
			raw:     "://bad url",
			wantErr: true,
		},
		{
			name: "strips fragment and tracking query params",
			raw:  "HTTPS://Docs.Example.com/docs/intro/?utm_source=newsletter&fbclid=123&lang=en#overview",
			want: "https://docs.example.com/docs/intro?lang=en",
		},
		{
			name: "stabilizes query ordering",
			raw:  "https://docs.example.com/docs/api?z=last&b=beta&a=alpha",
			want: "https://docs.example.com/docs/api?a=alpha&b=beta&z=last",
		},
		{
			name: "cleans docs style path shapes",
			raw:  "https://docs.example.com/docs/reference/../guides//getting-started/./",
			want: "https://docs.example.com/docs/guides/getting-started",
		},
		{
			name: "drops default https port",
			raw:  "https://docs.example.com:443/docs/",
			want: "https://docs.example.com/docs",
		},
		{
			name: "normalizes empty path to root",
			raw:  "https://docs.example.com",
			want: "https://docs.example.com/",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeEntryURL(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NormalizeEntryURL(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeEntryURL(%q) error = %v", tc.raw, err)
			}
			if got.String() != tc.want {
				t.Fatalf("NormalizeEntryURL(%q) = %q, want %q", tc.raw, got.String(), tc.want)
			}
		})
	}
}

func TestCanonicalKeyNormalize(t *testing.T) {
	t.Parallel()

	base, err := NormalizeEntryURL("https://docs.example.com/docs/")
	if err != nil {
		t.Fatalf("NormalizeEntryURL(base) error = %v", err)
	}

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{
			name: "resolves relative link against docs base",
			raw:  "../api/./reference/?utm_medium=email&topic=auth#methods",
			want: "https://docs.example.com/api/reference?topic=auth",
		},
		{
			name: "canonical query ordering makes equivalent urls stable",
			raw:  "https://docs.example.com/docs/api?b=2&a=1",
			want: "https://docs.example.com/docs/api?a=1&b=2",
		},
		{
			name:    "invalid relative candidate bubbles parse error",
			raw:     "http://[::1",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := CanonicalKey(tc.raw, base)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("CanonicalKey(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("CanonicalKey(%q) error = %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("CanonicalKey(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestSameDomainNormalize(t *testing.T) {
	t.Parallel()

	base, err := NormalizeEntryURL("https://docs.example.com:443/docs/index.html")
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
			name: "same absolute host after normalization",
			raw:  "https://DOCS.example.com/docs/guides/start?utm_campaign=launch",
			want: true,
		},
		{
			name: "relative docs link stays in domain",
			raw:  "../reference/faq#top",
			want: true,
		},
		{
			name: "different subdomain is off domain",
			raw:  "https://blog.example.com/post",
			want: false,
		},
		{
			name: "different registrable domain is off domain",
			raw:  "https://example.org/docs",
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

			got, err := SameDomain(tc.raw, base)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("SameDomain(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("SameDomain(%q) error = %v", tc.raw, err)
			}
			if got != tc.want {
				t.Fatalf("SameDomain(%q) = %t, want %t", tc.raw, got, tc.want)
			}
		})
	}
}
