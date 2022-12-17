package segments

import (
	"io/fs"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert"

	testify_mock "github.com/stretchr/testify/mock"
)

const (
	hasFiles = "HasFiles"
)

type MockDirEntry struct {
	name     string
	isDir    bool
	fileMode fs.FileMode
	fileInfo fs.FileInfo
	err      error
}

func (m *MockDirEntry) Name() string {
	return m.name
}

func (m *MockDirEntry) IsDir() bool {
	return m.isDir
}

func (m *MockDirEntry) Type() fs.FileMode {
	return m.fileMode
}

func (m *MockDirEntry) Info() (fs.FileInfo, error) {
	return m.fileInfo, m.err
}

func TestPackage(t *testing.T) {
	cases := []struct {
		Name            string
		Case            string
		File            string
		PackageContents string
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "1.0.0 node.js",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 1.0.0 test",
			Name:            "node",
			File:            "package.json",
			PackageContents: "{\"version\":\"1.0.0\",\"name\":\"test\"}",
		},
		{
			Case:            "1.0.0 php",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 1.0.0 test",
			Name:            "php",
			File:            "composer.json",
			PackageContents: "{\"version\":\"1.0.0\",\"name\":\"test\"}",
		},
		{
			Case:            "3.2.1 node.js",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 3.2.1 test",
			Name:            "node", File: "package.json",
			PackageContents: "{\"version\":\"3.2.1\",\"name\":\"test\"}",
		},
		{
			Case:            "1.0.0 cargo",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 1.0.0 test",
			Name:            "cargo",
			File:            "Cargo.toml",
			PackageContents: "[package]\nname=\"test\"\nversion=\"1.0.0\"\n",
		},
		{
			Case:            "3.2.1 cargo",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 3.2.1 test",
			Name:            "cargo",
			File:            "Cargo.toml",
			PackageContents: "[package]\nname=\"test\"\nversion=\"3.2.1\"\n",
		},
		{
			Case:            "1.0.0 poetry",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 1.0.0 test",
			Name:            "poetry",
			File:            "pyproject.toml",
			PackageContents: "[tool.poetry]\nname=\"test\"\nversion=\"1.0.0\"\n",
		},
		{
			Case:            "3.2.1 poetry",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 3.2.1 test",
			Name:            "poetry",
			File:            "pyproject.toml",
			PackageContents: "[tool.poetry]\nname=\"test\"\nversion=\"3.2.1\"\n",
		},
		{
			Case:            "No version present node.js",
			ExpectedEnabled: true,
			ExpectedString:  "test",
			Name:            "node",
			File:            "package.json",
			PackageContents: "{\"name\":\"test\"}",
		},
		{
			Case:            "No version present cargo",
			ExpectedEnabled: true,
			ExpectedString:  "test",
			Name:            "cargo",
			File:            "Cargo.toml",
			PackageContents: "[package]\nname=\"test\"\n",
		},
		{
			Case:            "No version present poetry",
			ExpectedEnabled: true,
			ExpectedString:  "test",
			Name:            "poetry",
			File:            "pyproject.toml",
			PackageContents: "[tool.poetry]\nname=\"test\"\n",
		},
		{
			Case:            "No name present node.js",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 1.0.0",
			Name:            "node",
			File:            "package.json",
			PackageContents: "{\"version\":\"1.0.0\"}",
		},
		{
			Case:            "No name present cargo",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 1.0.0",
			Name:            "cargo",
			File:            "Cargo.toml",
			PackageContents: "[package]\nversion=\"1.0.0\"\n",
		},
		{
			Case:            "No name present poetry",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 1.0.0",
			Name:            "poetry",
			File:            "pyproject.toml",
			PackageContents: "[tool.poetry]\nversion=\"1.0.0\"\n",
		},
		{
			Case:            "Empty project package node.js",
			ExpectedEnabled: false,
			Name:            "node",
			File:            "package.json",
			PackageContents: "{}",
		},
		{
			Case:            "Empty project package cargo",
			Name:            "cargo",
			File:            "Cargo.toml",
			PackageContents: "",
		},
		{
			Case:            "Empty project package poetry",
			Name:            "poetry",
			File:            "pyproject.toml",
			PackageContents: "",
		},
		{
			Case:            "Invalid json",
			ExpectedString:  "invalid character '}' looking for beginning of value",
			Name:            "node",
			File:            "package.json",
			PackageContents: "}",
		},
		{
			Case:            "Invalid toml",
			ExpectedString:  "toml: line 1: unexpected end of table name (table names cannot be empty)",
			Name:            "cargo",
			File:            "Cargo.toml",
			PackageContents: "[",
		},
		{
			Case:            "Julia project",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 0.1.0 ProjectEuler",
			Name:            "julia",
			File:            "JuliaProject.toml",
			PackageContents: "name = \"ProjectEuler\"\nversion = \"0.1.0\"",
		},
		{
			Case:            "Julia project no name",
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 0.1.0",
			Name:            "julia",
			File:            "JuliaProject.toml",
			PackageContents: "version = \"0.1.0\"",
		},
		{
			Case:            "Julia project no version",
			ExpectedEnabled: true,
			ExpectedString:  "ProjectEuler",
			Name:            "julia",
			File:            "JuliaProject.toml",
			PackageContents: "name = \"ProjectEuler\"",
		},
		{
			Case:            "Julia project invalid toml",
			ExpectedString:  "toml: line 1: unexpected end of table name (table names cannot be empty)",
			Name:            "julia",
			File:            "JuliaProject.toml",
			PackageContents: "[",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On(hasFiles, testify_mock.Anything).Run(func(args testify_mock.Arguments) {
			for _, c := range env.ExpectedCalls {
				if c.Method == hasFiles {
					c.ReturnArguments = testify_mock.Arguments{args.Get(0).(string) == tc.File}
				}
			}
		})
		env.On("FileContent", tc.File).Return(tc.PackageContents)
		pkg := &Project{}
		pkg.Init(properties.Map{}, env)
		assert.Equal(t, tc.ExpectedEnabled, pkg.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, pkg.Template(), pkg), tc.Case)
		}
	}
}

