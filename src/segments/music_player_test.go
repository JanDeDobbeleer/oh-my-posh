package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

func TestMusicPlayerResolveIcon(t *testing.T) {
	cases := []struct {
		Case         string
		Status       string
		ExpectedIcon template.Markup
	}{
		{Case: "playing", Status: playing, ExpectedIcon: "ICON_PLAYING "},
		{Case: "paused", Status: paused, ExpectedIcon: "ICON_PAUSED "},
		{Case: "stopped", Status: stopped, ExpectedIcon: "ICON_STOPPED "},
		{Case: "ad", Status: ad, ExpectedIcon: "ICON_AD "},
	}

	for _, tc := range cases {
		player := &MusicPlayer{Status: tc.Status}

		props := options.Map{
			PlayingIcon: template.RawMarkup("ICON_PLAYING "),
			PausedIcon:  template.RawMarkup("ICON_PAUSED "),
			StoppedIcon: template.RawMarkup("ICON_STOPPED "),
			AdIcon:      template.RawMarkup("ICON_AD "),
		}

		player.resolveIcon(props)

		assert.Equal(t, tc.ExpectedIcon, player.Icon, tc.Case)
	}
}

func TestMusicPlayerEnable(t *testing.T) {
	player := &MusicPlayer{}

	assert.True(t, player.Enable(func() error { return nil }))
	assert.False(t, player.Enable(func() error { return assert.AnError }))
}
