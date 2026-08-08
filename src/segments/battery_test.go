package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/battery"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestBatteryEnabledAtZeroPercent(t *testing.T) {
	env := new(mock.Environment)
	env.On("IsWsl").Return(false)
	env.On("BatteryState").Return(&battery.Info{
		Percentage: 0,
		State:      battery.NotCharging,
	}, nil)

	segment := &Battery{}
	segment.Init(options.Map{}, env)

	assert.True(t, segment.Enabled())
}
