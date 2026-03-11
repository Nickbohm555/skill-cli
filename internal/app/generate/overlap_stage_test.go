package generate

import (
	"strings"
	"testing"
	"time"

	"github.com/Nickbohm555/skill-cli/internal/overlap"
)

func TestOverlapStageSummaryIncludesSelectedModeAndReadyStatus(t *testing.T) {
	t.Parallel()

	selectedAt := time.Date(2026, time.March, 11, 19, 0, 0, 0, time.UTC)
	stage := OverlapStage{
		Detect: func(candidate overlap.SkillProfile, index overlap.InstalledIndex) overlap.OverlapReport {
			report := overlap.NewReport(candidate)
			report.AddFinding(overlap.OverlapFinding{
				RuleID:          "OVLP.NAME.EXACT",
				ExistingSkillID: "installed.docs",
				Severity:        overlap.SeverityHigh,
				Explanation:     "Candidate name exactly matches an installed skill name.",
			})
			return report
		},
		Decide: func(report overlap.OverlapReport) (overlap.OverlapReport, string) {
			report.Decision = &overlap.ConflictResolutionDecision{
				CandidateSkillID: report.Candidate.ID,
				TargetSkillID:    "installed.docs",
				Mode:             overlap.ResolutionMerge,
				Blocking:         false,
				SelectedAt:       &selectedAt,
			}
			return report, `Selected conflict resolution: merge with existing skill "installed.docs".`
		},
	}

	result := stage.Run(overlap.SkillProfile{ID: "candidate.docs"}, overlap.InstalledIndex{})

	if !result.ReadyForHandoff {
		t.Fatal("ReadyForHandoff = false, want true")
	}
	if !result.Gate.Allowed {
		t.Fatalf("Gate.Allowed = false, want true: %#v", result.Gate)
	}
	if result.Gate.Reason != overlapGateReasonResolvedDecision {
		t.Fatalf("Gate.Reason = %q, want %q", result.Gate.Reason, overlapGateReasonResolvedDecision)
	}
	if result.InstallHandoff == nil {
		t.Fatal("InstallHandoff = nil, want handoff payload")
	}
	if result.InstallHandoff.Decision.Mode != overlap.ResolutionMerge {
		t.Fatalf("InstallHandoff.Decision.Mode = %q, want %q", result.InstallHandoff.Decision.Mode, overlap.ResolutionMerge)
	}
	if result.PreInstallStatus != "READY for pre-install handoff." {
		t.Fatalf("PreInstallStatus = %q", result.PreInstallStatus)
	}
	if !strings.Contains(result.SummaryBlock, "Resolution Summary") {
		t.Fatalf("SummaryBlock = %q, want resolution summary header", result.SummaryBlock)
	}
	if !strings.Contains(result.SummaryBlock, "Selected mode: merge_with_existing") {
		t.Fatalf("SummaryBlock = %q, want selected mode line", result.SummaryBlock)
	}
	if !strings.Contains(result.SummaryBlock, "Target skill: installed.docs") {
		t.Fatalf("SummaryBlock = %q, want target skill line", result.SummaryBlock)
	}
	if !strings.Contains(result.SummaryBlock, "Status: READY for pre-install handoff.") {
		t.Fatalf("SummaryBlock = %q, want ready status line", result.SummaryBlock)
	}
	if !strings.Contains(result.SummaryBlock, "Next step: Proceed to Phase 06 install approval with the merge_with_existing handoff.") {
		t.Fatalf("SummaryBlock = %q, want next step line", result.SummaryBlock)
	}
}

