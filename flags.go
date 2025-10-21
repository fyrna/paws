package paws

type FlagType int

const (
	BoolType   FlagType = iota // Boolean flag (true/false)
	StringType                 // String flag
	IntType                    // Integer flag
	UintType                   // Unsigned integer flag
	FloatType                  // Floating point flag
)

// FlagTypeConstrait defines the allowed types for flag values
type FlagTypeConstrait interface {
	~string | ~bool | ~int | ~uint | ~float64
}

// Flag represents a command line flag definition
type Flag struct {
	Name       string   // Flag name (long form)
	Aliases    []string // Short aliases (single characters)
	Type       FlagType // Flag's type (string, bool, etc)
	DefValue   any      // Default value
	IsRequired bool     // Whether the flag is required
	ChoicesOpt []string // Allowed values for string flags
	Min, Max   int      // Range constraints for integer flags
	HelpText   string   // Help description
}

// Paw creates a new flag with the specified name and aliases
// The type is automatically determined from the generic type parameter T
func Paw[T FlagTypeConstrait](name string, aliases ...string) *Flag {
	var (
		t      FlagType
		defVal T
	)

	switch any(defVal).(type) {
	case bool:
		t = BoolType
		defVal = any(false).(T)
	case int:
		t = IntType
		defVal = any(0).(T)
	case uint:
		t = UintType
		defVal = any(uint(0)).(T)
	case float64:
		t = FloatType
		defVal = any(0.0).(T)
	case string:
		t = StringType
		defVal = any("").(T)
	}

	return &Flag{
		Name:     name,
		Aliases:  aliases,
		Type:     t,
		DefValue: defVal,
	}
}

// Default sets the default value for the flag
func (f *Flag) Default(value any) *Flag {
	f.DefValue = value
	return f
}

// Required marks the flag as required
func (f *Flag) Required() *Flag {
	f.IsRequired = true
	return f
}

// Help sets the help text for the flag
func (f *Flag) Help(text string) *Flag {
	f.HelpText = text
	return f
}

// Choices sets allowed values for string flags
func (f *Flag) Choices(choices ...string) *Flag {
	f.ChoicesOpt = choices
	return f
}

// Range sets minimum and maximum values for integer flags
func (f *Flag) Range(min, max int) *Flag {
	f.Min = min
	f.Max = max
	return f
}
