// Package cmdflag is a minimal POSIX-style command-line flag parser covering
// the API surface oh-my-posh uses: typed Var registration, POSIX-style
// parsing (--flag=value, --flag value, shorthands, the -- terminator,
// interspersed positionals), hidden flags, and usage rendering matching
// common CLI flag conventions byte for byte for the flag types in use.
package cmdflag

import (
	"fmt"
	"strconv"
	"strings"
)

type ErrorHandling int

const (
	ContinueOnError ErrorHandling = iota
	ExitOnError
)

const (
	boolType = "bool"
	trueStr  = "true"
)

// Value is the interface to the dynamic value stored in a flag.
type Value interface {
	String() string
	Set(string) error
	Type() string
}

type Flag struct {
	Name      string
	Shorthand string
	Usage     string
	Value     Value
	DefValue  string
	Hidden    bool
	Changed   bool
}

type FlagSet struct {
	name       string
	flags      map[string]*Flag
	shorthands map[byte]*Flag
	order      []*Flag
	args       []string

	ParseErrorsAllowlist struct {
		UnknownFlags bool
	}
}

func NewFlagSet(name string, _ ErrorHandling) *FlagSet {
	return &FlagSet{
		name:       name,
		flags:      make(map[string]*Flag),
		shorthands: make(map[byte]*Flag),
	}
}

// value implementations

type stringValue struct{ p *string }

func (v stringValue) String() string     { return *v.p }
func (v stringValue) Set(s string) error { *v.p = s; return nil }
func (v stringValue) Type() string       { return "string" }

type boolValue struct{ p *bool }

func (v boolValue) String() string { return strconv.FormatBool(*v.p) }
func (v boolValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	*v.p = b
	return err
}
func (v boolValue) Type() string { return boolType }

type intValue struct{ p *int }

func (v intValue) String() string { return strconv.Itoa(*v.p) }
func (v intValue) Set(s string) error {
	i, err := strconv.ParseInt(s, 0, 64)
	*v.p = int(i)
	return err
}
func (v intValue) Type() string { return "int" }

type float64Value struct{ p *float64 }

func (v float64Value) String() string { return strconv.FormatFloat(*v.p, 'g', -1, 64) }
func (v float64Value) Set(s string) error {
	f, err := strconv.ParseFloat(s, 64)
	*v.p = f
	return err
}
func (v float64Value) Type() string { return "float64" }

// registration

func (f *FlagSet) Var(value Value, name, shorthand, usage string) *Flag {
	flag := &Flag{
		Name:      name,
		Shorthand: shorthand,
		Usage:     usage,
		Value:     value,
		DefValue:  value.String(),
	}

	f.flags[name] = flag
	f.order = append(f.order, flag)

	if shorthand != "" {
		f.shorthands[shorthand[0]] = flag
	}

	return flag
}

func (f *FlagSet) StringVar(p *string, name, value, usage string) {
	f.StringVarP(p, name, "", value, usage)
}

func (f *FlagSet) StringVarP(p *string, name, shorthand, value, usage string) {
	*p = value
	f.Var(stringValue{p}, name, shorthand, usage)
}

func (f *FlagSet) String(name, value, usage string) *string {
	p := new(string)
	f.StringVar(p, name, value, usage)
	return p
}

func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	f.BoolVarP(p, name, "", value, usage)
}

func (f *FlagSet) BoolVarP(p *bool, name, shorthand string, value bool, usage string) {
	*p = value
	f.Var(boolValue{p}, name, shorthand, usage)
}

func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	f.IntVarP(p, name, "", value, usage)
}

func (f *FlagSet) IntVarP(p *int, name, shorthand string, value int, usage string) {
	*p = value
	f.Var(intValue{p}, name, shorthand, usage)
}

func (f *FlagSet) Float64Var(p *float64, name string, value float64, usage string) {
	*p = value
	f.Var(float64Value{p}, name, "", usage)
}

// inspection

func (f *FlagSet) Lookup(name string) *Flag {
	return f.flags[name]
}

func (f *FlagSet) ShorthandLookup(name string) *Flag {
	if name == "" {
		return nil
	}

	return f.shorthands[name[0]]
}

func (f *FlagSet) Changed(name string) bool {
	flag := f.flags[name]
	return flag != nil && flag.Changed
}

func (f *FlagSet) MarkHidden(name string) error {
	flag := f.flags[name]
	if flag == nil {
		return fmt.Errorf("flag %q does not exist", name)
	}

	flag.Hidden = true
	return nil
}

// VisitAll visits the flags in registration order rather than
// lexicographically: the sole caller formats a command line and is
// order-insensitive, and usage rendering sorts separately.
func (f *FlagSet) VisitAll(fn func(*Flag)) {
	for _, flag := range f.order {
		fn(flag)
	}
}

// AddFlagSet adds flags from another set that are not yet present.
func (f *FlagSet) AddFlagSet(other *FlagSet) {
	if other == nil {
		return
	}

	other.VisitAll(func(flag *Flag) {
		if f.flags[flag.Name] == nil {
			f.flags[flag.Name] = flag
			f.order = append(f.order, flag)

			if flag.Shorthand != "" && f.shorthands[flag.Shorthand[0]] == nil {
				f.shorthands[flag.Shorthand[0]] = flag
			}
		}
	})
}

func (f *FlagSet) HasFlags() bool {
	return len(f.order) > 0
}

func (f *FlagSet) HasAvailableFlags() bool {
	for _, flag := range f.order {
		if !flag.Hidden {
			return true
		}
	}

	return false
}

