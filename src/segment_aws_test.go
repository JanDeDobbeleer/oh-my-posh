package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAWSSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		Profile         string
		Vault           string
		Region          string
		DefaultRegion   string
		ConfigFile      string
		Template        string
	}{
		{Case: "disabled", ExpectedString: "", ExpectedEnabled: false},
		{Case: "enabled no region", ExpectedString: "company", ExpectedEnabled: true, Profile: "company"},
		{Case: "enabled with region", ExpectedString: "company@eu-west", ExpectedEnabled: true, Profile: "company", Region: "eu-west"},
		{
			Case:            "template: enabled no region",
			ExpectedString:  "profile: company",
			ExpectedEnabled: true,
			Profile:         "company",
			Template:        "profile: {{.Profile}}{{if .Region}} in {{.Region}}{{end}}",
		},
		{
			Case:            "template: enabled with region",
			ExpectedString:  "profile: company in eu-west",
			ExpectedEnabled: true,
			Profile:         "company",
			Region:          "eu-west",
			Template:        "profile: {{.Profile}}{{if .Region}} in {{.Region}}{{end}}",
		},
		{Case: "template: invalid", ExpectedString: invalidTemplate, ExpectedEnabled: true, Profile: "c", Template: "{{ .Burp"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getenv", "AWS_VAULT").Return(tc.Vault)
		env.On("getenv", "AWS_PROFILE").Return(tc.Profile)
		env.On("getenv", "AWS_REGION").Return(tc.Region)
		env.On("getenv", "AWS_DEFAULT_REGION").Return(tc.DefaultRegion)
		env.On("getenv", "AWS_CONFIG_FILE").Return(tc.ConfigFile)
		env.On("getFileContent", "/usr/home/.aws/config").Return("")
		env.On("homeDir", nil).Return("/usr/home")
		props := &properties{
			values: map[Property]interface{}{},
		}
		if tc.Template != "" {
			props.values[SegmentTemplate] = tc.Template
		}

		aws := &aws{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.ExpectedEnabled, aws.enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, aws.string(), tc.Case)
	}
}
