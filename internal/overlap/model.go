package overlap

import "time"

type OverlapSeverity string

const (
	SeverityNone   OverlapSeverity = "none"
	SeverityLow    OverlapSeverity = "low"
	SeverityMedium OverlapSeverity = "medium"
	SeverityHigh   OverlapSeverity = "high"
)

type ResolutionMode string

const (
	ResolutionNewInstall ResolutionMode = "new_install"
	ResolutionUpdate     ResolutionMode = "update_existing"
	ResolutionMerge      ResolutionMode = "merge_with_existing"
	ResolutionAbort      ResolutionMode = "abort"
)

type SkillProfile struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	InScope     []string `json:"in_scope,omitempty"`
	OutOfScope  []string `json:"out_of_scope,omitempty"`
	Commands    []string `json:"commands,omitempty"`
	SourcePath  string   `json:"source_path,omitempty"`
}

type OverlapSignal struct {
	Key   string  `json:"key"`
	Value string  `json:"value,omitempty"`
	Score float64 `json:"score,omitempty"`
}

type ExplanationMetadata struct {
	Summary string          `json:"summary,omitempty"`
	RuleIDs []string        `json:"rule_ids,omitempty"`
	Signals []OverlapSignal `json:"signals,omitempty"`
}

type OverlapFinding struct {
	RuleID          string              `json:"rule_id"`
	ExistingSkillID string              `json:"existing_skill_id"`
	Severity        OverlapSeverity     `json:"severity"`
	Score           float64             `json:"score,omitempty"`
	Signals         []OverlapSignal     `json:"signals,omitempty"`
	Explanation     string              `json:"explanation,omitempty"`
	ExplanationMeta ExplanationMetadata `json:"explanation_meta,omitempty"`
}

type ConflictResolutionDecision struct {
	CandidateSkillID string              `json:"candidate_skill_id,omitempty"`
	TargetSkillID    string              `json:"target_skill_id,omitempty"`
	Mode             ResolutionMode      `json:"mode"`
	Blocking         bool                `json:"blocking"`
	SelectedAt       *time.Time          `json:"selected_at,omitempty"`
	Explanation      string              `json:"explanation,omitempty"`
	ExplanationMeta  ExplanationMetadata `json:"explanation_meta,omitempty"`
}
