package paws

import (
	"errors"
	"fmt"
)

// Common error types for argument parsing
var (
	ErrUnknownFlag  = errors.New("unknown flag")
	ErrFlagValue    = errors.New("invalid flag value")
	ErrMissingValue = errors.New("flag requires value")
	ErrRequiredFlag = errors.New("required flag missing")
	ErrParse        = errors.New("parse error")
)

// ParseError represents a parsing error with context
type ParseError struct {
	Err   error  // Error type (ErrUnknownFlag, etc)
	Flag  string // Flag name involved
	Value string // Flag value if any
	Cause error  // Underlying error
}

// Error returns a formatted error message
func (e *ParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Err.Error(), e.Flag, e.Cause.Error())
	}
	if e.Value != "" {
		return fmt.Sprintf("%s: %s with value %s", e.Err.Error(), e.Flag, e.Value)
	}
	return fmt.Sprintf("%s: %s", e.Err.Error(), e.Flag)
}

// Unwrap returns the underlying error
func (e *ParseError) Unwrap() error { return e.Err }

// helpers to build typed ParseError
func errorUnknownFlag(flag string) *ParseError {
	return &ParseError{Err: ErrUnknownFlag, Flag: flag}
}

func errorMissingValue(flag string) *ParseError {
	return &ParseError{Err: ErrMissingValue, Flag: flag}
}

func errorRequiredFlag(flag string) *ParseError {
	return &ParseError{Err: ErrRequiredFlag, Flag: flag}
}
