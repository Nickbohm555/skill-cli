package overlap

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/huh/v2"
)

type ResolutionOption struct {
	Label string
	Mode  ResolutionMode
}

type DecisionPrompt struct {
	Title         string
	Message       string
	Severity      OverlapSeverity
	TargetSkillID string
	Options       []ResolutionOption
	Report        OverlapReport
}

type DecisionPrompter interface {
	PromptDecision(prompt DecisionPrompt) (ResolutionMode, error)
}

type DecisionPrompterFunc func(prompt DecisionPrompt) (ResolutionMode, error)

func (f DecisionPrompterFunc) PromptDecision(prompt DecisionPrompt) (ResolutionMode, error) {
	return f(prompt)
}

type DecisionFlow struct {
	Prompter DecisionPrompter
	Now      func() time.Time
}

type HuhDecisionPrompter struct{}

func (HuhDecisionPrompter) PromptDecision(prompt DecisionPrompt) (ResolutionMode, error) {
	selection := ResolutionAbort
	options := make([]huh.Option[ResolutionMode], 0, len(prompt.Options))
	for _, option := range prompt.Options {
		options = append(options, huh.NewOption(option.Label, option.Mode))
	}

	err := huh.NewSelect[ResolutionMode]().
		Title(prompt.Title).
		Description(prompt.Message).
		Options(options...).
		Value(&selection).
		Run()
	if err != nil {
		return "", err
	}

	return selection, nil
}

func NewDecisionFlow(prompter DecisionPrompter) DecisionFlow {
	return DecisionFlow{
		Prompter: prompter,
		Now:      time.Now,
	}
}

func (f DecisionFlow) Decide(report OverlapReport) (OverlapReport, string) {
	if f.Now == nil {
		f.Now = time.Now
	}

	if report.OverallSeverity == SeverityNone || len(report.Findings) == 0 {
		decision := ConflictResolutionDecision{
			CandidateSkillID: report.Candidate.ID,
			Mode:             ResolutionNewInstall,
			Blocking:         false,
			Explanation:      "No overlap detected; proceed as a new install candidate.",
		}
		selectedAt := f.Now()
		decision.SelectedAt = &selectedAt
		report.Decision = &decision
		return report, "No conflicts found. Proceeding with the new-install path."
	}

	prompt := buildDecisionPrompt(report)
	if f.Prompter == nil {
		decision := blockingAbortDecision(report, prompt.TargetSkillID, "No decision prompter was configured, so conflict resolution stayed blocked.")
		report.Decision = &decision
		return report, "Conflict resolution is blocked because no interactive decision handler is configured."
	}

	mode, err := f.Prompter.PromptDecision(prompt)
	if err != nil {
		decision := blockingAbortDecision(report, prompt.TargetSkillID, interruptedDecisionExplanation(err))
		report.Decision = &decision
		return report, "Conflict resolution was interrupted. Defaulting to abort and keeping install handoff blocked."
	}

	if !isExplicitResolutionChoice(mode) {
		decision := blockingAbortDecision(report, prompt.TargetSkillID, fmt.Sprintf("Invalid conflict-resolution selection %q; explicit resolution is still required.", mode))
		report.Decision = &decision
		return report, "Conflict resolution is blocked because no valid explicit choice was made."
	}

	selectedAt := f.Now()
	decision := ConflictResolutionDecision{
		CandidateSkillID: report.Candidate.ID,
		TargetSkillID:    prompt.TargetSkillID,
		Mode:             mode,
		Blocking:         mode == ResolutionAbort,
		SelectedAt:       &selectedAt,
		Explanation:      selectedDecisionExplanation(mode, prompt.TargetSkillID),
	}
	if len(report.Findings) > 0 {
		decision.ExplanationMeta = report.Findings[0].ExplanationMeta
	}
	report.Decision = &decision

	switch mode {
	case ResolutionUpdate:
		return report, fmt.Sprintf("Selected conflict resolution: update existing skill %q.", prompt.TargetSkillID)
	case ResolutionMerge:
		return report, fmt.Sprintf("Selected conflict resolution: merge with existing skill %q.", prompt.TargetSkillID)
	default:
		return report, fmt.Sprintf("Selected conflict resolution: abort changes for target %q.", prompt.TargetSkillID)
	}
}

func buildDecisionPrompt(report OverlapReport) DecisionPrompt {
	targetSkillID := report.Candidate.ID
	if len(report.Findings) > 0 && report.Findings[0].ExistingSkillID != "" {
		targetSkillID = report.Findings[0].ExistingSkillID
	}

	return DecisionPrompt{
		Title:         "Overlap detected. Choose a conflict-resolution path before install:",
		Message:       decisionPromptMessage(report, targetSkillID),
		Severity:      report.OverallSeverity,
		TargetSkillID: targetSkillID,
		Options:       explicitResolutionOptions(),
		Report:        report,
	}
}

func explicitResolutionOptions() []ResolutionOption {
	return []ResolutionOption{
		{Label: "Update existing skill", Mode: ResolutionUpdate},
		{Label: "Merge with existing skill", Mode: ResolutionMerge},
		{Label: "Abort", Mode: ResolutionAbort},
	}
}

func decisionPromptMessage(report OverlapReport, targetSkillID string) string {
	parts := []string{
		fmt.Sprintf("Severity: %s", strings.ToUpper(string(report.OverallSeverity))),
		fmt.Sprintf("Target: %s", targetSkillID),
	}
	if len(report.Findings) > 0 {
		parts = append(parts, fmt.Sprintf("Reason: %s", strings.TrimSpace(report.Findings[0].Explanation)))
	}
	return strings.Join(parts, "\n")
}

func blockingAbortDecision(report OverlapReport, targetSkillID string, explanation string) ConflictResolutionDecision {
	return ConflictResolutionDecision{
		CandidateSkillID: report.Candidate.ID,
		TargetSkillID:    targetSkillID,
		Mode:             ResolutionAbort,
		Blocking:         true,
		Explanation:      explanation,
	}
}

func isExplicitResolutionChoice(mode ResolutionMode) bool {
	switch mode {
	case ResolutionUpdate, ResolutionMerge, ResolutionAbort:
		return true
	default:
		return false
	}
}

func interruptedDecisionExplanation(err error) string {
	if err == nil {
		return "Conflict-resolution prompt did not complete; install handoff remains blocked."
	}
	if err == io.EOF {
		return "Conflict-resolution prompt ended with EOF; defaulting to abort and keeping install handoff blocked."
	}
	return fmt.Sprintf("Conflict-resolution prompt failed (%v); defaulting to abort and keeping install handoff blocked.", err)
}

func selectedDecisionExplanation(mode ResolutionMode, targetSkillID string) string {
	switch mode {
	case ResolutionUpdate:
		return fmt.Sprintf("User explicitly chose to update the existing skill %q.", targetSkillID)
	case ResolutionMerge:
		return fmt.Sprintf("User explicitly chose to merge the candidate with the existing skill %q.", targetSkillID)
	case ResolutionAbort:
		return fmt.Sprintf("User explicitly chose to abort overlap resolution for %q.", targetSkillID)
	default:
		return "Conflict resolution selection is unresolved."
	}
}
