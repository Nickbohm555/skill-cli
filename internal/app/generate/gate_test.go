package generate

import (
	"testing"

	"github.com/Nickbohm555/skill-cli/internal/validation"
)

func TestGateAllowsWarningOnlyReports(t *testing.T) {
	t.Parallel()

	report := validation.NewReport()
	report.AddIssue(validation.ValidationIssue{
		RuleID:   "VAL.SCOPE.BOUNDARY_STYLE_WARNING",
		Severity: validation.SeverityWarning,
		Path:     "sections.in_scope.items[0]",
		Message:  "boundary wording could be more specific",
		Priority: 300,
	})

	decision := CanProceed(report)
	if !decision.Allowed {
		t.Fatalf("CanProceed() allowed = false, want true: %#v", decision)
	}
	if decision.BlockingIssue != nil {
		t.Fatalf("CanProceed() blocking issue = %#v, want nil", *decision.BlockingIssue)
	}
	if decision.Reason != gateReasonAllowed {
		t.Fatalf("CanProceed() reason = %q, want %q", decision.Reason, gateReasonAllowed)
	}
	if len(decision.Report.Issues) != 1 {
		t.Fatalf("CanProceed() report issue count = %d, want 1", len(decision.Report.Issues))
	}
}

func TestGateBlocksOnFirstErrorDeterministically(t *testing.T) {
	t.Parallel()

	report := validation.NewReport()
	report.AddIssue(validation.ValidationIssue{
		RuleID:   "VAL.SCOPE.OUT_OF_SCOPE_ENTRY_TOO_BRIEF",
		Severity: validation.SeverityError,
		Path:     "sections.out_of_scope.items[0]",
		Message:  "out-of-scope boundary is too brief",
		Priority: 220,
	})
	report.AddIssue(validation.ValidationIssue{
		RuleID:   "VAL.STRUCT.METADATA_NAME_REQUIRED",
		Severity: validation.SeverityError,
		Path:     "metadata.name",
		Message:  "name missing",
		Priority: 10,
	})
	report.AddIssue(validation.ValidationIssue{
		RuleID:   "VAL.SCOPE.BOUNDARY_STYLE_WARNING",
		Severity: validation.SeverityWarning,
		Path:     "sections.in_scope.items[0]",
		Message:  "boundary wording could be more specific",
		Priority: 300,
	})

	decision := CanProceed(report)
	if decision.Allowed {
		t.Fatalf("CanProceed() allowed = true, want false: %#v", decision)
	}
	if decision.BlockingIssue == nil {
		t.Fatal("CanProceed() blocking issue = nil, want first blocking issue")
	}
	if decision.BlockingIssue.RuleID != "VAL.STRUCT.METADATA_NAME_REQUIRED" {
		t.Fatalf("CanProceed() blocking rule = %q, want %q", decision.BlockingIssue.RuleID, "VAL.STRUCT.METADATA_NAME_REQUIRED")
	}
	if decision.Reason != gateReasonBlockedByRule {
		t.Fatalf("CanProceed() reason = %q, want %q", decision.Reason, gateReasonBlockedByRule)
	}
	if len(decision.Report.Issues) != 3 {
		t.Fatalf("CanProceed() report issue count = %d, want 3", len(decision.Report.Issues))
	}
}

func TestGateMatchesValidateCandidateProgressionPolicy(t *testing.T) {
	t.Parallel()

	candidate := validation.CandidateSkill{
		Metadata: validation.SkillMetadata{
			Name:        "go-docs-skill",
			Description: "Generate a skill from a single Go docs URL.",
		},
		Title:          "Go Docs Skill",
		PurposeSummary: validation.TextSection{Heading: "Purpose", Body: "Generate a scoped skill from one docs URL."},
		PrimaryTasks: validation.ListSection{
			Heading: "Primary Tasks",
			Items:   []string{"Extract docs instructions from one source."},
		},
		SuccessCriteria: validation.ListSection{
			Heading: "Success Criteria",
			Items:   []string{"The generated skill stays scoped to one source."},
		},
		Constraints: validation.ListSection{
			Heading: "Constraints",
			Items:   []string{"Use only the provided docs URL as source material."},
		},
		Dependencies: validation.ListSection{
			Heading: "Dependencies",
			Items:   []string{"Go 1.25.x"},
		},
		ExampleRequests: validation.ListSection{
			Heading: "Example Requests",
			Items:   []string{"Build a skill from https://go.dev/doc/"},
		},
		ExampleOutputs: validation.ListSection{
			Heading: "Example Outputs",
			Items:   []string{"A SKILL.md with metadata, scope, and install steps."},
		},
		InScope: validation.ListSection{
			Heading: "In Scope",
			Items:   []string{"misc"},
		},
		OutOfScope: validation.ListSection{
			Heading: "Out Of Scope",
			Items:   []string{"Mixing unrelated documentation sets."},
		},
	}

	decision := CanProceed(ValidateCandidate(candidate))
	if decision.Allowed {
		t.Fatal("CanProceed(ValidateCandidate()) allowed = true, want false")
	}
	if decision.BlockingIssue == nil {
		t.Fatal("CanProceed(ValidateCandidate()) blocking issue = nil, want one")
	}
	if decision.BlockingIssue.RuleID != "VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF" {
		t.Fatalf("CanProceed(ValidateCandidate()) blocking rule = %q, want %q", decision.BlockingIssue.RuleID, "VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF")
	}
}
