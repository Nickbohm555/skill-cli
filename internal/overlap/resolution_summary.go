package overlap

import (
	"fmt"
	"strings"
)

type ResolutionSummary struct {
	Findings     string
	SelectedMode string
	TargetSkill  string
	NextStep     string
	Status       string
}

func BuildResolutionSummary(report OverlapReport) ResolutionSummary {
	targetSkillID := report.Candidate.ID
	if len(report.Findings) > 0 && report.Findings[0].ExistingSkillID != "" {
		targetSkillID = report.Findings[0].ExistingSkillID
	}

	selectedMode := "unresolved"
	nextStep := "Stop before Phase 06 install approval until conflict resolution is completed explicitly."
	status := "BLOCKED before pre-install handoff."
	if report.Decision != nil {
		if report.Decision.Mode != "" {
			selectedMode = string(report.Decision.Mode)
		}
		if strings.TrimSpace(report.Decision.TargetSkillID) != "" {
			targetSkillID = report.Decision.TargetSkillID
		}
		if report.Decision.IsResolved() {
			status = "READY for pre-install handoff."
		}
		nextStep = resolutionNextStep(*report.Decision)
	}

	return ResolutionSummary{
		Findings:     findingsSummary(report),
		SelectedMode: selectedMode,
		TargetSkill:  targetSkillID,
		NextStep:     nextStep,
		Status:       status,
	}
}

func (s ResolutionSummary) String() string {
	lines := []string{
		"Resolution Summary",
		fmt.Sprintf("Findings: %s", s.Findings),
		fmt.Sprintf("Selected mode: %s", s.SelectedMode),
		fmt.Sprintf("Target skill: %s", s.TargetSkill),
		fmt.Sprintf("Next step: %s", s.NextStep),
		fmt.Sprintf("Status: %s", s.Status),
	}
	return strings.Join(lines, "\n")
}

func findingsSummary(report OverlapReport) string {
	if len(report.Findings) == 0 {
		return "No overlap findings detected."
	}

	ruleIDs := make([]string, 0, len(report.Findings))
	for _, finding := range report.Findings {
		if strings.TrimSpace(finding.RuleID) == "" {
			continue
		}
		ruleIDs = append(ruleIDs, finding.RuleID)
	}

	summary := fmt.Sprintf(
		"%s overlap across %d finding(s)",
		strings.ToUpper(string(report.OverallSeverity)),
		len(report.Findings),
	)
	if len(ruleIDs) > 0 {
		summary += fmt.Sprintf(" [%s]", strings.Join(ruleIDs, ", "))
	}
	return summary + "."
}

func resolutionNextStep(decision ConflictResolutionDecision) string {
	switch decision.Mode {
	case ResolutionNewInstall:
		return "Proceed to Phase 06 install approval as a new install candidate."
	case ResolutionUpdate:
		return "Proceed to Phase 06 install approval with the update_existing handoff."
	case ResolutionMerge:
		return "Proceed to Phase 06 install approval with the merge_with_existing handoff."
	case ResolutionAbort:
		return "Stop before Phase 06 install approval; the selected outcome is abort."
	default:
		return "Stop before Phase 06 install approval until conflict resolution is completed explicitly."
	}
}
