//go:build !darwin && !windows

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyWsl(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		ExecOutput      string
	}{
		{
			Case:            "Spotify not running",
			ExpectedString:  " - ",
			ExpectedEnabled: false,
			ExecOutput:      "INFO: No tasks are running which match the specified criteria.\n",
		},
		{
			Case:            "Spotify stopped/paused",
			ExpectedString:  " - ",
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
			Case:            "tasklist.exe not in path",
			ExpectedString:  " - ",
			ExpectedEnabled: false,
			ExecOutput:      ""},
	}
	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("isWsl").Return(true)
		env.On("runCommand", "tasklist.exe", []string{"/V", "/FI", "Imagename eq Spotify.exe", "/FO", "CSV", "/NH"}).Return(tc.ExecOutput, nil)
		s := &spotify{
			env:   env,
			props: properties{},
		}
		assert.Equal(t, tc.ExpectedEnabled, s.enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, s.string(), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
