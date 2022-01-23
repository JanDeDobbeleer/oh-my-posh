package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyStringPlayingSong(t *testing.T) {
	expected := "\ue602 Candlemass - Spellbreaker"
	env := new(MockedEnvironment)
	s := &spotify{
		MusicPlayer: MusicPlayer{
			Artist: "Candlemass",
			Track:  "Spellbreaker",
			Status: "playing",
			Icon:   "\ue602 ",
		},
		props: properties{},
		env:   env,
	}
	assert.Equal(t, expected, renderTemplate(env, s.template(), s))
}

func TestSpotifyStringStoppedSong(t *testing.T) {
	expected := "\uf04d "
	env := new(MockedEnvironment)
	s := &spotify{
		MusicPlayer: MusicPlayer{
			Artist: "Candlemass",
			Track:  "Spellbreaker",
			Status: "stopped",
			Icon:   "\uf04d ",
		},
		props: properties{},
		env:   env,
	}
	assert.Equal(t, expected, renderTemplate(env, s.template(), s))
}
