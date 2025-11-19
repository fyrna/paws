package paws

import (
	"errors"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name:     "Only error and flag",
			err:      &ParseError{Err: ErrUnknownFlag, Flag: "verbose"},
			expected: "paws: unknown flag: verbose",
		},
		{
			name:     "With value",
			err:      &ParseError{Err: ErrFlagValue, Flag: "count", Value: "abc"},
			expected: "paws: invalid flag value: count with value abc",
		},
		{
			name:     "With cause",
			err:      &ParseError{Err: ErrFlagValue, Flag: "port", Cause: errors.New("invalid format")},
			expected: "paws: invalid flag value: port (invalid format)",
		},
		{
			name:     "With value and cause - cause takes precedence",
			err:      &ParseError{Err: ErrFlagValue, Flag: "test", Value: "123", Cause: errors.New("underlying error")},
			expected: "paws: invalid flag value: test (underlying error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ParseError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	err := &ParseError{Err: ErrMissingValue, Flag: "file"}
	if got := err.Unwrap(); got != ErrMissingValue {
		t.Errorf("ParseError.Unwrap() = %v, want %v", got, ErrMissingValue)
	}
}

func TestErrorHelpers(t *testing.T) {
	tests := []struct {
		name     string
		helper   func(string) *ParseError
		flag     string
		expected *ParseError
	}{
		{
			name:   "errorUnknownFlag",
			helper: errorUnknownFlag,
			flag:   "unknown",
			expected: &ParseError{
				Err:  ErrUnknownFlag,
				Flag: "unknown",
			},
		},
		{
			name:   "errorMissingValue",
			helper: errorMissingValue,
			flag:   "required",
			expected: &ParseError{
				Err:  ErrMissingValue,
				Flag: "required",
			},
		},
		{
			name:   "errorRequiredFlag",
			helper: errorRequiredFlag,
			flag:   "missing",
			expected: &ParseError{
				Err:  ErrRequiredFlag,
				Flag: "missing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.helper(tt.flag)
			if result.Err != tt.expected.Err {
				t.Errorf("%s() error = %v, want %v", tt.name, result.Err, tt.expected.Err)
			}
			if result.Flag != tt.expected.Flag {
				t.Errorf("%s() flag = %v, want %v", tt.name, result.Flag, tt.expected.Flag)
			}
		})
	}
}
