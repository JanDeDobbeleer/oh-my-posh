//go:build windows

package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyWindowsSMTC(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		Output          string
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "playing",
			Output:          "playing|Spellbreaker|Candlemass|Nightfall|3",
			ExpectedString:  " Candlemass - Spellbreaker",
			ExpectedEnabled: true,
		},
		{
			Case:            "paused",
			Output:          "paused|Spellbreaker|Candlemass|Nightfall|3",
			ExpectedString:  " Candlemass - Spellbreaker",
			ExpectedEnabled: true,
		},
		{
			Case:            "ad (empty album, track number 0)",
			Output:          "playing|君のこころが観たいもの　プライムビデオ|プライムビデオ||0",
			ExpectedString:  " プライムビデオ - 君のこころが観たいもの　プライムビデオ",
			ExpectedEnabled: true,
		},
		{
			Case:            "stopped",
			Output:          "stopped||||0",
			ExpectedEnabled: false,
		},
		{
			Case:            "closed (no Spotify session)",
			Output:          "closed||||0",
			ExpectedEnabled: false,
		},
		{
			Case:            "track with parentheses and quotes",
			Output:          `playing|Collapsing (feat. Björn "Speed" Strid)|Demon Hunter|The World Is a Thorn|9`,
			ExpectedString:  " Demon Hunter - Collapsing (feat. Björn \"Speed\" Strid)",
			ExpectedEnabled: true,
		},
		{
			Case:            "powershell error",
			Error:           errors.New("oops"),
			ExpectedEnabled: false,
		},
		{
			Case:            "malformed output (too few fields)",
			Output:          "playing|only|three|fields",
			ExpectedEnabled: false,
		},
		{
			Case:            "malformed output (no separators)",
			Output:          "garbage",
			ExpectedEnabled: false,
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("RunCommand", "powershell.exe", []string{"-NoProfile", "-NonInteractive", "-Command", spotifySMTCScript}).Return(tc.Output, tc.Error)
		// Edge PWA fallback never matches in this group of tests.
		env.On("QueryWindowTitles", "msedge.exe", `^(Spotify.*)`).Return("", &runtime.NotImplemented{})

		s := &Spotify{}
		s.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, s.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}

func TestSpotifyWindowsPWA(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		ExpectedString  string
		Title           string
		ExpectedEnabled bool
	}{
		{
			Case:            "playing",
			ExpectedString:  " Sarah, the Illstrumentalist - Snow in Stockholm",
			ExpectedEnabled: true,
			Title:           "Spotify - Snow in Stockholm • Sarah, the Illstrumentalist",
		},
		{
			Case:            "stopped",
			ExpectedEnabled: false,
			Title:           "Spotify - Web Player",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		// SMTC unavailable — falls back to Edge PWA window title parsing.
		env.On("RunCommand", "powershell.exe", []string{"-NoProfile", "-NonInteractive", "-Command", spotifySMTCScript}).Return("", errors.New("smtc unavailable"))
		env.On("QueryWindowTitles", "msedge.exe", `^(Spotify.*)`).Return(tc.Title, tc.Error)

		s := &Spotify{}
		s.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, s.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}
