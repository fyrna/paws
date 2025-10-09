package apawse

import "fmt"

var (
	ErrUnknownFlag  = fmt.Errorf("apawse: unknown flag")
	ErrFlagValue    = fmt.Errorf("apawse: invalid flag value")
	ErrMissingValue = fmt.Errorf("apawse: flag requires value")
	ErrRequiredFlag = fmt.Errorf("apawse: required flag missing")
	ErrParse        = fmt.Errorf("apawse: parse error")
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
