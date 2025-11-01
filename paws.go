package paws

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// CommandDef represents a command definition with its path and flags
type CommandDef struct {
	Path  []string // Command path (e.g., ["git", "commit"])
	Flags []*Flag  // Command-specific flags
}

// ParseResult contains the result of parsing command line arguments
type ParseResult struct {
	Command    *CommandDef       // Matched command (if any)
	Positional []string          // Positional arguments
	Flags      map[string]string // Parsed flag values
	GlobalFlag map[string]*Flag  // Global flag definitions
	DoubleDash bool              // Whether -- was encountered
	RawArgs    []string          // Original arguments
}

// Parser is the main argument parser
type Parser struct {
	Commands []*CommandDef // Registered commands
	Flags    []*Flag       // Global flags
}

// New creates a new argument parser
func New() *Parser {
	return &Parser{
		Commands: []*CommandDef{},
		Flags:    []*Flag{},
	}
}

// AddCommand registers a new command with the parser
func (p *Parser) AddCommand(path []string, flags []*Flag) {
	p.Commands = append(p.Commands, &CommandDef{
		Path:  path,
		Flags: flags,
	})
}

// AddFlags registers global flags with the parser
func (p *Parser) AddFlags(flags ...*Flag) {
	p.Flags = append(p.Flags, flags...)
}

// Parse parses command line arguments and returns a ParseResult
func (p *Parser) Parse(args []string) (*ParseResult, error) {
	result := &ParseResult{
		Flags:      make(map[string]string),
		GlobalFlag: make(map[string]*Flag),
		RawArgs:    args,
	}

	// Build global flag lookup table
	for _, flag := range p.Flags {
		result.GlobalFlag[flag.Name] = flag
		for _, alias := range flag.Aliases {
			result.GlobalFlag[alias] = flag
		}
	}

	var (
		cmd        *CommandDef
		positional []string
		i          = 0
	)

	// Step 1: Find command
	cmd, consumed := p.findCommand(args)
	if cmd != nil {
		result.Command = cmd
		i = consumed
	}

	// Step 2: Parse flags and collect positional args
	inPositional := false
	for i < len(args) {
		arg := args[i]

		if arg == "--" {
			inPositional = true
			i++
			result.DoubleDash = true
			continue
		}

		if !inPositional && strings.HasPrefix(arg, "-") {
			consumed, err := p.parseFlag(arg, args, i, cmd, result)
			if err != nil {
				return nil, err
			}
			i += consumed
			continue
		}

		// Positional argument
		positional = append(positional, arg)
		i++
	}

	result.Positional = positional
	return result, nil
}

// findCommand searches for the best matching command in the arguments
func (p *Parser) findCommand(args []string) (*CommandDef, int) {
	var bestMatch *CommandDef
	bestLength := 0

	for _, cmd := range p.Commands {
		if len(cmd.Path) <= len(args) {
			match := true
			for j, part := range cmd.Path {
				if j >= len(args) || args[j] != part {
					match = false
					break
				}
			}
			if match && len(cmd.Path) > bestLength {
				bestMatch = cmd
				bestLength = len(cmd.Path)
			}
		}
	}

	return bestMatch, bestLength
}

// parseFlag handles both long and short flag parsing
func (p *Parser) parseFlag(arg string, args []string, i int, cmd *CommandDef, result *ParseResult) (int, error) {
	long := strings.HasPrefix(arg, "--")
	nameStart := 2

	if !long {
		nameStart = 1
	}

	s := arg[nameStart:]

	// Handle grouped short flags like -abc
	if !long && len(s) > 1 {
		for _, c := range s {
			f := p.findFlag(string(c), cmd)
			if f == nil {
				return 0, errorUnknownFlag(string(c))
			}
			if f.Type != BoolType {
				return 0, fmt.Errorf("non-boolean flag -%c cannot be grouped", c)
			}
			result.Flags[f.Name] = "true"
		}
		return 1, nil
	}

	// Handle --flag=value
	var value string
	eq := strings.IndexByte(s, '=')
	if eq != -1 {
		value = s[eq+1:]
		s = s[:eq]
	}

	f := p.findFlag(s, cmd)
	if f == nil {
		return 0, errorUnknownFlag(s)
	}

	// Determine value if not via '='
	if value == "" {
		// If boolean, allow --flag or -f
		if f.Type == BoolType {
			// check if next token is value (e.g., --cute yes)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				v := args[i+1]
				if err := p.validateFlagValue(f, v); err != nil {
					return 0, &ParseError{Err: ErrFlagValue, Flag: f.Name, Value: v, Cause: err}
				}
				result.Flags[f.Name] = v
				return 2, nil
			}
			result.Flags[f.Name] = "true"
			return 1, nil
		}

		// Non-bool must have explicit value
		if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
			return 0, errorMissingValue(s)
		}
		value = args[i+1]
		consumed := 2
		if err := p.validateFlagValue(f, value); err != nil {
			return 0, &ParseError{Err: ErrFlagValue, Flag: f.Name, Value: value, Cause: err}
		}
		result.Flags[f.Name] = value
		return consumed, nil
	}

	// Validate --flag=value
	if err := p.validateFlagValue(f, value); err != nil {
		return 0, &ParseError{Err: ErrFlagValue, Flag: f.Name, Value: value, Cause: err}
	}
	result.Flags[f.Name] = value
	return 1, nil
}

