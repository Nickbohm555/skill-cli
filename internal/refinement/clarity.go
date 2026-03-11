package refinement

import (
	"fmt"
	"strings"
)

type DeepeningMode string

const (
	DeepeningModeNone             DeepeningMode = "none"
	DeepeningModeFreeText         DeepeningMode = "free_text"
	DeepeningModeStructuredChoice DeepeningMode = "structured_choice"
	DeepeningModeCapped           DeepeningMode = "capped"
)

const defaultMaxDeepeningAttempts = 2

var defaultClarityThresholds = map[FieldID]int{
	FieldPurposeSummary:  6,
	FieldPrimaryTasks:    6,
	FieldSuccessCriteria: 6,
	FieldConstraints:     6,
	FieldDependencies:    5,
	FieldExampleRequests: 5,
	FieldExampleOutputs:  5,
	FieldInScope:         5,
	FieldOutOfScope:      5,
}

var ambiguityPhrases = []string{
	"etc",
	"and so on",
	"something",
	"things",
	"stuff",
	"maybe",
	"probably",
	"kind of",
	"sort of",
	"whatever",
	"not sure",
	"tbd",
	"unknown",
	"depends",
}

var specificityPhrases = []string{
	"because",
	"for example",
	"including",
	"such as",
	"must",
	"should",
	"when",
	"if",
	"using",
	"via",
	"exclude",
	"excluding",
	"only",
}

type ClarityPolicy struct {
	thresholds           map[FieldID]int
	maxDeepeningAttempts int
}

type ClarityAssessment struct {
	FieldID     FieldID
	Score       int
	Threshold   int
	Pass        bool
	WordCount   int
	Signals     []string
	Penalties   []string
	AnswerValue string
}

type DeepeningDecision struct {
	FieldID              FieldID
	Mode                 DeepeningMode
	Attempt              int
	MaxAttempts          int
	RequireExplicitOther bool
	Reason               string
}

func DefaultClarityPolicy() ClarityPolicy {
	thresholds := make(map[FieldID]int, len(defaultClarityThresholds))
	for fieldID, threshold := range defaultClarityThresholds {
		thresholds[fieldID] = threshold
	}

	return ClarityPolicy{
		thresholds:           thresholds,
		maxDeepeningAttempts: defaultMaxDeepeningAttempts,
	}
}

func (p ClarityPolicy) Threshold(fieldID FieldID) (int, error) {
	threshold, ok := p.thresholds[fieldID]
	if !ok {
		return 0, fmt.Errorf("unknown clarity threshold for field %q", fieldID)
	}
	return threshold, nil
}

func (p ClarityPolicy) Assess(fieldID FieldID, answer string) (ClarityAssessment, error) {
	threshold, err := p.Threshold(fieldID)
	if err != nil {
		return ClarityAssessment{}, err
	}

	trimmed := strings.TrimSpace(answer)
	normalized := strings.ToLower(trimmed)
	words := strings.Fields(trimmed)

	score := 0
	signals := make([]string, 0, 4)
	penalties := make([]string, 0, 2)

	switch count := len(words); {
	case count == 0:
		penalties = append(penalties, "missing_answer")
	case count <= 3:
		score += 1
		penalties = append(penalties, "too_short")
	case count <= 7:
		score += 2
		signals = append(signals, "minimum_length")
	case count <= 15:
		score += 4
		signals = append(signals, "good_length")
	default:
		score += 5
		signals = append(signals, "strong_length")
	}

	if hasStructuredDetail(trimmed) {
		score++
		signals = append(signals, "structured_detail")
	}

	if hasConcreteDetail(trimmed) {
		score++
		signals = append(signals, "concrete_detail")
	}

	specificityHits := countPhraseHits(normalized, specificityPhrases)
	switch {
	case specificityHits >= 2:
		score += 2
		signals = append(signals, "specificity_markers", "multiple_specificity_markers")
	case specificityHits == 1:
		score++
		signals = append(signals, "specificity_markers")
	}

	ambiguityHits := countPhraseHits(normalized, ambiguityPhrases)
	switch {
	case ambiguityHits >= 2:
		score -= 4
		penalties = append(penalties, "ambiguous_language", "multiple_ambiguity_markers")
	case ambiguityHits == 1:
		score -= 2
		penalties = append(penalties, "ambiguous_language")
	}

	if score < 0 {
		score = 0
	}
	if score > 10 {
		score = 10
	}

	return ClarityAssessment{
		FieldID:     fieldID,
		Score:       score,
		Threshold:   threshold,
		Pass:        score >= threshold,
		WordCount:   len(words),
		Signals:     signals,
		Penalties:   penalties,
		AnswerValue: trimmed,
	}, nil
}

func (p ClarityPolicy) MeetsThreshold(fieldID FieldID, answer string) (bool, error) {
	assessment, err := p.Assess(fieldID, answer)
	if err != nil {
		return false, err
	}
	return assessment.Pass, nil
}

func (p ClarityPolicy) DeepeningDecision(fieldID FieldID, answer string, attempts int) (DeepeningDecision, error) {
	if attempts < 0 {
		return DeepeningDecision{}, fmt.Errorf("attempt count cannot be negative")
	}

	assessment, err := p.Assess(fieldID, answer)
	if err != nil {
		return DeepeningDecision{}, err
	}

	decision := DeepeningDecision{
		FieldID:     fieldID,
		Attempt:     attempts,
		MaxAttempts: p.maxDeepeningAttempts,
	}

	if assessment.Pass {
		decision.Mode = DeepeningModeNone
		decision.Reason = "clarity_threshold_met"
		return decision, nil
	}

	switch {
	case attempts == 0:
		decision.Mode = DeepeningModeFreeText
		decision.Reason = "low_clarity_targeted_follow_up"
	case attempts < p.maxDeepeningAttempts:
		decision.Mode = DeepeningModeStructuredChoice
		decision.RequireExplicitOther = true
		decision.Reason = "low_clarity_switch_to_structured_choice"
	default:
		decision.Mode = DeepeningModeCapped
		decision.RequireExplicitOther = true
		decision.Reason = "max_deepening_attempts_reached"
	}

	return decision, nil
}

func hasStructuredDetail(answer string) bool {
	return strings.ContainsAny(answer, "\n,;:")
}

func hasConcreteDetail(answer string) bool {
	return strings.ContainsAny(answer, "0123456789/`\"'()[]{}")
}

func countPhraseHits(answer string, phrases []string) int {
	hits := 0
	for _, phrase := range phrases {
		if strings.Contains(answer, phrase) {
			hits++
		}
	}
	return hits
}
