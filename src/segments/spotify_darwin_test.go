//go:build darwin

package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyDarwinEnabledAndSpotifyPlaying(t *testing.T) {
	cases := []struct {
		Error       error
		BatchedCase string
		Expected    string
		Enabled     bool
	}{
		{BatchedCase: "false|||||0", Expected: "", Enabled: false},
		{BatchedCase: "false||", Expected: "", Error: errors.New("oops"), Enabled: false},
		{BatchedCase: "true|playing|Candlemass|Spellbreaker|Nightfall|3", Expected: "\ue602 Candlemass - Spellbreaker", Enabled: true},
		{BatchedCase: "true|paused|Candlemass|Spellbreaker|Nightfall|3", Expected: "\uf04c Candlemass - Spellbreaker", Enabled: true},
		{BatchedCase: "true|playing||アコム【公式】||0", Expected: "\ueebb  - アコム【公式】", Enabled: true},
		{BatchedCase: "true|stopped||||0", Expected: "", Enabled: false},
	}
	batchedCommand := `
	if application "Spotify" is running then
		tell application "Spotify"
			set playerState to player state as string
			set artistName to ""
			set trackName to ""
			set albumName to ""
			set trackNumber to 0
			if playerState is not "stopped" then
				try
					set artistName to artist of current track as string
				end try
				try
					set trackName to name of current track as string
				end try
				try
					set albumName to album of current track as string
				end try
				try
					set trackNumber to track number of current track as integer
				end try
			end if
			return "true|" & playerState & "|" & artistName & "|" & trackName & "|" & albumName & "|" & trackNumber
		end tell
	else
		return "false|||||0"
	end if
	`
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("RunCommand", "osascript", []string{"-e", batchedCommand}).Return(tc.BatchedCase, tc.Error)

		s := &Spotify{}
		s.Init(options.Map{}, env)

		assert.Equal(t, tc.Enabled, s.Enabled())
		assert.Equal(t, tc.Expected, renderTemplate(env, s.Template(), s))
	}
}
