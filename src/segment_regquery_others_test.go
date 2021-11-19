//go:build !windows

package main

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestRegQueryNonWindows(t *testing.T) {
	r := regquery{}
	assert.EqualValues(t, r.enabled(), false)
}
