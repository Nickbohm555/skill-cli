package refinement

import (
	"reflect"
	"testing"
)

func TestValidatorEvaluateCommitGateBehavior(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		setup              func(t *testing.T) *SessionState
		wantCommitReady    bool
		wantMissing        []FieldID
		wantNeedsAttention []FieldID
		wantStatuses       map[FieldID]ReadinessStatus
		wantReasons        map[FieldID][]ValidationReason
	}{
		{
			name: "missing required fields fail closed",
			setup: func(t *testing.T) *SessionState {
				t.Helper()
				return newReadySession(t, map[FieldID]string{
					FieldOutOfScope: "",
				})
			},
			wantCommitReady:    false,
			wantMissing:        []FieldID{FieldOutOfScope},
			wantNeedsAttention: []FieldID{},
			wantStatuses: map[FieldID]ReadinessStatus{
				FieldOutOfScope: ReadinessMissing,
			},
			wantReasons: map[FieldID][]ValidationReason{
				FieldOutOfScope: {ValidationReasonRequiredMissing},
			},
		},
		{
			name: "low clarity required fields block commit",
			setup: func(t *testing.T) *SessionState {
				t.Helper()
				return newReadySession(t, map[FieldID]string{
					FieldConstraints: "Maybe keep it flexible and stuff.",
				})
			},
			wantCommitReady:    false,
			wantMissing:        []FieldID{},
			wantNeedsAttention: []FieldID{FieldConstraints},
			wantStatuses: map[FieldID]ReadinessStatus{
				FieldConstraints: ReadinessNeedsAttention,
			},
			wantReasons: map[FieldID][]ValidationReason{
				FieldConstraints: {ValidationReasonLowClarity},
			},
		},
		{
			name: "revision drift keeps impacted fields gated",
			setup: func(t *testing.T) *SessionState {
				t.Helper()
				state := newReadySession(t, nil)
				_, err := state.ReviseAnswer(FieldPurposeSummary, "Refocus the skill on database docs only, excluding generic examples.", DefaultFieldGraph())
				if err != nil {
					t.Fatalf("ReviseAnswer() error = %v", err)
				}
				return state
			},
			wantCommitReady:    false,
			wantMissing:        []FieldID{},
			wantNeedsAttention: []FieldID{FieldPurposeSummary, FieldPrimaryTasks, FieldSuccessCriteria, FieldExampleRequests, FieldExampleOutputs, FieldInScope, FieldOutOfScope},
			wantStatuses: map[FieldID]ReadinessStatus{
				FieldPurposeSummary:  ReadinessNeedsAttention,
				FieldPrimaryTasks:    ReadinessNeedsAttention,
				FieldSuccessCriteria: ReadinessNeedsAttention,
				FieldExampleRequests: ReadinessNeedsAttention,
				FieldExampleOutputs:  ReadinessNeedsAttention,
				FieldInScope:         ReadinessNeedsAttention,
				FieldOutOfScope:      ReadinessNeedsAttention,
			},
			wantReasons: map[FieldID][]ValidationReason{
				FieldPurposeSummary:  {ValidationReasonNeedsRevalidation},
				FieldPrimaryTasks:    {ValidationReasonNeedsRevalidation},
				FieldSuccessCriteria: {ValidationReasonNeedsRevalidation},
				FieldExampleRequests: {ValidationReasonNeedsRevalidation},
				FieldExampleOutputs:  {ValidationReasonNeedsRevalidation},
				FieldInScope:         {ValidationReasonNeedsRevalidation},
				FieldOutOfScope:      {ValidationReasonNeedsRevalidation},
			},
		},
		{
			name: "fully ready session passes commit gate",
			setup: func(t *testing.T) *SessionState {
				t.Helper()
				return newReadySession(t, nil)
			},
			wantCommitReady:    true,
			wantMissing:        []FieldID{},
			wantNeedsAttention: []FieldID{},
			wantStatuses: map[FieldID]ReadinessStatus{
				FieldPurposeSummary:  ReadinessReady,
				FieldPrimaryTasks:    ReadinessReady,
				FieldSuccessCriteria: ReadinessReady,
				FieldConstraints:     ReadinessReady,
				FieldDependencies:    ReadinessReady,
				FieldExampleRequests: ReadinessReady,
				FieldExampleOutputs:  ReadinessReady,
				FieldInScope:         ReadinessReady,
				FieldOutOfScope:      ReadinessReady,
			},
		},
	}

	validator := DefaultValidator()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state := tt.setup(t)
			report, err := validator.Evaluate(state)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}

			if report.CommitReady != tt.wantCommitReady {
				t.Fatalf("CommitReady = %v, want %v", report.CommitReady, tt.wantCommitReady)
			}
			if !reflect.DeepEqual(report.MissingFields, tt.wantMissing) {
				t.Fatalf("MissingFields = %v, want %v", report.MissingFields, tt.wantMissing)
			}
			if !reflect.DeepEqual(report.NeedsAttention, tt.wantNeedsAttention) {
				t.Fatalf("NeedsAttention = %v, want %v", report.NeedsAttention, tt.wantNeedsAttention)
			}

			for fieldID, wantStatus := range tt.wantStatuses {
				got, ok := report.Fields[fieldID]
				if !ok {
					t.Fatalf("report.Fields missing %q", fieldID)
				}
				if got.Status != wantStatus {
					t.Fatalf("field %q status = %q, want %q", fieldID, got.Status, wantStatus)
				}
			}

			for fieldID, wantReasons := range tt.wantReasons {
				got := report.Fields[fieldID]
				if !reflect.DeepEqual(got.Reasons, wantReasons) {
					t.Fatalf("field %q reasons = %v, want %v", fieldID, got.Reasons, wantReasons)
				}
			}

			if len(report.Sections) != len(state.OrderedSections()) {
				t.Fatalf("Sections length = %d, want %d", len(report.Sections), len(state.OrderedSections()))
			}
		})
	}
}

