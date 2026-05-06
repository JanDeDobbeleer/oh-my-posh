//go:build windows

package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyWindowsSMTC(t *testing.T) {
	cases := []struct {
		Case            string
		SMTCOutput      string
		SMTCError       error
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:      "no session",
			SMTCError: errors.New("no Spotify SMTC session"),
		},
		{
			Case:       "stopped",
			SMTCOutput: "Candlemass\nSpellbreaker\nStopped",
		},
		{
			Case:            "playing",
			SMTCOutput:      "Candlemass\nSpellbreaker\nPlaying",
			ExpectedEnabled: true,
			ExpectedString:  "\ue602 Candlemass - Spellbreaker",
		},
		{
			Case:            "paused",
			SMTCOutput:      "Candlemass\nSpellbreaker\nPaused",
			ExpectedEnabled: true,
			ExpectedString:  "\uf04c Candlemass - Spellbreaker",
		},
		{
			Case:            "special characters in title",
			SMTCOutput:      "Demon Hunter\nCollapsing (feat. Björn \"Speed\" Strid)\nPlaying",
			ExpectedEnabled: true,
			ExpectedString:  "\ue602 Demon Hunter - Collapsing (feat. Björn \"Speed\" Strid)",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("QuerySMTC").Return(tc.SMTCOutput, tc.SMTCError)
		// SMTC is the only path for native Spotify; msedge.exe is only reached
		// when SMTC fails, which is covered by TestSpotifyWindowsPWA.
		if tc.SMTCError != nil {
			env.On("QueryWindowTitles", "msedge.exe", `^(Spotify.*)`).Return("", errors.New("not found"))
		}

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
		Case            string
		ExpectedString  string
		Title           string
		TitleError      error
		ExpectedEnabled bool
	}{
		{
			Case:            "playing",
			ExpectedString:  "\ue602 Sarah, the Illstrumentalist - Snow in Stockholm",
			ExpectedEnabled: true,
			Title:           "Spotify - Snow in Stockholm • Sarah, the Illstrumentalist",
		},
		{
			Case:            "playing 2",
			ExpectedString:  "\ue602 Main one - Bring the drama",
			ExpectedEnabled: true,
			Title:           "Spotify - Bring the drama • Main one",
		},
		{
			Case:  "stopped",
			Title: "Spotify - Web Player",
		},
		{
			Case:       "not running",
			TitleError: errors.New("not found"),
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		// SMTC returns no active Spotify session, triggering the Edge PWA fallback.
		env.On("QuerySMTC").Return("", errors.New("no Spotify SMTC session"))
		env.On("QueryWindowTitles", "msedge.exe", `^(Spotify.*)`).Return(tc.Title, tc.TitleError)

		s := &Spotify{}
		s.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, s.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}