func TestNuspecPackage(t *testing.T) {
	cases := []struct {
		Case            string
		HasFiles        bool
		FileName        string
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "valid file",
			FileName:        "../test/valid.nuspec",
			HasFiles:        true,
			ExpectedEnabled: true,
			ExpectedString:  "\uf487 0.1.0 Az.Compute",
		},
		{
			Case:            "invalid file",
			FileName:        "../test/invalid.nuspec",
			HasFiles:        true,
			ExpectedEnabled: false,
		},
		{
			Case:            "no info in file",
			FileName:        "../test/empty.nuspec",
			HasFiles:        true,
			ExpectedEnabled: false,
		},
		{
			Case:            "no files",
			HasFiles:        false,
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On(hasFiles, testify_mock.Anything).Run(func(args testify_mock.Arguments) {
			for _, c := range env.ExpectedCalls {
				if c.Method != hasFiles {
					continue
				}
				if args.Get(0).(string) == "*.nuspec" {
					c.ReturnArguments = testify_mock.Arguments{tc.HasFiles}
					continue
				}
				c.ReturnArguments = testify_mock.Arguments{false}
			}
		})
		env.On("Pwd").Return("posh")
		env.On("LsDir", "posh").Return([]fs.DirEntry{
			&MockDirEntry{
				name: tc.FileName,
			},
		})
		content, _ := os.ReadFile(tc.FileName)
		env.On("FileContent", tc.FileName).Return(string(content))
		pkg := &Project{}
		pkg.Init(properties.Map{}, env)
		assert.Equal(t, tc.ExpectedEnabled, pkg.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, pkg.Template(), pkg), tc.Case)
		}
	}
}

func TestDotnetProject(t *testing.T) {
	cases := []struct {
		Case            string
		FileName        string
		HasFiles        bool
		ProjectContents string
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "valid .csproj file",
			FileName:        "Valid.csproj",
			HasFiles:        true,
			ProjectContents: "...<TargetFramework>net7.0</TargetFramework>...",
			ExpectedEnabled: true,
			ExpectedString:  "Valid \uf9fd net7.0",
		},
		{
			Case:            "valid .fsproj file",
			FileName:        "Valid.fsproj",
			HasFiles:        true,
			ProjectContents: "...<TargetFramework>net6.0</TargetFramework>...",
			ExpectedEnabled: true,
			ExpectedString:  "Valid \uf9fd net6.0",
		},
		{
			Case:            "valid .vbproj file",
			FileName:        "Valid.vbproj",
			HasFiles:        true,
			ProjectContents: "...<TargetFramework>net5.0</TargetFramework>...",
			ExpectedEnabled: true,
			ExpectedString:  "Valid \uf9fd net5.0",
		},
		{
			Case:            "invalid or empty contents",
			FileName:        "Invalid.csproj",
			HasFiles:        true,
			ExpectedEnabled: false,
			ExpectedString:  "cannot extract TFM from Invalid project file",
		},
		{
			Case:            "no files",
			HasFiles:        false,
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On(hasFiles, testify_mock.Anything).Run(func(args testify_mock.Arguments) {
			for _, c := range env.ExpectedCalls {
				if c.Method == hasFiles {
					pattern := "*" + filepath.Ext(tc.FileName)
					c.ReturnArguments = testify_mock.Arguments{args.Get(0).(string) == pattern}
				}
			}
		})
		env.On("Pwd").Return("posh")
		env.On("LsDir", "posh").Return([]fs.DirEntry{
			&MockDirEntry{
				name: tc.FileName,
			},
		})
		env.On("FileContent", tc.FileName).Return(tc.ProjectContents)
		pkg := &Project{}
		pkg.Init(properties.Map{}, env)
		assert.Equal(t, tc.ExpectedEnabled, pkg.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, pkg.Template(), pkg), tc.Case)
		}
	}
}
