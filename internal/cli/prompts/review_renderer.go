package prompts

import (
	"fmt"
	"strings"

	"github.com/Nickbohm555/skill-cli/internal/refinement"
)

type ReviewModel struct {
	Sections    []ReviewSection
	CommitReady bool
	Summary     string
}

type ReviewSection struct {
	ID          refinement.SectionID
	Title       string
	CommitReady bool
	Fields      []ReviewField
}

type ReviewField struct {
	ID          refinement.FieldID
	Label       string
	Answer      string
	Status      refinement.ReadinessStatus
	StatusLabel string
	Hints       []string
}

func BuildReviewModel(report refinement.ValidationReport) ReviewModel {
	model := ReviewModel{
		Sections:    make([]ReviewSection, 0, len(report.Sections)),
		CommitReady: report.CommitReady,
		Summary:     reviewSummary(report.CommitReady),
	}

	for _, section := range report.Sections {
		viewSection := ReviewSection{
			ID:          section.Section,
			Title:       sectionTitle(section.Section),
			CommitReady: section.CommitReady,
			Fields:      make([]ReviewField, 0, len(section.Fields)),
		}

		for _, field := range section.Fields {
			viewSection.Fields = append(viewSection.Fields, ReviewField{
				ID:          field.Field.Definition.ID,
				Label:       field.Field.Definition.Label,
				Answer:      reviewAnswer(field.Field.Answer.Value),
				Status:      field.Status,
				StatusLabel: reviewStatusLabel(field.Status),
				Hints:       reviewHints(field),
			})
		}

		model.Sections = append(model.Sections, viewSection)
	}

	return model
}

func RenderReview(report refinement.ValidationReport) string {
	model := BuildReviewModel(report)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Commit readiness: %s\n", readinessBanner(model.CommitReady)))
	b.WriteString(model.Summary)

	for _, section := range model.Sections {
		b.WriteString("\n\n")
		b.WriteString(fmt.Sprintf("%s [%s]\n", section.Title, sectionStatusLabel(section.CommitReady)))

		for _, field := range section.Fields {
			b.WriteString(fmt.Sprintf("- %s [%s]\n", field.Label, field.StatusLabel))
			b.WriteString(fmt.Sprintf("  %s\n", field.Answer))
			for _, hint := range field.Hints {
				b.WriteString(fmt.Sprintf("  Hint: %s\n", hint))
			}
		}
	}

	return b.String()
}

func reviewSummary(commitReady bool) string {
	if commitReady {
		return "All required fields are ready. You can commit this refinement session."
	}
	return "Commit is blocked until every required field is complete, specific, and revalidated after revisions."
}

func readinessBanner(commitReady bool) string {
	if commitReady {
		return "ready"
	}
	return "blocked"
}

func sectionStatusLabel(commitReady bool) string {
	if commitReady {
		return "ready"
	}
	return "needs attention"
}

func reviewStatusLabel(status refinement.ReadinessStatus) string {
	switch status {
	case refinement.ReadinessReady:
		return "ready"
	case refinement.ReadinessMissing:
		return "missing"
	default:
		return "needs attention"
	}
}

func reviewAnswer(answer string) string {
	trimmed := strings.TrimSpace(answer)
	if trimmed == "" {
		return "(no answer yet)"
	}
	return trimmed
}

func sectionTitle(section refinement.SectionID) string {
	switch section {
	case refinement.SectionPurpose:
		return "Purpose"
	case refinement.SectionConstraints:
		return "Constraints"
	case refinement.SectionExamples:
		return "Examples"
	case refinement.SectionBoundaries:
		return "Boundaries"
	default:
		return strings.Title(strings.ReplaceAll(string(section), "_", " "))
	}
}

func reviewHints(field refinement.FieldValidation) []string {
	hints := make([]string, 0, len(field.Reasons))
	for _, reason := range field.Reasons {
		switch reason {
		case refinement.ValidationReasonRequiredMissing:
			hints = append(hints, "Add an answer before commit.")
		case refinement.ValidationReasonLowClarity:
			hints = append(hints, "Add concrete details so this field can move to ready.")
		case refinement.ValidationReasonNeedsRevalidation:
			hints = append(hints, "Re-opened because a related answer changed.")
		}
	}
	return hints
}
