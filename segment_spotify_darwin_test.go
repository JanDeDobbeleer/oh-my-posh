// +build darwin

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type spotifyArgs struct {
	spotifyDarwinTitle   string
	spotifyDarwinRunning string
	spotifyDarwinStatus  string
	spotifyDarwinArtist  string
	spotifyDarwinTrack   string
}

func bootStrapSpotifyDarwinTest(args *spotifyArgs) *spotify {
	env := new(MockedEnvironment)
	env.On("runCommand", "osascript", []string{"-e", "application \"Spotify\" is running"}).Return(args.spotifyDarwinRunning, nil)
	env.On("runCommand", "osascript", []string{"-e", "tell application \"Spotify\" to player state as string"}).Return(args.spotifyDarwinStatus, nil)
	env.On("runCommand", "osascript", []string{"-e", "tell application \"Spotify\" to artist of current track as string"}).Return(args.spotifyDarwinArtist, nil)
	env.On("runCommand", "osascript", []string{"-e", "tell application \"Spotify\" to name of current track as string"}).Return(args.spotifyDarwinTrack, nil)
	props := &properties{}
	s := &spotify{
		env:   env,
		props: props,
	}
	return s
}

func TestSpotifyDarwinEnabledAndSpotifyNotRunning(t *testing.T) {
	args := &spotifyArgs{
		spotifyDarwinRunning: "false",
	}
	s := bootStrapSpotifyDarwinTest(args)
	assert.Equal(t, false, s.enabled())
}

func TestSpotifyDarwinEnabledAndSpotifyPlaying(t *testing.T) {
	args := &spotifyArgs{
		spotifyDarwinRunning: "true",
		spotifyDarwinStatus:  "playing",
		spotifyDarwinArtist:  "Candlemass",
		spotifyDarwinTrack:   "Spellbreaker",
	}
	s := bootStrapSpotifyDarwinTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\ue602 Candlemass - Spellbreaker", s.string())
}

func TestSpotifyDarwinEnabledAndSpotifyPaused(t *testing.T) {
	args := &spotifyArgs{
		spotifyDarwinRunning: "true",
		spotifyDarwinStatus:  "paused",
		spotifyDarwinArtist:  "Candlemass",
		spotifyDarwinTrack:   "Spellbreaker",
	}
	s := bootStrapSpotifyDarwinTest(args)
	assert.Equal(t, true, s.enabled())
	assert.Equal(t, "\uF8E3 Candlemass - Spellbreaker", s.string())
}
