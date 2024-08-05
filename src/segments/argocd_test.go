package segments

import (
	"fmt"
	"path"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	poshHome = "/Users/posh"
)

func TestArgocdGetConfigFromOpts(t *testing.T) {
	configFile := "/Users/posh/.config/argocd/config"
	cases := []struct {
		Case     string
		Opts     string
		Expected string
	}{
		{Case: "invalid flag in opts", Opts: "--invalid", Expected: ""},
		{Case: "no config in opts", Opts: "--grpc-web", Expected: ""},
		{
			Case:     "config in opts",
			Opts:     fmt.Sprintf("--grpc-web --config %s --plaintext", configFile),
			Expected: configFile,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", argocdOptsEnv).Return(tc.Opts)

		argocd := &Argocd{
			env:   env,
			props: properties.Map{},
		}
		config := argocd.getConfigFromOpts()
		assert.Equal(t, tc.Expected, config, tc.Case)
	}
}

func TestArgocdGetConfigPath(t *testing.T) {
	configFile := path.Join(poshHome, ".config", "argocd", "config")
	cases := []struct {
		Case          string
		Opts          string
		Expected      string
		ExpectedError string
	}{
		{Case: "without opts", Expected: configFile},
		{Case: "with opts", Opts: "--config /etc/argocd/config", Expected: "/etc/argocd/config"},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return(poshHome)
		env.On("Getenv", argocdOptsEnv).Return(tc.Opts)

		argocd := &Argocd{
			env:   env,
			props: properties.Map{},
		}
		assert.Equal(t, tc.Expected, argocd.getConfigPath())
	}
}

func TestArgocdParseConfig(t *testing.T) {
	configFile := "/Users/posh/.config/argocd/config"
	cases := []struct {
		ExpectedContext ArgocdContext
		Case            string
		Config          string
		ExpectedError   string
		Expected        bool
	}{
		{Case: "missing or empty yaml", Config: "", ExpectedError: argocdInvalidYaml},
		{
			Case:          "invalid yaml",
			ExpectedError: argocdInvalidYaml,
			Config: `
[context]
context
`,
		},
		{
			Case:          "invalid config",
			ExpectedError: argocdInvalidYaml,
			Config: `
contexts:
  - name: context1
    server: server1
    user: user1
  - name: context2
    server: server2
    userr: user2
current-context: context2
servers:
  - grpc-web: true
    server: server1
  - grpc-web: false
    server: serve2
`,
		},
		{
			Case:          "no current context found",
			ExpectedError: argocdNoCurrent,
			Config: `
contexts:
  - name: context1
    server: server1
    user: user1
  - name: context2
    server: server2
    user: user2
`,
		},
		{
			Case:     "current context found",
			Expected: true,
			Config: `
contexts:
  - name: context1
    server: server1
    user: user1
  - name: context2
    server: server2
    user: user2
current-context: context2
servers:
  - grpc-web: true
    server: server1
  - grpc-web: false
    server: serve2
users:
  - auth-token: authtoken1
    name: user1
    refresh-token: refreshtoken1
  - auth-token: authtoken2
    name: user2
    refresh-token: refreshtoken2
`,
			ExpectedContext: ArgocdContext{
				Name:   "context2",
				Server: "server2",
				User:   "user2",
			},
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("FileContent", configFile).Return(tc.Config)
		env.On("Error", testify_.Anything).Return()

		argocd := &Argocd{
			env:   env,
			props: properties.Map{},
		}
		if len(tc.ExpectedError) > 0 {
			_, err := argocd.parseConfig(configFile)
			assert.EqualError(t, err, tc.ExpectedError, tc.Case)
			continue
		}
		config, err := argocd.parseConfig(configFile)
		assert.NoErrorf(t, err, tc.Case)
		assert.Equal(t, tc.Expected, config, tc.Case)
		assert.Equal(t, tc.ExpectedContext, argocd.ArgocdContext, tc.Case)
	}
}

func TestArgocdSegment(t *testing.T) {
	configFile := path.Join(poshHome, ".config", "argocd", "config")
	cases := []struct {
		ExpectedContext ArgocdContext
		Case            string
		Opts            string
		Config          string
		Template        string
		ExpectedString  string
		ExpectedError   string
		ExpectedEnabled bool
	}{
		{
			Case: "default template",
			Opts: "",
			Config: `
contexts:
  - name: context1
    server: server1
    user: user1
  - name: context2
    server: server2
    user: user2
current-context: context2
servers:
  - grpc-web: true
    server: server1
  - grpc-web: false
    server: serve2
`,
			ExpectedString:  "context2",
			ExpectedEnabled: true,
			ExpectedContext: ArgocdContext{
				Name:   "context2",
				Server: "server2",
				User:   "user2",
			},
		},
		{
			Case: "full template",
			Opts: "",
			Config: `
contexts:
  - name: context1
    server: server1
    user: user1
  - name: context2
    server: server2
    user: user2
current-context: context2
servers:
  - grpc-web: true
    server: server1
  - grpc-web: false
    server: serve2
`,
			Template:        "{{ .Name }}:{{ .User}}@{{ .Server }}",
			ExpectedString:  "context2:user2@server2",
			ExpectedEnabled: true,
			ExpectedContext: ArgocdContext{
				Name:   "context2",
				Server: "server2",
				User:   "user2",
			},
		},
		{
			Case:            "broken config",
			Config:          `}`,
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return(poshHome)
		env.On("Getenv", argocdOptsEnv).Return(tc.Opts)
		env.On("FileContent", configFile).Return(tc.Config)
		env.On("Error", testify_.Anything).Return()
		env.On("Flags").Return(&runtime.Flags{})

		argocd := &Argocd{
			env:   env,
			props: properties.Map{},
		}

		assert.Equal(t, tc.ExpectedEnabled, argocd.Enabled(), tc.Case)

		if !tc.ExpectedEnabled {
			continue
		}

		assert.Equal(t, tc.ExpectedContext, argocd.ArgocdContext, tc.Case)
		if len(tc.Template) > 0 {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, argocd), tc.Case)
		} else {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, argocd.Template(), argocd), tc.Case)
		}
	}
}