// findFlag searches for a flag definition by name
func (p *Parser) findFlag(name string, cmd *CommandDef) *Flag {
	// Search in global flags first
	for _, flag := range p.Flags {
		if flag.Name == name || slices.Contains(flag.Aliases, name) {
			return flag
		}
	}

	// Search in command flags
	if cmd != nil {
		for _, flag := range cmd.Flags {
			if flag.Name == name || slices.Contains(flag.Aliases, name) {
				return flag
			}
		}
	}

	return nil
}

// ValidateRequired checks if all required flags are provided
func (p *Parser) ValidateRequired(result *ParseResult) error {
	allFlags := p.Flags

	if result.Command != nil {
		allFlags = append(allFlags, result.Command.Flags...)
	}

	for _, flag := range allFlags {
		if flag.IsRequired && flag.Type != BoolType {
			value, exists := result.Flags[flag.Name]
			if !exists || value == "" {
				return errorRequiredFlag(flag.Name)
			}
		}
	}
	return nil
}

// validateFlagValue validates flag value based on its constraints
func (p *Parser) validateFlagValue(flag *Flag, value string) error {
	switch flag.Type {
	case StringType:
		if len(flag.ChoicesOpt) > 0 && !slices.Contains(flag.ChoicesOpt, value) {
			return fmt.Errorf("value '%s' not in allowed choices: %v", value, flag.ChoicesOpt)
		}

	case IntType:
		val, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: '%s'", value)
		}
		if flag.Min != 0 || flag.Max != 0 {
			if val < flag.Min || val > flag.Max {
				return fmt.Errorf("value %d out of range [%d, %d]", val, flag.Min, flag.Max)
			}
		}

	case UintType:
		val, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value: '%s'", value)
		}
		if flag.Min != 0 || flag.Max != 0 {
			// Ensure min is not negative for uint
			m := max(0, flag.Min)
			if val < uint64(m) || val > uint64(flag.Max) {
				return fmt.Errorf("value %d out of range [%d, %d]", val, m, flag.Max)
			}
		}

	case FloatType:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: '%s'", value)
		}
		if flag.Min != 0 || flag.Max != 0 {
			if val < float64(flag.Min) || val > float64(flag.Max) {
				return fmt.Errorf("value %f out of range [%f, %f]", val, float64(flag.Min), float64(flag.Max))
			}
		}

	case BoolType:
		// Boolean flags accept various truthy/falsy values
		if !isValidBoolValue(value) {
			return fmt.Errorf("invalid boolean value: '%s' (allowed: true/t/yes/y/false/f/no/n)", value)
		}
	}

	return nil
}

// Bool returns boolean value for flag, using default value if not provided
func (r *ParseResult) Bool(n string) bool {
	if val, exists := r.Flags[n]; exists {
		return parseBoolValue(val)
	}

	// Look for default value from flag definition
	flag := r.findFlag(n)
	if flag != nil && flag.DefValue != nil {
		if b, ok := flag.DefValue.(bool); ok {
			return b
		}
	}
	return false
}

// String returns string value for flag, using default value if not provided
func (r *ParseResult) String(n string) string {
	if val, exists := r.Flags[n]; exists {
		return val
	}

	flag := r.findFlag(n)
	if flag != nil && flag.DefValue != nil {
		if s, ok := flag.DefValue.(string); ok {
			return s
		}
	}
	return ""
}

// Int returns integer value for flag, using default value if not provided
func (r *ParseResult) Int(n string) int {
	if val, exists := r.Flags[n]; exists {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}

	flag := r.findFlag(n)
	if flag != nil && flag.DefValue != nil {
		if i, ok := flag.DefValue.(int); ok {
			return i
		}
	}
	return 0
}

