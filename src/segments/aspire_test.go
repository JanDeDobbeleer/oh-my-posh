package segments

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/stretchr/testify/assert"
)

func TestAspireSegment(t *testing.T) {
	cases := []struct {
		CLIError        error
		PSError         error
		FallbackFile    string
		ExpectedLang    string
		CLIOutput       string
		PSOutput        string
		PropsContent    string
		ExpectedName    string
		Case            string
		ExpectedVersion string
		ExpectedEnabled bool
		ExpectedRunning bool
		FetchRunning    bool
		HasProps        bool
		HasAspire       bool
	}{
		{
			Case:            "no apphost file",
			HasAspire:       false,
			ExpectedEnabled: false,
		},
		{
			Case:            "apphost resolved by aspire cli",
			HasAspire:       true,
			CLIOutput:       `{"selected_project_file":"test/MyApp.AppHost/MyApp.AppHost.csproj","all_project_file_candidates":["test/MyApp.AppHost/MyApp.AppHost.csproj"]}`,
			PSOutput:        "[]",
			ExpectedEnabled: true,
			ExpectedName:    "MyApp.AppHost",
			ExpectedLang:    "cs",
			FetchRunning:    true,
		},
		{
			Case:            "apphost running from aspire ps",
			HasAspire:       true,
			CLIOutput:       `{"selected_project_file":"test/MyApp.AppHost/MyApp.AppHost.csproj","all_project_file_candidates":["test/MyApp.AppHost/MyApp.AppHost.csproj"]}`,
			PSOutput:        `[{"appHostPath":"test/MyApp.AppHost/MyApp.AppHost.csproj","appHostPid":123}]`,
			ExpectedEnabled: true,
			ExpectedRunning: true,
			ExpectedName:    "MyApp.AppHost",
			ExpectedLang:    "cs",
			FetchRunning:    true,
		},
		{
			Case:            "aspire ps with version",
			HasAspire:       true,
			CLIOutput:       `{"selected_project_file":"test/MyApp.AppHost/MyApp.AppHost.csproj","all_project_file_candidates":["test/MyApp.AppHost/MyApp.AppHost.csproj"]}`,
			PSOutput:        `[{"appHostPath":"test/MyApp.AppHost/MyApp.AppHost.csproj","appHostPid":123}]`,
			ExpectedEnabled: true,
			ExpectedRunning: true,
			ExpectedName:    "MyApp.AppHost",
			ExpectedLang:    "cs",
			ExpectedVersion: "9.0.0",
			FetchRunning:    true,
			HasProps:        true,
			PropsContent:    `<PackageVersion Include="Aspire.Hosting.AppHost" Version="9.0.0" />`,
		},
		{
			Case:            "running lookup can be disabled",
			HasAspire:       true,
			CLIOutput:       `{"selected_project_file":"test/MyApp.AppHost/MyApp.AppHost.csproj","all_project_file_candidates":["test/MyApp.AppHost/MyApp.AppHost.csproj"]}`,
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
		{
			Case:            "fallback to parent *.AppHost.csproj when aspire cli unavailable",
			HasAspire:       false,
			FallbackFile:    "MyApp.AppHost.csproj",
			ExpectedEnabled: true,
			ExpectedName:    "MyApp.AppHost",
			ExpectedLang:    "cs",
			FetchRunning:    false,
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
				env.On("RunCommand", aspireCommand, []string{"ps", "--format", "json"}).Return(tc.PSOutput, tc.PSError)
			}
		} else {
			switch tc.FallbackFile {
			case "apphost.cs":
				env.On("HasParentFilePath", "apphost.cs", false).Return(&runtime.FileInfo{Path: appHostCS, ParentFolder: filepath.Dir(appHostCS)}, nil)
			case "apphost.ts":
				env.On("HasParentFilePath", "apphost.cs", false).Return(&runtime.FileInfo{}, errors.New("not found"))
				env.On("HasParentFilePath", "apphost.ts", false).Return(&runtime.FileInfo{Path: appHostTS, ParentFolder: filepath.Dir(appHostTS)}, nil)
			case "MyApp.AppHost.csproj":
				env.On("HasParentFilePath", "apphost.cs", false).Return(&runtime.FileInfo{}, errors.New("not found"))
				env.On("HasParentFilePath", "apphost.ts", false).Return(&runtime.FileInfo{}, errors.New("not found"))
				env.On("LsDir", pwd).Return([]fs.DirEntry{})
				env.On("LsDir", filepath.Join("test", "MyApp.AppHost")).Return([]fs.DirEntry{
					&MockDirEntry{name: "MyApp.AppHost.csproj", isDir: false},
				})
			default:
				env.On("HasParentFilePath", "apphost.cs", false).Return(&runtime.FileInfo{}, errors.New("not found"))
				env.On("HasParentFilePath", "apphost.ts", false).Return(&runtime.FileInfo{}, errors.New("not found"))
				env.On("LsDir", pwd).Return([]fs.DirEntry{})
				env.On("LsDir", filepath.Join("test", "MyApp.AppHost")).Return([]fs.DirEntry{})
				env.On("LsDir", "test").Return([]fs.DirEntry{})
				env.On("LsDir", ".").Return([]fs.DirEntry{})
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