func TestOverlapStageSummaryStaysBlockedWhenDecisionIsMissing(t *testing.T) {
	t.Parallel()

	stage := OverlapStage{
		Detect: func(candidate overlap.SkillProfile, index overlap.InstalledIndex) overlap.OverlapReport {
			report := overlap.NewReport(candidate)
			report.AddFinding(overlap.OverlapFinding{
				RuleID:          "OVLP.STRUCTURAL.OVERLAP",
				ExistingSkillID: "installed.docs",
				Severity:        overlap.SeverityMedium,
				Explanation:     "Candidate moderately overlaps an installed skill and requires explicit resolution.",
			})
			return report
		},
		Decide: func(report overlap.OverlapReport) (overlap.OverlapReport, string) {
			return report, "Conflict resolution is still required."
		},
	}

	result := stage.Run(overlap.SkillProfile{ID: "candidate.docs"}, overlap.InstalledIndex{})

	if result.ReadyForHandoff {
		t.Fatal("ReadyForHandoff = true, want false")
	}
	if result.Gate.Allowed {
		t.Fatalf("Gate.Allowed = true, want false: %#v", result.Gate)
	}
	if result.Gate.Reason != overlapGateReasonMissingDecision {
		t.Fatalf("Gate.Reason = %q, want %q", result.Gate.Reason, overlapGateReasonMissingDecision)
	}
	if result.InstallHandoff != nil {
		t.Fatalf("InstallHandoff = %#v, want nil", result.InstallHandoff)
	}
	if result.PreInstallStatus != "BLOCKED before pre-install handoff." {
		t.Fatalf("PreInstallStatus = %q", result.PreInstallStatus)
	}
	if !strings.Contains(result.SummaryBlock, "Selected mode: unresolved") {
		t.Fatalf("SummaryBlock = %q, want unresolved selected mode line", result.SummaryBlock)
	}
	if !strings.Contains(result.SummaryBlock, "Status: BLOCKED before pre-install handoff.") {
		t.Fatalf("SummaryBlock = %q, want blocked status line", result.SummaryBlock)
	}
}

func TestOverlapStageGateAllowsNoOverlapNewInstallHandoff(t *testing.T) {
	t.Parallel()

	stage := OverlapStage{
		Detect: func(candidate overlap.SkillProfile, index overlap.InstalledIndex) overlap.OverlapReport {
			return overlap.NewReport(candidate)
		},
	}

	result := stage.Run(overlap.SkillProfile{ID: "candidate.docs"}, overlap.InstalledIndex{})

	if !result.ReadyForHandoff {
		t.Fatal("ReadyForHandoff = false, want true")
	}
	if !result.Gate.Allowed {
		t.Fatalf("Gate.Allowed = false, want true: %#v", result.Gate)
	}
	if result.Gate.Reason != overlapGateReasonNoOverlap {
		t.Fatalf("Gate.Reason = %q, want %q", result.Gate.Reason, overlapGateReasonNoOverlap)
	}
	if result.InstallHandoff == nil {
		t.Fatal("InstallHandoff = nil, want handoff payload")
	}
	if result.InstallHandoff.Decision.Mode != overlap.ResolutionNewInstall {
		t.Fatalf("InstallHandoff.Decision.Mode = %q, want %q", result.InstallHandoff.Decision.Mode, overlap.ResolutionNewInstall)
	}
}

func TestOverlapStageGateBlocksAbortDecision(t *testing.T) {
	t.Parallel()

	selectedAt := time.Date(2026, time.March, 11, 20, 0, 0, 0, time.UTC)
	stage := OverlapStage{
		Detect: func(candidate overlap.SkillProfile, index overlap.InstalledIndex) overlap.OverlapReport {
			report := overlap.NewReport(candidate)
			report.AddFinding(overlap.OverlapFinding{
				RuleID:          "OVLP.NAME.EXACT",
				ExistingSkillID: "installed.docs",
				Severity:        overlap.SeverityHigh,
				Explanation:     "Candidate name exactly matches an installed skill name.",
			})
			return report
		},
		Decide: func(report overlap.OverlapReport) (overlap.OverlapReport, string) {
			report.Decision = &overlap.ConflictResolutionDecision{
				CandidateSkillID: report.Candidate.ID,
				TargetSkillID:    "installed.docs",
				Mode:             overlap.ResolutionAbort,
				Blocking:         true,
				SelectedAt:       &selectedAt,
			}
			return report, `Selected conflict resolution: abort changes for target "installed.docs".`
		},
	}

	result := stage.Run(overlap.SkillProfile{ID: "candidate.docs"}, overlap.InstalledIndex{})

	if result.ReadyForHandoff {
		t.Fatal("ReadyForHandoff = true, want false")
	}
	if result.Gate.Allowed {
		t.Fatalf("Gate.Allowed = true, want false: %#v", result.Gate)
	}
	if result.Gate.Reason != overlapGateReasonAbortDecision {
		t.Fatalf("Gate.Reason = %q, want %q", result.Gate.Reason, overlapGateReasonAbortDecision)
	}
	if result.InstallHandoff != nil {
		t.Fatalf("InstallHandoff = %#v, want nil", result.InstallHandoff)
	}
}
