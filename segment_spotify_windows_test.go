// +build windows

package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type spotifyArgs struct {
	spotifyWindowsTitle    string
	spotifyNotRunningError error
}

func bootStrapSpotifyWindowsTest(args *spotifyArgs) *spotify {
	env := new(MockedEnvironment)
	env.On("getWindowTitle", "spotify.exe").Return(args.spotifyWindowsTitle, args.spotifyNotRunningError)
	props := &properties{}
	s := &spotify{
		env:   env,
		props: props,
	}
	return s
}

func TestSpotifyWindowsEnabledAndSpotifyNotRunning(t *testing.T) {
	args := &spotifyArgs{
		spotifyNotRunningError: errors.New(""),
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, false, s.enabled())
}

func TestSpotifyWindowsEnabledAndSpotifyPlaying(t *testing.T) {
	args := &spotifyArgs{
		spotifyWindowsTitle: "Candlemass - Spellbreaker",
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\ue602 Candlemass - Spellbreaker", s.string())
}

func TestSpotifyWindowsEnabledAndSpotifyStopped(t *testing.T) {
	args := &spotifyArgs{
		spotifyWindowsTitle: "Spotify premium",
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\uf04d ", s.string())
}
