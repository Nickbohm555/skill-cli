package generate

import "github.com/Nickbohm555/skill-cli/internal/validation"

// GateDecision is the centralized progression decision for validated output.
type GateDecision struct {
	Allowed       bool
	BlockingIssue *validation.ValidationIssue
	Reason        string
	Report        validation.ValidationReport
}

const (
	gateReasonAllowed       = "validation passed with no blocking issues"
	gateReasonBlockedByRule = "validation blocked by blocking issue"
)

// CanProceed is the single authority for allowing downstream progression.
func CanProceed(report validation.ValidationReport) GateDecision {
	decision := GateDecision{
		Allowed: false,
		Reason:  gateReasonAllowed,
		Report:  report,
	}

	if issue, ok := report.NextBlockingIssue(); ok {
		issueCopy := issue
		decision.Reason = gateReasonBlockedByRule
		decision.BlockingIssue = &issueCopy
		return decision
	}

	decision.Allowed = true
	return decision
}
