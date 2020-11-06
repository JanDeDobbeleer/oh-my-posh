// +build windows

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyWindowsEnabledWithoutAutoHotkey(t *testing.T) {
	env := new(MockedEnvironment)
	props := &properties{}
	env.On("hasCommand", "AutoHotkey").Return(false)
	s := &spotify{
		env:   env,
		props: props,
	}
	assert.Equal(t, false, s.enabled())
}

func TestSpotifyWindowsEnabledWithAutoHotkeyAndSpotifyPlaying(t *testing.T) {
	env := new(MockedEnvironment)
	props := &properties{}
	env.On("hasCommand", "AutoHotkey").Return(true)
	env.On("runCommand", "AutoHotkey", []string{""}).Return("Candlemass - Spellbreaker", nil)
	s := &spotify{
		env:   env,
		props: props,
	}
	expected := &spotify{
		env:    env,
		props:  props,
		artist: "Candlemass",
		track:  "Spellbreaker",
		status: "playing",
	}
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, expected, s)
}

func TestSpotifyWindowsEnabledWithAutoHotkeyAndSpotifyStopped(t *testing.T) {
	env := new(MockedEnvironment)
	props := &properties{}
	env.On("hasCommand", "AutoHotkey").Return(true)
	env.On("runCommand", "AutoHotkey", []string{""}).Return("Spotify premium", nil)
	s := &spotify{
		env:   env,
		props: props,
	}
	expected := &spotify{
		env:    env,
		props:  props,
		artist: "",
		track:  "",
		status: "stopped",
	}
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, expected, s)
}
