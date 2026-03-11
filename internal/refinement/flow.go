package refinement

import (
	"fmt"
	"strings"
)

type FlowState string

const (
	FlowStateCollecting FlowState = "collecting"
	FlowStateReview     FlowState = "review"
	FlowStateCommitted  FlowState = "committed"
)

type FlowEventType string

const (
	FlowEventAskPrimary             FlowEventType = "ask_primary"
	FlowEventSummarizeFirstHandoff  FlowEventType = "summarize_first_handoff"
	FlowEventAskDeepening           FlowEventType = "ask_deepening"
	FlowEventFieldReady             FlowEventType = "field_ready"
	FlowEventEnterReview            FlowEventType = "enter_review"
	FlowEventDeepeningAttemptCapped FlowEventType = "deepening_attempt_capped"
	FlowEventCommit                 FlowEventType = "commit"
)

type FlowEvent struct {
	Type    FlowEventType
	FieldID FieldID
	Section SectionID
	Attempt int
	Detail  string
}

type SummarizeFirstSignal struct {
	FieldID  FieldID
	Section  SectionID
	Answer   string
	Summary  string
	Decision DeepeningDecision
}

type QuestionAsker interface {
	AskPrimary(field FieldState) (string, error)
	AskDeepening(field FieldState, decision DeepeningDecision, signal SummarizeFirstSignal) (string, error)
}

type SummarizeFirstHandler interface {
	SummarizeFirst(field FieldState, decision DeepeningDecision) (SummarizeFirstSignal, error)
}

type FlowResult struct {
	State  FlowState
	Report ValidationReport
	Events []FlowEvent
}

type Flow struct {
	state              *SessionState
	validator          Validator
	policy             ClarityPolicy
	asker              QuestionAsker
	handoff            SummarizeFirstHandler
	runtimeState       FlowState
	deepeningAttempts  map[FieldID]int
	events             []FlowEvent
	lastValidation     ValidationReport
	reviewStateEntered bool
}

func NewFlow(state *SessionState, asker QuestionAsker, handoff SummarizeFirstHandler) (*Flow, error) {
	if state == nil {
		return nil, fmt.Errorf("flow session state is required")
	}
	if asker == nil {
		return nil, fmt.Errorf("flow question asker is required")
	}
	if handoff == nil {
		return nil, fmt.Errorf("flow summarize-first handler is required")
	}

	return &Flow{
		state:             state,
		validator:         DefaultValidator(),
		policy:            DefaultClarityPolicy(),
		asker:             asker,
		handoff:           handoff,
		runtimeState:      FlowStateCollecting,
		deepeningAttempts: make(map[FieldID]int, len(state.RequiredFields())),
		events:            make([]FlowEvent, 0, len(state.RequiredFields())*2),
	}, nil
}

func (f *Flow) State() FlowState {
	return f.runtimeState
}

func (f *Flow) Events() []FlowEvent {
	return append([]FlowEvent(nil), f.events...)
}

func (f *Flow) Attempts(fieldID FieldID) int {
	return f.deepeningAttempts[fieldID]
}

func (f *Flow) Run() (FlowResult, error) {
	if f.runtimeState == FlowStateCommitted {
		return FlowResult{}, fmt.Errorf("flow already committed")
	}

	for _, sectionID := range f.state.OrderedSections() {
		for _, fieldID := range f.state.SectionFields(sectionID) {
			field, ok := f.state.Field(fieldID)
			if !ok {
				return FlowResult{}, fmt.Errorf("flow field %q not found", fieldID)
			}
			if field.Status == ReadinessReady {
				continue
			}
			if err := f.processField(field); err != nil {
				return FlowResult{}, err
			}
		}
	}

	report, err := f.validator.Evaluate(f.state)
	if err != nil {
		return FlowResult{}, err
	}
	f.lastValidation = report

	if len(report.MissingFields) > 0 {
		f.runtimeState = FlowStateCollecting
		return FlowResult{
			State:  f.runtimeState,
			Report: report,
			Events: f.Events(),
		}, nil
	}

	f.runtimeState = FlowStateReview
	if !f.reviewStateEntered {
		f.events = append(f.events, FlowEvent{
			Type:   FlowEventEnterReview,
			Detail: "all required fields reached baseline completion",
		})
		f.reviewStateEntered = true
	}

	return FlowResult{
		State:  f.runtimeState,
		Report: report,
		Events: f.Events(),
	}, nil
}

