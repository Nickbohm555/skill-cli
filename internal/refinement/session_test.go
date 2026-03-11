package refinement

import (
	"reflect"
	"testing"
)

func TestSessionStateInitializesRequiredFieldsAndSections(t *testing.T) {
	t.Parallel()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	wantSections := []SectionID{
		SectionPurpose,
		SectionConstraints,
		SectionExamples,
		SectionBoundaries,
	}
	if got := state.OrderedSections(); !reflect.DeepEqual(got, wantSections) {
		t.Fatalf("OrderedSections() = %v, want %v", got, wantSections)
	}

	wantPurposeFields := []FieldID{
		FieldPurposeSummary,
		FieldPrimaryTasks,
		FieldSuccessCriteria,
	}
	if got := state.SectionFields(SectionPurpose); !reflect.DeepEqual(got, wantPurposeFields) {
		t.Fatalf("SectionFields(%q) = %v, want %v", SectionPurpose, got, wantPurposeFields)
	}

	for _, fieldID := range state.RequiredFields() {
		field, ok := state.Field(fieldID)
		if !ok {
			t.Fatalf("Field(%q) missing from session", fieldID)
		}
		if !field.Definition.Required {
			t.Fatalf("Field(%q) should be required", fieldID)
		}
		if field.Status != ReadinessMissing {
			t.Fatalf("Field(%q) status = %q, want %q", fieldID, field.Status, ReadinessMissing)
		}
		if field.Answer.Value != "" || field.Answer.Revision != 0 {
			t.Fatalf("Field(%q) answer = %+v, want zero value", fieldID, field.Answer)
		}
	}
}

func TestSessionFieldGraphRevisionMarksImpactedFieldsNeedsAttention(t *testing.T) {
	t.Parallel()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	graph := DefaultFieldGraph()
	allFields := []FieldID{
		FieldPurposeSummary,
		FieldPrimaryTasks,
		FieldSuccessCriteria,
		FieldConstraints,
		FieldDependencies,
		FieldExampleRequests,
		FieldExampleOutputs,
		FieldInScope,
		FieldOutOfScope,
	}

	for _, fieldID := range allFields {
		if err := state.SetAnswer(fieldID, "filled "+string(fieldID)); err != nil {
			t.Fatalf("SetAnswer(%q) error = %v", fieldID, err)
		}
		if err := state.MarkReady(fieldID); err != nil {
			t.Fatalf("MarkReady(%q) error = %v", fieldID, err)
		}
	}

	impacted, err := state.ReviseAnswer(FieldPurposeSummary, "updated purpose", graph)
	if err != nil {
		t.Fatalf("ReviseAnswer() error = %v", err)
	}

	wantImpacted := []FieldID{
		FieldPrimaryTasks,
		FieldSuccessCriteria,
		FieldExampleRequests,
		FieldExampleOutputs,
		FieldInScope,
		FieldOutOfScope,
	}
	if !reflect.DeepEqual(impacted, wantImpacted) {
		t.Fatalf("ReviseAnswer() impacted = %v, want %v", impacted, wantImpacted)
	}

	revised, _ := state.Field(FieldPurposeSummary)
	if revised.Status != ReadinessNeedsAttention {
		t.Fatalf("revised field status = %q, want %q", revised.Status, ReadinessNeedsAttention)
	}
	if revised.Answer.Revision == 0 {
		t.Fatalf("revised field revision was not incremented: %+v", revised.Answer)
	}

	for _, fieldID := range wantImpacted {
		field, _ := state.Field(fieldID)
		if field.Status != ReadinessNeedsAttention {
			t.Fatalf("impacted field %q status = %q, want %q", fieldID, field.Status, ReadinessNeedsAttention)
		}
	}

	constraints, _ := state.Field(FieldConstraints)
	if constraints.Status != ReadinessReady {
		t.Fatalf("unrelated field %q status = %q, want %q", FieldConstraints, constraints.Status, ReadinessReady)
	}

	dependencies, _ := state.Field(FieldDependencies)
	if dependencies.Status != ReadinessReady {
		t.Fatalf("unrelated field %q status = %q, want %q", FieldDependencies, dependencies.Status, ReadinessReady)
	}
}
