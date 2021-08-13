package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringPlayingMedia(t *testing.T) {
	expected := "\ue602 Candlemass - Spellbreaker"
	s := &media{}
	s.info.MediaInfo.Title = "Candlemass"
	s.info.MediaInfo.Artist = "Spellbreaker"
	s.info.Playback.PlaybackState = 5
	assert.Equal(t, expected, s.string())
}

func TestStringPausedMedia(t *testing.T) {
	expected := "\uF8E3 Candlemass - Spellbreaker"
	s := &media{}
	s.info.MediaInfo.Title = "Candlemass"
	s.info.MediaInfo.Artist = "Spellbreaker"
	s.info.Playback.PlaybackState = 6
	assert.Equal(t, expected, s.string())
}

func TestStringStoppedMedia(t *testing.T) {
	expected := "\uF04D Candlemass - Spellbreaker"
	s := &media{}
	s.info.MediaInfo.Title = "Candlemass"
	s.info.MediaInfo.Artist = "Spellbreaker"
	s.info.Playback.PlaybackState = 4
	assert.Equal(t, expected, s.string())
}

func TestStringOthersMedia(t *testing.T) {
	expected := "Candlemass - Spellbreaker"
	s := &media{}
	s.other = "Candlemass - Spellbreaker"
	assert.Equal(t, expected, s.string())
}
