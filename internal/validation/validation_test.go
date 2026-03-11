package validation

import (
	"reflect"
	"testing"
)

func TestParseSkillNormalizesFrontmatterAndSections(t *testing.T) {
	t.Parallel()

	input := []byte(`---
name: go-docs-skill
description: Turn one Go docs URL into a scoped Codex skill.
metadata:
  short-description: Go docs skill
---

# Go Docs Skill

## Purpose

Generate a Codex skill from one docs URL and keep the result installable.

## Primary Tasks

- Fetch the docs page.
- Extract the instructions.

## Success Criteria

- The skill stays focused on one docs source.

## Constraints

- Use one source URL only.

## Dependencies

- Go 1.25.x
- OpenAI Codex

## Example Requests

- Build a Go docs skill from https://go.dev/doc/

## Example Outputs

- A SKILL.md with install steps and scope limits.

## In Scope

- Extracting instructions from the chosen docs source.

## Out of Scope

- Mixing unrelated documentation sets.
`)

	got, err := ParseSkill(input)
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	if got.Metadata.Name != "go-docs-skill" {
		t.Fatalf("metadata.name = %q, want %q", got.Metadata.Name, "go-docs-skill")
	}
	if got.Metadata.Description != "Turn one Go docs URL into a scoped Codex skill." {
		t.Fatalf("metadata.description = %q", got.Metadata.Description)
	}
	if got.Metadata.Extra["short-description"] != "Go docs skill" {
		t.Fatalf("metadata.extra[short-description] = %q", got.Metadata.Extra["short-description"])
	}
	if got.Title != "Go Docs Skill" {
		t.Fatalf("title = %q, want %q", got.Title, "Go Docs Skill")
	}
	if got.PurposeSummary.Body != "Generate a Codex skill from one docs URL and keep the result installable." {
		t.Fatalf("purpose body = %q", got.PurposeSummary.Body)
	}

	wantTasks := []string{"Fetch the docs page.", "Extract the instructions."}
	if !reflect.DeepEqual(got.PrimaryTasks.Items, wantTasks) {
		t.Fatalf("primary tasks = %#v, want %#v", got.PrimaryTasks.Items, wantTasks)
	}

	wantScope := []string{"Extracting instructions from the chosen docs source."}
	if !reflect.DeepEqual(got.InScope.Items, wantScope) {
		t.Fatalf("in scope = %#v, want %#v", got.InScope.Items, wantScope)
	}
}

func TestParseSkillLeavesMissingSectionsEmpty(t *testing.T) {
	t.Parallel()

	got, err := ParseSkill([]byte(`---
name: minimal
description: minimal example
---

# Minimal Skill

## Purpose

Do one thing well.
`))
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	if got.PrimaryTasks.Heading != "Primary Tasks" {
		t.Fatalf("primary tasks heading = %q, want default heading", got.PrimaryTasks.Heading)
	}
	if len(got.PrimaryTasks.Items) != 0 {
		t.Fatalf("primary tasks items = %#v, want empty", got.PrimaryTasks.Items)
	}
	if len(got.OutOfScope.Items) != 0 {
		t.Fatalf("out of scope items = %#v, want empty", got.OutOfScope.Items)
	}
}

func TestValidationReportOrderingIsDeterministic(t *testing.T) {
	t.Parallel()

	issues := []ValidationIssue{
		{RuleID: "VAL.SCOPE.B", Severity: SeverityWarning, Path: "sections.out_of_scope", Message: "warning"},
		{RuleID: "VAL.META.NAME", Severity: SeverityError, Path: "metadata.name", Message: "missing", Priority: 20},
		{RuleID: "VAL.META.DESCRIPTION", Severity: SeverityError, Path: "metadata.description", Message: "missing", Priority: 10},
		{RuleID: "VAL.SCOPE.A", Severity: SeverityError, Path: "sections.in_scope", Message: "missing", Priority: 10},
	}

	const runs = 5
	var first []ValidationIssue
	for i := 0; i < runs; i++ {
		report := NewReport()
		report.AddIssues(issues...)

		if !report.HasBlockingIssues() {
			t.Fatal("HasBlockingIssues() = false, want true")
		}

		next, ok := report.NextBlockingIssue()
		if !ok {
			t.Fatal("NextBlockingIssue() = none, want first error")
		}
		if next.RuleID != "VAL.META.DESCRIPTION" {
			t.Fatalf("NextBlockingIssue() = %q, want %q", next.RuleID, "VAL.META.DESCRIPTION")
		}

		if i == 0 {
			first = append([]ValidationIssue(nil), report.Issues...)
			continue
		}
		if !reflect.DeepEqual(report.Issues, first) {
			t.Fatalf("ordered issues run %d = %#v, want %#v", i, report.Issues, first)
		}
	}
}

