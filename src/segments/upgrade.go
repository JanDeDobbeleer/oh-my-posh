package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/upgrade"
)

type upgradeData struct {
	Latest  string `json:"latest"`
	Current string `json:"current"`
}

type Upgrade struct {
	props properties.Properties
	env   runtime.Environment

	// deprecated
	Version string

	upgradeData
}

func (u *Upgrade) Template() string {
	return " \uf019 "
}

func (u *Upgrade) Init(props properties.Properties, env runtime.Environment) {
	u.props = props
	u.env = env
}

func (u *Upgrade) Enabled() bool {
	u.Current = build.Version

	latest, err := u.checkUpdate(u.Current)

	if err != nil || u.Current == latest.Latest {
		return false
	}

	u.upgradeData = *latest
	u.Version = u.Latest
	return true
}

func (u *Upgrade) checkUpdate(current string) (*upgradeData, error) {
	tag, err := upgrade.Latest(u.env)
	if err != nil {
		return nil, err
	}

	return &upgradeData{
		// strip leading v
		Latest:  tag[1:],
		Current: current,
	}, nil
}
