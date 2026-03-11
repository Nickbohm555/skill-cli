package refinement

import (
	"reflect"
	"strings"
	"testing"
)

func TestFlowRunProgressesToCommitReadyReview(t *testing.T) {
	t.Parallel()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	asker := &stubQuestionAsker{
		primaryAnswers: readyAnswers(),
	}
	handoff := &stubSummarizeFirstHandler{}

	flow, err := NewFlow(state, asker, handoff)
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}

	result, err := flow.Run()
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.State != FlowStateReview {
		t.Fatalf("Run().State = %q, want %q", result.State, FlowStateReview)
	}
	if !result.Report.CommitReady {
		t.Fatalf("Run().Report.CommitReady = false, want true")
	}
	if len(result.Report.MissingFields) != 0 {
		t.Fatalf("Run().Report.MissingFields = %v, want empty", result.Report.MissingFields)
	}
	if len(handoff.calls) != 0 {
		t.Fatalf("summarize-first calls = %d, want 0 for clear primary answers", len(handoff.calls))
	}

	wantPrimaryOrder := state.RequiredFields()
	if !reflect.DeepEqual(asker.primaryOrder, wantPrimaryOrder) {
		t.Fatalf("primary order = %v, want %v", asker.primaryOrder, wantPrimaryOrder)
	}

	if !hasEvent(result.Events, FlowEventEnterReview, "") {
		t.Fatalf("Run().Events missing %q event: %v", FlowEventEnterReview, result.Events)
	}
}

func TestFlowHandoffOccursBeforeDeepeningSequence(t *testing.T) {
	t.Parallel()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	answers := readyAnswers()
	answers[FieldPurposeSummary] = []string{"Things maybe."}

	asker := &stubQuestionAsker{
		primaryAnswers: answers,
		deepeningAnswers: map[FieldID][]string{
			FieldPurposeSummary: {
				"Generate a skill from one docs URL because the workflow must stay deterministic and focused on installable output.",
			},
		},
	}
	handoff := &stubSummarizeFirstHandler{asker: asker}

	flow, err := NewFlow(state, asker, handoff)
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}

	result, err := flow.Run()
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.State != FlowStateReview {
		t.Fatalf("Run().State = %q, want %q", result.State, FlowStateReview)
	}
	if !result.Report.CommitReady {
		t.Fatalf("Run().Report.CommitReady = false, want true")
	}

	wantSequence := []string{
		"primary:" + string(FieldPurposeSummary),
		"handoff:" + string(FieldPurposeSummary),
		"deepening:" + string(FieldPurposeSummary) + ":" + string(DeepeningModeFreeText),
	}
	if !reflect.DeepEqual(asker.sequence[:len(wantSequence)], wantSequence) {
		t.Fatalf("sequence prefix = %v, want %v", asker.sequence[:len(wantSequence)], wantSequence)
	}

	handoffIndex := indexOfEvent(result.Events, FlowEventSummarizeFirstHandoff, FieldPurposeSummary)
	deepeningIndex := indexOfEvent(result.Events, FlowEventAskDeepening, FieldPurposeSummary)
	if handoffIndex == -1 || deepeningIndex == -1 {
		t.Fatalf("expected handoff and deepening events for %q: %v", FieldPurposeSummary, result.Events)
	}
	if handoffIndex > deepeningIndex {
		t.Fatalf("handoff event index = %d, deepening event index = %d; want handoff before deepening", handoffIndex, deepeningIndex)
	}
}

func TestFlowSequenceStopsAtDeepeningAttemptCap(t *testing.T) {
	t.Parallel()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	answers := readyAnswers()
	answers[FieldConstraints] = []string{"Maybe keep it flexible and stuff."}

	asker := &stubQuestionAsker{
		primaryAnswers: answers,
		deepeningAnswers: map[FieldID][]string{
			FieldConstraints: {
				"Things maybe.",
				"Stuff and so on.",
				"Unknown, depends.",
			},
		},
	}
	handoff := &stubSummarizeFirstHandler{}

	flow, err := NewFlow(state, asker, handoff)
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}

	result, err := flow.Run()
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.State != FlowStateReview {
		t.Fatalf("Run().State = %q, want %q", result.State, FlowStateReview)
	}
	if result.Report.CommitReady {
		t.Fatalf("Run().Report.CommitReady = true, want false")
	}
	if flow.Attempts(FieldConstraints) != 3 {
		t.Fatalf("Attempts(%q) = %d, want 3", FieldConstraints, flow.Attempts(FieldConstraints))
	}
	if !hasEvent(result.Events, FlowEventDeepeningAttemptCapped, FieldConstraints) {
		t.Fatalf("Run().Events missing %q event for %q: %v", FlowEventDeepeningAttemptCapped, FieldConstraints, result.Events)
	}

	fieldReport := result.Report.Fields[FieldConstraints]
	if fieldReport.Status != ReadinessNeedsAttention {
		t.Fatalf("field status = %q, want %q", fieldReport.Status, ReadinessNeedsAttention)
	}
	if !reflect.DeepEqual(fieldReport.Reasons, []ValidationReason{ValidationReasonNeedsRevalidation, ValidationReasonLowClarity}) {
		t.Fatalf("field reasons = %v, want revalidation + low clarity", fieldReport.Reasons)
	}
}