func TestValidationReportWarningsDoNotBlock(t *testing.T) {
	t.Parallel()

	report := NewReport()
	report.AddIssue(ValidationIssue{
		RuleID:   "VAL.SCOPE.WARNING",
		Severity: SeverityWarning,
		Path:     "sections.in_scope",
		Message:  "scope could be tighter",
	})

	if report.HasBlockingIssues() {
		t.Fatal("HasBlockingIssues() = true, want false")
	}
	if _, ok := report.NextBlockingIssue(); ok {
		t.Fatal("NextBlockingIssue() returned issue for warning-only report")
	}
}

func TestPromptForRuleCoversAllBlockingPhaseFourRules(t *testing.T) {
	t.Parallel()

	blockingRules := []string{
		ruleStructuralInternal,
		"VAL.STRUCT.METADATA_NAME_REQUIRED",
		"VAL.STRUCT.METADATA_DESCRIPTION_REQUIRED",
		"VAL.STRUCT.TITLE_REQUIRED",
		"VAL.STRUCT.PURPOSE_SUMMARY_REQUIRED",
		"VAL.STRUCT.PRIMARY_TASKS_REQUIRED",
		"VAL.STRUCT.SUCCESS_CRITERIA_REQUIRED",
		"VAL.STRUCT.CONSTRAINTS_REQUIRED",
		"VAL.STRUCT.DEPENDENCIES_REQUIRED",
		"VAL.STRUCT.EXAMPLE_REQUESTS_REQUIRED",
		"VAL.STRUCT.EXAMPLE_OUTPUTS_REQUIRED",
		"VAL.STRUCT.IN_SCOPE_REQUIRED",
		"VAL.STRUCT.OUT_OF_SCOPE_REQUIRED",
		ruleInScopeEntryTooBrief,
		ruleInScopeCatchAll,
		ruleOutOfScopeEntryTooBrief,
		ruleOutOfScopeCatchAll,
	}

	for _, ruleID := range blockingRules {
		prompt := PromptForRule(ruleID)
		if prompt == "" {
			t.Fatalf("PromptForRule(%q) = empty, want targeted prompt", ruleID)
		}
		if prompt == unknownRulePrompt {
			t.Fatalf("PromptForRule(%q) returned fallback prompt, want targeted prompt", ruleID)
		}
	}
}

func TestPromptForRuleFallsBackDeterministically(t *testing.T) {
	t.Parallel()

	got := PromptForRule("VAL.UNKNOWN.RULE")
	if got != unknownRulePrompt {
		t.Fatalf("PromptForRule() fallback = %q, want %q", got, unknownRulePrompt)
	}
}

func TestStructuralValidationAcceptsValidCandidate(t *testing.T) {
	t.Parallel()

	candidate, err := ParseSkill(validSkillFixture())
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	report := ValidateStructural(candidate)
	if report.HasBlockingIssues() {
		t.Fatalf("ValidateStructural() blocking issues = %#v, want none", report.Issues)
	}
}

