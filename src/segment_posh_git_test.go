package main

import (
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
		env := new(MockedEnvironment)
		env.On("Getenv", poshGitEnv).Return(tc.PoshGitPrompt)
		p := &poshgit{
			env:   env,
			props: &properties{},
		}
		assert.Equal(t, tc.Enabled, p.enabled(), tc.Case)
		if tc.Enabled {
			assert.Equal(t, tc.Expected, renderTemplate(env, p.template(), p), tc.Case)
		}
	}
}
