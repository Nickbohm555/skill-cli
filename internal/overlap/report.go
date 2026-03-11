package overlap

import "sort"

type IndexWarning struct {
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type OverlapReport struct {
	Candidate       SkillProfile                `json:"candidate"`
	OverallSeverity OverlapSeverity             `json:"overall_severity"`
	Findings        []OverlapFinding            `json:"findings"`
	Warnings        []IndexWarning              `json:"warnings,omitempty"`
	Decision        *ConflictResolutionDecision `json:"decision,omitempty"`
}

func NewReport(candidate SkillProfile) OverlapReport {
	return OverlapReport{
		Candidate:       candidate,
		OverallSeverity: SeverityNone,
		Findings:        make([]OverlapFinding, 0),
		Warnings:        make([]IndexWarning, 0),
	}
}

func (r *OverlapReport) AddFinding(finding OverlapFinding) {
	r.Findings = append(r.Findings, finding)
	r.SortFindings()
	r.OverallSeverity = maxSeverity(r.OverallSeverity, highestSeverity(r.Findings))
}

func (r *OverlapReport) AddWarning(warning IndexWarning) {
	r.Warnings = append(r.Warnings, warning)
	sort.SliceStable(r.Warnings, func(i, j int) bool {
		left := r.Warnings[i]
		right := r.Warnings[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		return left.Message < right.Message
	})
}

func (r *OverlapReport) SortFindings() {
	sort.SliceStable(r.Findings, func(i, j int) bool {
		left := r.Findings[i]
		right := r.Findings[j]
		if severityRank(left.Severity) != severityRank(right.Severity) {
			return severityRank(left.Severity) < severityRank(right.Severity)
		}
		if left.RuleID != right.RuleID {
			return left.RuleID < right.RuleID
		}
		if left.ExistingSkillID != right.ExistingSkillID {
			return left.ExistingSkillID < right.ExistingSkillID
		}
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		return left.Explanation < right.Explanation
	})
}

func (d ConflictResolutionDecision) IsResolved() bool {
	if d.Blocking {
		return false
	}
	switch d.Mode {
	case ResolutionNewInstall, ResolutionUpdate, ResolutionMerge:
		return d.SelectedAt != nil
	default:
		return false
	}
}

func highestSeverity(findings []OverlapFinding) OverlapSeverity {
	severity := SeverityNone
	for _, finding := range findings {
		severity = maxSeverity(severity, finding.Severity)
	}
	return severity
}

func maxSeverity(left, right OverlapSeverity) OverlapSeverity {
	if severityRank(right) < severityRank(left) {
		return right
	}
	return left
}

func severityRank(severity OverlapSeverity) int {
	switch severity {
	case SeverityHigh:
		return 0
	case SeverityMedium:
		return 1
	case SeverityLow:
		return 2
	case SeverityNone:
		return 3
	default:
		return 4
	}
}