func TestFlowReviseReasksDirectDependentsOnly(t *testing.T) {
	t.Parallel()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	asker := &stubQuestionAsker{primaryAnswers: readyAnswers()}
	handoff := &stubSummarizeFirstHandler{}

	flow, err := NewFlow(state, asker, handoff)
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}
	if _, err := flow.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	asker.primaryOrder = nil
	asker.sequence = nil
	asker.primaryAnswers = revisedDirectDependentAnswers()

	result, err := flow.Revise(
		"revise purpose_summary",
		"Refocus the skill on Go documentation only, with deterministic extraction, install steps, and clear source boundaries.",
	)
	if err != nil {
		t.Fatalf("Revise() error = %v", err)
	}

	wantReasked := []FieldID{
		FieldPrimaryTasks,
		FieldSuccessCriteria,
		FieldExampleRequests,
		FieldInScope,
		FieldOutOfScope,
	}
	if !reflect.DeepEqual(asker.primaryOrder, wantReasked) {
		t.Fatalf("reasked order = %v, want %v", asker.primaryOrder, wantReasked)
	}

	transitive, _ := state.Field(FieldExampleOutputs)
	if transitive.Status != ReadinessNeedsAttention {
		t.Fatalf("transitive impacted status = %q, want %q", transitive.Status, ReadinessNeedsAttention)
	}
	if containsField(asker.primaryOrder, FieldExampleOutputs) {
		t.Fatalf("transitive impacted field %q should not be re-asked directly", FieldExampleOutputs)
	}
	if result.Report.CommitReady {
		t.Fatalf("Revise().Report.CommitReady = true, want false with transitive follow-up still reopened")
	}
}

func TestFlowReviseRejectsInvalidTarget(t *testing.T) {
	t.Parallel()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	flow, err := NewFlow(state, &stubQuestionAsker{}, &stubSummarizeFirstHandler{})
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}

	if _, err := flow.Revise("revise does_not_exist", "new answer"); err == nil {
		t.Fatal("Revise() error = nil, want invalid target error")
	}
	if _, err := flow.Revise("revise purpose_summary now", "new answer"); err == nil {
		t.Fatal("Revise() error = nil, want strict command-form error")
	}
}

func TestFlowReviseBlocksCommitUntilImpactedFieldsAreResolved(t *testing.T) {
	t.Parallel()

	state, err := NewSessionState()
	if err != nil {
		t.Fatalf("NewSessionState() error = %v", err)
	}

	asker := &stubQuestionAsker{primaryAnswers: readyAnswers()}
	handoff := &stubSummarizeFirstHandler{}

	flow, err := NewFlow(state, asker, handoff)
	if err != nil {
		t.Fatalf("NewFlow() error = %v", err)
	}
	if _, err := flow.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	asker.primaryAnswers = revisedDirectDependentAnswers()
	result, err := flow.Revise(
		"revise purpose_summary",
		"Refocus the skill on Go documentation only, with deterministic extraction, install steps, and clear source boundaries.",
	)
	if err != nil {
		t.Fatalf("Revise() error = %v", err)
	}

	if result.Report.CommitReady {
		t.Fatalf("Revise().Report.CommitReady = true, want false")
	}
	if _, err := flow.Commit(); err == nil {
		t.Fatal("Commit() error = nil, want blocked commit after revision drift")
	}

	fieldReport := result.Report.Fields[FieldExampleOutputs]
	if fieldReport.Status != ReadinessNeedsAttention {
		t.Fatalf("field status = %q, want %q", fieldReport.Status, ReadinessNeedsAttention)
	}
	if !containsReason(fieldReport.Reasons, ValidationReasonNeedsRevalidation) {
		t.Fatalf("field reasons = %v, want %q included", fieldReport.Reasons, ValidationReasonNeedsRevalidation)
	}
}

type stubQuestionAsker struct {
	primaryAnswers   map[FieldID][]string
	deepeningAnswers map[FieldID][]string
	primaryOrder     []FieldID
	sequence         []string
}

func (s *stubQuestionAsker) AskPrimary(field FieldState) (string, error) {
	s.primaryOrder = append(s.primaryOrder, field.Definition.ID)
	s.sequence = append(s.sequence, "primary:"+string(field.Definition.ID))
	return shiftAnswer(s.primaryAnswers, field.Definition.ID), nil
}

