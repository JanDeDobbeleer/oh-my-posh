package segments

import "github.com/jandedobbeleer/oh-my-posh/src/properties"

type Distrobox struct {
	base

    ContainerID string
    Icon        string
}

const (
    DistroboxIcon properties.Property = "icon"
)

func (d *Distrobox) Enabled() bool {
    d.ContainerID = d.env.Getenv("CONTAINER_ID")
    d.Icon = d.props.GetString(DistroboxIcon, "\uED95")

    return d.ContainerID != ""
}

func (d *Distrobox) Template() string {
    return "{{ .Icon }} {{ .ContainerID }} "
}
