package segments

import (
	"io/ioutil"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerraform(t *testing.T) {
	cases := []struct {
		Case            string
		Template        string
		HasTfCommand    bool
		HasTfFolder     bool
		HasTfFiles      bool
		HasTfStateFile  bool
		FetchVersion    bool
		WorkspaceName   string
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "default workspace",
			ExpectedString:  "default",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			HasTfFolder:     true,
			HasTfCommand:    true,
		},
		{
			Case:           "no command",
			ExpectedString: "",
			WorkspaceName:  "default",
			HasTfFolder:    true,
		},
		{
			Case:           "no directory, no files",
			ExpectedString: "",
			WorkspaceName:  "default",
			HasTfCommand:   true,
		},
		{
			Case:           "no files",
			ExpectedString: "",
			WorkspaceName:  "default",
			HasTfCommand:   true,
			FetchVersion:   true,
		},
		{
			Case:            "files",
			ExpectedString:  ">= 1.0.10",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			HasTfFiles:      true,
			HasTfCommand:    true,
			FetchVersion:    true,
		},
		{
			Case:            "files",
			ExpectedString:  "0.12.24",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			HasTfStateFile:  true,
			HasTfCommand:    true,
			FetchVersion:    true,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)

		env.On("HasCommand", "terraform").Return(tc.HasTfCommand)
		env.On("HasFolder", ".terraform").Return(tc.HasTfFolder)
		env.On("Pwd").Return("")
		env.On("RunCommand", "terraform", []string{"workspace", "show"}).Return(tc.WorkspaceName, nil)
		env.On("HasFiles", "versions.tf").Return(tc.HasTfFiles)
		env.On("HasFiles", "main.tf").Return(tc.HasTfFiles)
		env.On("HasFiles", "terraform.tfstate").Return(tc.HasTfStateFile)
		if tc.HasTfFiles {
			content, _ := ioutil.ReadFile("../test/versions.tf")
			env.On("FileContent", "versions.tf").Return(string(content))
		}
		if tc.HasTfStateFile {
			content, _ := ioutil.ReadFile("../test/terraform.tfstate")
			env.On("FileContent", "terraform.tfstate").Return(string(content))
		}
		tf := &Terraform{
			env: env,
			props: properties.Map{
				properties.FetchVersion: tc.FetchVersion,
			},
		}
		template := tc.Template
		if len(template) == 0 {
			template = tf.Template()
		}
		assert.Equal(t, tc.ExpectedEnabled, tf.Enabled(), tc.Case)
		var got = renderTemplate(env, template, tf)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
	}
}
