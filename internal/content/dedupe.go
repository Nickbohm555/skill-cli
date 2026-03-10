package content

import "strings"

const (
	DuplicateReasonSourceChecksumMatch = "source_checksum_match"
	DuplicateReasonNormalizedFormMatch = "normalized_form_match"
)

// ApplyConservativeDedupe suppresses only high-confidence duplicates: exact
// source checksum matches or exact strict-normalized content matches.
func ApplyConservativeDedupe(pages []NormalizedPage) []NormalizedPage {
	if len(pages) == 0 {
		return nil
	}

	deduped := make([]NormalizedPage, len(pages))
	copy(deduped, pages)

	seenBySourceChecksum := make(map[string]string, len(pages))
	seenByNormalizedForm := make(map[string]string, len(pages))

	for i := range deduped {
		page := deduped[i]
		page.Deduped = false
		page.DuplicateOf = ""
		page.DuplicateReason = ""

		sourceChecksum := strings.TrimSpace(page.Metadata.SourceChecksum)
		if sourceChecksum != "" {
			if originalID, exists := seenBySourceChecksum[sourceChecksum]; exists {
				page.Deduped = true
				page.DuplicateOf = originalID
				page.DuplicateReason = DuplicateReasonSourceChecksumMatch
				deduped[i] = page
				continue
			}
		}

		normalizedChecksum := StrictNormalizedChecksum(page)
		if normalizedChecksum != "" {
			if originalID, exists := seenByNormalizedForm[normalizedChecksum]; exists {
				page.Deduped = true
				page.DuplicateOf = originalID
				page.DuplicateReason = DuplicateReasonNormalizedFormMatch
				deduped[i] = page
				continue
			}
		}

		if sourceChecksum != "" {
			seenBySourceChecksum[sourceChecksum] = page.ID
		}
		if normalizedChecksum != "" {
			seenByNormalizedForm[normalizedChecksum] = page.ID
		}

		deduped[i] = page
	}

	return deduped
}

// StrictNormalizedChecksum returns the checksum key for the exact normalized
// content form used by conservative duplicate suppression.
func StrictNormalizedChecksum(page NormalizedPage) string {
	markdown := normalizeForDedupe(page.Markdown)
	plainText := normalizeForDedupe(page.PlainText)
	if markdown == "" && plainText == "" {
		return ""
	}

	return checksum(markdown + "\n---\n" + plainText)
}

func normalizeForDedupe(input string) string {
	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(input), "\r\n", "\n"), "\n")
	cleaned := make([]string, 0, len(lines))
	blankCount := 0

	for _, line := range lines {
		trimmed := strings.TrimRight(strings.Join(strings.Fields(line), " "), " ")
		if trimmed == "" {
			blankCount++
			if blankCount > 1 {
				continue
			}
			cleaned = append(cleaned, "")
			continue
		}

		blankCount = 0
		cleaned = append(cleaned, trimmed)
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}
