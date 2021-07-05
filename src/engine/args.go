package engine

// Args holds the arguments passed to the engine.
type Args struct {
	ErrorCode      *int
	PrintConfig    *bool
	ConfigFormat   *string
	PrintShell     *bool
	Config         *string
	Shell          *string
	PWD            *string
	PSWD           *string
	Version        *bool
	Debug          *bool
	ExecutionTime  *float64
	Millis         *bool
	Eval           *bool
	Init           *bool
	PrintInit      *bool
	ExportPNG      *bool
	Author         *string
	CursorPadding  *int
	RPromptOffset  *int
	StackCount     *int
	Command        *string
	PrintTransient *bool
}
