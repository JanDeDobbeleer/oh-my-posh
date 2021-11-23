// TODO:
//
//  * reg value -> string formatting to own fn?
//  * support for other reg types?
//  * custom formatting for values?  (custom sprintf is too dangerous, what else?)
//

package main

type winreg struct {
	props         *properties
	env           environmentInfo
	segmentString string

	KeyValue    string
	GotKeyValue bool
}

const (

	// Properties defining how to get to the key.
	RegistryPath Property = "registry_path" // path from the supplied root under which the key exists
	RegistryKey  Property = "registry_key"  // key within full reg path formed from two above
)

func (wr *winreg) init(props *properties, env environmentInfo) {
	wr.props = props
	wr.env = env
	wr.KeyValue = ""
}

func (wr *winreg) enabled() bool {
	// Definitely no point going further if it's not windows.
	if wr.env.getRuntimeGOOS() != windowsPlatform {
		return false
	}

	registryPath := wr.props.getString(RegistryPath, "")
	registryKey := wr.props.getString(RegistryKey, "")

	keyValue, err := wr.env.getWindowsRegistryKeyValue(registryPath, registryKey)

	wr.KeyValue = keyValue
	wr.GotKeyValue = (err == nil)

	// Need to do the template work here, not in string() to determine whether to display the segment
	// If there is a template and the resulting output is not empty, then enable the segment.
	segmentTemplate := wr.props.getString(SegmentTemplate, "")
	if len(segmentTemplate) > 0 {
		displayString := wr.templateString(segmentTemplate)
		if len(displayString) > 0 {
			wr.segmentString = displayString
			return true
		}
	}

	// failing that, enabled if we got a key value
	wr.segmentString = wr.KeyValue
	return wr.GotKeyValue
}

func (wr *winreg) string() string {
	return wr.segmentString
}

func (wr *winreg) templateString(segmentTemplate string) string {
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  wr,
		Env:      wr.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}
	return text
}
