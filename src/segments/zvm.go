package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

const (
	// DefaultZigIcon is the default icon used if none is specified
	DefaultZigIcon = "ZVM"

	// PropertyZigIcon is the property key for the zig icon
	PropertyZigIcon properties.Property = "zigicon"
)

// Zvm represents a Zig Version Manager segment
type Zvm struct {
	language
	Version  string // Public for template access
	ZigIcon  string // Public for template access
	colorCmd *colorCommand
}

type colorCommand struct {
	env runtime.Environment
}

// colorState represents the ZVM color configuration state
type colorState struct {
	enabled bool
	valid   bool
}

func (c *colorCommand) detectColorState() colorState {
	output, err := c.env.RunCommand("zvm", "--color")
	if err != nil {
		return colorState{valid: false}
	}

	output = strings.ToLower(strings.TrimSpace(output))
	switch output {
	case "on", "yes", "y", "enabled", "true":
		return colorState{enabled: true, valid: true}
	case "off", "no", "n", "disabled", "false":
		return colorState{enabled: false, valid: true}
	default:
		return colorState{valid: false}
	}
}

func (c *colorCommand) setColor(enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	_, err := c.env.RunCommand("zvm", "--color", value)
	return err
}

// SetText sets the version text
func (z *Zvm) SetText(text string) {
	z.Version = text
}

// Text returns the current version
func (z *Zvm) Text() string {
	return z.Version
}

// Template returns the template string for the segment
func (z *Zvm) Template() string {
	return " {{ if .ZigIcon }}{{ .ZigIcon }} {{ end }}{{ .Version }} "
}

// Init initializes the segment with the given properties and environment
func (z *Zvm) Init(props properties.Properties, env runtime.Environment) {
	z.props = props
	z.env = env
	z.colorCmd = &colorCommand{env: env}

	z.ZigIcon = z.props.GetString(PropertyZigIcon, DefaultZigIcon)

	// Only try to get version if zvm command exists
	if z.env.HasCommand("zvm") {
		z.Version = z.getZvmVersion()
	}
}

// Enabled returns true if the segment should be enabled
func (z *Zvm) Enabled() bool {
	if !z.env.HasCommand("zvm") {
		return false
	}
	return z.Version != ""
}

// getZvmVersion returns the current active Zvm version
func (z *Zvm) getZvmVersion() string {
	// Detect current color state
	originalState := z.colorCmd.detectColorState()

	// If we couldn't detect the state, proceed with color disabled
	if !originalState.valid {
		if err := z.colorCmd.setColor(false); err != nil {
			return ""
		}
		defer func() {
			_ = z.colorCmd.setColor(true) // Best effort to restore color
		}()
	} else if originalState.enabled {
		// Temporarily disable colors if they were enabled
		if err := z.colorCmd.setColor(false); err != nil {
			return ""
		}
		defer func() {
			_ = z.colorCmd.setColor(originalState.enabled) // Restore original state
		}()
	}

	// Get version list
	output, err := z.env.RunCommand("zvm", "list")
	if err != nil {
		return ""
	}

	return parseActiveVersion(output)
}

// parseActiveVersion extracts the active version from zvm list output
func parseActiveVersion(output string) string {
	words := strings.Fields(output)
	for _, word := range words {
		if !strings.Contains(word, "[x]") {
			continue
		}
		return strings.TrimSpace(strings.ReplaceAll(word, "[x]", ""))
	}
	return ""
}
