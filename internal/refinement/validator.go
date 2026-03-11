package refinement

import "strings"

type ValidationReason string

const (
	ValidationReasonRequiredMissing   ValidationReason = "required_missing"
	ValidationReasonLowClarity        ValidationReason = "low_clarity"
	ValidationReasonNeedsRevalidation ValidationReason = "needs_revalidation"
)

type FieldValidation struct {
	Field       FieldState
	Status      ReadinessStatus
	Required    bool
	Clarity     ClarityAssessment
	Reasons     []ValidationReason
	CommitReady bool
}

type SectionValidation struct {
	Section     SectionID
	Fields      []FieldValidation
	CommitReady bool
}

type ValidationReport struct {
	Sections       []SectionValidation
	Fields         map[FieldID]FieldValidation
	CommitReady    bool
	MissingFields  []FieldID
	NeedsAttention []FieldID
}

type Validator struct {
	clarityPolicy ClarityPolicy
}

func DefaultValidator() Validator {
	return Validator{clarityPolicy: DefaultClarityPolicy()}
}

func NewValidator(policy ClarityPolicy) Validator {
	return Validator{clarityPolicy: policy}
}

func (v Validator) Evaluate(state *SessionState) (ValidationReport, error) {
	report := ValidationReport{
		Sections:       make([]SectionValidation, 0, len(state.OrderedSections())),
		Fields:         make(map[FieldID]FieldValidation, len(state.RequiredFields())),
		CommitReady:    true,
		MissingFields:  make([]FieldID, 0),
		NeedsAttention: make([]FieldID, 0),
	}

	for _, sectionID := range state.OrderedSections() {
		section := SectionValidation{
			Section:     sectionID,
			Fields:      make([]FieldValidation, 0, len(state.SectionFields(sectionID))),
			CommitReady: true,
		}

		for _, fieldID := range state.SectionFields(sectionID) {
			field, ok := state.Field(fieldID)
			if !ok {
				continue
			}

			validation, err := v.evaluateField(field)
			if err != nil {
				return ValidationReport{}, err
			}

			section.Fields = append(section.Fields, validation)
			report.Fields[fieldID] = validation

			if !validation.CommitReady {
				section.CommitReady = false
				report.CommitReady = false
				switch validation.Status {
				case ReadinessMissing:
					report.MissingFields = append(report.MissingFields, fieldID)
				case ReadinessNeedsAttention:
					report.NeedsAttention = append(report.NeedsAttention, fieldID)
				}
			}
		}

		report.Sections = append(report.Sections, section)
	}

	return report, nil
}

func (v Validator) evaluateField(field FieldState) (FieldValidation, error) {
	assessment, err := v.clarityPolicy.Assess(field.Definition.ID, field.Answer.Value)
	if err != nil {
		return FieldValidation{}, err
	}

	validation := FieldValidation{
		Field:       field,
		Status:      field.Status,
		Required:    field.Definition.Required,
		Clarity:     assessment,
		Reasons:     make([]ValidationReason, 0, 2),
		CommitReady: true,
	}

	if !field.Definition.Required {
		return validation, nil
	}

	if strings.TrimSpace(field.Answer.Value) == "" || field.Status == ReadinessMissing {
		validation.Status = ReadinessMissing
		validation.Reasons = append(validation.Reasons, ValidationReasonRequiredMissing)
		validation.CommitReady = false
		return validation, nil
	}

	if field.Status == ReadinessNeedsAttention {
		validation.Status = ReadinessNeedsAttention
		validation.Reasons = append(validation.Reasons, ValidationReasonNeedsRevalidation)
		validation.CommitReady = false
		if !assessment.Pass {
			validation.Reasons = append(validation.Reasons, ValidationReasonLowClarity)
		}
		return validation, nil
	}

	if !assessment.Pass {
		validation.Status = ReadinessNeedsAttention
		validation.Reasons = append(validation.Reasons, ValidationReasonLowClarity)
		validation.CommitReady = false
		return validation, nil
	}

	validation.Status = ReadinessReady
	return validation, nil
}
