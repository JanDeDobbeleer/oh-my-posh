package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyStringPlayingSong(t *testing.T) {
	expected := "\ue602 Candlemass - Spellbreaker"
	env := new(mock.Environment)

	s := &Spotify{
		MusicPlayer: MusicPlayer{
			Artist: "Candlemass",
			Track:  "Spellbreaker",
			Status: "playing",
			Icon:   "\ue602 ",
		},
	}
	s.Init(properties.Map{}, env)

	assert.Equal(t, expected, renderTemplate(env, s.Template(), s))
}

func TestSpotifyStringStoppedSong(t *testing.T) {
	expected := "\uf04d"
	env := new(mock.Environment)

	s := &Spotify{
		MusicPlayer: MusicPlayer{
			Artist: "Candlemass",
			Track:  "Spellbreaker",
			Status: "stopped",
			Icon:   "\uf04d ",
		},
	}
	s.Init(properties.Map{}, env)

	assert.Equal(t, expected, renderTemplate(env, s.Template(), s))
}
