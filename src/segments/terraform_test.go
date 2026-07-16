package segments

import (
	"os"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestTerraform(t *testing.T) {
	cases := []struct {
		Case              string
		Template          string
		Command           string
		WorkspaceName     string
		TenvEnv           string
		TenvFile          string
		TenvFileContent   string
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
		{
			Case:            "terraform version from .terraform-version file",
			ExpectedString:  "1.7.0",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			HasTfFolder:     true,
			HasTfCommand:    true,
			FetchVersion:    true,
			TenvFile:        ".terraform-version",
			TenvFileContent: "1.7.0\n",
		},
		{
			Case:            "terraform version from TFENV_TERRAFORM_VERSION env",
			ExpectedString:  "1.8.0",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			HasTfFolder:     true,
			HasTfCommand:    true,
			FetchVersion:    true,
			TenvEnv:         "1.8.0",
		},
		{
			Case:            "env takes precedence over .terraform-version file",
			ExpectedString:  "1.8.0",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			HasTfFolder:     true,
			HasTfCommand:    true,
			FetchVersion:    true,
			TenvEnv:         "1.8.0",
			TenvFile:        ".terraform-version",
			TenvFileContent: "1.7.0",
		},
		{
			Case:              "tenv file takes precedence over terraform files",
			ExpectedString:    "1.7.0",
			ExpectedEnabled:   true,
			WorkspaceName:     "default",
			Template:          "{{ .Version }}",
			HasTfVersionFiles: true,
			HasTfCommand:      true,
			FetchVersion:      true,
			TenvFile:          ".terraform-version",
			TenvFileContent:   "1.7.0",
		},
		{
			Case:            "enabled by .terraform-version alone",
			ExpectedString:  "1.7.0",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			HasTfCommand:    true,
			FetchVersion:    true,
			TenvFile:        ".terraform-version",
			TenvFileContent: "1.7.0",
		},
		{
			Case:            "tofu version from .opentofu-version file",
			ExpectedString:  "1.8.5",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			Command:         "tofu",
			HasTfFolder:     true,
			HasTfCommand:    true,
			FetchVersion:    true,
			TenvFile:        ".opentofu-version",
			TenvFileContent: "1.8.5",
		},
		{
			Case:            "tofu version from TOFUENV_TOFU_VERSION env",
			ExpectedString:  "1.9.0",
			ExpectedEnabled: true,
			WorkspaceName:   "default",
			Template:        "{{ .Version }}",
			Command:         "tofu",
			HasTfFolder:     true,
			HasTfCommand:    true,
			FetchVersion:    true,
			TenvEnv:         "1.9.0",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		cmd := tc.Command
		if cmd == "" {
			cmd = "terraform"
		}

		tenvEnvVar := "TFENV_TERRAFORM_VERSION"
		tenvFile := ".terraform-version"
		if cmd == "tofu" {
			tenvEnvVar = "TOFUENV_TOFU_VERSION"
			tenvFile = ".opentofu-version"
		}

		env.On("HasCommand", cmd).Return(tc.HasTfCommand)
		env.On("HasFolder", ".terraform").Return(tc.HasTfFolder)
		env.On("HasFiles", ".tf").Return(tc.HasTfFiles)
		env.On("HasFiles", ".tfplan").Return(tc.HasTfFiles)
		env.On("HasFiles", ".tfstate").Return(tc.HasTfFiles)
		env.On("Pwd").Return("")
		env.On("RunCommand", cmd, []string{"workspace", "show"}).Return(tc.WorkspaceName, nil)
		env.On("HasFiles", "versions.tf").Return(tc.HasTfVersionFiles)
		env.On("HasFiles", "main.tf").Return(tc.HasTfVersionFiles)
		env.On("HasFiles", "terraform.tfstate").Return(tc.HasTfStateFile)
		env.On("Getenv", tenvEnvVar).Return(tc.TenvEnv)
		env.On("HasFiles", tenvFile).Return(tc.TenvFile == tenvFile)
		env.On("FileContent", tenvFile).Return(tc.TenvFileContent)
		if tc.HasTfVersionFiles {
			content, _ := os.ReadFile("../test/versions.tf")
			env.On("FileContent", "versions.tf").Return(string(content))
		}
		if tc.HasTfStateFile {
			content, _ := os.ReadFile("../test/terraform.tfstate")
			env.On("FileContent", "terraform.tfstate").Return(string(content))
		}

		props := options.Map{
			options.FetchVersion: tc.FetchVersion,
			Command:              cmd,
		}

		tf := &Terraform{}
		tf.Init(props, env)

		template := tc.Template
		if template == "" {
			template = tf.Template()
		}
		assert.Equal(t, tc.ExpectedEnabled, tf.Enabled(), tc.Case)
		var got = renderTemplate(env, template, tf)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
	}
}
