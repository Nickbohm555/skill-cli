package prompts

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Nickbohm555/skill-cli/internal/refinement"
)

func TestPromptPrimaryPlansCoverRequiredFields(t *testing.T) {
	t.Parallel()

	state, err := refinement.NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	adapter := DefaultRefinementFormAdapter()

	for _, fieldID := range state.RequiredFields() {
		field, ok := state.Field(fieldID)
		if !ok {
			t.Fatalf("Field(%q) missing", fieldID)
		}

		plan, err := adapter.PrimaryPlan(field)
		if err != nil {
			t.Fatalf("PrimaryPlan(%q) error = %v", fieldID, err)
		}

		if plan.Kind != PromptKindPrimary {
			t.Fatalf("PrimaryPlan(%q) kind = %q, want %q", fieldID, plan.Kind, PromptKindPrimary)
		}
		if len(plan.Prompts) != 1 {
			t.Fatalf("PrimaryPlan(%q) prompts = %d, want 1", fieldID, len(plan.Prompts))
		}

		spec := plan.Prompts[0]
		if spec.Control != ControlTypeInput {
			t.Fatalf("PrimaryPlan(%q) control = %q, want %q", fieldID, spec.Control, ControlTypeInput)
		}
		if spec.Title != field.Definition.Label {
			t.Fatalf("PrimaryPlan(%q) title = %q, want %q", fieldID, spec.Title, field.Definition.Label)
		}
		if !strings.Contains(spec.Description, "Use concrete details") {
			t.Fatalf("PrimaryPlan(%q) description = %q, want standard guidance", fieldID, spec.Description)
		}
		if strings.TrimSpace(spec.Placeholder) == "" {
			t.Fatalf("PrimaryPlan(%q) placeholder was empty", fieldID)
		}
	}
}

func TestPromptDeepeningRoutingIsDeterministic(t *testing.T) {
	t.Parallel()

	adapter := DefaultRefinementFormAdapter()
	field := fieldWithAnswer(t, refinement.FieldConstraints, "Maybe keep it flexible and stuff.")

	firstPlan, err := adapter.DeepeningPlan(field, 0)
	if err != nil {
		t.Fatalf("DeepeningPlan(attempt=0) error = %v", err)
	}
	if firstPlan.Kind != PromptKindDeepeningFreeText {
		t.Fatalf("DeepeningPlan(attempt=0) kind = %q, want %q", firstPlan.Kind, PromptKindDeepeningFreeText)
	}
	if len(firstPlan.Prompts) != 1 || firstPlan.Prompts[0].Control != ControlTypeInput {
		t.Fatalf("DeepeningPlan(attempt=0) prompts = %+v, want one input", firstPlan.Prompts)
	}

	secondPlan, err := adapter.DeepeningPlan(field, 1)
	if err != nil {
		t.Fatalf("DeepeningPlan(attempt=1) error = %v", err)
	}
	if secondPlan.Kind != PromptKindDeepeningSelect {
		t.Fatalf("DeepeningPlan(attempt=1) kind = %q, want %q", secondPlan.Kind, PromptKindDeepeningSelect)
	}
	if !secondPlan.RequireOther {
		t.Fatalf("DeepeningPlan(attempt=1) RequireOther = false, want true")
	}

	gotOptions := secondPlan.Prompts[0].Options
	wantOptions := []PromptOption{
		{Label: "Input or source limitations", Value: "input_limitations"},
		{Label: "Behavior the skill must avoid", Value: "forbidden_behavior"},
		{Label: "Environment or tooling requirement", Value: "environment_requirement"},
		{Label: "Other (describe)", Value: OtherOptionValue},
	}
	if !reflect.DeepEqual(gotOptions, wantOptions) {
		t.Fatalf("DeepeningPlan(attempt=1) options = %v, want %v", gotOptions, wantOptions)
	}

	finalPlan, err := adapter.DeepeningPlan(field, 2)
	if err != nil {
		t.Fatalf("DeepeningPlan(attempt=2) error = %v", err)
	}
	if finalPlan.Kind != PromptKindDeepeningFallback {
		t.Fatalf("DeepeningPlan(attempt=2) kind = %q, want %q", finalPlan.Kind, PromptKindDeepeningFallback)
	}
	if !strings.Contains(finalPlan.Prompts[0].Description, "Pick the closest structured option") {
		t.Fatalf("DeepeningPlan(attempt=2) description = %q, want final-attempt guidance", finalPlan.Prompts[0].Description)
	}
}

func TestPromptDeepeningSkipsWhenClarityPasses(t *testing.T) {
	t.Parallel()

	adapter := DefaultRefinementFormAdapter()
	field := fieldWithAnswer(t, refinement.FieldPurposeSummary, "Generate a Codex skill from one docs URL, including install steps, scope boundaries, and example requests for later review.")

	needsDeepening, decision, err := adapter.NeedsDeepening(field, 0)
	if err != nil {
		t.Fatalf("NeedsDeepening() error = %v", err)
	}
	if needsDeepening {
		t.Fatalf("NeedsDeepening() = true, want false")
	}
	if decision.Mode != refinement.DeepeningModeNone {
		t.Fatalf("NeedsDeepening() mode = %q, want %q", decision.Mode, refinement.DeepeningModeNone)
	}

	plan, err := adapter.DeepeningPlan(field, 0)
	if err != nil {
		t.Fatalf("DeepeningPlan() error = %v", err)
	}
	if plan.Kind != PromptKindNoop {
		t.Fatalf("DeepeningPlan() kind = %q, want %q", plan.Kind, PromptKindNoop)
	}
	if len(plan.Prompts) != 0 {
		t.Fatalf("DeepeningPlan() prompts = %d, want 0", len(plan.Prompts))
	}
}

func TestPromptBuildDeepeningFieldsSupportsOtherPath(t *testing.T) {
	t.Parallel()

	adapter := DefaultRefinementFormAdapter()
	field := fieldWithAnswer(t, refinement.FieldExampleOutputs, "Maybe examples.")

	var choice string
	var other string
	plan, fields, err := adapter.BuildDeepeningFields(field, 1, DeepeningBindings{
		Choice: &choice,
		Other:  &other,
	})
	if err != nil {
		t.Fatalf("BuildDeepeningFields() error = %v", err)
	}

	if plan.Kind != PromptKindDeepeningSelect {
		t.Fatalf("BuildDeepeningFields() kind = %q, want %q", plan.Kind, PromptKindDeepeningSelect)
	}
	if len(fields) != 2 {
		t.Fatalf("BuildDeepeningFields() fields = %d, want 2", len(fields))
	}
	if plan.Prompts[0].Options[len(plan.Prompts[0].Options)-1].Value != OtherOptionValue {
		t.Fatalf("BuildDeepeningFields() last option = %q, want %q", plan.Prompts[0].Options[len(plan.Prompts[0].Options)-1].Value, OtherOptionValue)
	}
}

func fieldWithAnswer(t *testing.T, fieldID refinement.FieldID, answer string) refinement.FieldState {
	t.Helper()

	state, err := refinement.NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}
	if err := state.SetAnswer(fieldID, answer); err != nil {
		t.Fatalf("SetAnswer(%q) error = %v", fieldID, err)
	}
	field, ok := state.Field(fieldID)
	if !ok {
		t.Fatalf("Field(%q) missing", fieldID)
	}
	return field
}
