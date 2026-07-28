//go:build illumos || solaris

package battery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNoBattery(t *testing.T) {
	info, err := Get()

	assert.Nil(t, info)
	assert.IsType(t, &NoBatteryError{}, err)
}
