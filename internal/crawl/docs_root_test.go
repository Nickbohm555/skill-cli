package crawl

import "testing"

func TestDeriveDocsRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{
			name: "top level docs root",
			raw:  "https://docs.example.com/docs/getting-started/install",
			want: "https://docs.example.com/docs",
		},
		{
			name: "nested documentation root keeps parent prefix",
			raw:  "https://www.example.com/developer/documentation/api/auth?lang=en",
			want: "https://www.example.com/developer/documentation",
		},
		{
			name: "guide root uses singular guide segment",
			raw:  "https://example.com/platform/guide/intro/",
			want: "https://example.com/platform/guide",
		},
		{
			name: "highest precedence docs marker wins deterministically",
			raw:  "https://example.com/learn/guide/docs/reference",
			want: "https://example.com/learn/guide/docs",
		},
		{
			name: "fallback to site root when no docs marker exists",
			raw:  "https://example.com/help/reference/index.html#overview",
			want: "https://example.com/",
		},
		{
			name:    "invalid entry url returns error",
			raw:     "://bad url",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := DeriveDocsRoot(tc.raw)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("DeriveDocsRoot(%q) error = nil, want error", tc.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("DeriveDocsRoot(%q) error = %v", tc.raw, err)
			}
			if got.String() != tc.want {
				t.Fatalf("DeriveDocsRoot(%q) = %q, want %q", tc.raw, got.String(), tc.want)
			}
		})
	}
}
