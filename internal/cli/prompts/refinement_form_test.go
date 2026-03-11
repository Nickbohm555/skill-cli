package prompts

import (
	"fmt"
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
	testCases := []struct {
		name             string
		fieldID          refinement.FieldID
		answer           string
		attempts         int
		wantKind         PromptKind
		wantPromptCount  int
		wantControl      ControlType
		wantRequireOther bool
		wantDescription  string
	}{
		{
			name:            "low clarity first pass stays in targeted free text",
			fieldID:         refinement.FieldConstraints,
			answer:          "Maybe keep it flexible and stuff.",
			attempts:        0,
			wantKind:        PromptKindDeepeningFreeText,
			wantPromptCount: 1,
			wantControl:     ControlTypeInput,
			wantDescription: "Add the missing specificity",
		},
		{
			name:             "second low clarity pass switches to structured options",
			fieldID:          refinement.FieldConstraints,
			answer:           "Maybe keep it flexible and stuff.",
			attempts:         1,
			wantKind:         PromptKindDeepeningSelect,
			wantPromptCount:  2,
			wantControl:      ControlTypeSelect,
			wantRequireOther: true,
			wantDescription:  "Add the missing specificity",
		},
		{
			name:             "attempt cap forces fallback wording",
			fieldID:          refinement.FieldConstraints,
			answer:           "Maybe keep it flexible and stuff.",
			attempts:         2,
			wantKind:         PromptKindDeepeningFallback,
			wantPromptCount:  2,
			wantControl:      ControlTypeSelect,
			wantRequireOther: true,
			wantDescription:  "Pick the closest structured option",
		},
		{
			name:            "high clarity answer skips deepening",
			fieldID:         refinement.FieldPurposeSummary,
			answer:          "Generate a Codex skill from one docs URL, including install steps, scope boundaries, and example requests for later review.",
			attempts:        0,
			wantKind:        PromptKindNoop,
			wantPromptCount: 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			field := fieldWithAnswer(t, tc.fieldID, tc.answer)
			plan, err := adapter.DeepeningPlan(field, tc.attempts)
			if err != nil {
				t.Fatalf("DeepeningPlan() error = %v", err)
			}

			if plan.Kind != tc.wantKind {
				t.Fatalf("DeepeningPlan() kind = %q, want %q", plan.Kind, tc.wantKind)
			}
			if len(plan.Prompts) != tc.wantPromptCount {
				t.Fatalf("DeepeningPlan() prompts = %d, want %d", len(plan.Prompts), tc.wantPromptCount)
			}
			if tc.wantPromptCount == 0 {
				return
			}
			if plan.Prompts[0].Control != tc.wantControl {
				t.Fatalf("DeepeningPlan() first control = %q, want %q", plan.Prompts[0].Control, tc.wantControl)
			}
			if plan.RequireOther != tc.wantRequireOther {
				t.Fatalf("DeepeningPlan() RequireOther = %t, want %t", plan.RequireOther, tc.wantRequireOther)
			}
			if !strings.Contains(plan.Prompts[0].Description, tc.wantDescription) {
				t.Fatalf("DeepeningPlan() description = %q, want substring %q", plan.Prompts[0].Description, tc.wantDescription)
			}
		})
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

func TestPromptStructuredChoiceOptionsStayStable(t *testing.T) {
	t.Parallel()

	adapter := DefaultRefinementFormAdapter()
	testCases := []struct {
		fieldID      refinement.FieldID
		answer       string
		wantOptions  []PromptOption
		wantKind     PromptKind
		wantOtherKey string
	}{
		{
			fieldID:  refinement.FieldConstraints,
			answer:   "Maybe keep it flexible and stuff.",
			wantKind: PromptKindDeepeningSelect,
			wantOptions: []PromptOption{
				{Label: "Input or source limitations", Value: "input_limitations"},
				{Label: "Behavior the skill must avoid", Value: "forbidden_behavior"},
				{Label: "Environment or tooling requirement", Value: "environment_requirement"},
				{Label: "Other (describe)", Value: OtherOptionValue},
			},
			wantOtherKey: "constraints_other",
		},
		{
			fieldID:  refinement.FieldOutOfScope,
			answer:   "Things maybe.",
			wantKind: PromptKindDeepeningSelect,
			wantOptions: []PromptOption{
				{Label: "Disallowed source expansion", Value: "source_exclusion"},
				{Label: "Unsupported workflow or capability", Value: "workflow_exclusion"},
				{Label: "Speculative or invented content", Value: "speculation_exclusion"},
				{Label: "Other (describe)", Value: OtherOptionValue},
			},
			wantOtherKey: "out_of_scope_other",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(string(tc.fieldID), func(t *testing.T) {
			t.Parallel()

			field := fieldWithAnswer(t, tc.fieldID, tc.answer)
			plan, err := adapter.DeepeningPlan(field, 1)
			if err != nil {
				t.Fatalf("DeepeningPlan() error = %v", err)
			}

			if plan.Kind != tc.wantKind {
				t.Fatalf("DeepeningPlan() kind = %q, want %q", plan.Kind, tc.wantKind)
			}
			if len(plan.Prompts) != 2 {
				t.Fatalf("DeepeningPlan() prompts = %d, want 2", len(plan.Prompts))
			}
			if !reflect.DeepEqual(plan.Prompts[0].Options, tc.wantOptions) {
				t.Fatalf("DeepeningPlan() options = %v, want %v", plan.Prompts[0].Options, tc.wantOptions)
			}
			if plan.Prompts[1].Key != tc.wantOtherKey {
				t.Fatalf("DeepeningPlan() other key = %q, want %q", plan.Prompts[1].Key, tc.wantOtherKey)
			}
		})
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

func TestPromptOtherPathValidationIsSafe(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		selected  *string
		value     string
		wantError string
	}{
		{
			name:     "non other choice allows empty custom detail",
			selected: stringPtr("usage_output"),
			value:    "",
		},
		{
			name:      "other choice requires explicit detail",
			selected:  stringPtr(OtherOptionValue),
			value:     "   ",
			wantError: fmt.Sprintf("%s requires additional detail", refinement.FieldExampleOutputs),
		},
		{
			name:     "other choice accepts custom detail",
			selected: stringPtr(OtherOptionValue),
			value:    "Include one shell example plus one unsupported-path example.",
		},
		{
			name:  "nil selection does not block plain input validation",
			value: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := requiredOther(refinement.FieldExampleOutputs, tc.selected)(tc.value)
			switch {
			case tc.wantError == "" && err != nil:
				t.Fatalf("requiredOther() error = %v, want nil", err)
			case tc.wantError != "" && err == nil:
				t.Fatalf("requiredOther() error = nil, want %q", tc.wantError)
			case tc.wantError != "" && err.Error() != tc.wantError:
				t.Fatalf("requiredOther() error = %q, want %q", err.Error(), tc.wantError)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
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
