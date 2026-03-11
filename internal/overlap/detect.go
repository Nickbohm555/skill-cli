package overlap

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

func Detect(candidate SkillProfile, index InstalledIndex) OverlapReport {
	report := NewReport(candidate)

	for _, warning := range index.Warnings {
		report.AddWarning(warning)
	}

	profiles := slices.Clone(index.Profiles)
	sort.SliceStable(profiles, func(i, j int) bool {
		left := normalizedProfile(profiles[i])
		right := normalizedProfile(profiles[j])
		if left.SourcePath != right.SourcePath {
			return left.SourcePath < right.SourcePath
		}
		if left.ID != right.ID {
			return left.ID < right.ID
		}
		return left.Name < right.Name
	})

	normalizedCandidate := normalizedProfile(candidate)
	for _, existing := range profiles {
		finding, ok := detectFinding(normalizedCandidate, normalizedProfile(existing))
		if !ok {
			continue
		}
		report.AddFinding(finding)
	}

	return report
}

func detectFinding(candidate, existing SkillProfile) (OverlapFinding, bool) {
	if candidate.SourcePath != "" && candidate.SourcePath == existing.SourcePath {
		return OverlapFinding{
			RuleID:          "OVLP.PATH.EXACT",
			ExistingSkillID: existing.ID,
			Severity:        SeverityHigh,
			Score:           1,
			Signals: []OverlapSignal{
				{Key: "path_exact", Value: existing.SourcePath, Score: 1},
			},
			Explanation: fmt.Sprintf(
				"Candidate source path exactly matches installed skill path %q.",
				existing.SourcePath,
			),
			ExplanationMeta: ExplanationMetadata{
				Summary: "Exact source-path collision with an installed skill.",
				RuleIDs: []string{"OVLP.PATH.EXACT"},
				Signals: []OverlapSignal{
					{Key: "path_exact", Value: existing.SourcePath, Score: 1},
				},
			},
		}, true
	}

	if profilesExactlyMatch(candidate, existing) {
		signals := []OverlapSignal{
			{Key: "content_exact", Value: existing.ID, Score: 1},
		}
		if candidate.Name != "" && candidate.Name == existing.Name {
			signals = append(signals, OverlapSignal{Key: "name_exact", Value: candidate.Name, Score: 1})
		}
		if len(candidate.InScope) > 0 || len(existing.InScope) > 0 {
			signals = append(signals, OverlapSignal{Key: "in_scope_jaccard", Value: "1.00", Score: 1})
		}
		if len(candidate.Commands) > 0 || len(existing.Commands) > 0 {
			signals = append(signals, OverlapSignal{Key: "command_overlap", Value: "1.00", Score: 1})
		}

		return OverlapFinding{
			RuleID:          "OVLP.CONTENT.EXACT",
			ExistingSkillID: existing.ID,
			Severity:        SeverityHigh,
			Score:           1,
			Signals:         signals,
			Explanation:     "Candidate skill exactly matches an installed skill after normalization.",
			ExplanationMeta: ExplanationMetadata{
				Summary: "Candidate and installed skill content are identical after normalization.",
				RuleIDs: []string{"OVLP.CONTENT.EXACT"},
				Signals: signals,
			},
		}, true
	}

	if candidate.Name != "" && candidate.Name == existing.Name {
		signals := []OverlapSignal{
			{Key: "name_exact", Value: candidate.Name, Score: 1},
		}
		return OverlapFinding{
			RuleID:          "OVLP.NAME.EXACT",
			ExistingSkillID: existing.ID,
			Severity:        SeverityHigh,
			Score:           1,
			Signals:         signals,
			Explanation:     "Candidate name exactly matches an installed skill name.",
			ExplanationMeta: ExplanationMetadata{
				Summary: "Exact collision on skill name.",
				RuleIDs: []string{"OVLP.NAME.EXACT"},
				Signals: signals,
			},
		}, true
	}

	descScore := descriptionSimilarity(candidate.Description, existing.Description)
	inScopeScore := listJaccard(candidate.InScope, existing.InScope)
	outOfScopeScore := outOfScopeConflict(candidate, existing)
	commandScore := listJaccard(candidate.Commands, existing.Commands)

	score := weightedScore(descScore, inScopeScore, outOfScopeScore, commandScore)
	severity := ClassifyScore(score)
	if severity == SeverityNone {
		return OverlapFinding{}, false
	}

	signals := make([]OverlapSignal, 0, 4)
	if descScore > 0 {
		signals = append(signals, OverlapSignal{
			Key:   "description_similarity",
			Value: formatScore(descScore),
			Score: descScore,
		})
	}
	if inScopeScore > 0 {
		signals = append(signals, OverlapSignal{
			Key:   "in_scope_jaccard",
			Value: formatScore(inScopeScore),
			Score: inScopeScore,
		})
	}
	if outOfScopeScore > 0 {
		signals = append(signals, OverlapSignal{
			Key:   "out_of_scope_conflict",
			Value: formatScore(outOfScopeScore),
			Score: outOfScopeScore,
		})
	}
	if commandScore > 0 {
		signals = append(signals, OverlapSignal{
			Key:   "command_overlap",
			Value: formatScore(commandScore),
			Score: commandScore,
		})
	}

	explanation := "Candidate overlaps an installed skill across normalized description, scope, or command signals."
	switch severity {
	case SeverityHigh:
		explanation = "Candidate strongly overlaps an installed skill across normalized description, scope, or command signals."
	case SeverityMedium:
		explanation = "Candidate moderately overlaps an installed skill and requires explicit resolution."
	case SeverityLow:
		explanation = "Candidate has limited but detectable overlap with an installed skill."
	}

	return OverlapFinding{
		RuleID:          "OVLP.STRUCTURAL.OVERLAP",
		ExistingSkillID: existing.ID,
		Severity:        severity,
		Score:           score,
		Signals:         signals,
		Explanation:     explanation,
		ExplanationMeta: ExplanationMetadata{
			Summary: "Weighted overlap classification from description, in-scope, out-of-scope, and command signals.",
			RuleIDs: []string{"OVLP.STRUCTURAL.OVERLAP"},
			Signals: signals,
		},
	}, true
}

