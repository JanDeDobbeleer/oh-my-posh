package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYTMStringPlayingSong(t *testing.T) {
	expected := "\ue602 Candlemass - Spellbreaker"
	y := &ytm{
		status: &ytmStatus{
			state:  playing,
			author: "Candlemass",
			title:  "Spellbreaker",
		},
	}
	assert.Equal(t, expected, y.string())
}

func TestYTMStringPausedSong(t *testing.T) {
	expected := "\uF8E3 Candlemass - Spellbreaker"
	y := &ytm{
		status: &ytmStatus{
			state:  paused,
			author: "Candlemass",
			title:  "Spellbreaker",
		},
	}
	assert.Equal(t, expected, y.string())
}

func TestYTMStringStoppedSong(t *testing.T) {
	expected := "\uf04d "
	y := &ytm{
		status: &ytmStatus{
			state:  stopped,
			author: "Candlemass",
			title:  "Spellbreaker",
		},
	}
	assert.Equal(t, expected, y.string())
}

func TestGetPlaying(t *testing.T) {
	json := `{ "player": { "hasSong": true, "isPaused": false }, "track": { "author": "Candlemass", "title": "Spellbreaker" } }`
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	getDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	client = &mockClient{}

	s := newYTMDAStatusService("example.com")
	status, err := s.Get()

	assert.NotNil(t, status)
	assert.Nil(t, err)

	assert.Equal(t, status.state, playing)
	assert.Equal(t, "Candlemass", status.author)
	assert.Equal(t, "Spellbreaker", status.title)
}

func TestGetPaused(t *testing.T) {
	json := `{ "player": { "hasSong": true, "isPaused": true }, "track": { "author": "Candlemass", "title": "Spellbreaker" } }`
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	getDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	client = &mockClient{}

	s := newYTMDAStatusService("example.com")
	status, err := s.Get()

	assert.NotNil(t, status)
	assert.Nil(t, err)

	assert.Equal(t, status.state, paused)
	assert.Equal(t, "Candlemass", status.author)
	assert.Equal(t, "Spellbreaker", status.title)
}

func TestGetStopped(t *testing.T) {
	json := `{ "player": { "hasSong": false }, "track": { "author": "", "title": "" } }`
	r := ioutil.NopCloser(bytes.NewReader([]byte(json)))
	getDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	client = &mockClient{}

	s := newYTMDAStatusService("example.com")
	status, err := s.Get()

	assert.NotNil(t, status)
	assert.Nil(t, err)

	assert.Equal(t, status.state, stopped)
	assert.Equal(t, "", status.author)
	assert.Equal(t, "", status.title)
}
