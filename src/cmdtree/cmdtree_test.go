package cmdtree

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testTree() (root *Command, ran *string) {
	ran = new(string)

	root = &Command{
		Use: "root",
		Run: func(_ *Command, _ []string) { *ran = "root" },
	}

	child := &Command{
		Use: "child",
		Run: func(_ *Command, args []string) { *ran = "child:" + join(args) },
	}

	grandchild := &Command{
		Use: "grandchild",
		Run: func(_ *Command, args []string) { *ran = "grandchild:" + join(args) },
	}

	root.AddCommand(child)
	child.AddCommand(grandchild)

	return root, ran
}

func join(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

func TestRouting(t *testing.T) {
	root, ran := testTree()
	root.SetArgs([]string{"child", "grandchild", "pos"})
	assert.NoError(t, root.Execute())
	assert.Equal(t, "grandchild:pos", *ran)
}

func TestRoutingWithFlagBeforeCommand(t *testing.T) {
	root, ran := testTree()
	var config string
	root.PersistentFlags().StringVarP(&config, "config", "c", "", "")

	root.SetArgs([]string{"--config", "x", "child", "positional"})
	assert.NoError(t, root.Execute())
	assert.Equal(t, "child:positional", *ran)
	assert.Equal(t, "x", config)
}

func TestUnknownCommand(t *testing.T) {
	root, _ := testTree()
	root.SetArgs([]string{"bogus"})
	assert.EqualError(t, root.Execute(), `unknown command "bogus" for "root"`)
}

func TestPersistentFlagInheritance(t *testing.T) {
	root, _ := testTree()
	var trace bool
	root.PersistentFlags().BoolVar(&trace, "trace", false, "")

	root.SetArgs([]string{"child", "--trace"})
	assert.NoError(t, root.Execute())
	assert.True(t, trace)
}

func TestPersistentHooksRunOnce(t *testing.T) {
	var order []string

	root := &Command{
		Use:              "root",
		PersistentPreRun: func(_ *Command, _ []string) { order = append(order, "pre") },
		PersistentPostRun: func(_ *Command, _ []string) {
			order = append(order, "post")
		},
	}

	child := &Command{
		Use: "child",
		Run: func(_ *Command, _ []string) { order = append(order, "run") },
	}

	root.AddCommand(child)
	root.SetArgs([]string{"child"})
	assert.NoError(t, root.Execute())
	assert.Equal(t, []string{"pre", "run", "post"}, order)
}

func TestArgsValidators(t *testing.T) {
	cmd := &Command{Use: "cmd"}

	assert.NoError(t, NoArgs(cmd, nil))
	assert.EqualError(t, NoArgs(cmd, []string{"x"}), `unknown command "x" for "cmd"`)

	assert.NoError(t, ExactArgs(1)(cmd, []string{"a"}))
	assert.EqualError(t, ExactArgs(1)(cmd, nil), "accepts 1 arg(s), received 0")

	assert.EqualError(t, MinimumNArgs(1)(cmd, nil), "requires at least 1 arg(s), only received 0")
	assert.EqualError(t, RangeArgs(1, 2)(cmd, nil), "accepts between 1 and 2 arg(s), received 0")

	cmd.ValidArgs = []string{"get", "set"}
	assert.NoError(t, OnlyValidArgs(cmd, []string{"get"}))
	assert.EqualError(t, OnlyValidArgs(cmd, []string{"bogus"}), `invalid argument "bogus" for "cmd"`)
}

func TestShorthandShadowing(t *testing.T) {
	// print defines -s stack-count while root defines -s shell; the
	// subcommand's local flag must win
	root := &Command{Use: "root"}
	var shell string
	root.Flags().StringVarP(&shell, "shell", "s", "", "")

	var stack int
	child := &Command{Use: "child", Run: func(_ *Command, _ []string) {}}
	child.Flags().IntVarP(&stack, "stack-count", "s", 0, "")
	root.AddCommand(child)

	root.SetArgs([]string{"child", "-s", "3"})
	assert.NoError(t, root.Execute())
	assert.Equal(t, 3, stack)
	assert.Equal(t, "", shell)
}
