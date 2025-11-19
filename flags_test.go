package paws

import (
	"testing"
)

func TestPaw(t *testing.T) {
	tests := []struct {
		name     string
		flagType any
		expected FlagType
	}{
		{"bool flag", true, BoolType},
		{"string flag", "default", StringType},
		{"int flag", 42, IntType},
		{"uint flag", uint(42), UintType},
		{"float flag", 3.14, FloatType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var flag *Flag
			switch tt.flagType.(type) {
			case bool:
				flag = Paw[bool]("test")
			case string:
				flag = Paw[string]("test")
			case int:
				flag = Paw[int]("test")
			case uint:
				flag = Paw[uint]("test")
			case float64:
				flag = Paw[float64]("test")
			}

			if flag.Type != tt.expected {
				t.Errorf("Paw() type = %v, want %v", flag.Type, tt.expected)
			}
			if flag.Name != "test" {
				t.Errorf("Paw() name = %v, want test", flag.Name)
			}
		})
	}
}

func TestFlagMethods(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		flag := Paw[int]("count").Default(5)
		if flag.DefValue != 5 {
			t.Errorf("Default() = %v, want 5", flag.DefValue)
		}
	})

	t.Run("Required", func(t *testing.T) {
		flag := Paw[string]("file").Required()
		if !flag.IsRequired {
			t.Error("Required() did not set IsRequired to true")
		}
	})

	t.Run("Help", func(t *testing.T) {
		helpText := "input file path"
		flag := Paw[string]("file").Help(helpText)
		if flag.HelpText != helpText {
			t.Errorf("Help() = %v, want %v", flag.HelpText, helpText)
		}
	})

	t.Run("Choices", func(t *testing.T) {
		flag := Paw[string]("mode").Choices("fast", "slow", "medium")
		if flag.Type != StringType {
			t.Error("Choices should only work on string flags")
		}
		if len(flag.choices) != 3 {
			t.Errorf("Choices() created %d choices, want 3", len(flag.choices))
		}
		if len(flag.ChoicesOpt) != 3 {
			t.Errorf("ChoicesOpt has %d options, want 3", len(flag.ChoicesOpt))
		}
	})

	t.Run("Choices panic on non-string", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Choices() should panic on non-string flag")
			}
		}()
		Paw[int]("count").Choices("1", "2", "3")
	})

	t.Run("Range", func(t *testing.T) {
		flag := Paw[int]("size").Range(1, 100)
		if flag.Min != 1 || flag.Max != 100 {
			t.Errorf("Range() = [%d, %d], want [1, 100]", flag.Min, flag.Max)
		}
	})

	t.Run("Range panic on non-numeric", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Range() should panic on non-numeric flag")
			}
		}()
		Paw[string]("name").Range(1, 100)
	})
}

func TestFlagAliases(t *testing.T) {
	flag := Paw[bool]("verbose", "v", "V")
	if len(flag.Aliases) != 2 {
		t.Errorf("Aliases count = %d, want 2", len(flag.Aliases))
	}
	if flag.Aliases[0] != "v" || flag.Aliases[1] != "V" {
		t.Errorf("Aliases = %v, want [v, V]", flag.Aliases)
	}
}
