package errors

import (
	"fmt"
	"runtime/debug"
)

type ErrorType string

const (
	TypeBadRequest ErrorType = "BAD_REQUEST"
	TypeNotFound   ErrorType = "NOT_FOUND"
	TypeConflict   ErrorType = "CONFLICT"
	TypeInternal   ErrorType = "INTERNAL"
)

const (
	CodeTeamExists    = "TEAM_EXISTS"
	CodePRExists      = "PR_EXISTS"
	CodePRMerged      = "PR_MERGED"
	CodeNotAssigned   = "NOT_ASSIGNED"
	CodeNoCandidate   = "NO_CANDIDATE"
	CodeNotFound      = "NOT_FOUND"
	CodeBadRequest    = "BAD_REQUEST"
	CodeInternalError = "INTERNAL_ERROR"
)

type AppError struct {
	Type       ErrorType
	Code       string
	Message    string
	Cause      error
	Stack      []byte
	HTTPStatus int
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func NewTeamExists(message string) *AppError {
	return &AppError{
		Type:       TypeConflict,
		Code:       CodeTeamExists,
		Message:    message,
		HTTPStatus: 400,
		Stack:      debug.Stack(),
	}
}

func NewPRExists(message string) *AppError {
	return &AppError{
		Type:       TypeConflict,
		Code:       CodePRExists,
		Message:    message,
		HTTPStatus: 409,
		Stack:      debug.Stack(),
	}
}

func NewPRMerged(message string) *AppError {
	return &AppError{
		Type:       TypeBadRequest,
		Code:       CodePRMerged,
		Message:    message,
		HTTPStatus: 409,
		Stack:      debug.Stack(),
	}
}

func NewNotAssigned(message string) *AppError {
	return &AppError{
		Type:       TypeBadRequest,
		Code:       CodeNotAssigned,
		Message:    message,
		HTTPStatus: 409,
		Stack:      debug.Stack(),
	}
}

func NewNoCandidate(message string) *AppError {
	return &AppError{
		Type:       TypeBadRequest,
		Code:       CodeNoCandidate,
		Message:    message,
		HTTPStatus: 409,
		Stack:      debug.Stack(),
	}
}

func NewNotFound(message string) *AppError {
	return &AppError{
		Type:       TypeNotFound,
		Code:       CodeNotFound,
		Message:    message,
		HTTPStatus: 404,
		Stack:      debug.Stack(),
	}
}

func NewValidation(message string) *AppError {
	return &AppError{
		Type:       TypeBadRequest,
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: 400,
		Stack:      debug.Stack(),
	}
}

func WrapInternal(err error, message string) *AppError {
	return &AppError{
		Type:       TypeInternal,
		Code:       CodeInternalError,
		Message:    message,
		Cause:      err,
		HTTPStatus: 500,
		Stack:      debug.Stack(),
	}
}

var (
	ErrTeamExists     = NewTeamExists("team_name already exists")
	ErrPRExists       = NewPRExists("PR id already exists")
	ErrPRMerged       = NewPRMerged("cannot reassign on merged PR")
	ErrNotAssigned    = NewNotAssigned("reviewer is not assigned to this PR")
	ErrNoCandidate    = NewNoCandidate("no active replacement candidate in team")
	ErrTeamNotFound   = NewNotFound("team not found")
	ErrUserNotFound   = NewNotFound("user not found")
	ErrPRNotFound     = NewNotFound("PR not found")
	ErrAuthorNotFound = NewNotFound("author not found")
)