func TestStructuralValidationFailsClosedOnMissingRequiredSections(t *testing.T) {
	t.Parallel()

	candidate, err := ParseSkill([]byte(`---
name: minimal
description: minimal example
---

# Minimal Skill

## Purpose

Do one thing well.
`))
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	report := ValidateStructural(candidate)
	if !report.HasBlockingIssues() {
		t.Fatal("ValidateStructural() blocking issues = false, want true")
	}

	next, ok := report.NextBlockingIssue()
	if !ok {
		t.Fatal("NextBlockingIssue() = none, want first blocking issue")
	}
	if next.RuleID != "VAL.STRUCT.PRIMARY_TASKS_REQUIRED" {
		t.Fatalf("NextBlockingIssue().RuleID = %q, want %q", next.RuleID, "VAL.STRUCT.PRIMARY_TASKS_REQUIRED")
	}
}

func TestStructuralValidationRejectsMalformedValues(t *testing.T) {
	t.Parallel()

	candidate, err := ParseSkill(validSkillFixture())
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	candidate.Metadata.Name = ""
	candidate.PrimaryTasks.Items = []string{"ship the skill", ""}

	report := ValidateStructural(candidate)
	if !report.HasBlockingIssues() {
		t.Fatal("ValidateStructural() blocking issues = false, want true")
	}

	want := []ValidationIssue{
		{
			RuleID:   "VAL.STRUCT.METADATA_NAME_REQUIRED",
			Severity: SeverityError,
			Path:     "metadata.name",
			Message:  "metadata.name must not be blank.",
			Priority: 10,
		},
		{
			RuleID:   "VAL.STRUCT.PRIMARY_TASKS_REQUIRED",
			Severity: SeverityError,
			Path:     "sections.primary_tasks.items[1]",
			Message:  "Primary Tasks entries must not be blank.",
			Priority: 50,
		},
	}

	for _, wantIssue := range want {
		if !containsIssue(report.Issues, wantIssue) {
			t.Fatalf("ValidateStructural() issues = %#v, want %#v present", report.Issues, wantIssue)
		}
	}
}

func TestStructuralValidationOrderingIsDeterministic(t *testing.T) {
	t.Parallel()

	candidate, err := ParseSkill(validSkillFixture())
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	candidate.Metadata.Description = ""
	candidate.OutOfScope.Items = nil

	const runs = 5
	var first []ValidationIssue
	for i := 0; i < runs; i++ {
		report := ValidateStructural(candidate)
		next, ok := report.NextBlockingIssue()
		if !ok {
			t.Fatal("NextBlockingIssue() = none, want first blocking issue")
		}
		if next.RuleID != "VAL.STRUCT.METADATA_DESCRIPTION_REQUIRED" {
			t.Fatalf("NextBlockingIssue().RuleID = %q, want %q", next.RuleID, "VAL.STRUCT.METADATA_DESCRIPTION_REQUIRED")
		}

		if i == 0 {
			first = append([]ValidationIssue(nil), report.Issues...)
			continue
		}
		if !reflect.DeepEqual(report.Issues, first) {
			t.Fatalf("ordered issues run %d = %#v, want %#v", i, report.Issues, first)
		}
	}
}

func TestSemanticValidationAcceptsSpecificBoundaries(t *testing.T) {
	t.Parallel()

	candidate, err := ParseSkill(validSkillFixture())
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	report := ValidateSemantic(candidate)
	if report.HasBlockingIssues() {
		t.Fatalf("ValidateSemantic() blocking issues = %#v, want none", report.Issues)
	}
}

func TestSemanticValidationRejectsBriefBoundaryEntries(t *testing.T) {
	t.Parallel()

	candidate, err := ParseSkill(validSkillFixture())
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	candidate.InScope.Items = []string{"Go docs"}
	candidate.OutOfScope.Items = []string{"Other topics"}

	report := ValidateSemantic(candidate)
	if !report.HasBlockingIssues() {
		t.Fatal("ValidateSemantic() blocking issues = false, want true")
	}

	want := []ValidationIssue{
		{
			RuleID:   "VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF",
			Severity: SeverityError,
			Path:     "sections.in_scope.items[0]",
			Message:  "In Scope entries must be specific enough to define a concrete boundary.",
			Priority: 130,
		},
		{
			RuleID:   "VAL.SCOPE.OUT_OF_SCOPE_ENTRY_TOO_BRIEF",
			Severity: SeverityError,
			Path:     "sections.out_of_scope.items[0]",
			Message:  "Out Of Scope entries must be specific enough to define a concrete boundary.",
			Priority: 150,
		},
	}

	for _, wantIssue := range want {
		if !containsIssue(report.Issues, wantIssue) {
			t.Fatalf("ValidateSemantic() issues = %#v, want %#v present", report.Issues, wantIssue)
		}
	}
}

