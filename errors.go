package paws

import "fmt"

var (
	ErrUnknownFlag  = fmt.Errorf("paws: unknown flag")
	ErrFlagValue    = fmt.Errorf("paws: invalid flag value")
	ErrMissingValue = fmt.Errorf("paws: flag requires value")
	ErrRequiredFlag = fmt.Errorf("paws: required flag missing")
	ErrParse        = fmt.Errorf("paws: parse error")
)

type ParseError struct {
	Err   error  // Error type (ErrUnknownFlag, etc)
	Flag  string // Flag name involved
	Value string // Flag value if any
	Cause error  // Underlying error
}

func (e *ParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Err.Error(), e.Flag, e.Cause.Error())
	}
	if e.Value != "" {
		return fmt.Sprintf("%s: %s with value %s", e.Err.Error(), e.Flag, e.Value)
	}
	return fmt.Sprintf("%s: %s", e.Err.Error(), e.Flag)
}

func (e *ParseError) Unwrap() error { return e.Err }

func errorUnknownFlag(flag string) *ParseError {
	return &ParseError{Err: ErrUnknownFlag, Flag: flag}
}
func errorMissingValue(flag string) *ParseError {
	return &ParseError{Err: ErrMissingValue, Flag: flag}
}
func errorRequiredFlag(flag string) *ParseError {
	return &ParseError{Err: ErrRequiredFlag, Flag: flag}
}
