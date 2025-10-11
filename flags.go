package paws

type FlagType int

const (
	BoolType FlagType = iota
	StringType
	IntType
	UintType
	FloatType
)

// Flag represents a command line flag definition
type Flag struct {
	Name       string
	Aliases    []string
	Type       FlagType
	DefValue   any
	IsRequired bool
	ChoicesOpt []string
	Min, Max   int
	HelpText   string
}

// some litte sugar function :3

// PawBool creates a new boolean flag
func PawBool(name string, aliases ...string) *Flag {
	return &Flag{
		Name:     name,
		Aliases:  aliases,
		Type:     BoolType,
		DefValue: false,
	}
}

// PawString creates a new string flag
func PawString(name string, aliases ...string) *Flag {
	return &Flag{
		Name:    name,
		Aliases: aliases,
		Type:    StringType,
	}
}

// PawInt creates a new integer flag
func PawInt(name string, aliases ...string) *Flag {
	return &Flag{
		Name:    name,
		Aliases: aliases,
		Type:    IntType,
	}
}

// PawUint creates a new unsigned integer flag
func PawUint(name string, aliases ...string) *Flag {
	return &Flag{
		Name:    name,
		Aliases: aliases,
		Type:    UintType,
	}
}

// PawFloat creates a new float flag
func PawFloat(name string, aliases ...string) *Flag {
	return &Flag{
		Name:    name,
		Aliases: aliases,
		Type:    FloatType,
	}
}

func (f *Flag) Default(value any) *Flag {
	f.DefValue = value
	return f
}

func (f *Flag) Required() *Flag {
	f.IsRequired = true
	return f
}

func (f *Flag) Help(text string) *Flag {
	f.HelpText = text
	return f
}

func (f *Flag) Choices(choices ...string) *Flag {
	f.ChoicesOpt = choices
	return f
}

func (f *Flag) Range(min, max int) *Flag {
	f.Min = min
	f.Max = max
	return f
}