func newReadySession(t *testing.T, overrides map[FieldID]string) *SessionState {
	t.Helper()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	answers := map[FieldID]string{
		FieldPurposeSummary:  "Generate a Codex skill from one docs URL, including install steps, scope boundaries, and review-ready examples.",
		FieldPrimaryTasks:    "Capture the docs source, extract implementation guidance, and turn it into a focused skill with explicit installation steps.",
		FieldSuccessCriteria: "The generated skill is installable, scoped to one domain, and includes concrete usage examples plus constraints.",
		FieldConstraints:     "Use one docs URL only, keep the skill deterministic, and exclude unsupported setup steps or speculative workflows.",
		FieldDependencies:    "Requires network access for docs fetches, a reachable documentation site, and OpenAI credentials only when structured summarization is enabled.",
		FieldExampleRequests: "Examples should include generating a skill from Go docs and refining boundaries when the docs mix tutorials with API references.",
		FieldExampleOutputs:  "Output examples must show install commands, supported inputs, and one explicit out-of-scope case for the final skill.",
		FieldInScope:         "In scope: extracting skill instructions, installation notes, supported commands, and concrete examples from the chosen docs set.",
		FieldOutOfScope:      "Out of scope: building unrelated tooling, inventing missing APIs, or merging content from multiple unrelated documentation sites.",
	}

	for fieldID, override := range overrides {
		answers[fieldID] = override
	}

	for _, fieldID := range state.RequiredFields() {
		if err := state.SetAnswer(fieldID, answers[fieldID]); err != nil {
			t.Fatalf("SetAnswer(%q) error = %v", fieldID, err)
		}
		if answers[fieldID] == "" {
			continue
		}
		if err := state.MarkReady(fieldID); err != nil {
			t.Fatalf("MarkReady(%q) error = %v", fieldID, err)
		}
	}

	return state
}
