package main

import (
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPoshGitSegment(t *testing.T) {
	cases := []struct {
		Case          string
		PoshGitPrompt string
		Expected      string
		Enabled       bool
	}{
		{Case: "regular prompt", PoshGitPrompt: "my prompt", Expected: "my prompt", Enabled: true},
		{Case: "prompt with spaces", PoshGitPrompt: "   my prompt", Expected: "my prompt", Enabled: true},
		{Case: "no prompt", PoshGitPrompt: "", Enabled: false},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Getenv", poshGitEnv).Return(tc.PoshGitPrompt)
		p := &PoshGit{
			env:   env,
			props: &properties.Map{},
		}
		assert.Equal(t, tc.Enabled, p.Enabled(), tc.Case)
		if tc.Enabled {
			assert.Equal(t, tc.Expected, renderTemplate(env, p.Template(), p), tc.Case)
		}
	}
}
