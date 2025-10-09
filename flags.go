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

// FlagBuilder provides a fluent interface for flag creation
type FlagBuilder struct {
	flag   *Flag
	parser *Parser
}

// some litte sugar function :3

// PawBool creates a new boolean flag builder
func (p *Parser) PawBool(name string, aliases ...string) *FlagBuilder {
	return &FlagBuilder{
		flag: &Flag{
			Name:     name,
			Aliases:  aliases,
			Type:     BoolType,
			DefValue: false,
		},
		parser: p,
	}
}

// PawString creates a new string flag builder
func (p *Parser) PawString(name string, aliases ...string) *FlagBuilder {
	return &FlagBuilder{
		flag: &Flag{
			Name:    name,
			Aliases: aliases,
			Type:    StringType,
		},
		parser: p,
	}
}

// PawInt creates a new integer flag builder
func (p *Parser) PawInt(name string, aliases ...string) *FlagBuilder {
	return &FlagBuilder{
		flag: &Flag{
			Name:    name,
			Aliases: aliases,
			Type:    IntType,
		},
		parser: p,
	}
}

// PawUint creates a new unsigned integer flag builder
func (p *Parser) PawUint(name string, aliases ...string) *FlagBuilder {
	return &FlagBuilder{
		flag: &Flag{
			Name:    name,
			Aliases: aliases,
			Type:    UintType,
		},
		parser: p,
	}
}

// PawFloat creates a new float flag builder
func (p *Parser) PawFloat(name string, aliases ...string) *FlagBuilder {
	return &FlagBuilder{
		flag: &Flag{
			Name:    name,
			Aliases: aliases,
			Type:    FloatType,
		},
		parser: p,
	}
}

func (b *FlagBuilder) Default(value any) *FlagBuilder {
	b.flag.DefValue = value
	return b
}

func (b *FlagBuilder) Required() *FlagBuilder {
	b.flag.IsRequired = true
	return b
}

func (b *FlagBuilder) Help(text string) *FlagBuilder {
	b.flag.HelpText = text
	return b
}

func (b *FlagBuilder) Choices(choices ...string) *FlagBuilder {
	b.flag.ChoicesOpt = choices
	return b
}

func (b *FlagBuilder) Range(min, max int) *FlagBuilder {
	b.flag.Min = min
	b.flag.Max = max
	return b
}

// for chaining
func (b *FlagBuilder) END() *Parser {
	b.parser.AddFlags(b.flag)
	return b.parser
}
