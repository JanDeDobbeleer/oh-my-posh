//go:build !darwin && !windows

package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyWsl(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExecOutput      string
		ExpectedEnabled bool
	}{
		{
			Case:            "Spotify not running",
			ExpectedString:  "-",
			ExpectedEnabled: false,
			ExecOutput:      "INFO: No tasks are running which match the specified criteria.\n",
		},
		{
			Case:            "Spotify stopped/paused",
			ExpectedString:  "-",
			ExpectedEnabled: false,
			ExecOutput: `"Spotify.exe","21824","Console","1","124,928 K","Running","PC\user","0:09:44","Spotify Premium"
"Spotify.exe","21876","Console","1","25,520 K","Running","PC\user","0:00:00","N/A"
"Spotify.exe","21988","Console","1","60,840 K","Not Responding","PC\user","0:04:56","AngleHiddenWindow"
"Spotify.exe","22052","Console","1","29,040 K","Unknown","PC\user","0:00:00","N/A"
"Spotify.exe","22072","Console","1","43,960 K","Unknown","PC\user","0:01:50","N/A"
"Spotify.exe","10404","Console","1","256,924 K","Unknown","PC\user","0:10:49","N/A"`,
		},
		{
			Case:            "Spotify playing",
			ExpectedString:  "\ue602 Candlemass - Spellbreaker",
			ExpectedEnabled: true,
			ExecOutput: `"Spotify.exe","21824","Console","1","124,928 K","Running","PC\user","0:09:44","Candlemass - Spellbreaker"
"Spotify.exe","21876","Console","1","25,520 K","Running","PC\user","0:00:00","N/A"
"Spotify.exe","21988","Console","1","60,840 K","Not Responding","PC\user","0:04:56","AngleHiddenWindow"
"Spotify.exe","22052","Console","1","29,040 K","Unknown","PC\user","0:00:00","N/A"
"Spotify.exe","22072","Console","1","43,960 K","Unknown","PC\user","0:01:50","N/A"
"Spotify.exe","10404","Console","1","256,924 K","Unknown","PC\user","0:10:49","N/A"`,
		},
		{
			Case:            "Spotify playing",
			ExpectedString:  "\ue602 Grabbitz - Another Form Of \"Goodbye\"",
			ExpectedEnabled: true,
			ExecOutput: `"Spotify.exe","13748","Console","1","303.744 K","Running","GARMIN\elderbroekowe","0:03:58","Grabbitz - Another Form Of "Goodbye""
"Spotify.exe","4208","Console","1","31.544 K","Running","GARMIN\elderbroekowe","0:00:00","N/A"
"Spotify.exe","14528","Console","1","184.020 K","Running","GARMIN\elderbroekowe","0:02:54","N/A"
"Spotify.exe","14488","Console","1","53.828 K","Unknown","GARMIN\elderbroekowe","0:00:08","N/A"
"Spotify.exe","14800","Console","1","29.576 K","Unknown","GARMIN\elderbroekowe","0:00:00","N/A"
"Spotify.exe","19836","Console","1","237.360 K","Unknown","GARMIN\elderbroekowe","0:07:46","N/A"`,
		},
		{
			Case:            "tasklist.exe not in path",
			ExpectedString:  "-",
			ExpectedEnabled: false,
			ExecOutput:      "",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("IsWsl").Return(true)
		env.On("RunCommand", "tasklist.exe", []string{"/V", "/FI", "Imagename eq Spotify.exe", "/FO", "CSV", "/NH"}).Return(tc.ExecOutput, nil)
		s := &Spotify{
			env:   env,
			props: properties.Map{},
		}
		assert.Equal(t, tc.ExpectedEnabled, s.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
