package segments

import (
	"path"
	"path/filepath"
	"testing"

	"oh-my-posh/environment"
	"oh-my-posh/mock"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestGcpSegment(t *testing.T) {
	standardTemplate := "{{ if .Error }}{{ .Error }}{{ else }}{{ .Project }}{{ end }}"
	allTemplate := "{{.Project}} :: {{.Region}} :: {{.Account}}"

	cases := []struct {
		Case            string
		Template        string
		ConfigPath      string
		ActiveConfig    string
		ExpectedEnabled bool
		ExpectedString  string
	}{
		{
			Case:            "all information",
			Template:        allTemplate,
			ConfigPath:      "../test/",
			ActiveConfig:    "gcptest",
			ExpectedEnabled: true,
			ExpectedString:  "test-test-test :: europe-test1 :: test@example.com",
		},
		{
			Case:         "non-existent config file",
			Template:     standardTemplate,
			ConfigPath:   "../invalid/",
			ActiveConfig: "nofile",
		},
		{
			Case:       "invalid active config file",
			Template:   standardTemplate,
			ConfigPath: "../invalid/",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Getenv", "CLOUDSDK_CONFIG").Return(tc.ConfigPath)
		fcPath, _ := filepath.Abs(path.Join(tc.ConfigPath, "active_config"))
		env.On("FileContent", fcPath).Return(tc.ActiveConfig)
		env.On("Log", environment.Error, "Gcp.Enabled()", mock2.Anything).Return()
		g := &Gcp{
			env: env,
		}
		assert.Equal(t, tc.ExpectedEnabled, g.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, g), tc.Case)
		}
	}
}
