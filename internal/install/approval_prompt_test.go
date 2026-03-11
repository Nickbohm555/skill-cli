package install

import (
	"errors"
	"io"
	"testing"
	"time"
)

func TestApprovalInteractiveApprove(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 11, 23, 0, 0, 0, time.UTC)
	var seenPrompt ApprovalPrompt

	collector := NewApprovalCollector(ApprovalPrompterFunc(func(prompt ApprovalPrompt) (bool, error) {
		seenPrompt = prompt
		return true, nil
	}))
	collector.Now = func() time.Time { return now }

	decision, err := collector.Collect(ApprovalPolicy{
		Interactive: true,
		Prompt: ApprovalPrompt{
			Title:   "Approve install to skills/go-docs-skill?",
			Message: "Preview is ready for review.",
		},
	})
	if err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}
	if !decision.Approved {
		t.Fatal("Collect().Approved = false, want true")
	}
	if decision.ApprovalSource != ApprovalSourceInteractiveConfirm {
		t.Fatalf("Collect().ApprovalSource = %q, want %q", decision.ApprovalSource, ApprovalSourceInteractiveConfirm)
	}
	if decision.DecisionAt == nil || !decision.DecisionAt.Equal(now) {
		t.Fatalf("Collect().DecisionAt = %v, want %v", decision.DecisionAt, now)
	}
	if !decision.IsExplicitApproval() {
		t.Fatal("Collect() did not produce an explicit approval")
	}
	if seenPrompt.Affirmative != defaultApprovalAffirmative {
		t.Fatalf("Collect() prompt affirmative = %q, want default %q", seenPrompt.Affirmative, defaultApprovalAffirmative)
	}
	if seenPrompt.Negative != defaultApprovalNegative {
		t.Fatalf("Collect() prompt negative = %q, want default %q", seenPrompt.Negative, defaultApprovalNegative)
	}
}

func TestApprovalInteractiveDecline(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 11, 23, 1, 0, 0, time.UTC)
	collector := NewApprovalCollector(ApprovalPrompterFunc(func(prompt ApprovalPrompt) (bool, error) {
		return false, nil
	}))
	collector.Now = func() time.Time { return now }

	decision, err := collector.Collect(ApprovalPolicy{Interactive: true})
	if !errors.Is(err, ErrInstallDeclined) {
		t.Fatalf("Collect() error = %v, want %v", err, ErrInstallDeclined)
	}
	if decision.Approved {
		t.Fatal("Collect().Approved = true, want false")
	}
	if decision.ApprovalSource != ApprovalSourceDeclined {
		t.Fatalf("Collect().ApprovalSource = %q, want %q", decision.ApprovalSource, ApprovalSourceDeclined)
	}
	if decision.DecisionAt == nil || !decision.DecisionAt.Equal(now) {
		t.Fatalf("Collect().DecisionAt = %v, want %v", decision.DecisionAt, now)
	}
}

func TestApprovalInteractiveInterruptedPrompt(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 11, 23, 2, 0, 0, time.UTC)
	collector := NewApprovalCollector(ApprovalPrompterFunc(func(prompt ApprovalPrompt) (bool, error) {
		return false, io.EOF
	}))
	collector.Now = func() time.Time { return now }

	decision, err := collector.Collect(ApprovalPolicy{Interactive: true})
	if !errors.Is(err, ErrInstallDeclined) {
		t.Fatalf("Collect() error = %v, want %v", err, ErrInstallDeclined)
	}
	if decision.Approved {
		t.Fatal("Collect().Approved = true, want false")
	}
	if decision.ApprovalSource != ApprovalSourceInterrupted {
		t.Fatalf("Collect().ApprovalSource = %q, want %q", decision.ApprovalSource, ApprovalSourceInterrupted)
	}
	if decision.DecisionAt == nil || !decision.DecisionAt.Equal(now) {
		t.Fatalf("Collect().DecisionAt = %v, want %v", decision.DecisionAt, now)
	}
}

func TestApprovalNonInteractiveRequiresExplicitFlag(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 11, 23, 3, 0, 0, time.UTC)
	collector := NewApprovalCollector(nil)
	collector.Now = func() time.Time { return now }

	decision, err := collector.Collect(ApprovalPolicy{Interactive: false})
	if !errors.Is(err, ErrInstallApprovalRequiredNonInteractive) {
		t.Fatalf("Collect() error = %v, want %v", err, ErrInstallApprovalRequiredNonInteractive)
	}
	if decision.Approved {
		t.Fatal("Collect().Approved = true, want false")
	}
	if decision.ApprovalSource != ApprovalSourceNone {
		t.Fatalf("Collect().ApprovalSource = %q, want %q", decision.ApprovalSource, ApprovalSourceNone)
	}
	if decision.DecisionAt == nil || !decision.DecisionAt.Equal(now) {
		t.Fatalf("Collect().DecisionAt = %v, want %v", decision.DecisionAt, now)
	}
}

func TestApprovalNonInteractiveExplicitApproveAllowed(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 11, 23, 4, 0, 0, time.UTC)
	collector := NewApprovalCollector(nil)
	collector.Now = func() time.Time { return now }

	decision, err := collector.Collect(ApprovalPolicy{
		Interactive:            false,
		ExplicitApprovalByFlag: true,
	})
	if err != nil {
		t.Fatalf("Collect() error = %v, want nil", err)
	}
	if !decision.Approved {
		t.Fatal("Collect().Approved = false, want true")
	}
	if decision.ApprovalSource != ApprovalSourceNonInteractiveFlag {
		t.Fatalf("Collect().ApprovalSource = %q, want %q", decision.ApprovalSource, ApprovalSourceNonInteractiveFlag)
	}
	if decision.DecisionAt == nil || !decision.DecisionAt.Equal(now) {
		t.Fatalf("Collect().DecisionAt = %v, want %v", decision.DecisionAt, now)
	}
	if !decision.IsExplicitApproval() {
		t.Fatal("Collect() did not produce an explicit approval")
	}
}
