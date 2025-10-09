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
	DoubleDash bool
	RawArgs    []string
	flag       []*Flag // for global flag getter
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

// Parse parses command line arguments and returns the result.
func (p *Parser) Parse(args []string) (*ParseResult, error) {
	result := &ParseResult{
		Flags:   make(map[string]string),
		RawArgs: args,
		flag:    p.Flags,
	}

	cmd, consumed, doubleDash, err := p.parseArgs(args, result)
	if err != nil {
		return nil, err
	}

	result.Command = cmd
	result.DoubleDash = doubleDash

	// Capture positional arguments
	if doubleDash {
		if consumed < len(args) && args[consumed] == "--" {
			consumed++
		}
		result.Positional = args[consumed:]
	} else {
		result.Positional = p.extractPositional(args, consumed)
	}

	return result, nil
}

// parseArgs handles the main argument parsing logic.
func (p *Parser) parseArgs(args []string, result *ParseResult) (*CommandDef, int, bool, error) {
	var (
		cmd          *CommandDef
		inPositional = false
		doubleDash   = false
		consumed     = 0
	)

	// Phase 1: Find command first
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			doubleDash = true
			if cmd == nil {
				return nil, i + 1, true, nil
			}
			break
		}

		if strings.HasPrefix(arg, "-") {
			// Skip flags during phase 1
			continue
		}

		// Find matching command
		if cmd == nil {
			found, cmdConsumed := p.findCommand(args[i:])
			if found != nil {
				cmd = found
				i += cmdConsumed - 1
				consumed = i + 1
			}
		}
	}

	// Reset for phase 2: Parse all args with command context
	consumed = 0
	inPositional = false
	doubleDash = false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			doubleDash = true
			inPositional = true
			consumed = i + 1
			continue
		}

		if inPositional {
			continue
		}

		if strings.HasPrefix(arg, "-") {
			// Parse flag with command context
			flagConsumed, err := p.parseFlag(arg, args, i, cmd, result)
			if err != nil {
				return nil, 0, false, err
			}
			i += flagConsumed
			consumed = i + 1
			continue
		}

		// Skip command parts that were already processed
		if cmd != nil && i < len(cmd.Path) {
			consumed = i + 1
			continue
		}

		// Positional argument - we'll capture these later
		consumed = i + 1
	}

	return cmd, consumed, doubleDash, nil
}

// extractPositional extracts positional arguments after command and flags.
func (p *Parser) extractPositional(args []string, consumed int) []string {
	if consumed >= len(args) {
		return []string{}
	}

	var positional []string
	inFlags := false

	for i := consumed; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			inFlags = false
			continue
		}

		if strings.HasPrefix(arg, "-") && !inFlags {
			// Skip flags
			continue
		}

		positional = append(positional, arg)
	}

	return positional
}

// findCommand searches for a command that matches the provided arguments.
func (p *Parser) findCommand(args []string) (*CommandDef, int) {
	var bestMatch *CommandDef
	var bestConsumed int

	for _, cmd := range p.Commands {
		if len(cmd.Path) <= len(args) {
			match := true
			for j, part := range cmd.Path {
				if args[j] != part {
					match = false
					break
				}
			}
			if match && len(cmd.Path) > bestConsumed {
				bestMatch = cmd
				bestConsumed = len(cmd.Path)
			}
		}
	}
	return bestMatch, bestConsumed
}

// parseFlag handles flag parsing for both long and short formats.
func (p *Parser) parseFlag(arg string, args []string, i int, cmd *CommandDef, result *ParseResult) (int, error) {
	if strings.HasPrefix(arg, "--") {
		return p.parseLongFlag(arg, args, i, cmd, result)
	}
	return p.parseShortFlag(arg, args, i, cmd, result)
}

// parseLongFlag handles long format flags (--flag).
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
		return 0, nil
	}

	flagDef := p.findFlag(flagName, cmd)
	if flagDef == nil {
		return 0, errorUnknownFlag(flagName)
	}

	if flagDef.Type == BoolType {
		result.Flags[flagDef.Name] = "true"
		return 0, nil
	}

	if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
		return 0, errorMissingValue(flagName)
	}

	result.Flags[flagDef.Name] = args[i+1]
	return 1, nil
}

// parseShortFlag handles short format flags (-f).
func (p *Parser) parseShortFlag(arg string, args []string, i int, cmd *CommandDef, result *ParseResult) (int, error) {
	flagChars := strings.TrimPrefix(arg, "-")
	consumed := 0

	for j, char := range flagChars {
		flagStr := string(char)
		flagDef := p.findFlag(flagStr, cmd)

		if flagDef == nil {
			return 0, errorUnknownFlag(flagStr)
		}

		if flagDef.Type == BoolType {
			result.Flags[flagDef.Name] = "true"
			continue
		}

		if j < len(flagChars)-1 {
			return 0, fmt.Errorf("flag -%s requires value and cannot be grouped", flagStr)
		}

		if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
			return 0, errorMissingValue(flagStr)
		}

		result.Flags[flagDef.Name] = args[i+1]
		consumed = 1
	}

	return consumed, nil
}

// findFlag searches for flag definition by name or alias.
func (p *Parser) findFlag(name string, cmd *CommandDef) *Flag {
	// Search in global flags first
	for _, flag := range p.Flags {
		if flag.Name == name {
			return flag
		}
		if slices.Contains(flag.Aliases, name) {
			return flag
		}
	}

	// Search in command flags if command context exists
	if cmd != nil {
		for _, flag := range cmd.Flags {
			if flag.Name == name {
				return flag
			}
			if slices.Contains(flag.Aliases, name) {
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
			if value, exists := result.Flags[flag.Name]; !exists || value == "" {
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
	if flag := r.findFlag(n); flag != nil && flag.DefValue != nil {
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

	if flag := r.findFlag(n); flag != nil && flag.DefValue != nil {
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

	if flag := r.findFlag(n); flag != nil && flag.DefValue != nil {
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

	if flag := r.findFlag(n); flag != nil && flag.DefValue != nil {
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

	if flag := r.findFlag(n); flag != nil && flag.DefValue != nil {
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
	for _, flag := range r.flag {
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
