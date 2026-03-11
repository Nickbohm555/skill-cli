package validation

import "sort"

type Severity string

const (
	SeverityError   Severity = "Error"
	SeverityWarning Severity = "Warning"
)

type ValidationIssue struct {
	RuleID   string   `json:"rule_id"`
	Severity Severity `json:"severity"`
	Path     string   `json:"path,omitempty"`
	Message  string   `json:"message"`
	Priority int      `json:"priority,omitempty"`
}

type ValidationReport struct {
	Issues []ValidationIssue `json:"issues"`
}

func NewReport() ValidationReport {
	return ValidationReport{Issues: make([]ValidationIssue, 0)}
}

func (r *ValidationReport) AddIssue(issue ValidationIssue) {
	r.Issues = append(r.Issues, issue)
	r.SortIssues()
}

func (r *ValidationReport) AddIssues(issues ...ValidationIssue) {
	r.Issues = append(r.Issues, issues...)
	r.SortIssues()
}

func (r *ValidationReport) SortIssues() {
	sort.SliceStable(r.Issues, func(i, j int) bool {
		left := r.Issues[i]
		right := r.Issues[j]

		if severityRank(left.Severity) != severityRank(right.Severity) {
			return severityRank(left.Severity) < severityRank(right.Severity)
		}
		if left.Priority != right.Priority {
			return left.Priority < right.Priority
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.RuleID != right.RuleID {
			return left.RuleID < right.RuleID
		}
		return left.Message < right.Message
	})
}

func (r ValidationReport) HasBlockingIssues() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

func (r ValidationReport) NextBlockingIssue() (ValidationIssue, bool) {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			return issue, true
		}
	}
	return ValidationIssue{}, false
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityError:
		return 0
	case SeverityWarning:
		return 1
	default:
		return 2
	}
}
