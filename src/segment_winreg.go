package main

type winreg struct {
	props *properties
	env   environmentInfo

	Value string
}

const (
	// path from the supplied root under which the key exists
	RegistryPath Property = "path"
	// key within full reg path formed from two above
	RegistryKey Property = "key"
)

func (wr *winreg) init(props *properties, env environmentInfo) {
	wr.props = props
	wr.env = env
}

func (wr *winreg) enabled() bool {
	if wr.env.getRuntimeGOOS() != windowsPlatform {
		return false
	}

	registryPath := wr.props.getString(RegistryPath, "")
	registryKey := wr.props.getString(RegistryKey, "")

	var err error
	wr.Value, err = wr.env.getWindowsRegistryKeyValue(registryPath, registryKey)
	return err == nil
}

func (wr *winreg) string() string {
	segmentTemplate := wr.props.getString(SegmentTemplate, "{{ .Value }}")
	return wr.templateString(segmentTemplate)
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
