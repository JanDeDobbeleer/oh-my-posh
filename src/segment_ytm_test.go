package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func bootstrapYTMDATest(json string, err error) *ytm {
	url := "http://127.0.0.1:9863"
	env := new(MockedEnvironment)
	env.On("HTTPRequest", url+"/query").Return([]byte(json), err)
	ytm := &ytm{
		env: env,
		props: properties{
			APIURL: url,
		},
	}
	return ytm
}

func TestYTMDAPlaying(t *testing.T) {
	json := `{ "player": { "hasSong": true, "isPaused": false }, "track": { "author": "Candlemass", "title": "Spellbreaker" } }`
	ytm := bootstrapYTMDATest(json, nil)
	err := ytm.setStatus()
	assert.NoError(t, err)
	assert.Equal(t, "playing", ytm.Status)
	assert.Equal(t, "Candlemass", ytm.Artist)
	assert.Equal(t, "Spellbreaker", ytm.Track)
}

func TestYTMDAPaused(t *testing.T) {
	json := `{ "player": { "hasSong": true, "isPaused": true }, "track": { "author": "Candlemass", "title": "Spellbreaker" } }`
	ytm := bootstrapYTMDATest(json, nil)
	err := ytm.setStatus()
	assert.NoError(t, err)
	assert.Equal(t, "paused", ytm.Status)
	assert.Equal(t, "Candlemass", ytm.Artist)
	assert.Equal(t, "Spellbreaker", ytm.Track)
}

func TestYTMDAStopped(t *testing.T) {
	json := `{ "player": { "hasSong": false }, "track": { "author": "", "title": "" } }`
	ytm := bootstrapYTMDATest(json, nil)
	err := ytm.setStatus()
	assert.NoError(t, err)
	assert.Equal(t, "stopped", ytm.Status)
	assert.Equal(t, "", ytm.Artist)
	assert.Equal(t, "", ytm.Track)
}

func TestYTMDAError(t *testing.T) {
	json := `{ "player": { "hasSong": false }, "track": { "author": "", "title": "" } }`
	ytm := bootstrapYTMDATest(json, errors.New("Oh noes"))
	enabled := ytm.enabled()
	assert.False(t, enabled)
}
