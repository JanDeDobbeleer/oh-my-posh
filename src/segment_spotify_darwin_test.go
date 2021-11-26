//go:build darwin

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type spotifyArgs struct {
	running  string
	status   string
	artist   string
	track    string
	runError error
}

func bootStrapSpotifyDarwinTest(args *spotifyArgs) *spotify {
	env := new(MockedEnvironment)
	env.On("runCommand", "osascript", []string{"-e", "application \"Spotify\" is running"}).Return(args.running, args.runError)
	env.On("runCommand", "osascript", []string{"-e", "tell application \"Spotify\" to player state as string"}).Return(args.status, nil)
	env.On("runCommand", "osascript", []string{"-e", "tell application \"Spotify\" to artist of current track as string"}).Return(args.artist, nil)
	env.On("runCommand", "osascript", []string{"-e", "tell application \"Spotify\" to name of current track as string"}).Return(args.track, nil)
	s := &spotify{
		env: env,
	}
	return s
}

func TestSpotifyDarwinEnabledAndSpotifyNotRunning(t *testing.T) {
	args := &spotifyArgs{
		running: "false",
	}
	s := bootStrapSpotifyDarwinTest(args)
	assert.Equal(t, false, s.enabled())
}

func TestSpotifyDarwinEnabledAndSpotifyPlaying(t *testing.T) {
	args := &spotifyArgs{
		running: "true",
		status:  "playing",
		artist:  "Candlemass",
		track:   "Spellbreaker",
	}
	s := bootStrapSpotifyDarwinTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\ue602 Candlemass - Spellbreaker", s.string())
}

func TestSpotifyDarwinEnabledAndSpotifyPaused(t *testing.T) {
	args := &spotifyArgs{
		running: "true",
		status:  "paused",
		artist:  "Candlemass",
		track:   "Spellbreaker",
	}
	s := bootStrapSpotifyDarwinTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\uF8E3 Candlemass - Spellbreaker", s.string())
}