func TestSemanticValidationRejectsVagueCatchAllPhrasing(t *testing.T) {
	t.Parallel()

	candidate, err := ParseSkill(validSkillFixture())
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	candidate.InScope.Items = []string{"Handle Go standard library setup and anything else users need."}
	candidate.OutOfScope.Items = []string{"Exclude unrelated deployment platforms and miscellaneous requests."}

	report := ValidateSemantic(candidate)
	if !report.HasBlockingIssues() {
		t.Fatal("ValidateSemantic() blocking issues = false, want true")
	}

	want := []ValidationIssue{
		{
			RuleID:   "VAL.SCOPE.IN_SCOPE_VAGUE_CATCH_ALL",
			Severity: SeverityError,
			Path:     "sections.in_scope.items[0]",
			Message:  "In Scope entries must avoid vague catch-all phrasing.",
			Priority: 140,
		},
		{
			RuleID:   "VAL.SCOPE.OUT_OF_SCOPE_VAGUE_CATCH_ALL",
			Severity: SeverityError,
			Path:     "sections.out_of_scope.items[0]",
			Message:  "Out Of Scope entries must avoid vague catch-all phrasing.",
			Priority: 160,
		},
	}

	for _, wantIssue := range want {
		if !containsIssue(report.Issues, wantIssue) {
			t.Fatalf("ValidateSemantic() issues = %#v, want %#v present", report.Issues, wantIssue)
		}
	}
}

func TestSemanticValidationOrderingIsDeterministic(t *testing.T) {
	t.Parallel()

	candidate, err := ParseSkill(validSkillFixture())
	if err != nil {
		t.Fatalf("ParseSkill returned error: %v", err)
	}

	candidate.InScope.Items = []string{"General docs"}
	candidate.OutOfScope.Items = []string{"Handle unrelated product areas and anything else."}

	const runs = 5
	var first []ValidationIssue
	for i := 0; i < runs; i++ {
		report := ValidateSemantic(candidate)
		next, ok := report.NextBlockingIssue()
		if !ok {
			t.Fatal("NextBlockingIssue() = none, want first blocking issue")
		}
		if next.RuleID != "VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF" {
			t.Fatalf("NextBlockingIssue().RuleID = %q, want %q", next.RuleID, "VAL.SCOPE.IN_SCOPE_ENTRY_TOO_BRIEF")
		}

		if i == 0 {
			first = append([]ValidationIssue(nil), report.Issues...)
			continue
		}
		if !reflect.DeepEqual(report.Issues, first) {
			t.Fatalf("ordered issues run %d = %#v, want %#v", i, report.Issues, first)
		}
	}
}

func containsIssue(issues []ValidationIssue, want ValidationIssue) bool {
	for _, issue := range issues {
		if reflect.DeepEqual(issue, want) {
			return true
		}
	}
	return false
}

func validSkillFixture() []byte {
	return []byte(`---
name: go-docs-skill
description: Turn one Go docs URL into a scoped Codex skill.
metadata:
  short-description: Go docs skill
---

# Go Docs Skill

## Purpose

Generate a Codex skill from one docs URL and keep the result installable.

## Primary Tasks

- Fetch the docs page.
- Extract the instructions.

## Success Criteria

- The skill stays focused on one docs source.

## Constraints

- Use one source URL only.

## Dependencies

- Go 1.25.x
- OpenAI Codex

## Example Requests

- Build a Go docs skill from https://go.dev/doc/

## Example Outputs

- A SKILL.md with install steps and scope limits.

## In Scope

- Extracting instructions from the chosen docs source.

## Out of Scope

- Mixing unrelated documentation sets.
`)
}
