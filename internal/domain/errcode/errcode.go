package errcode

import (
	"errors"
	"fmt"
)

// Code classifies a domain or application error across all driving surfaces.
type Code string

const (
	CodeValidation Code = "VALIDATION_ERROR"
	CodeAuth       Code = "AUTH_ERROR"
	CodeNotFound   Code = "NOT_FOUND"
	CodeProvision  Code = "PROVISION_ERROR"
	CodeExternal   Code = "EXTERNAL_ERROR"
)

// Error represents a structured error carrying a domain classification code.
type Error struct {
	Code    Code
	Message string
	Cause   error
}

// Error returns the formatted error message.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for errors.Is and errors.As traversal.
func (e *Error) Unwrap() error {
	return e.Cause
}

// New constructs a domain Error without an underlying cause.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap constructs a domain Error wrapping an underlying cause.
func Wrap(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: message, Cause: cause}
}

// IsCode reports whether err or any error in its chain has the given Code.
func IsCode(err error, code Code) bool {
	var target *Error
	if errors.As(err, &target) {
		return target.Code == code
	}
	return false
}

// CodeOf extracts the Code of err, or returns CodeExternal if unclassified.
func CodeOf(err error) Code {
	var target *Error
	if errors.As(err, &target) {
		return target.Code
	}
	return CodeExternal
}
