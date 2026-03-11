package generate

import "github.com/Nickbohm555/skill-cli/internal/overlap"

type OverlapDetectorFunc func(candidate overlap.SkillProfile, index overlap.InstalledIndex) overlap.OverlapReport

type OverlapDeciderFunc func(report overlap.OverlapReport) (overlap.OverlapReport, string)

type OverlapStageResult struct {
	Report           overlap.OverlapReport
	DecisionMessage  string
	Summary          overlap.ResolutionSummary
	SummaryBlock     string
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

	return OverlapStageResult{
		Report:           report,
		DecisionMessage:  decisionMessage,
		Summary:          summary,
		SummaryBlock:     summary.String(),
		ReadyForHandoff:  report.Decision != nil && report.Decision.IsResolved(),
		PreInstallStatus: summary.Status,
	}
}
