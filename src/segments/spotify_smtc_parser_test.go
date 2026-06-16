//go:build linux && !darwin

package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

// TestParseSMTCLineAndApply exercises the WSL PowerShell path: a
// "<status>|<title>|<artist>|<album>|<trackNumber>" line is parsed into a
// *runtime.MediaInfo and applied to the segment. The expected strings carry
// the default Nerd Font icons resolveIcon assigns (playing , paused
// , ad ).
func TestParseSMTCLineAndApply(t *testing.T) {
	cases := []struct {
		Case            string
		Output          string
		ExpectedString  string
		ExpectedStatus  string
		ExpectedEnabled bool
		ExpectedParseOK bool
	}{
		{
			Case:            "playing",
			Output:          "playing|Spellbreaker|Candlemass|Nightfall|3",
			ExpectedString:  " Candlemass - Spellbreaker",
			ExpectedStatus:  playing,
			ExpectedEnabled: true,
			ExpectedParseOK: true,
		},
		{
			Case:            "paused",
			Output:          "paused|Spellbreaker|Candlemass|Nightfall|3",
			ExpectedString:  " Candlemass - Spellbreaker",
			ExpectedStatus:  paused,
			ExpectedEnabled: true,
			ExpectedParseOK: true,
		},
		{
			Case:            "ad (empty album, track number 0)",
			Output:          "playing|君のこころが観たいもの　プライムビデオ|プライムビデオ||0",
			ExpectedString:  " プライムビデオ - 君のこころが観たいもの　プライムビデオ",
			ExpectedStatus:  ad,
			ExpectedEnabled: true,
			ExpectedParseOK: true,
		},
		{
			Case:            "track with parentheses and quotes",
			Output:          `playing|Collapsing (feat. Björn "Speed" Strid)|Demon Hunter|The World Is a Thorn|9`,
			ExpectedString:  " Demon Hunter - Collapsing (feat. Björn \"Speed\" Strid)",
			ExpectedStatus:  playing,
			ExpectedEnabled: true,
			ExpectedParseOK: true,
		},
		{
			Case:            "stopped",
			Output:          "stopped||||0",
			ExpectedEnabled: false,
			ExpectedParseOK: true,
		},
		{
			Case:            "closed (no session)",
			Output:          "closed||||0",
			ExpectedEnabled: false,
			ExpectedParseOK: true,
		},
		{
			Case:            "malformed (too few fields)",
			Output:          "playing|only|three|fields",
			ExpectedEnabled: false,
			ExpectedParseOK: false,
		},
		{
			Case:            "malformed (no separators)",
			Output:          "garbage",
			ExpectedEnabled: false,
			ExpectedParseOK: false,
		},
		{
			Case:            "empty",
			Output:          "",
			ExpectedEnabled: false,
			ExpectedParseOK: false,
		},
	}

	for _, tc := range cases {
		info, ok := parseSMTCLine(tc.Output)
		assert.Equal(t, tc.ExpectedParseOK, ok, tc.Case)
		if !ok {
			continue
		}

		env := new(mock.Environment)
		s := &Spotify{}
		s.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, s.applyMediaInfo(info), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedStatus, s.Status, tc.Case)
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}
