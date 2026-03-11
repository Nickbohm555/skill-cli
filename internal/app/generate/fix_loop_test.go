package generate

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Nickbohm555/skill-cli/internal/validation"
)

func TestFixLoopPromptsOneBlockingIssuePerIteration(t *testing.T) {
	t.Parallel()

	candidate := validation.CandidateSkill{}

	validateCalls := 0
	askedPrompts := make([]string, 0)
	askedRuleIDs := make([]string, 0)
	appliedAnswers := make([]string, 0)

	loop := FixLoop{
		Validate: func(candidate validation.CandidateSkill) validation.ValidationReport {
			validateCalls++

			report := validation.NewReport()
			switch {
			case candidate.Metadata.Name == "":
				report.AddIssue(validation.ValidationIssue{
					RuleID:   "VAL.STRUCT.METADATA_NAME_REQUIRED",
					Severity: validation.SeverityError,
					Path:     "metadata.name",
					Message:  "name missing",
					Priority: 10,
				})
			case len(candidate.InScope.Items) == 0:
				report.AddIssue(validation.ValidationIssue{
					RuleID:   "VAL.STRUCT.IN_SCOPE_REQUIRED",
					Severity: validation.SeverityError,
					Path:     "sections.in_scope.items",
					Message:  "in scope missing",
					Priority: 110,
				})
			}

			return report
		},
		Prompt: func(issue validation.ValidationIssue, prompt string) (string, error) {
			askedRuleIDs = append(askedRuleIDs, issue.RuleID)
			askedPrompts = append(askedPrompts, prompt)

			switch issue.RuleID {
			case "VAL.STRUCT.METADATA_NAME_REQUIRED":
				return "go-docs-skill", nil
			case "VAL.STRUCT.IN_SCOPE_REQUIRED":
				return "Extract instructions from the selected docs source only.", nil
			default:
				return "", errors.New("unexpected prompt")
			}
		},
		Apply: func(candidate validation.CandidateSkill, issue validation.ValidationIssue, answer string) (validation.CandidateSkill, error) {
			appliedAnswers = append(appliedAnswers, answer)

			switch issue.RuleID {
			case "VAL.STRUCT.METADATA_NAME_REQUIRED":
				candidate.Metadata.Name = answer
			case "VAL.STRUCT.IN_SCOPE_REQUIRED":
				candidate.InScope.Items = []string{answer}
			default:
				return candidate, errors.New("unexpected apply")
			}

			return candidate, nil
		},
	}

	result, err := loop.Run(candidate)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Report.HasBlockingIssues() {
		t.Fatalf("Run() final report has blocking issues = %#v, want none", result.Report.Issues)
	}
	if result.Iterations != 2 {
		t.Fatalf("Run() iterations = %d, want 2", result.Iterations)
	}
	if result.ValidationPasses != 3 {
		t.Fatalf("Run() validation passes = %d, want 3", result.ValidationPasses)
	}
	if validateCalls != 3 {
		t.Fatalf("validator calls = %d, want 3", validateCalls)
	}

	wantRules := []string{
		"VAL.STRUCT.METADATA_NAME_REQUIRED",
		"VAL.STRUCT.IN_SCOPE_REQUIRED",
	}
	if !reflect.DeepEqual(result.PromptedRuleIDs, wantRules) {
		t.Fatalf("Run() prompted rules = %#v, want %#v", result.PromptedRuleIDs, wantRules)
	}
	if !reflect.DeepEqual(askedRuleIDs, wantRules) {
		t.Fatalf("asked rule ids = %#v, want %#v", askedRuleIDs, wantRules)
	}

	wantPrompts := []string{
		validation.PromptForRule("VAL.STRUCT.METADATA_NAME_REQUIRED"),
		validation.PromptForRule("VAL.STRUCT.IN_SCOPE_REQUIRED"),
	}
	if !reflect.DeepEqual(askedPrompts, wantPrompts) {
		t.Fatalf("asked prompts = %#v, want %#v", askedPrompts, wantPrompts)
	}

	wantAnswers := []string{
		"go-docs-skill",
		"Extract instructions from the selected docs source only.",
	}
	if !reflect.DeepEqual(appliedAnswers, wantAnswers) {
		t.Fatalf("applied answers = %#v, want %#v", appliedAnswers, wantAnswers)
	}
	if result.Candidate.Metadata.Name != "go-docs-skill" {
		t.Fatalf("final candidate metadata.name = %q, want %q", result.Candidate.Metadata.Name, "go-docs-skill")
	}
	if !reflect.DeepEqual(result.Candidate.InScope.Items, []string{wantAnswers[1]}) {
		t.Fatalf("final candidate in-scope = %#v, want %#v", result.Candidate.InScope.Items, []string{wantAnswers[1]})
	}
}

func TestFixLoopReturnsUserCanceledAfterFirstBlockingIssue(t *testing.T) {
	t.Parallel()

	validateCalls := 0

	loop := FixLoop{
		Validate: func(candidate validation.CandidateSkill) validation.ValidationReport {
			validateCalls++

			report := validation.NewReport()
			report.AddIssue(validation.ValidationIssue{
				RuleID:   "VAL.STRUCT.METADATA_NAME_REQUIRED",
				Severity: validation.SeverityError,
				Path:     "metadata.name",
				Message:  "name missing",
				Priority: 10,
			})
			report.AddIssue(validation.ValidationIssue{
				RuleID:   "VAL.STRUCT.IN_SCOPE_REQUIRED",
				Severity: validation.SeverityError,
				Path:     "sections.in_scope.items",
				Message:  "in scope missing",
				Priority: 110,
			})
			return report
		},
		Prompt: func(issue validation.ValidationIssue, prompt string) (string, error) {
			if issue.RuleID != "VAL.STRUCT.METADATA_NAME_REQUIRED" {
				t.Fatalf("Prompt() issue = %q, want first blocking issue only", issue.RuleID)
			}
			return "", ErrUserCanceled
		},
		Apply: func(candidate validation.CandidateSkill, issue validation.ValidationIssue, answer string) (validation.CandidateSkill, error) {
			t.Fatal("Apply() should not run after cancel")
			return candidate, nil
		},
	}

	result, err := loop.Run(validation.CandidateSkill{})
	if !errors.Is(err, ErrUserCanceled) {
		t.Fatalf("Run() error = %v, want %v", err, ErrUserCanceled)
	}
	if result.Iterations != 0 {
		t.Fatalf("Run() iterations = %d, want 0", result.Iterations)
	}
	if result.ValidationPasses != 1 {
		t.Fatalf("Run() validation passes = %d, want 1", result.ValidationPasses)
	}
	if validateCalls != 1 {
		t.Fatalf("validator calls = %d, want 1", validateCalls)
	}
	if !result.Report.HasBlockingIssues() {
		t.Fatal("Run() final report blocking issues = false, want true")
	}
	if !reflect.DeepEqual(result.PromptedRuleIDs, []string{}) {
		t.Fatalf("Run() prompted rules = %#v, want empty", result.PromptedRuleIDs)
	}
}

func TestValidateCandidateCombinesStructuralAndSemanticReports(t *testing.T) {
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

	report := ValidateCandidate(candidate)
	if !report.HasBlockingIssues() {
		t.Fatal("ValidateCandidate() blocking issues = false, want true")
	}

	next, ok := report.NextBlockingIssue()
	if !ok {
		t.Fatal("ValidateCandidate() next blocking issue = none, want one")
	}
	if next.RuleID != "VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF" {
		t.Fatalf("ValidateCandidate() next blocking issue = %q, want %q", next.RuleID, "VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF")
	}
}