func normalizedProfile(profile SkillProfile) SkillProfile {
	return SkillProfile{
		ID:          normalizeComparableText(profile.ID),
		Name:        normalizeComparableText(profile.Name),
		Description: normalizeComparableText(profile.Description),
		InScope:     normalizeList(profile.InScope),
		OutOfScope:  normalizeList(profile.OutOfScope),
		Commands:    normalizeList(profile.Commands),
		SourcePath:  normalizeSourcePath(profile.SourcePath),
	}
}

func profilesExactlyMatch(left, right SkillProfile) bool {
	if !hasComparableContent(left) || !hasComparableContent(right) {
		return false
	}
	return left.Name == right.Name &&
		left.Description == right.Description &&
		slices.Equal(left.InScope, right.InScope) &&
		slices.Equal(left.OutOfScope, right.OutOfScope) &&
		slices.Equal(left.Commands, right.Commands)
}

func hasComparableContent(profile SkillProfile) bool {
	return profile.Description != "" ||
		len(profile.InScope) > 0 ||
		len(profile.OutOfScope) > 0 ||
		len(profile.Commands) > 0
}

func descriptionSimilarity(left, right string) float64 {
	return textJaccard(left, right)
}

func outOfScopeConflict(candidate, existing SkillProfile) float64 {
	return maxFloat(
		listJaccard(candidate.InScope, existing.OutOfScope),
		listJaccard(candidate.OutOfScope, existing.InScope),
	)
}

func weightedScore(descScore, inScopeScore, outOfScopeScore, commandScore float64) float64 {
	return clampScore(
		(descScore * descriptionWeight) +
			(inScopeScore * inScopeJaccardWeight) +
			(outOfScopeScore * outOfScopeWeight) +
			(commandScore * commandOverlapWeight),
	)
}

func textJaccard(left, right string) float64 {
	return setJaccard(tokenSet(left), tokenSet(right))
}

func listJaccard(left, right []string) float64 {
	return setJaccard(left, right)
}

func setJaccard(left, right []string) float64 {
	if len(left) == 0 && len(right) == 0 {
		return 0
	}

	leftSet := make(map[string]struct{}, len(left))
	rightSet := make(map[string]struct{}, len(right))
	for _, item := range left {
		if item == "" {
			continue
		}
		leftSet[item] = struct{}{}
	}
	for _, item := range right {
		if item == "" {
			continue
		}
		rightSet[item] = struct{}{}
	}

	if len(leftSet) == 0 && len(rightSet) == 0 {
		return 0
	}

	intersection := 0
	union := len(leftSet)
	for item := range rightSet {
		if _, exists := leftSet[item]; exists {
			intersection++
			continue
		}
		union++
	}
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func tokenSet(input string) []string {
	fields := strings.FieldsFunc(normalizeComparableText(input), func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return false
		case r >= '0' && r <= '9':
			return false
		}
		return true
	})
	if len(fields) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(fields))
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		tokens = append(tokens, field)
	}
	sort.Strings(tokens)
	return tokens
}

func clampScore(score float64) float64 {
	switch {
	case score < 0:
		return 0
	case score > 1:
		return 1
	default:
		return score
	}
}

func maxFloat(left, right float64) float64 {
	if right > left {
		return right
	}
	return left
}

func formatScore(score float64) string {
	return fmt.Sprintf("%.2f", score)
}
