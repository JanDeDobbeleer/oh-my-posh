// Package cmdtree is a minimal command-line command tree implementation covering
// the API surface oh-my-posh uses: a command tree with persistent flags and
// hooks, POSIX flag parsing via the sibling cmdflag package, positional
// argument validators, and help/usage output matching common CLI framework
// conventions byte for byte for the features in use.
package cmdtree

import (
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cmdflag"
)

// ExplorerLaunchHelpText is shown on Windows when the binary is launched from
// Explorer instead of a terminal. Setting it to "" disables the check.
var ExplorerLaunchHelpText = `This is a command line tool.

You need to open cmd.exe and run it from there.
`

type PositionalArgs func(cmd *Command, args []string) error

type Command struct {
	out               io.Writer
	flags             *cmdflag.FlagSet
	parent            *Command
	PersistentPostRun func(cmd *Command, args []string)
	pflags            *cmdflag.FlagSet
	PersistentPreRun  func(cmd *Command, args []string)
	Args              PositionalArgs
	Run               func(cmd *Command, args []string)
	Example           string
	Long              string
	Short             string
	Use               string
	ValidArgs         []string
	commands          []*Command
	requiredFlags     []string
	Aliases           []string
	setArgs           []string
	Hidden            bool
	helpRequested     bool
	CompletionOptions struct{ DisableDefaultCmd bool }
}

// SetOut redirects help output to a caller-supplied writer instead of stdout.
func (c *Command) SetOut(w io.Writer) {
	c.out = w
}

func (c *Command) outWriter() io.Writer {
	for cmd := c; cmd != nil; cmd = cmd.parent {
		if cmd.out != nil {
			return cmd.out
		}
	}

	return os.Stdout
}

func (c *Command) Name() string {
	name, _, _ := strings.Cut(c.Use, " ")
	return name
}

func (c *Command) CommandPath() string {
	if c.parent == nil {
		return c.Name()
	}

	return c.parent.CommandPath() + " " + c.Name()
}

func (c *Command) UseLine() string {
	useLine := c.Use
	if c.parent != nil {
		useLine = c.parent.CommandPath() + " " + c.Use
	}

	if c.Flags().HasAvailableFlags() && !strings.Contains(useLine, "[flags]") {
		useLine += " [flags]"
	}

	return useLine
}

func (c *Command) Root() *Command {
	if c.parent == nil {
		return c
	}

	return c.parent.Root()
}

func (c *Command) AddCommand(cmds ...*Command) {
	for _, cmd := range cmds {
		cmd.parent = c
		c.commands = append(c.commands, cmd)
	}
}

func (c *Command) Flags() *cmdflag.FlagSet {
	if c.flags == nil {
		c.flags = cmdflag.NewFlagSet(c.Name(), cmdflag.ContinueOnError)
	}

	return c.flags
}

func (c *Command) PersistentFlags() *cmdflag.FlagSet {
	if c.pflags == nil {
		c.pflags = cmdflag.NewFlagSet(c.Name(), cmdflag.ContinueOnError)
	}

	return c.pflags
}

func (c *Command) SetArgs(args []string) {
	c.setArgs = args
}

// MarkPersistentFlagRequired only applies to a flag registered on this
// command's own persistent set and errors otherwise.
func (c *Command) MarkPersistentFlagRequired(name string) error {
	flag := c.PersistentFlags().Lookup(name)
	if flag == nil {
		return fmt.Errorf("no such flag -%v", name)
	}

	c.requiredFlags = append(c.requiredFlags, name)
	return nil
}

func (c *Command) hasSubCommands() bool {
	for _, cmd := range c.commands {
		if !cmd.Hidden && cmd.Name() != "help" {
			return true
		}
	}

	return false
}

func (c *Command) findChild(name string) *Command {
	for _, cmd := range c.commands {
		if cmd.Name() == name {
			return cmd
		}

		if slices.Contains(cmd.Aliases, name) {
			return cmd
		}
	}

	return nil
}

// mergedFlags returns the command's local flags plus every ancestor's
// persistent flags, with the auto help flag registered.
func (c *Command) mergedFlags() *cmdflag.FlagSet {
	merged := cmdflag.NewFlagSet(c.Name(), cmdflag.ContinueOnError)
	merged.AddFlagSet(c.Flags())

	for cmd := c; cmd != nil; cmd = cmd.parent {
		merged.AddFlagSet(cmd.pflags)
	}

	return merged
}