// Uint returns unsigned integer value for flag, using default value if not provided
func (r *ParseResult) Uint(n string) uint {
	if val, exists := r.Flags[n]; exists {
		if u, err := strconv.ParseUint(val, 10, 64); err == nil {
			return uint(u)
		}
	}

	flag := r.findFlag(n)
	if flag != nil && flag.DefValue != nil {
		if u, ok := flag.DefValue.(uint); ok {
			return u
		}
		if i, ok := flag.DefValue.(int); ok && i >= 0 {
			return uint(i)
		}
	}
	return 0
}

// Float returns float64 value for flag, using default value if not provided
func (r *ParseResult) Float(n string) float64 {
	if val, exists := r.Flags[n]; exists {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}

	flag := r.findFlag(n)
	if flag != nil && flag.DefValue != nil {
		if f, ok := flag.DefValue.(float64); ok {
			return f
		}
		if i, ok := flag.DefValue.(int); ok {
			return float64(i)
		}
	}
	return 0.0
}

// findFlag searches for flag definition in both global and command flags
func (r *ParseResult) findFlag(name string) *Flag {
	// Search in global flags
	if flag, exists := r.GlobalFlag[name]; exists {
		return flag
	}

	// Search in command flags
	if r.Command != nil {
		for _, flag := range r.Command.Flags {
			if flag.Name == name || slices.Contains(flag.Aliases, name) {
				return flag
			}
		}
	}

	return nil
}

// isValidBoolValue checks if a string represents a valid boolean value
func isValidBoolValue(v string) bool {
	if len(v) == 0 {
		return false
	}

	if len(v) == 1 {
		c := v[0]
		return c == 't' || c == 'f' ||
			c == 'T' || c == 'F' ||
			c == 'y' || c == 'n' ||
			c == 'Y' || c == 'N' ||
			c == '1' || c == '0'
	}

	switch len(v) {
	case 2: // "no", "NO", "No", "nO"
		return (v[0] == 'n' || v[0] == 'N') &&
			(v[1] == 'o' || v[1] == 'O')

	case 3: // "yes", "YES", "Yes", etc
		if (v[0] == 'y' || v[0] == 'Y') &&
			(v[1] == 'e' || v[1] == 'E') &&
			(v[2] == 's' || v[2] == 'S') {
			return true
		}

	case 4: // "true", "TRUE", "True", etc
		if (v[0] == 't' || v[0] == 'T') &&
			(v[1] == 'r' || v[1] == 'R') &&
			(v[2] == 'u' || v[2] == 'U') &&
			(v[3] == 'e' || v[3] == 'E') {
			return true
		}

	case 5: // "false", "FALSE", "False", etc
		if (v[0] == 'f' || v[0] == 'F') &&
			(v[1] == 'a' || v[1] == 'A') &&
			(v[2] == 'l' || v[2] == 'L') &&
			(v[3] == 's' || v[3] == 'S') &&
			(v[4] == 'e' || v[4] == 'E') {
			return true
		}
	}

	return false
}

// parseBoolValue converts various string representations to boolean
func parseBoolValue(value string) bool {
	if len(value) == 0 {
		return false
	}

	if len(value) == 1 {
		switch value[0] {
		case 't', 'T', 'y', 'Y', '1':
			return true
		default: // 'f', 'F', 'n', 'N', '0'
			return false
		}
	}

	switch len(value) {
	case 2: // "no" - false
		if (value[0] == 'n' || value[0] == 'N') &&
			(value[1] == 'o' || value[1] == 'O') {
			return false
		}

	case 3: // "yes" - true
		if (value[0] == 'y' || value[0] == 'Y') &&
			(value[1] == 'e' || value[1] == 'E') &&
			(value[2] == 's' || value[2] == 'S') {
			return true
		}

	case 4: // "true" - true
		if (value[0] == 't' || value[0] == 'T') &&
			(value[1] == 'r' || value[1] == 'R') &&
			(value[2] == 'u' || value[2] == 'U') &&
			(value[3] == 'e' || value[3] == 'E') {
			return true
		}

	case 5: // "false" - false
		if (value[0] == 'f' || value[0] == 'F') &&
			(value[1] == 'a' || value[1] == 'A') &&
			(value[2] == 'l' || value[2] == 'L') &&
			(value[3] == 's' || value[3] == 'S') &&
			(value[4] == 'e' || value[4] == 'E') {
			return false
		}
	}

	return false
}
