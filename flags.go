package paws

type FlagType int

const (
	BoolType   FlagType = iota // Boolean flag (true/false)
	StringType                 // String flag
	IntType                    // Integer flag
	UintType                   // Unsigned integer flag
	FloatType                  // Floating point flag
)

// FlagTypeConstraint defines the allowed types for flag values
type FlagTypeConstraint interface {
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

	choices map[string]struct{}
}

// Paw creates a new flag with the specified name and aliases
// The type is automatically determined from the generic type parameter T
func Paw[T FlagTypeConstraint](name string, aliases ...string) *Flag {
	var (
		def T
		t   FlagType
	)

	switch any(def).(type) {
	case bool:
		t = BoolType
	case int:
		t = IntType
	case uint:
		t = UintType
	case float64:
		t = FloatType
	case string:
		t = StringType
	}

	return &Flag{
		Name:     name,
		Aliases:  aliases,
		Type:     t,
		DefValue: def,
	}
}

// Default sets the default value for the flag
func (f *Flag) Default(v any) *Flag {
	f.DefValue = v
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

// Choices only valid for string flags.
func (f *Flag) Choices(opts ...string) *Flag {
	if f.Type != StringType {
		panic("Choices can only be used on string flags")
	}

	f.choices = make(map[string]struct{}, len(opts))
	for _, o := range opts {
		f.choices[o] = struct{}{}
	}
	f.ChoicesOpt = opts
	return f
}

// Range only valid for numeric flags.
func (f *Flag) Range(min, max int) *Flag {
	if f.Type != IntType && f.Type != UintType && f.Type != FloatType {
		panic("Range can only be used on int/uint/float flags")
	}
	f.Min = min
	f.Max = max
	return f
}