// inheritedFlags returns the ancestors' persistent flags, rendered in help
// output as the "Global Flags" section.
func (c *Command) inheritedFlags() *cmdflag.FlagSet {
	inherited := cmdflag.NewFlagSet(c.Name(), cmdflag.ContinueOnError)

	for cmd := c.parent; cmd != nil; cmd = cmd.parent {
		inherited.AddFlagSet(cmd.pflags)
	}

	return inherited
}

// localFlags returns the command's own flags: declared local flags, its own
// persistent flags, and - only when already initialized - the help flag,
// so the help flag is only listed once a caller has actually requested help.
// The unknown-help-topic path renders root usage without a help flag.
func (c *Command) localFlags() *cmdflag.FlagSet {
	local := cmdflag.NewFlagSet(c.Name(), cmdflag.ContinueOnError)
	local.AddFlagSet(c.Flags())
	local.AddFlagSet(c.pflags)

	return local
}

func (c *Command) registerHelpFlag() {
	if c.Flags().Lookup("help") != nil {
		return
	}

	c.Flags().BoolVarP(&c.helpRequested, "help", "h", false, "help for "+c.Name())
}

// Execute parses os.Args (or SetArgs), routes to the addressed subcommand,
// validates positionals, and runs the hooks and the command.
func (c *Command) Execute() error {
	if c.parent == nil {
		checkExplorerLaunch()
		c.ensureHelpCommand()
	}

	args := c.setArgs
	if args == nil {
		args = os.Args[1:]
	}

	cmd, remaining, err := c.route(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Run '%s --help' for usage.\n", c.CommandPath())
		return err
	}

	return cmd.execute(remaining)
}

// route walks the command tree: at each level the first token that is not a
// flag (accounting for flags that consume a value) addresses a child.
func (c *Command) route(args []string) (*Command, []string, error) {
	current := c

	for {
		name, ok := firstPositional(current, args)
		if !ok {
			return current, args, nil
		}

		child := current.findChild(name)
		if child == nil {
			if current == c && len(current.commands) > 0 {
				return nil, nil, fmt.Errorf("unknown command %q for %q", name, c.CommandPath())
			}

			return current, args, nil
		}

		args = removeFirst(args, name)
		current = child
	}
}

// firstPositional finds the first argument that cannot be a flag or a flag
// value at this command level.
func firstPositional(c *Command, args []string) (string, bool) {
	flags := c.mergedFlags()

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--":
			return "", false
		case strings.HasPrefix(arg, "--"):
			name, _, hasValue := strings.Cut(arg[2:], "=")
			if hasValue {
				continue
			}

			if flag := flags.Lookup(name); flag != nil && flag.Value.Type() != "bool" {
				i++ // skip the flag's value
			}
		case strings.HasPrefix(arg, "-") && len(arg) > 1:
			if strings.Contains(arg, "=") {
				continue
			}

			shorthand := arg[len(arg)-1:]
			if flag := flags.ShorthandLookup(shorthand); flag != nil && flag.Value.Type() != "bool" {
				i++ // skip the flag's value
			}
		default:
			return arg, true
		}
	}

	return "", false
}

func removeFirst(args []string, value string) []string {
	result := make([]string, 0, len(args)-1)
	removed := false

	for _, arg := range args {
		if !removed && arg == value {
			removed = true
			continue
		}

		result = append(result, arg)
	}

	return result
}

func (c *Command) execute(args []string) error {
	c.registerHelpFlag()
	flags := c.mergedFlags()

	if err := flags.Parse(args); err != nil {
		return c.flagError(err)
	}

	if c.helpRequested {
		return c.Help()
	}

	for _, name := range c.requiredFlags {
		if !flags.Changed(name) {
			return c.flagError(fmt.Errorf("required flag(s) \"%s\" not set", name))
		}
	}

	positionals := flags.Args()

	if c.Args != nil {
		if err := c.Args(c, positionals); err != nil {
			return c.flagError(err)
		}
	}

	if c.Run == nil {
		return c.Help()
	}

	if hook := c.findHook(func(cmd *Command) func(*Command, []string) { return cmd.PersistentPreRun }); hook != nil {
		hook(c, positionals)
	}

	c.Run(c, positionals)

	if hook := c.findHook(func(cmd *Command) func(*Command, []string) { return cmd.PersistentPostRun }); hook != nil {
		hook(c, positionals)
	}

	return nil
}

