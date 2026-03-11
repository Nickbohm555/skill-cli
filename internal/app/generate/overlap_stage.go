package generate

import "github.com/Nickbohm555/skill-cli/internal/overlap"

type OverlapDetectorFunc func(candidate overlap.SkillProfile, index overlap.InstalledIndex) overlap.OverlapReport

type OverlapDeciderFunc func(report overlap.OverlapReport) (overlap.OverlapReport, string)

type OverlapGateDecision struct {
	Allowed bool
	Reason  string
}

type OverlapInstallHandoff struct {
	Decision overlap.ConflictResolutionDecision
	Summary  overlap.ResolutionSummary
}

type OverlapStageResult struct {
	Report           overlap.OverlapReport
	DecisionMessage  string
	Summary          overlap.ResolutionSummary
	SummaryBlock     string
	Gate             OverlapGateDecision
	InstallHandoff   *OverlapInstallHandoff
	ReadyForHandoff  bool
	PreInstallStatus string
}

type OverlapStage struct {
	Detect OverlapDetectorFunc
	Decide OverlapDeciderFunc
}

func (s OverlapStage) Run(candidate overlap.SkillProfile, index overlap.InstalledIndex) OverlapStageResult {
	detect := s.Detect
	if detect == nil {
		detect = overlap.Detect
	}

	decide := s.Decide
	if decide == nil {
		defaultFlow := overlap.NewDecisionFlow(nil)
		decide = defaultFlow.Decide
	}

	report := detect(candidate, index)
	report, decisionMessage := decide(report)
	summary := overlap.BuildResolutionSummary(report)
	gate, handoff := overlapGateDecision(report, summary)

	return OverlapStageResult{
		Report:           report,
		DecisionMessage:  decisionMessage,
		Summary:          summary,
		SummaryBlock:     summary.String(),
		Gate:             gate,
		InstallHandoff:   handoff,
		ReadyForHandoff:  gate.Allowed,
		PreInstallStatus: summary.Status,
	}
}

const (
	overlapGateReasonNoOverlap          = "no overlap detected; Phase 06 install approval handoff is allowed"
	overlapGateReasonResolvedDecision   = "explicit conflict resolution recorded; Phase 06 install approval handoff is allowed"
	overlapGateReasonMissingDecision    = "overlap detected but no conflict resolution decision was recorded; Phase 06 install approval handoff is blocked"
	overlapGateReasonBlockingDecision   = "conflict resolution decision is still blocking; Phase 06 install approval handoff is blocked"
	overlapGateReasonAbortDecision      = "conflict resolution selected abort; Phase 06 install approval handoff is blocked"
	overlapGateReasonUnresolvedDecision = "conflict resolution decision is incomplete; Phase 06 install approval handoff is blocked"
)

func overlapGateDecision(report overlap.OverlapReport, summary overlap.ResolutionSummary) (OverlapGateDecision, *OverlapInstallHandoff) {
	if report.Decision == nil {
		if report.OverallSeverity == overlap.SeverityNone && len(report.Findings) == 0 {
			return OverlapGateDecision{
				Allowed: true,
				Reason:  overlapGateReasonNoOverlap,
			}, nil
		}

		return OverlapGateDecision{
			Allowed: false,
			Reason:  overlapGateReasonMissingDecision,
		}, nil
	}

	switch {
	case report.Decision.Mode == overlap.ResolutionAbort:
		return OverlapGateDecision{
			Allowed: false,
			Reason:  overlapGateReasonAbortDecision,
		}, nil
	case report.Decision.Blocking:
		return OverlapGateDecision{
			Allowed: false,
			Reason:  overlapGateReasonBlockingDecision,
		}, nil
	case !report.Decision.IsResolved():
		return OverlapGateDecision{
			Allowed: false,
			Reason:  overlapGateReasonUnresolvedDecision,
		}, nil
	default:
		return OverlapGateDecision{
				Allowed: true,
				Reason:  resolvedGateReason(*report.Decision),
			}, &OverlapInstallHandoff{
				Decision: *report.Decision,
				Summary:  summary,
			}
	}
}

func resolvedGateReason(decision overlap.ConflictResolutionDecision) string {
	if decision.Mode == overlap.ResolutionNewInstall {
		return overlapGateReasonNoOverlap
	}
	return overlapGateReasonResolvedDecision
}
