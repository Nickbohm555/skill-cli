package refinement

import (
	"fmt"
	"strings"
)

type SectionID string

const (
	SectionPurpose     SectionID = "purpose"
	SectionConstraints SectionID = "constraints"
	SectionExamples    SectionID = "examples"
	SectionBoundaries  SectionID = "boundaries"
)

type FieldID string

const (
	FieldPurposeSummary  FieldID = "purpose_summary"
	FieldPrimaryTasks    FieldID = "primary_tasks"
	FieldSuccessCriteria FieldID = "success_criteria"
	FieldConstraints     FieldID = "constraints"
	FieldDependencies    FieldID = "dependencies"
	FieldExampleRequests FieldID = "example_requests"
	FieldExampleOutputs  FieldID = "example_outputs"
	FieldInScope         FieldID = "in_scope"
	FieldOutOfScope      FieldID = "out_of_scope"
)

type ReadinessStatus string

const (
	ReadinessReady          ReadinessStatus = "ready"
	ReadinessNeedsAttention ReadinessStatus = "needs_attention"
	ReadinessMissing        ReadinessStatus = "missing"
)

var defaultSectionOrder = []SectionID{
	SectionPurpose,
	SectionConstraints,
	SectionExamples,
	SectionBoundaries,
}

// FieldDefinition declares one deterministic refinement field and the section
// it belongs to for later grouped review rendering.
type FieldDefinition struct {
	ID       FieldID
	Section  SectionID
	Label    string
	Required bool
}

// AnswerRecord stores the latest answer plus deterministic revision metadata.
type AnswerRecord struct {
	Value    string
	Revision int
}

// FieldState is the current session snapshot for one field.
type FieldState struct {
	Definition FieldDefinition
	Answer     AnswerRecord
	Status     ReadinessStatus
}

// SessionState owns required field contracts, answer storage, and explicit
// readiness tracking without depending on any prompt or transport code.
type SessionState struct {
	fieldOrder      []FieldID
	sectionOrder    []SectionID
	sectionFields   map[SectionID][]FieldID
	fields          map[FieldID]FieldState
	revisionCounter int
}

// DefaultFieldRegistry returns the required phase-3 field set in stable order.
func DefaultFieldRegistry() []FieldDefinition {
	return []FieldDefinition{
		{ID: FieldPurposeSummary, Section: SectionPurpose, Label: "Purpose Summary", Required: true},
		{ID: FieldPrimaryTasks, Section: SectionPurpose, Label: "Primary Tasks", Required: true},
		{ID: FieldSuccessCriteria, Section: SectionPurpose, Label: "Success Criteria", Required: true},
		{ID: FieldConstraints, Section: SectionConstraints, Label: "Constraints", Required: true},
		{ID: FieldDependencies, Section: SectionConstraints, Label: "Dependencies", Required: true},
		{ID: FieldExampleRequests, Section: SectionExamples, Label: "Example Requests", Required: true},
		{ID: FieldExampleOutputs, Section: SectionExamples, Label: "Example Outputs", Required: true},
		{ID: FieldInScope, Section: SectionBoundaries, Label: "In Scope", Required: true},
		{ID: FieldOutOfScope, Section: SectionBoundaries, Label: "Out Of Scope", Required: true},
	}
}

func NewSessionState() (*SessionState, error) {
	return NewSessionStateWithRegistry(DefaultFieldRegistry())
}

func NewSessionStateWithRegistry(registry []FieldDefinition) (*SessionState, error) {
	if len(registry) == 0 {
		return nil, fmt.Errorf("session field registry cannot be empty")
	}

	state := &SessionState{
		fieldOrder:    make([]FieldID, 0, len(registry)),
		sectionOrder:  append([]SectionID(nil), defaultSectionOrder...),
		sectionFields: make(map[SectionID][]FieldID, len(defaultSectionOrder)),
		fields:        make(map[FieldID]FieldState, len(registry)),
	}

	seen := make(map[FieldID]struct{}, len(registry))
	for _, def := range registry {
		if err := validateFieldDefinition(def); err != nil {
			return nil, err
		}
		if _, exists := seen[def.ID]; exists {
			return nil, fmt.Errorf("duplicate field definition %q", def.ID)
		}
		seen[def.ID] = struct{}{}

		state.fieldOrder = append(state.fieldOrder, def.ID)
		state.sectionFields[def.Section] = append(state.sectionFields[def.Section], def.ID)
		state.fields[def.ID] = FieldState{
			Definition: def,
			Status:     ReadinessMissing,
		}
	}

	for _, section := range state.sectionOrder {
		if len(state.sectionFields[section]) == 0 {
			return nil, fmt.Errorf("section %q has no registered fields", section)
		}
	}

	return state, nil
}

