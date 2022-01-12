package main

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert"
)

func TestNbgv(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedEnabled bool
		ExpectedString  string
		Response        string
		HasNbgv         bool
		SegmentTemplate string
		Error           error
	}{
		{Case: "nbgv not installed"},
		{Case: "nbgv installed, no version file", HasNbgv: true, Response: "{ \"VersionFileFound\": false }"},
		{Case: "nbgv installed with version file", ExpectedEnabled: true, HasNbgv: true, Response: "{ \"VersionFileFound\": true }"},
		{
			Case:            "invalid template",
			ExpectedEnabled: true,
			ExpectedString:  "invalid template text",
			HasNbgv:         true,
			Response:        "{ \"VersionFileFound\": true }",
			SegmentTemplate: "{{ err }}",
		},
		{
			Case:    "command error",
			HasNbgv: true,
			Error:   errors.New("oh noes"),
		},
		{
			Case:     "invalid json",
			HasNbgv:  true,
			Response: "><<<>>>",
		},
		{
			Case:            "Version",
			ExpectedEnabled: true,
			ExpectedString:  "bump",
			HasNbgv:         true,
			Response:        "{ \"VersionFileFound\": true, \"Version\": \"bump\" }",
			SegmentTemplate: "{{ .Version }}",
		},
		{
			Case:            "AssemblyVersion",
			ExpectedEnabled: true,
			ExpectedString:  "bump",
			HasNbgv:         true,
			Response:        "{ \"VersionFileFound\": true, \"AssemblyVersion\": \"bump\" }",
			SegmentTemplate: "{{ .AssemblyVersion }}",
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("hasCommand", "nbgv").Return(tc.HasNbgv)
		env.On("runCommand", "nbgv", []string{"get-version", "--format=json"}).Return(tc.Response, tc.Error)
		env.onTemplate()
		nbgv := &nbgv{
			env: env,
			props: properties{
				SegmentTemplate: tc.SegmentTemplate,
			},
		}
		enabled := nbgv.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if enabled {
			assert.Equal(t, tc.ExpectedString, nbgv.string(), tc.Case)
		}
	}
}