// findHook returns the nearest defined hook walking up the tree, so only the
// closest one runs.
func (c *Command) findHook(get func(*Command) func(*Command, []string)) func(*Command, []string) {
	for cmd := c; cmd != nil; cmd = cmd.parent {
		if hook := get(cmd); hook != nil {
			return hook
		}
	}

	return nil
}

func (c *Command) flagError(err error) error {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	// print the usage string via Println, leaving a trailing blank line
	fmt.Fprintln(os.Stderr, c.usageString())
	return err
}

// ensureHelpCommand adds the implicit help command if none is registered.
func (c *Command) ensureHelpCommand() {
	if c.findChild("help") != nil {
		return
	}

	c.AddCommand(&Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
Simply type ` + c.Name() + ` help [path to command] for full details.`,
		Run: func(_ *Command, args []string) {
			target := c
			for _, name := range args {
				child := target.findChild(name)
				if child == nil {
					fmt.Printf("Unknown help topic %#q\n", args)
					fmt.Fprint(os.Stderr, c.usageString())
					return
				}

				target = child
			}

			_ = target.Help()
		},
	})
}

// help output

func (c *Command) Help() error {
	c.registerHelpFlag()

	long := c.Long
	if long == "" {
		long = c.Short
	}

	if long != "" {
		fmt.Fprintln(c.outWriter(), strings.TrimRight(long, "\n"))
		fmt.Fprintln(c.outWriter())
	}

	fmt.Fprint(c.outWriter(), c.usageString())
	return nil
}

func (c *Command) usageString() string {
	var sb strings.Builder

	sb.WriteString("Usage:\n")
	if c.Run != nil || !c.hasSubCommands() {
		sb.WriteString("  " + c.UseLine() + "\n")
	}

	if c.hasSubCommands() {
		sb.WriteString("  " + c.CommandPath() + " [command]\n")
	}

	if len(c.Aliases) > 0 {
		sb.WriteString("\nAliases:\n")
		sb.WriteString("  " + c.Name() + ", " + strings.Join(c.Aliases, ", ") + "\n")
	}

	if c.Example != "" {
		sb.WriteString("\nExamples:\n")
		sb.WriteString(c.Example + "\n")
	}

	if c.hasSubCommands() {
		sb.WriteString("\nAvailable Commands:\n")

		commands := make([]*Command, 0, len(c.commands))
		for _, cmd := range c.commands {
			if !cmd.Hidden {
				commands = append(commands, cmd)
			}
		}

		sort.Slice(commands, func(i, j int) bool { return commands[i].Name() < commands[j].Name() })

		padding := 11
		for _, cmd := range commands {
			if len(cmd.Name())+2 > padding {
				padding = len(cmd.Name()) + 2
			}
		}

		for _, cmd := range commands {
			fmt.Fprintf(&sb, "  %-*s %s\n", padding, cmd.Name(), cmd.Short)
		}
	}

	if local := c.localFlags(); local.HasAvailableFlags() {
		sb.WriteString("\nFlags:\n")
		sb.WriteString(local.FlagUsages())
	}

	if inherited := c.inheritedFlags(); inherited.HasAvailableFlags() {
		sb.WriteString("\nGlobal Flags:\n")
		sb.WriteString(inherited.FlagUsages())
	}

	if c.hasSubCommands() {
		sb.WriteString("\nUse \"" + c.CommandPath() + " [command] --help\" for more information about a command.\n")
	}

	return sb.String()
}

// positional argument validators

func NoArgs(cmd *Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}

	return nil
}

func ExactArgs(n int) PositionalArgs {
	return func(_ *Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
		}

		return nil
	}
}

func MinimumNArgs(n int) PositionalArgs {
	return func(_ *Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("requires at least %d arg(s), only received %d", n, len(args))
		}

		return nil
	}
}

func RangeArgs(minimum, maximum int) PositionalArgs {
	return func(_ *Command, args []string) error {
		if len(args) < minimum || len(args) > maximum {
			return fmt.Errorf("accepts between %d and %d arg(s), received %d", minimum, maximum, len(args))
		}

		return nil
	}
}

func OnlyValidArgs(cmd *Command, args []string) error {
	if len(cmd.ValidArgs) == 0 {
		return nil
	}

	for _, arg := range args {
		if !contains(cmd.ValidArgs, arg) {
			return fmt.Errorf("invalid argument %q for %q", arg, cmd.CommandPath())
		}
	}

	return nil
}

func contains(values []string, value string) bool {
	return slices.Contains(values, value)
}
