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
