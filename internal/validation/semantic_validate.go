package validation

import (
	"strconv"
	"strings"
)

const (
	ruleInScopeEntryTooBrief    = "VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF"
	ruleInScopeCatchAll         = "VAL.SCOPE.IN_SCOPE_VAGUE_CATCH_ALL"
	ruleOutOfScopeEntryTooBrief = "VAL.SCOPE.OUT_OF_SCOPE_ENTRY_TOO_BRIEF"
	ruleOutOfScopeCatchAll      = "VAL.SCOPE.OUT_OF_SCOPE_VAGUE_CATCH_ALL"
)

type scopeRule struct {
	ruleID   string
	path     string
	message  string
	priority int
}

var vagueCatchAllPhrases = []string{
	"and more",
	"anything else",
	"anything related",
	"etc",
	"everything else",
	"general help",
	"miscellaneous",
	"other stuff",
	"various topics",
	"whatever you need",
}

func ValidateSemantic(candidate CandidateSkill) ValidationReport {
	report := NewReport()
	validateScopeEntries(&report, candidate.InScope.Items, "sections.in_scope", scopeRule{
		ruleID:   ruleInScopeEntryTooBrief,
		message:  "In Scope entries must be specific enough to define a concrete boundary.",
		priority: 130,
	}, scopeRule{
		ruleID:   ruleInScopeCatchAll,
		message:  "In Scope entries must avoid vague catch-all phrasing.",
		priority: 140,
	})
	validateScopeEntries(&report, candidate.OutOfScope.Items, "sections.out_of_scope", scopeRule{
		ruleID:   ruleOutOfScopeEntryTooBrief,
		message:  "Out Of Scope entries must be specific enough to define a concrete boundary.",
		priority: 150,
	}, scopeRule{
		ruleID:   ruleOutOfScopeCatchAll,
		message:  "Out Of Scope entries must avoid vague catch-all phrasing.",
		priority: 160,
	})
	return report
}

func validateScopeEntries(report *ValidationReport, items []string, basePath string, briefRule scopeRule, catchAllRule scopeRule) {
	for index, raw := range items {
		entry := strings.TrimSpace(raw)
		if entry == "" {
			continue
		}

		path := indexedScopePath(basePath, index)
		if isNonTrivialScopeEntry(entry) {
			if containsVagueCatchAll(entry) {
				report.AddIssue(ValidationIssue{
					RuleID:   catchAllRule.ruleID,
					Severity: SeverityError,
					Path:     path,
					Message:  catchAllRule.message,
					Priority: catchAllRule.priority,
				})
			}
			continue
		}

		report.AddIssue(ValidationIssue{
			RuleID:   briefRule.ruleID,
			Severity: SeverityError,
			Path:     path,
			Message:  briefRule.message,
			Priority: briefRule.priority,
		})
	}
}

func isNonTrivialScopeEntry(entry string) bool {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(entry)))
	if len(words) < 3 {
		return false
	}

	substantive := 0
	for _, word := range words {
		if len(strings.Trim(word, ".,:;()[]{}\"'`")) >= 4 {
			substantive++
		}
	}

	return substantive >= 2
}

func containsVagueCatchAll(entry string) bool {
	normalized := normalizeScopeEntry(entry)
	for _, phrase := range vagueCatchAllPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func normalizeScopeEntry(entry string) string {
	lowered := strings.ToLower(strings.TrimSpace(entry))
	replacer := strings.NewReplacer(
		".", " ",
		",", " ",
		";", " ",
		":", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"/", " ",
		"-", " ",
		"_", " ",
		"\n", " ",
		"\t", " ",
	)
	return strings.Join(strings.Fields(replacer.Replace(lowered)), " ")
}

func indexedScopePath(basePath string, index int) string {
	return basePath + ".items[" + strconv.Itoa(index) + "]"
}
