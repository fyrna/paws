package paws

import (
	"testing"
)

func TestNewParser(t *testing.T) {
	parser := New()
	if parser == nil {
		t.Error("New() returned nil")
	}
	if len(parser.Commands) != 0 {
		t.Error("New parser should have no commands")
	}
	if len(parser.Flags) != 0 {
		t.Error("New parser should have no flags")
	}
}

func TestAddCommand(t *testing.T) {
	parser := New()
	cmdPath := []string{"git", "commit"}
	flags := []*Flag{
		Paw[string]("message", "m"),
		Paw[bool]("amend", "a"),
	}

	parser.AddCommand(cmdPath, flags)

	if len(parser.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(parser.Commands))
	}

	cmd := parser.Commands[0]
	if len(cmd.Path) != 2 || cmd.Path[0] != "git" || cmd.Path[1] != "commit" {
		t.Errorf("Command path = %v, want [git commit]", cmd.Path)
	}
	if len(cmd.Flags) != 2 {
		t.Errorf("Command flags count = %d, want 2", len(cmd.Flags))
	}
}

func TestAddFlags(t *testing.T) {
	parser := New()
	flags := []*Flag{
		Paw[bool]("verbose", "v"),
		Paw[string]("config", "c"),
	}

	parser.AddFlags(flags...)

	if len(parser.Flags) != 2 {
		t.Fatalf("Expected 2 flags, got %d", len(parser.Flags))
	}
	if parser.flagIndex == nil {
		t.Error("flagIndex should be initialized after AddFlags")
	}
}

func TestParseGlobalFlags(t *testing.T) {
	parser := New()
	parser.AddFlags(
		Paw[bool]("verbose", "v"),
		Paw[string]("file", "f"),
		Paw[int]("count"),
	)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(*ParseResult) bool
	}{
		{
			name:    "boolean flag",
			args:    []string{"--verbose"},
			wantErr: false,
			check:   func(r *ParseResult) bool { return r.Bool("verbose") },
		},
		{
			name:    "boolean short flag",
			args:    []string{"-v"},
			wantErr: false,
			check:   func(r *ParseResult) bool { return r.Bool("verbose") },
		},
		{
			name:    "string flag",
			args:    []string{"--file", "test.txt"},
			wantErr: false,
			check:   func(r *ParseResult) bool { return r.String("file") == "test.txt" },
		},
		{
			name:    "string flag with equals",
			args:    []string{"--file=test.txt"},
			wantErr: false,
			check:   func(r *ParseResult) bool { return r.String("file") == "test.txt" },
		},
		{
			name:    "int flag",
			args:    []string{"--count", "42"},
			wantErr: false,
			check:   func(r *ParseResult) bool { return r.Int("count") == 42 },
		},
		{
			name:    "missing value",
			args:    []string{"--file"},
			wantErr: true,
		},
		{
			name:    "unknown flag",
			args:    []string{"--unknown"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("Parse() result check failed")
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	parser := New()

	parser.AddCommand([]string{"git", "commit"}, []*Flag{
		Paw[string]("message", "m"),
		Paw[bool]("amend"),
	})

	parser.AddCommand([]string{"git", "push"}, []*Flag{
		Paw[bool]("force", "f"),
	})

	tests := []struct {
		name        string
		args        []string
		wantCommand string
		wantErr     bool
	}{
		{
			name:        "exact command match",
			args:        []string{"git", "commit", "--message", "test"},
			wantCommand: "commit",
			wantErr:     false,
		},
		{
			name:        "partial command no match",
			args:        []string{"git", "status"},
			wantCommand: "",
			wantErr:     false,
		},
		{
			name:        "command with flags",
			args:        []string{"git", "push", "--force"},
			wantCommand: "push",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.wantCommand == "" && result.Command != nil {
					t.Error("Expected no command match")
				}
				if tt.wantCommand != "" && (result.Command == nil || result.Command.Path[len(result.Command.Path)-1] != tt.wantCommand) {
					t.Errorf("Expected command %s, got %v", tt.wantCommand, result.Command)
				}
			}
		})
	}
}

func TestGroupedShortFlags(t *testing.T) {
	parser := New()

	parser.AddFlags(
		Paw[bool]("verbose", "v"),
		Paw[bool]("interactive", "i"),
		Paw[bool]("force", "f"),
	)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
		check   func(*ParseResult) bool
	}{
		{
			name:    "grouped boolean flags",
			args:    []string{"-vif"},
			wantErr: false,
			check: func(r *ParseResult) bool {
				return r.Bool("verbose") && r.Bool("interactive") && r.Bool("force")
			},
		},
		{
			name:    "single boolean flag",
			args:    []string{"-v"},
			wantErr: false,
			check:   func(r *ParseResult) bool { return r.Bool("verbose") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(result) {
				t.Error("Parse() result check failed")
			}
		})
	}
}

