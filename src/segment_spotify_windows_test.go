//go:build windows

package main

import (
	"errors"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

type spotifyArgs struct {
	title    string
	runError error
}

func bootStrapSpotifyWindowsTest(args *spotifyArgs) *spotify {
	env := new(mock.MockedEnvironment)
	env.On("WindowTitle", "spotify.exe").Return(args.title, args.runError)
	s := &spotify{
		env:   env,
		props: properties.Map{},
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
	env := new(mock.MockedEnvironment)
	env.On("WindowTitle", "spotify.exe").Return(args.title, args.runError)
	s := &spotify{
		env:   env,
		props: properties.Map{},
	}
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\ue602 Candlemass - Spellbreaker", renderTemplate(env, s.template(), s))
}

func TestSpotifyWindowsEnabledAndSpotifyStopped(t *testing.T) {
	args := &spotifyArgs{
		title: "Spotify premium",
	}
	s := bootStrapSpotifyWindowsTest(args)
	assert.Equal(t, false, s.enabled())
}
