package crawl

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
)

var defaultTrackingQueryKeys = map[string]struct{}{
	"gclid":  {},
	"fbclid": {},
}

// NormalizeEntryURL canonicalizes the user-provided entry URL for crawl startup.
func NormalizeEntryURL(raw string) (*url.URL, error) {
	normalized, err := normalizeURL(raw, nil)
	if err != nil {
		return nil, err
	}

	return normalized, nil
}

// CanonicalKey returns the stable dedupe key for a candidate URL.
func CanonicalKey(raw string, base *url.URL) (string, error) {
	normalized, err := normalizeURL(raw, base)
	if err != nil {
		return "", err
	}

	return normalized.String(), nil
}

// SameDomain reports whether a candidate resolves onto the same normalized host as the crawl root.
func SameDomain(raw string, base *url.URL) (bool, error) {
	if base == nil {
		return false, fmt.Errorf("same-domain check requires a base URL")
	}

	normalizedBase, err := normalizeURL(base.String(), nil)
	if err != nil {
		return false, err
	}

	normalizedCandidate, err := normalizeURL(raw, normalizedBase)
	if err != nil {
		return false, err
	}

	return normalizedHost(normalizedCandidate) == normalizedHost(normalizedBase), nil
}

// normalizeURL centralizes URL policy for the crawl boundary:
// fragments are always removed, known tracking query params are dropped,
// remaining query params are preserved in stable encoded order, and paths are cleaned.
func normalizeURL(raw string, base *url.URL) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse url %q: %w", raw, err)
	}

	if base != nil {
		parsed = base.ResolveReference(parsed)
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("url %q must resolve to an absolute URL", raw)
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = normalizeHostPort(parsed)
	parsed.Fragment = ""
	parsed.Path = normalizePath(parsed.EscapedPath())
	parsed.RawPath = ""
	parsed.RawQuery = normalizeQuery(parsed.Query())

	return parsed, nil
}

func normalizeHostPort(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if isDefaultPort(u.Scheme, port) || port == "" {
		return host
	}

	return net.JoinHostPort(host, port)
}

func normalizePath(rawPath string) string {
	if rawPath == "" {
		return "/"
	}

	cleaned := path.Clean(rawPath)
	if cleaned == "." {
		return "/"
	}
	if !strings.HasPrefix(cleaned, "/") {
		return "/" + cleaned
	}

	return cleaned
}

func normalizeQuery(values url.Values) string {
	filtered := url.Values{}
	for key, entries := range values {
		if isTrackingQueryKey(key) {
			continue
		}
		for _, entry := range entries {
			filtered.Add(key, entry)
		}
	}

	return filtered.Encode()
}

func isTrackingQueryKey(key string) bool {
	lowerKey := strings.ToLower(key)
	if strings.HasPrefix(lowerKey, "utm_") {
		return true
	}

	_, found := defaultTrackingQueryKeys[lowerKey]
	return found
}

func isDefaultPort(scheme, port string) bool {
	return (strings.EqualFold(scheme, "http") && port == "80") ||
		(strings.EqualFold(scheme, "https") && port == "443")
}

func normalizedHost(u *url.URL) string {
	return normalizeHostPort(u)
}
