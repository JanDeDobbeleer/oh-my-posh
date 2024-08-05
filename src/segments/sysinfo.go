package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type SystemInfo struct {
	props properties.Properties
	env   runtime.Environment
	runtime.SystemInfo
	Precision int
}

const (
	// Precision number of decimal places to show
	Precision properties.Property = "precision"
)

func (s *SystemInfo) Template() string {
	return " {{ round .PhysicalPercentUsed .Precision }} "
}

func (s *SystemInfo) Enabled() bool {
	if s.PhysicalPercentUsed == 0 && s.SwapPercentUsed == 0 {
		return false
	}
	return true
}

func (s *SystemInfo) Init(props properties.Properties, env runtime.Environment) {
	s.props = props
	s.env = env
	s.Precision = s.props.GetInt(Precision, 2)
	sysInfo, err := env.SystemInfo()
	if err != nil {
		return
	}
	s.SystemInfo = *sysInfo
}
