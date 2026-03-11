package generate

import (
	"errors"
	"fmt"

	"github.com/Nickbohm555/skill-cli/internal/validation"
)

var ErrUserCanceled = errors.New("user canceled remediation")

type ValidatorFunc func(candidate validation.CandidateSkill) validation.ValidationReport

type PromptFunc func(issue validation.ValidationIssue, prompt string) (string, error)

type ApplyEditFunc func(candidate validation.CandidateSkill, issue validation.ValidationIssue, answer string) (validation.CandidateSkill, error)

type FixLoop struct {
	Validate ValidatorFunc
	Prompt   PromptFunc
	Apply    ApplyEditFunc
}

type FixLoopResult struct {
	Candidate        validation.CandidateSkill
	Report           validation.ValidationReport
	Iterations       int
	PromptedRuleIDs  []string
	ValidationPasses int
}

func NewFixLoop(prompt PromptFunc, apply ApplyEditFunc) (FixLoop, error) {
	if prompt == nil {
		return FixLoop{}, fmt.Errorf("fix loop prompt handler is required")
	}
	if apply == nil {
		return FixLoop{}, fmt.Errorf("fix loop edit handler is required")
	}

	return FixLoop{
		Validate: ValidateCandidate,
		Prompt:   prompt,
		Apply:    apply,
	}, nil
}

func ValidateCandidate(candidate validation.CandidateSkill) validation.ValidationReport {
	report := validation.NewReport()
	structural := validation.ValidateStructural(candidate)
	report.AddIssues(structural.Issues...)
	semantic := validation.ValidateSemantic(candidate)
	report.AddIssues(semantic.Issues...)
	return report
}

func (l FixLoop) Run(candidate validation.CandidateSkill) (FixLoopResult, error) {
	if l.Validate == nil {
		return FixLoopResult{}, fmt.Errorf("fix loop validator is required")
	}
	if l.Prompt == nil {
		return FixLoopResult{}, fmt.Errorf("fix loop prompt handler is required")
	}
	if l.Apply == nil {
		return FixLoopResult{}, fmt.Errorf("fix loop edit handler is required")
	}

	result := FixLoopResult{
		Candidate:       candidate,
		PromptedRuleIDs: make([]string, 0),
	}

	for {
		report := l.Validate(result.Candidate)
		result.Report = report
		result.ValidationPasses++

		decision := CanProceed(report)
		if decision.Allowed {
			return result, nil
		}
		if decision.BlockingIssue == nil {
			return result, fmt.Errorf("progression gate blocked without blocking issue")
		}
		issue := *decision.BlockingIssue

		prompt := validation.PromptForRule(issue.RuleID)
		answer, err := l.Prompt(issue, prompt)
		if err != nil {
			if errors.Is(err, ErrUserCanceled) {
				return result, ErrUserCanceled
			}
			return result, err
		}

		updated, err := l.Apply(result.Candidate, issue, answer)
		if err != nil {
			if errors.Is(err, ErrUserCanceled) {
				return result, ErrUserCanceled
			}
			return result, err
		}

		result.Candidate = updated
		result.Iterations++
		result.PromptedRuleIDs = append(result.PromptedRuleIDs, issue.RuleID)
	}
}
