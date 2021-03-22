package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYTMStringPlayingSong(t *testing.T) {
	expected := "\ue602 Candlemass - Spellbreaker"
	y := &ytm{
		artist: "Candlemass",
		track:  "Spellbreaker",
		status: playing,
	}
	assert.Equal(t, expected, y.string())
}

func TestYTMStringPausedSong(t *testing.T) {
	expected := "\uF8E3 Candlemass - Spellbreaker"
	y := &ytm{
		artist: "Candlemass",
		track:  "Spellbreaker",
		status: paused,
	}
	assert.Equal(t, expected, y.string())
}

func TestYTMStringStoppedSong(t *testing.T) {
	expected := "\uf04d "
	y := &ytm{
		artist: "Candlemass",
		track:  "Spellbreaker",
		status: stopped,
	}
	assert.Equal(t, expected, y.string())
}

func bootstrapYTMDATest(json string, err error) *ytm {
	url := "http://127.0.0.1:9863"
	env := new(MockedEnvironment)
	env.On("doGet", url+"/query").Return([]byte(json), err)
	props := &properties{
		values: map[Property]interface{}{
			APIURL: url,
		},
	}
	ytm := &ytm{
		env:   env,
		props: props,
	}
	return ytm
}

func TestYTMDAPlaying(t *testing.T) {
	json := `{ "player": { "hasSong": true, "isPaused": false }, "track": { "author": "Candlemass", "title": "Spellbreaker" } }`
	ytm := bootstrapYTMDATest(json, nil)
	err := ytm.setStatus()
	assert.NoError(t, err)
	assert.Equal(t, playing, ytm.status)
	assert.Equal(t, "Candlemass", ytm.artist)
	assert.Equal(t, "Spellbreaker", ytm.track)
}

func TestYTMDAPaused(t *testing.T) {
	json := `{ "player": { "hasSong": true, "isPaused": true }, "track": { "author": "Candlemass", "title": "Spellbreaker" } }`
	ytm := bootstrapYTMDATest(json, nil)
	err := ytm.setStatus()
	assert.NoError(t, err)
	assert.Equal(t, paused, ytm.status)
	assert.Equal(t, "Candlemass", ytm.artist)
	assert.Equal(t, "Spellbreaker", ytm.track)
}

func TestYTMDAStopped(t *testing.T) {
	json := `{ "player": { "hasSong": false }, "track": { "author": "", "title": "" } }`
	ytm := bootstrapYTMDATest(json, nil)
	err := ytm.setStatus()
	assert.NoError(t, err)
	assert.Equal(t, stopped, ytm.status)
	assert.Equal(t, "", ytm.artist)
	assert.Equal(t, "", ytm.track)
}

func TestYTMDAError(t *testing.T) {
	json := `{ "player": { "hasSong": false }, "track": { "author": "", "title": "" } }`
	ytm := bootstrapYTMDATest(json, errors.New("Oh noes"))
	enabled := ytm.enabled()
	assert.False(t, enabled)
}
