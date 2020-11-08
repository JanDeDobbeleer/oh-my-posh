// +build windows

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type spotifyArgs struct {
	hasAutoHoykey       bool
	spotifyWindowsTitle string
}

func bootStrapSpotifyWindowsTest(args *spotifyArgs) *spotify {
	env := new(MockedEnvironment)
	env.On("hasCommand", "AutoHotkey").Return(args.hasAutoHoykey)
	env.On("runCommand", "AutoHotkey", []string{""}).Return(args.spotifyWindowsTitle, nil)
	props := &properties{}
	s := &spotify{
		env:   env,
		props: props,
	}
	return s
}

func TestSpotifyWindowsEnabledWithoutAutoHotkey(t *testing.T) {
	args := &spotifyArgs{
		hasAutoHoykey: false,
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, false, s.enabled())
}

func TestSpotifyWindowsEnabledWithAutoHotkeyAndSpotifyPlaying(t *testing.T) {
	args := &spotifyArgs{
		hasAutoHoykey:       true,
		spotifyWindowsTitle: "Candlemass - Spellbreaker",
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\ue602 Candlemass - Spellbreaker", s.string())
}

func TestSpotifyWindowsEnabledWithAutoHotkeyAndSpotifyStopped(t *testing.T) {
	args := &spotifyArgs{
		hasAutoHoykey:       true,
		spotifyWindowsTitle: "Spotify premium",
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\uf04d ", s.string())
}