func validateFieldDefinition(def FieldDefinition) error {
	if strings.TrimSpace(string(def.ID)) == "" {
		return fmt.Errorf("field definition id cannot be empty")
	}
	if strings.TrimSpace(string(def.Section)) == "" {
		return fmt.Errorf("field definition %q section cannot be empty", def.ID)
	}
	if !isAllowedSection(def.Section) {
		return fmt.Errorf("field definition %q uses unsupported section %q", def.ID, def.Section)
	}
	if strings.TrimSpace(def.Label) == "" {
		return fmt.Errorf("field definition %q label cannot be empty", def.ID)
	}
	return nil
}

func isAllowedSection(section SectionID) bool {
	for _, candidate := range defaultSectionOrder {
		if candidate == section {
			return true
		}
	}
	return false
}

func (s *SessionState) OrderedSections() []SectionID {
	return append([]SectionID(nil), s.sectionOrder...)
}

func (s *SessionState) RequiredFields() []FieldID {
	fields := make([]FieldID, 0, len(s.fieldOrder))
	for _, fieldID := range s.fieldOrder {
		if s.fields[fieldID].Definition.Required {
			fields = append(fields, fieldID)
		}
	}
	return fields
}

func (s *SessionState) SectionFields(section SectionID) []FieldID {
	return append([]FieldID(nil), s.sectionFields[section]...)
}

func (s *SessionState) Field(fieldID FieldID) (FieldState, bool) {
	field, ok := s.fields[fieldID]
	if !ok {
		return FieldState{}, false
	}
	return field, true
}

func (s *SessionState) Snapshot() []FieldState {
	snapshot := make([]FieldState, 0, len(s.fieldOrder))
	for _, fieldID := range s.fieldOrder {
		snapshot = append(snapshot, s.fields[fieldID])
	}
	return snapshot
}

func (s *SessionState) Revision() int {
	return s.revisionCounter
}

func (s *SessionState) SetAnswer(fieldID FieldID, value string) error {
	field, ok := s.fields[fieldID]
	if !ok {
		return fmt.Errorf("unknown field %q", fieldID)
	}

	s.revisionCounter++
	field.Answer = AnswerRecord{
		Value:    strings.TrimSpace(value),
		Revision: s.revisionCounter,
	}

	if field.Answer.Value == "" {
		field.Status = ReadinessMissing
	} else {
		field.Status = ReadinessNeedsAttention
	}

	s.fields[fieldID] = field
	return nil
}

func (s *SessionState) MarkReady(fieldID FieldID) error {
	return s.updateStatus(fieldID, ReadinessReady)
}

func (s *SessionState) MarkMissing(fieldID FieldID) error {
	return s.updateStatus(fieldID, ReadinessMissing)
}

func (s *SessionState) MarkNeedsAttention(fieldIDs ...FieldID) error {
	for _, fieldID := range fieldIDs {
		field, ok := s.fields[fieldID]
		if !ok {
			return fmt.Errorf("unknown field %q", fieldID)
		}
		if strings.TrimSpace(field.Answer.Value) == "" {
			field.Status = ReadinessMissing
		} else {
			field.Status = ReadinessNeedsAttention
		}
		s.fields[fieldID] = field
	}
	return nil
}

func (s *SessionState) ReviseAnswer(fieldID FieldID, value string, graph FieldGraph) ([]FieldID, error) {
	if err := s.SetAnswer(fieldID, value); err != nil {
		return nil, err
	}

	impacted := graph.ImpactedBy(fieldID)
	if err := s.MarkNeedsAttention(impacted...); err != nil {
		return nil, err
	}

	return impacted, nil
}

func (s *SessionState) updateStatus(fieldID FieldID, status ReadinessStatus) error {
	field, ok := s.fields[fieldID]
	if !ok {
		return fmt.Errorf("unknown field %q", fieldID)
	}
	field.Status = status
	s.fields[fieldID] = field
	return nil
}
