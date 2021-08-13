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

func TestStringMediaWithTime(t *testing.T) {
	expected := "\ue602 [3:23/5:26] Candlemass - Spellbreaker"
	s := &media{}
	s.info.MediaInfo.Title = "Candlemass"
	s.info.MediaInfo.Artist = "Spellbreaker"
	s.info.Playback.PlaybackState = 5

	s.info.Timeline.Position.TotalSeconds = 3*60 + 23
	s.info.Timeline.Position.TotalMinutes = 3
	s.info.Timeline.Position.Seconds = 23

	s.info.Timeline.EndTime.TotalSeconds = 5*60 + 26
	s.info.Timeline.EndTime.TotalMinutes = 5
	s.info.Timeline.EndTime.Seconds = 26
	assert.Equal(t, expected, s.string())
}

func TestStringOthersMedia(t *testing.T) {
	expected := "Candlemass - Spellbreaker"
	s := &media{}
	s.other = "Candlemass - Spellbreaker"
	assert.Equal(t, expected, s.string())
}
