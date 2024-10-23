package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type SystemInfo struct {
	base

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
	s.Precision = s.props.GetInt(Precision, 2)

	sysInfo, err := s.env.SystemInfo()
	if err != nil {
		return false
	}

	s.SystemInfo = *sysInfo

	if s.PhysicalPercentUsed == 0 && s.SwapPercentUsed == 0 {
		return false
	}

	return true
}
