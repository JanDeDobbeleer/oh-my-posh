package segments

import (
	"os"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestTerraform(t *testing.T) {
	cases := []struct {
		Case              string
		Template          string
		WorkspaceName     string
		ExpectedString    string
		HasTfCommand      bool
		HasTfFolder       bool
		HasTfFiles        bool
		HasTfVersionFiles bool
		HasTfStateFile    bool
		FetchVersion      bool
		ExpectedEnabled   bool
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
			Case:              "files",
			ExpectedString:    ">= 1.0.10",
			ExpectedEnabled:   true,
			WorkspaceName:     "default",
			Template:          "{{ .Version }}",
			HasTfVersionFiles: true,
			HasTfCommand:      true,
			FetchVersion:      true,
		},
		{
			Case:            "version files",
			ExpectedString:  "0.12.24",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			HasTfStateFile:  true,
			HasTfCommand:    true,
			FetchVersion:    true,
		},
		{
			Case:            "context files",
			ExpectedString:  "default",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			HasTfFiles:      true,
			HasTfCommand:    true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		env.On("HasCommand", "terraform").Return(tc.HasTfCommand)
		env.On("HasFolder", ".terraform").Return(tc.HasTfFolder)
		env.On("HasFiles", ".tf").Return(tc.HasTfFiles)
		env.On("HasFiles", ".tfplan").Return(tc.HasTfFiles)
		env.On("HasFiles", ".tfstate").Return(tc.HasTfFiles)
		env.On("Pwd").Return("")
		env.On("RunCommand", "terraform", []string{"workspace", "show"}).Return(tc.WorkspaceName, nil)
		env.On("HasFiles", "versions.tf").Return(tc.HasTfVersionFiles)
		env.On("HasFiles", "main.tf").Return(tc.HasTfVersionFiles)
		env.On("HasFiles", "terraform.tfstate").Return(tc.HasTfStateFile)
		if tc.HasTfVersionFiles {
			content, _ := os.ReadFile("../test/versions.tf")
			env.On("FileContent", "versions.tf").Return(string(content))
		}
		if tc.HasTfStateFile {
			content, _ := os.ReadFile("../test/terraform.tfstate")
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