func TestDoubleDash(t *testing.T) {
	parser := New()

	parser.AddFlags(Paw[bool]("verbose", "v"))

	args := []string{"--verbose", "--", "positional", "--looks-like-flag"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !result.DoubleDash {
		t.Error("DoubleDash should be true")
	}
	if len(result.Positional) != 2 {
		t.Errorf("Expected 2 positional args, got %d", len(result.Positional))
	}
	if result.Positional[0] != "positional" || result.Positional[1] != "--looks-like-flag" {
		t.Errorf("Positional args = %v, want [positional --looks-like-flag]", result.Positional)
	}
}

func TestValidateRequired(t *testing.T) {
	parser := New()

	parser.AddFlags(
		Paw[string]("required-flag").Required(),
		Paw[string]("optional-flag"),
	)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "required flag provided",
			args:    []string{"--required-flag", "value"},
			wantErr: false,
		},
		{
			name:    "required flag missing",
			args:    []string{"--optional-flag", "value"},
			wantErr: true,
		},
		{
			name:    "no flags provided",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			err = parser.ValidateRequired(result)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlagValidation(t *testing.T) {
	parser := New()

	tests := []struct {
		name    string
		flag    *Flag
		value   string
		wantErr bool
	}{
		{
			name:    "valid string choice",
			flag:    Paw[string]("mode").Choices("fast", "slow"),
			value:   "fast",
			wantErr: false,
		},
		{
			name:    "invalid string choice",
			flag:    Paw[string]("mode").Choices("fast", "slow"),
			value:   "medium",
			wantErr: true,
		},
		{
			name:    "valid int in range",
			flag:    Paw[int]("count").Range(1, 10),
			value:   "5",
			wantErr: false,
		},
		{
			name:    "int below range",
			flag:    Paw[int]("count").Range(1, 10),
			value:   "0",
			wantErr: true,
		},
		{
			name:    "int above range",
			flag:    Paw[int]("count").Range(1, 10),
			value:   "11",
			wantErr: true,
		},
		{
			name:    "invalid int",
			flag:    Paw[int]("count"),
			value:   "not-a-number",
			wantErr: true,
		},
		{
			name:    "valid uint",
			flag:    Paw[uint]("size").Range(1, 100),
			value:   "50",
			wantErr: false,
		},
		{
			name:    "uint below range",
			flag:    Paw[uint]("size").Range(1, 100),
			value:   "0",
			wantErr: true,
		},
		{
			name:    "valid float",
			flag:    Paw[float64]("ratio").Range(0, 1),
			value:   "0.5",
			wantErr: false,
		},
		{
			name:    "invalid float",
			flag:    Paw[float64]("ratio"),
			value:   "not-a-float",
			wantErr: true,
		},
		{
			name:    "valid boolean true",
			flag:    Paw[bool]("flag"),
			value:   "true",
			wantErr: false,
		},
		{
			name:    "valid boolean yes",
			flag:    Paw[bool]("flag"),
			value:   "yes",
			wantErr: false,
		},
		{
			name:    "valid boolean 1",
			flag:    Paw[bool]("flag"),
			value:   "1",
			wantErr: false,
		},
		{
			name:    "invalid boolean",
			flag:    Paw[bool]("flag"),
			value:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser.AddFlags(tt.flag)
			err := parser.validateFlagValue(tt.flag, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFlagValue() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Clean up for next test
			parser.Flags = nil
			parser.flagIndex = nil
		})
	}
}

func TestResultGetters(t *testing.T) {
	parser := New()

	parser.AddFlags(
		Paw[bool]("bool-flag").Default(true),
		Paw[string]("string-flag").Default("default"),
		Paw[int]("int-flag").Default(42),
		Paw[uint]("uint-flag").Default(uint(100)),
		Paw[float64]("float-flag").Default(3.14),
	)

	args := []string{"--bool-flag", "false", "--string-flag", "custom", "--int-flag", "99"}
	result, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"Bool with override", result.Bool("bool-flag"), false},
		{"Bool with default", result.Bool("nonexistent"), false},
		{"String with override", result.String("string-flag"), "custom"},
		{"String with default", result.String("nonexistent"), ""},
		{"Int with override", result.Int("int-flag"), 99},
		{"Int with default", result.Int("nonexistent"), 0},
		{"Uint with default", result.Uint("uint-flag"), uint(100)},
		{"Float with default", result.Float("float-flag"), 3.14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestBooleanFlagVariations(t *testing.T) {
	parser := New()

	parser.AddFlags(Paw[bool]("flag"))

	// Test boolean flag without value (should be true)
	t.Run("boolean_flag_without_value", func(t *testing.T) {
		result, err := parser.Parse([]string{"--flag"})
		if err != nil {
			t.Errorf("Parse() with --flag error = %v", err)
		}
		if !result.Bool("flag") {
			t.Error("Boolean flag without value should be true")
		}
	})

	// Test boolean flag with explicit values
	truthyValues := []string{"true", "t", "yes", "y", "1", "TRUE", "YES"}
	falsyValues := []string{"false", "f", "no", "n", "0", "FALSE", "NO"}

	for _, value := range truthyValues {
		t.Run("truthy_"+value, func(t *testing.T) {
			result, err := parser.Parse([]string{"--flag", value})
			if err != nil {
				t.Errorf("Parse() with --flag %s error = %v", value, err)
			}
			if !result.Bool("flag") {
				t.Errorf("Boolean value %s should be true", value)
			}
		})
	}

	for _, value := range falsyValues {
		t.Run("falsy_"+value, func(t *testing.T) {
			result, err := parser.Parse([]string{"--flag", value})
			if err != nil {
				t.Errorf("Parse() with --flag %s error = %v", value, err)
			}
			if result.Bool("flag") {
				t.Errorf("Boolean value %s should be false", value)
			}
		})
	}
}

func TestBooleanFlagBehavior(t *testing.T) {
	parser := New()

	parser.AddFlags(Paw[bool]("verbose", "v"))

	tests := []struct {
		name        string
		args        []string
		wantVerbose bool
		wantPos     []string
	}{
		{
			name:        "boolean flag without value",
			args:        []string{"--verbose", "positional"},
			wantVerbose: true,
			wantPos:     []string{"positional"},
		},
		{
			name:        "boolean flag with explicit true",
			args:        []string{"--verbose", "true", "positional"},
			wantVerbose: true,
			wantPos:     []string{"positional"},
		},
		{
			name:        "boolean flag with explicit false",
			args:        []string{"--verbose", "false", "positional"},
			wantVerbose: false,
			wantPos:     []string{"positional"},
		},
		{
			name:        "short boolean flag without value",
			args:        []string{"-v", "positional"},
			wantVerbose: true,
			wantPos:     []string{"positional"},
		},
		{
			name:        "boolean flag with non-boolean next arg",
			args:        []string{"--verbose", "positional1", "positional2"},
			wantVerbose: true,
			wantPos:     []string{"positional1", "positional2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if result.Bool("verbose") != tt.wantVerbose {
				t.Errorf("Verbose flag = %v, want %v", result.Bool("verbose"), tt.wantVerbose)
			}

			if len(result.Positional) != len(tt.wantPos) {
				t.Errorf("Positional count = %d, want %d", len(result.Positional), len(tt.wantPos))
				return
			}

			for i, pos := range result.Positional {
				if pos != tt.wantPos[i] {
					t.Errorf("Positional[%d] = %s, want %s", i, pos, tt.wantPos[i])
				}
			}
		})
	}
}

func TestPositionalArguments(t *testing.T) {
	parser := New()

	parser.AddFlags(Paw[bool]("verbose", "v"))

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "only positional",
			args:     []string{"arg1", "arg2", "arg3"},
			expected: []string{"arg1", "arg2", "arg3"},
		},
		{
			name:     "mixed flags and positional - boolean without value",
			args:     []string{"--verbose", "arg1", "-v", "arg2"},
			expected: []string{"arg1", "arg2"},
		},
		{
			name:     "mixed flags and positional - boolean with explicit value",
			args:     []string{"--verbose", "true", "arg1", "-v", "false", "arg2"},
			expected: []string{"arg1", "arg2"},
		},
		{
			name:     "positional after double dash",
			args:     []string{"--verbose", "--", "arg1", "--looks-like-flag"},
			expected: []string{"arg1", "--looks-like-flag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(result.Positional) != len(tt.expected) {
				t.Errorf("Positional count = %d, want %d", len(result.Positional), len(tt.expected))
			}

			for i, pos := range result.Positional {
				if pos != tt.expected[i] {
					t.Errorf("Positional[%d] = %s, want %s", i, pos, tt.expected[i])
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty args", func(t *testing.T) {
		parser := New()
		result, err := parser.Parse([]string{})
		if err != nil {
			t.Fatalf("Parse() with empty args error = %v", err)
		}
		if len(result.Positional) != 0 {
			t.Error("Empty args should produce no positional arguments")
		}
	})

	t.Run("only double dash", func(t *testing.T) {
		parser := New()
		result, err := parser.Parse([]string{"--"})
		if err != nil {
			t.Fatalf("Parse() with only -- error = %v", err)
		}
		if !result.DoubleDash {
			t.Error("DoubleDash should be true")
		}
		if len(result.Positional) != 0 {
			t.Error("Only -- should produce no positional arguments")
		}
	})

	t.Run("non-boolean flag in group", func(t *testing.T) {
		parser := New()
		parser.AddFlags(
			Paw[bool]("verbose", "v"),
			Paw[string]("file", "f"),
		)

		_, err := parser.Parse([]string{"-vf", "filename"})
		if err == nil {
			t.Error("Expected error for non-boolean flag in group")
		}
	})
}
