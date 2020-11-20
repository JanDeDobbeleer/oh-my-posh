package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockYTMStatusService struct {
	HasSong  bool
	IsPaused bool
	Author   string
	Title    string
}

func (s *mockYTMStatusService) Get() (*ytmStatus, error) {
	state := playing
	if !s.HasSong {
		state = stopped
	} else if s.IsPaused {
		state = paused
	}

	status := &ytmStatus{
		state:  state,
		author: "Candlemass",
		title:  "Spellbreaker",
	}

	return status, nil
}

func getMockYTMStatusService(hasSong, isPaused bool, author, title string) *mockYTMStatusService {
	return &mockYTMStatusService{
		HasSong:  hasSong,
		IsPaused: isPaused,
		Author:   author,
		Title:    title,
	}
}

func TestYTMStringPlayingSong(t *testing.T) {
	expected := "\ue602 Candlemass - Spellbreaker"
	y := &ytm{
		service: getMockYTMStatusService(true, false, "Candlemass", "Spellbreaker"),
	}
	assert.Equal(t, expected, y.string())
}

func TestYTMStringPausedSong(t *testing.T) {
	expected := "\uF8E3 Candlemass - Spellbreaker"
	y := &ytm{
		service: getMockYTMStatusService(true, true, "Candlemass", "Spellbreaker"),
	}
	assert.Equal(t, expected, y.string())
}

func TestYTMStringStoppedSong(t *testing.T) {
	expected := "\uf04d "
	y := &ytm{
		service: getMockYTMStatusService(false, false, "", ""),
	}
	assert.Equal(t, expected, y.string())
}
