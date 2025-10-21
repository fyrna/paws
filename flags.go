package paws

type FlagType int

const (
	BoolType FlagType = iota
	StringType
	IntType
	UintType
	FloatType
)

type FlagTypeConstrait interface {
	~string | ~bool | ~int | ~uint | ~float64
}

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
