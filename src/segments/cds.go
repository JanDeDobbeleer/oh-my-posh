package segments

import (
	"encoding/json"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Cds struct {
	language
	HasDependency bool
}

func (c *Cds) Template() string {
	return languageTemplate
}

func (c *Cds) Init(props properties.Properties, env environment.Environment) {
	c.language = language{
		env:        env,
		props:      props,
		extensions: []string{".cdsrc.json", ".cdsrc-private.json", "*.cds"},
		commands: []*cmd{
			{
				executable: "cds",
				args:       []string{"--version"},
				regex:      `@sap/cds: (?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		loadContext: c.loadContext,
		inContext:   c.inContext,
		displayMode: props.GetString(DisplayMode, DisplayModeContext),
	}
}

func (c *Cds) Enabled() bool {
	return c.language.Enabled()
}

func (c *Cds) loadContext() {
	if !c.language.env.HasFiles("package.json") {
		return
	}

	content := c.language.env.FileContent("package.json")
	objmap := map[string]json.RawMessage{}

	if err := json.Unmarshal([]byte(content), &objmap); err != nil {
		return
	}

	dependencies := map[string]json.RawMessage{}

	if err := json.Unmarshal(objmap["dependencies"], &dependencies); err != nil {
		return
	}

	for d := range dependencies {
		if d == "@sap/cds" {
			c.HasDependency = true
			break
		}
	}
}

func (c *Cds) inContext() bool {
	return c.HasDependency
}