func (f *FlagSet) Args() []string {
	return f.args
}

// parsing

func (f *FlagSet) Parse(arguments []string) error {
	f.args = make([]string, 0, len(arguments))

	for len(arguments) > 0 {
		arg := arguments[0]
		arguments = arguments[1:]

		switch {
		case arg == "--":
			f.args = append(f.args, arguments...)
			return nil
		case strings.HasPrefix(arg, "--"):
			var err error
			arguments, err = f.parseLong(arg[2:], arguments)
			if err != nil {
				return err
			}
		case strings.HasPrefix(arg, "-") && len(arg) > 1:
			var err error
			arguments, err = f.parseShort(arg[1:], arguments)
			if err != nil {
				return err
			}
		default:
			f.args = append(f.args, arg)
		}
	}

	return nil
}

func (f *FlagSet) parseLong(name string, rest []string) ([]string, error) {
	value := ""
	hasValue := false

	if i := strings.Index(name, "="); i >= 0 {
		value = name[i+1:]
		name = name[:i]
		hasValue = true
	}

	flag := f.flags[name]
	if flag == nil {
		if !f.ParseErrorsAllowlist.UnknownFlags {
			return rest, fmt.Errorf("unknown flag: --%s", name)
		}

		// an unknown flag given as "--flag value" swallows the
		// value token unless the next token is itself a flag
		if !hasValue && len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
			return rest[1:], nil
		}

		return rest, nil
	}

	switch {
	case hasValue:
	case flag.Value.Type() == boolType:
		value = trueStr
	case len(rest) > 0:
		value = rest[0]
		rest = rest[1:]
	default:
		return rest, fmt.Errorf("flag needs an argument: --%s", name)
	}

	if err := flag.Value.Set(value); err != nil {
		return rest, fmt.Errorf("invalid argument %q for \"--%s\" flag: %v", value, flag.Name, err)
	}

	flag.Changed = true
	return rest, nil
}

func (f *FlagSet) parseShort(shorthands string, rest []string) ([]string, error) {
	for len(shorthands) > 0 {
		c := shorthands[0]

		flag := f.shorthands[c]
		if flag == nil {
			if f.ParseErrorsAllowlist.UnknownFlags {
				// drop the remainder of the group and a
				// separate value token unless it is itself a flag
				if len(shorthands) == 1 && len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
					return rest[1:], nil
				}

				return rest, nil
			}

			return rest, fmt.Errorf("unknown shorthand flag: %q in -%s", c, shorthands)
		}

		value := ""

		switch {
		case len(shorthands) > 2 && shorthands[1] == '=':
			value = shorthands[2:]
			shorthands = ""
		case flag.Value.Type() == boolType:
			value = trueStr
			shorthands = shorthands[1:]
		case len(shorthands) > 1:
			value = shorthands[1:]
			shorthands = ""
		case len(rest) > 0:
			value = rest[0]
			rest = rest[1:]
			shorthands = ""
		default:
			return rest, fmt.Errorf("flag needs an argument: %q in -%s", c, shorthands)
		}

		if err := flag.Value.Set(value); err != nil {
			return rest, fmt.Errorf("invalid argument %q for \"-%s, --%s\" flag: %v", value, flag.Shorthand, flag.Name, err)
		}

		flag.Changed = true
	}

	return rest, nil
}

// usage rendering

// FlagUsages renders the non-hidden flags sorted by name, shorthand column,
// type name after the flag, defaults in parentheses when they differ from
// the zero value, usage aligned in a single column.
func (f *FlagSet) FlagUsages() string {
	lines := make([]string, 0, len(f.order))
	maxlen := 0

	flags := make([]*Flag, len(f.order))
	copy(flags, f.order)
	sortFlags(flags)

	for _, flag := range flags {
		if flag.Hidden {
			continue
		}

		var line string
		if flag.Shorthand != "" {
			line = fmt.Sprintf("  -%s, --%s", flag.Shorthand, flag.Name)
		}

		if flag.Shorthand == "" {
			line = fmt.Sprintf("      --%s", flag.Name)
		}

		if name := typeName(flag); name != "" {
			line += " " + name
		}

		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}

		line += flag.Usage
		if defValue, ok := defaultValue(flag); ok {
			line += fmt.Sprintf(" (default %s)", defValue)
		}

		lines = append(lines, line)
	}

	var sb strings.Builder
	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		sb.WriteString(line[:sidx])
		// the gap is maxlen-sidx+2 spaces wide to match the implicit
		// separators of a Fprintln(left, spacing, usage)-style layout
		sb.WriteString(strings.Repeat(" ", maxlen-sidx+2))
		sb.WriteString(line[sidx+1:])
		sb.WriteString("\n")
	}

	return sb.String()
}

func sortFlags(flags []*Flag) {
	for i := 1; i < len(flags); i++ {
		for j := i; j > 0 && flags[j].Name < flags[j-1].Name; j-- {
			flags[j], flags[j-1] = flags[j-1], flags[j]
		}
	}
}

func typeName(flag *Flag) string {
	switch flag.Value.Type() {
	case boolType:
		return ""
	case "float64":
		return "float"
	case "int64":
		return "int"
	default:
		return flag.Value.Type()
	}
}

func defaultValue(flag *Flag) (string, bool) {
	switch flag.Value.Type() {
	case boolType:
		return flag.DefValue, flag.DefValue == trueStr
	case "string":
		return fmt.Sprintf("%q", flag.DefValue), flag.DefValue != ""
	default:
		return flag.DefValue, flag.DefValue != "0"
	}
}