func (f *Flow) Commit() (FlowResult, error) {
	if f.runtimeState != FlowStateReview {
		return FlowResult{}, fmt.Errorf("flow must be in review before commit")
	}

	report, err := f.validator.Evaluate(f.state)
	if err != nil {
		return FlowResult{}, err
	}
	f.lastValidation = report
	if !report.CommitReady {
		return FlowResult{}, fmt.Errorf("commit blocked: required fields are missing, unclear, or need revalidation")
	}

	f.runtimeState = FlowStateCommitted
	f.events = append(f.events, FlowEvent{
		Type:   FlowEventCommit,
		Detail: "commit gate passed",
	})

	return FlowResult{
		State:  f.runtimeState,
		Report: report,
		Events: f.Events(),
	}, nil
}

func (f *Flow) processField(field FieldState) error {
	if strings.TrimSpace(field.Answer.Value) == "" {
		answer, err := f.asker.AskPrimary(field)
		if err != nil {
			return fmt.Errorf("ask primary for %q: %w", field.Definition.ID, err)
		}
		f.events = append(f.events, FlowEvent{
			Type:    FlowEventAskPrimary,
			FieldID: field.Definition.ID,
			Section: field.Definition.Section,
			Detail:  "primary question answered",
		})
		if err := f.state.SetAnswer(field.Definition.ID, answer); err != nil {
			return err
		}
	}

	for {
		current, ok := f.state.Field(field.Definition.ID)
		if !ok {
			return fmt.Errorf("flow field %q not found after update", field.Definition.ID)
		}

		decision, err := f.policy.DeepeningDecision(current.Definition.ID, current.Answer.Value, f.deepeningAttempts[current.Definition.ID])
		if err != nil {
			return err
		}

		if decision.Mode == DeepeningModeNone {
			if err := f.state.MarkReady(current.Definition.ID); err != nil {
				return err
			}
			f.events = append(f.events, FlowEvent{
				Type:    FlowEventFieldReady,
				FieldID: current.Definition.ID,
				Section: current.Definition.Section,
				Attempt: f.deepeningAttempts[current.Definition.ID],
				Detail:  "clarity threshold met",
			})
			return nil
		}

		signal, err := f.handoff.SummarizeFirst(current, decision)
		if err != nil {
			return fmt.Errorf("summarize-first handoff for %q: %w", current.Definition.ID, err)
		}
		if signal.FieldID == "" {
			signal.FieldID = current.Definition.ID
		}
		if signal.Section == "" {
			signal.Section = current.Definition.Section
		}
		if strings.TrimSpace(signal.Answer) == "" {
			signal.Answer = current.Answer.Value
		}
		signal.Decision = decision

		f.events = append(f.events, FlowEvent{
			Type:    FlowEventSummarizeFirstHandoff,
			FieldID: current.Definition.ID,
			Section: current.Definition.Section,
			Attempt: decision.Attempt,
			Detail:  strings.TrimSpace(signal.Summary),
		})

		answer, err := f.asker.AskDeepening(current, decision, signal)
		if err != nil {
			return fmt.Errorf("ask deepening for %q: %w", current.Definition.ID, err)
		}
		f.deepeningAttempts[current.Definition.ID]++
		f.events = append(f.events, FlowEvent{
			Type:    FlowEventAskDeepening,
			FieldID: current.Definition.ID,
			Section: current.Definition.Section,
			Attempt: f.deepeningAttempts[current.Definition.ID],
			Detail:  string(decision.Mode),
		})

		if err := f.state.SetAnswer(current.Definition.ID, answer); err != nil {
			return err
		}

		if decision.Mode == DeepeningModeCapped {
			f.events = append(f.events, FlowEvent{
				Type:    FlowEventDeepeningAttemptCapped,
				FieldID: current.Definition.ID,
				Section: current.Definition.Section,
				Attempt: f.deepeningAttempts[current.Definition.ID],
				Detail:  decision.Reason,
			})
			return nil
		}
	}
}