func (s *stubQuestionAsker) AskDeepening(field FieldState, decision DeepeningDecision, signal SummarizeFirstSignal) (string, error) {
	s.sequence = append(s.sequence, "deepening:"+string(field.Definition.ID)+":"+string(decision.Mode))
	if signal.FieldID != field.Definition.ID {
		return "", errString("signal field mismatch")
	}
	if strings.TrimSpace(signal.Answer) == "" {
		return "", errString("signal answer missing")
	}
	return shiftAnswer(s.deepeningAnswers, field.Definition.ID), nil
}

type stubSummarizeFirstHandler struct {
	calls []FieldID
	asker *stubQuestionAsker
}

func (s *stubSummarizeFirstHandler) SummarizeFirst(field FieldState, decision DeepeningDecision) (SummarizeFirstSignal, error) {
	s.calls = append(s.calls, field.Definition.ID)
	if s.asker != nil {
		s.asker.sequence = append(s.asker.sequence, "handoff:"+string(field.Definition.ID))
	}
	return SummarizeFirstSignal{
		FieldID:  field.Definition.ID,
		Section:  field.Definition.Section,
		Answer:   field.Answer.Value,
		Summary:  "Current answer needs more specificity before review.",
		Decision: decision,
	}, nil
}

type errString string

func (e errString) Error() string { return string(e) }

func readyAnswers() map[FieldID][]string {
	return map[FieldID][]string{
		FieldPurposeSummary: {
			"Generate a Codex skill from one docs URL, including install steps, scope boundaries, and review-ready examples.",
		},
		FieldPrimaryTasks: {
			"Capture the docs source, extract implementation guidance, and turn it into a focused skill with explicit installation steps.",
		},
		FieldSuccessCriteria: {
			"The generated skill is installable, scoped to one domain, and includes concrete usage examples plus constraints.",
		},
		FieldConstraints: {
			"Use one docs URL only, keep the skill deterministic, and exclude unsupported setup steps or speculative workflows.",
		},
		FieldDependencies: {
			"Requires network access for docs fetches, a reachable documentation site, and OpenAI credentials only when structured summarization is enabled.",
		},
		FieldExampleRequests: {
			"Examples should include generating a skill from Go docs and refining boundaries when the docs mix tutorials with API references.",
		},
		FieldExampleOutputs: {
			"Output examples must show install commands, supported inputs, and one explicit out-of-scope case for the final skill.",
		},
		FieldInScope: {
			"In scope: extracting skill instructions, installation notes, supported commands, and concrete examples from the chosen docs set.",
		},
		FieldOutOfScope: {
			"Out of scope: building unrelated tooling, inventing missing APIs, or merging content from multiple unrelated documentation sites.",
		},
	}
}

func shiftAnswer(answers map[FieldID][]string, fieldID FieldID) string {
	if answers == nil {
		return ""
	}
	values := answers[fieldID]
	if len(values) == 0 {
		return ""
	}
	answer := values[0]
	answers[fieldID] = values[1:]
	return answer
}

func hasEvent(events []FlowEvent, wantType FlowEventType, fieldID FieldID) bool {
	return indexOfEvent(events, wantType, fieldID) >= 0
}

func indexOfEvent(events []FlowEvent, wantType FlowEventType, fieldID FieldID) int {
	for index, event := range events {
		if event.Type != wantType {
			continue
		}
		if fieldID == "" || event.FieldID == fieldID {
			return index
		}
	}
	return -1
}

func revisedDirectDependentAnswers() map[FieldID][]string {
	return map[FieldID][]string{
		FieldPrimaryTasks: {
			"Extract the Go docs guidance from one source, turn it into a Codex skill, and keep the generated instructions installable and scoped.",
		},
		FieldSuccessCriteria: {
			"The skill installs cleanly, stays anchored to Go documentation, and includes concrete operating constraints plus usable examples.",
		},
		FieldExampleRequests: {
			"Examples should cover generating a Go docs skill and tightening scope when the source mixes reference material with tutorials.",
		},
		FieldInScope: {
			"In scope: Go documentation extraction, skill instruction synthesis, install steps, and supported request examples from that source.",
		},
		FieldOutOfScope: {
			"Out of scope: mixing non-Go sources, inventing undocumented capabilities, or broadening the skill beyond the chosen Go docs site.",
		},
	}
}

func containsField(fields []FieldID, want FieldID) bool {
	for _, field := range fields {
		if field == want {
			return true
		}
	}
	return false
}

func containsReason(reasons []ValidationReason, want ValidationReason) bool {
	for _, reason := range reasons {
		if reason == want {
			return true
		}
	}
	return false
}
