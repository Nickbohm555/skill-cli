package install

import "errors"

type ErrorCode string

const (
	ErrorBlockedValidation              ErrorCode = "install_blocked_validation"
	ErrorBlockedConflict                ErrorCode = "install_blocked_conflict"
	ErrorApprovalDeclined               ErrorCode = "install_approval_declined"
	ErrorApprovalRequiredNonInteractive ErrorCode = "install_non_interactive_approval_required"
)

type InstallError struct {
	Code    ErrorCode
	Message string
}

func (e *InstallError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *InstallError) Is(target error) bool {
	other, ok := target.(*InstallError)
	if !ok {
		return false
	}
	return e.Code == other.Code
}

var (
	ErrInstallBlockedValidation = &InstallError{
		Code:    ErrorBlockedValidation,
		Message: "install blocked: validation still has blocking issues",
	}
	ErrInstallBlockedConflict = &InstallError{
		Code:    ErrorBlockedConflict,
		Message: "install blocked: conflict resolution is unresolved or aborting",
	}
	ErrInstallDeclined = &InstallError{
		Code:    ErrorApprovalDeclined,
		Message: "install blocked: approval was declined",
	}
	ErrInstallApprovalRequiredNonInteractive = &InstallError{
		Code:    ErrorApprovalRequiredNonInteractive,
		Message: "install blocked: non-interactive approval flag is required",
	}
)

func ErrorCodeOf(err error) ErrorCode {
	var installErr *InstallError
	if !errors.As(err, &installErr) {
		return ""
	}
	return installErr.Code
}

func IsBlockedValidation(err error) bool {
	return errors.Is(err, ErrInstallBlockedValidation)
}

func IsBlockedConflict(err error) bool {
	return errors.Is(err, ErrInstallBlockedConflict)
}

func IsApprovalDeclined(err error) bool {
	return errors.Is(err, ErrInstallDeclined)
}

func IsApprovalRequiredNonInteractive(err error) bool {
	return errors.Is(err, ErrInstallApprovalRequiredNonInteractive)
}
