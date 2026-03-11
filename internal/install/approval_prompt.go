package install

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/huh/v2"
)

const (
	defaultApprovalTitle       = "Approve install?"
	defaultApprovalAffirmative = "Approve install"
	defaultApprovalNegative    = "Cancel"
)

type ApprovalPrompt struct {
	Title       string
	Message     string
	Affirmative string
	Negative    string
}

type ApprovalPolicy struct {
	Interactive            bool
	ExplicitApprovalByFlag bool
	Prompt                 ApprovalPrompt
}

type ApprovalPrompter interface {
	PromptApproval(prompt ApprovalPrompt) (bool, error)
}

type ApprovalPrompterFunc func(prompt ApprovalPrompt) (bool, error)

func (f ApprovalPrompterFunc) PromptApproval(prompt ApprovalPrompt) (bool, error) {
	return f(prompt)
}

type HuhApprovalPrompter struct{}

func (HuhApprovalPrompter) PromptApproval(prompt ApprovalPrompt) (bool, error) {
	approved := false

	err := huh.NewConfirm().
		Title(prompt.Title).
		Description(prompt.Message).
		Affirmative(prompt.Affirmative).
		Negative(prompt.Negative).
		Value(&approved).
		Run()
	if err != nil {
		return false, err
	}

	return approved, nil
}

type ApprovalCollector struct {
	Prompter ApprovalPrompter
	Now      func() time.Time
}

func NewApprovalCollector(prompter ApprovalPrompter) ApprovalCollector {
	return ApprovalCollector{
		Prompter: prompter,
		Now:      time.Now,
	}
}

func (c ApprovalCollector) Collect(policy ApprovalPolicy) (ApprovalDecision, error) {
	if c.Now == nil {
		c.Now = time.Now
	}

	prompt := normalizedApprovalPrompt(policy.Prompt)

	if !policy.Interactive {
		if !policy.ExplicitApprovalByFlag {
			decisionAt := c.Now()
			return ApprovalDecision{
				Approved:       false,
				ApprovalSource: ApprovalSourceNone,
				DecisionAt:     &decisionAt,
				Explanation:    "Install is non-interactive and no explicit approval flag was provided.",
			}, ErrInstallApprovalRequiredNonInteractive
		}

		decisionAt := c.Now()
		return ApprovalDecision{
			Approved:       true,
			ApprovalSource: ApprovalSourceNonInteractiveFlag,
			DecisionAt:     &decisionAt,
			Explanation:    "Install was explicitly approved by non-interactive flag.",
		}, nil
	}

	if c.Prompter == nil {
		return deniedApprovalDecision(c.Now(), interruptedApprovalExplanation(nil)), ErrInstallDeclined
	}

	approved, err := c.Prompter.PromptApproval(prompt)
	if err != nil {
		return deniedApprovalDecision(c.Now(), interruptedApprovalExplanation(err)), ErrInstallDeclined
	}

	decisionAt := c.Now()
	if !approved {
		return ApprovalDecision{
			Approved:       false,
			ApprovalSource: ApprovalSourceDeclined,
			DecisionAt:     &decisionAt,
			Explanation:    "User explicitly declined install approval.",
		}, ErrInstallDeclined
	}

	return ApprovalDecision{
		Approved:       true,
		ApprovalSource: ApprovalSourceInteractiveConfirm,
		DecisionAt:     &decisionAt,
		Explanation:    interactiveApprovalExplanation(prompt),
	}, nil
}

func normalizedApprovalPrompt(prompt ApprovalPrompt) ApprovalPrompt {
	if strings.TrimSpace(prompt.Title) == "" {
		prompt.Title = defaultApprovalTitle
	}
	if strings.TrimSpace(prompt.Affirmative) == "" {
		prompt.Affirmative = defaultApprovalAffirmative
	}
	if strings.TrimSpace(prompt.Negative) == "" {
		prompt.Negative = defaultApprovalNegative
	}
	return prompt
}

func deniedApprovalDecision(decisionAt time.Time, explanation string) ApprovalDecision {
	return ApprovalDecision{
		Approved:       false,
		ApprovalSource: ApprovalSourceInterrupted,
		DecisionAt:     &decisionAt,
		Explanation:    explanation,
	}
}

func interactiveApprovalExplanation(prompt ApprovalPrompt) string {
	if strings.TrimSpace(prompt.Message) == "" {
		return "User explicitly approved install in interactive mode."
	}
	return fmt.Sprintf("User explicitly approved install in interactive mode after reviewing: %s", strings.TrimSpace(prompt.Message))
}

func interruptedApprovalExplanation(err error) string {
	if err == nil {
		return "Approval prompt could not run, so install approval defaulted to deny."
	}
	if err == io.EOF {
		return "Approval prompt ended with EOF, so install approval defaulted to deny."
	}
	return fmt.Sprintf("Approval prompt failed (%v), so install approval defaulted to deny.", err)
}
