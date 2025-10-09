package apawse

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type CommandDef struct {
	Path  []string
	Flags []*Flag
}

type ParseResult struct {
	Command    *CommandDef
	Positional []string
	Flags      map[string]string
	GlobalFlag map[string]*Flag
	DoubleDash bool
	RawArgs    []string
}

type Parser struct {
	Commands []*CommandDef
	Flags    []*Flag
}

func New() *Parser {
	return &Parser{
		Commands: []*CommandDef{},
		Flags:    []*Flag{},
	}
}

func (p *Parser) AddCommand(path []string, flags []*Flag) {
	p.Commands = append(p.Commands, &CommandDef{
		Path:  path,
		Flags: flags,
	})
}

func (p *Parser) AddFlags(flags ...*Flag) {
	p.Flags = append(p.Flags, flags...)
}

func (p *Parser) Parse(args []string) (*ParseResult, error) {
	result := &ParseResult{
		Flags:      make(map[string]string),
		GlobalFlag: make(map[string]*Flag),
		RawArgs:    args,
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

// parseFlag - consolidated flag parsing
func (p *Parser) parseFlag(arg string, args []string, i int, cmd *CommandDef, result *ParseResult) (int, error) {
	if strings.HasPrefix(arg, "--") {
		return p.parseLongFlag(arg, args, i, cmd, result)
	}
	return p.parseShortFlag(arg, args, i, cmd, result)
}

// parseLongFlag - handle --flag and --flag=value
func (p *Parser) parseLongFlag(arg string, args []string, i int, cmd *CommandDef, result *ParseResult) (int, error) {
	flagName := strings.TrimPrefix(arg, "--")

	// Handle --flag=value format
	if strings.Contains(flagName, "=") {
		parts := strings.SplitN(flagName, "=", 2)
		flagDef := p.findFlag(parts[0], cmd)

		if flagDef == nil {
			return 0, errorUnknownFlag(parts[0])
		}

		result.Flags[flagDef.Name] = parts[1]
		return 1, nil
	}

	flagDef := p.findFlag(flagName, cmd)
	if flagDef == nil {
		return 0, errorUnknownFlag(flagName)
	}

	if flagDef.Type == BoolType {
		result.Flags[flagDef.Name] = "true"
		return 1, nil
	}

	// Non-boolean flag requires value
	if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
		return 0, errorMissingValue(flagName)
	}

	result.Flags[flagDef.Name] = args[i+1]
	return 2, nil
}

// parseShortFlag - handle -f and -f value
func (p *Parser) parseShortFlag(arg string, args []string, i int, cmd *CommandDef, result *ParseResult) (int, error) {
	flagChars := strings.TrimPrefix(arg, "-")

	// Handle single flag: -f
	if len(flagChars) == 1 {
		flagDef := p.findFlag(flagChars, cmd)

		if flagDef == nil {
			return 0, errorUnknownFlag(flagChars)
		}

		if flagDef.Type == BoolType {
			result.Flags[flagDef.Name] = "true"
			return 1, nil
		}

		// Non-boolean flag requires value
		if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
			return 0, errorMissingValue(flagChars)
		}

		result.Flags[flagDef.Name] = args[i+1]
		return 2, nil
	}

	// Handle grouped flags: -abc (only boolean flags allowed)
	for _, char := range flagChars {
		flagStr := string(char)
		flagDef := p.findFlag(flagStr, cmd)

		if flagDef == nil {
			return 0, errorUnknownFlag(flagStr)
		}

		if flagDef.Type != BoolType {
			return 0, fmt.Errorf("non-boolean flag -%s cannot be grouped", flagStr)
		}

		result.Flags[flagDef.Name] = "true"
	}

	return 1, nil
}

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

// ValidateRequired checks if all required flags are provided.
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

// if res.Flag["paw"] == "true" ? that stupid.

// Bool returns boolean value for flag, using default value if not provided
func (r *ParseResult) Bool(n string) bool {
	if val, exists := r.Flags[n]; exists {
		return val == "true"
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
	for _, flag := range r.GlobalFlag {
		if flag.Name == name {
			return flag
		}
	}

	// Search in command flags
	if r.Command != nil {
		for _, flag := range r.Command.Flags {
			if flag.Name == name {
				return flag
			}
		}
	}

	return nil
}
