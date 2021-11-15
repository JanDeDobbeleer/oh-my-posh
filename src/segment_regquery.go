package main

type regquery struct {
    props          *properties
    env            environmentInfo
	content			string;
}

const (
     //NewProp switches something
     RegistryPath Property = "registry_path"
	 RegistryKey Property = "registry_key"
)

func (r *regquery) string() string {
    newText := "Yes"//r.content;
    return newText
}

func (r *regquery) init(props *properties, env environmentInfo) {
    r.props = props
    r.env = env
}