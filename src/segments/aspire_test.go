package segments

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/stretchr/testify/assert"
)

func TestAspireSegment(t *testing.T) {
	cases := []struct {
		Case              string
		HasAspire         bool
		CLIOutput         string
		CLIError          error
		FallbackFile      string
		PSResourcesOutput string
		PSResourcesError  error
		PSOutput          string
		PSError           error
		ExpectedName      string
		ExpectedLang      string
		ExpectedVersion   string
		ExpectedEnabled   bool
		ExpectedRunning   bool
		FetchRunning      bool
		HasProps          bool
		PropsContent      string
	}{
		{
			Case:            "no apphost file",
			HasAspire:       false,
			ExpectedEnabled: false,
		},
		{
			Case:              "apphost resolved by aspire cli",
			HasAspire:         true,
			CLIOutput:         "{\"selected_project_file\":\"test/MyApp.AppHost/apphost.cs\",\"all_project_file_candidates\":[\"test/MyApp.AppHost/apphost.cs\"]}",
			PSResourcesOutput: "[]",
			ExpectedEnabled:   true,
			ExpectedName:      "MyApp.AppHost",
			ExpectedLang:      "cs",
			FetchRunning:      true,
		},
		{
			Case:              "apphost running from aspire ps resources",
			HasAspire:         true,
			CLIOutput:         "{\"selected_project_file\":\"test/MyApp.AppHost/apphost.cs\",\"all_project_file_candidates\":[\"test/MyApp.AppHost/apphost.cs\"]}",
			PSResourcesOutput: "[{\"appHostPath\":\"test/MyApp.AppHost/apphost.cs\",\"appHostPid\":123}]",
			ExpectedEnabled:   true,
			ExpectedRunning:   true,
			ExpectedName:      "MyApp.AppHost",
			ExpectedLang:      "cs",
			FetchRunning:      true,
		},
		{
			Case:             "aspire ps falls back without resources",
			HasAspire:        true,
			CLIOutput:        "{\"selected_project_file\":\"test/MyApp.AppHost/apphost.cs\",\"all_project_file_candidates\":[\"test/MyApp.AppHost/apphost.cs\"]}",
			PSResourcesError: errors.New("unknown flag: --resources"),
			PSOutput:         "[{\"appHostPath\":\"test/MyApp.AppHost/apphost.cs\",\"appHostPid\":123}]",
			ExpectedEnabled:  true,
			ExpectedRunning:  true,
			ExpectedName:     "MyApp.AppHost",
			ExpectedLang:     "cs",
			ExpectedVersion:  "9.0.0",
			FetchRunning:     true,
			HasProps:         true,
			PropsContent:     `<PackageVersion Include="Aspire.Hosting.AppHost" Version="9.0.0" />`,
		},
		{
			Case:            "running lookup can be disabled",
			HasAspire:       true,
			CLIOutput:       "{\"selected_project_file\":\"test/MyApp.AppHost/apphost.cs\",\"all_project_file_candidates\":[\"test/MyApp.AppHost/apphost.cs\"]}",
			ExpectedEnabled: true,
			ExpectedName:    "MyApp.AppHost",
			ExpectedLang:    "cs",
			FetchRunning:    false,
		},
		{
			Case:            "fallback to parent apphost.ts when aspire cli unavailable",
			HasAspire:       false,
			FallbackFile:    "apphost.ts",
			ExpectedEnabled: true,
			ExpectedName:    "MyApp.AppHost",
			ExpectedLang:    "ts",
			FetchRunning:    true,
		},
	}

	pwd := filepath.Join("test", "MyApp.AppHost", "nested")
	appHostCS := filepath.Join("test", "MyApp.AppHost", "apphost.cs")
	appHostTS := filepath.Join("test", "MyApp.AppHost", "apphost.ts")

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Flags").Return(&runtime.Flags{})
		env.On("Pwd").Return(pwd)
		env.On("HasCommand", aspireCommand).Return(tc.HasAspire)

		if tc.HasAspire {
			env.On("RunCommand", aspireCommand, []string{"extension", "get-apphosts"}).Return(tc.CLIOutput, tc.CLIError)
			if tc.ExpectedEnabled && tc.FetchRunning {
				env.On("RunCommand", aspireCommand, []string{"ps", "--format", "json", "--resources"}).Return(tc.PSResourcesOutput, tc.PSResourcesError)
				if tc.PSResourcesError != nil {
					env.On("RunCommand", aspireCommand, []string{"ps", "--format", "json"}).Return(tc.PSOutput, tc.PSError)
				}
			}
		} else {
			switch tc.FallbackFile {
			case "apphost.cs":
				env.On("HasParentFilePath", "apphost.cs", false).Return(&runtime.FileInfo{Path: appHostCS, ParentFolder: filepath.Dir(appHostCS)}, nil)
			case "apphost.ts":
				env.On("HasParentFilePath", "apphost.cs", false).Return(&runtime.FileInfo{}, errors.New("not found"))
				env.On("HasParentFilePath", "apphost.ts", false).Return(&runtime.FileInfo{Path: appHostTS, ParentFolder: filepath.Dir(appHostTS)}, nil)
			default:
				env.On("HasParentFilePath", "apphost.cs", false).Return(&runtime.FileInfo{}, errors.New("not found"))
				env.On("HasParentFilePath", "apphost.ts", false).Return(&runtime.FileInfo{}, errors.New("not found"))
			}
		}

		if tc.HasProps {
			fileInfo := &runtime.FileInfo{
				Path:         filepath.Join("test", "Directory.Packages.props"),
				ParentFolder: "test",
			}
			env.On("HasParentFilePath", "Directory.Packages.props", false).Return(fileInfo, nil)
			env.On("FileContent", filepath.Join("test", "Directory.Packages.props")).Return(tc.PropsContent)
		} else {
			env.On("HasParentFilePath", "Directory.Packages.props", false).Return(&runtime.FileInfo{}, errors.New("not found"))
		}

		aspire := &Aspire{}
		props := options.Map{FetchRunning: tc.FetchRunning}
		aspire.Init(props, env)

		assert.Equal(t, tc.ExpectedEnabled, aspire.Enabled(), tc.Case)

		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedName, aspire.Name, tc.Case)
			assert.Equal(t, tc.ExpectedLang, aspire.Lang, tc.Case)
			assert.Equal(t, tc.ExpectedVersion, aspire.Version, tc.Case)
			assert.Equal(t, tc.ExpectedRunning, aspire.Running, tc.Case)
			assert.Equal(t, tc.ExpectedName, renderTemplate(env, "{{ .Name }}", aspire), tc.Case)
			assert.Equal(t, tc.ExpectedRunning, renderTemplate(env, "{{ .Running }}", aspire) == "true", tc.Case)
		}
	}
}
