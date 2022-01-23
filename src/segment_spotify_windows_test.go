//go:build windows

package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type spotifyArgs struct {
	title    string
	runError error
}

func bootStrapSpotifyWindowsTest(args *spotifyArgs) *spotify {
	env := new(MockedEnvironment)
	env.On("WindowTitle", "spotify.exe").Return(args.title, args.runError)
	env.onTemplate()
	s := &spotify{
		env:   env,
		props: properties{},
	}
	return s
}

func TestSpotifyWindowsEnabledAndSpotifyNotRunning(t *testing.T) {
	args := &spotifyArgs{
		runError: errors.New(""),
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, false, s.enabled())
}

func TestSpotifyWindowsEnabledAndSpotifyPlaying(t *testing.T) {
	args := &spotifyArgs{
		title: "Candlemass - Spellbreaker",
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\ue602 Candlemass - Spellbreaker", s.string())
}

func TestSpotifyWindowsEnabledAndSpotifyStopped(t *testing.T) {
	args := &spotifyArgs{
		title: "Spotify premium",
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, false, s.enabled())
}
