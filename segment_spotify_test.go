// +build windows

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyPlayingSong(t *testing.T) {
	expected := "\ue602 Candlemass - Spellbreaker"
	env := new(MockedEnvironment)
	env.On("string", nil).Return(expected)
	s := &spotify{
		artist: "Candlemass",
		track:  "Spellbreaker",
		status: "playing",
	}
	assert.Equal(t, expected, s.string())
}

func TestSpotifyPausedSong(t *testing.T) {
	expected := "\uF8E3 Candlemass - Spellbreaker"
	env := new(MockedEnvironment)
	env.On("string", nil).Return(expected)
	s := &spotify{
		artist: "Candlemass",
		track:  "Spellbreaker",
		status: "paused",
	}
	assert.Equal(t, expected, s.string())
}

func TestSpotifyStoppedSong(t *testing.T) {
	expected := "\uf04d Candlemass - Spellbreaker"
	env := new(MockedEnvironment)
	env.On("string", nil).Return(expected)
	s := &spotify{
		artist: "Candlemass",
		track:  "Spellbreaker",
		status: "stopped",
	}
	assert.Equal(t, expected, s.string())
}
