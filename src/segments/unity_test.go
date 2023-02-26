package segments

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	mock2 "github.com/stretchr/testify/mock"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

type CacheGet struct {
	key   string
	val   string
	found bool
}

type CacheSet struct {
	key string
	val string
}

type HTTPResponse struct {
	body string
	err  error
}

func TestUnitySegment(t *testing.T) {
	cases := []struct {
		Case                string
		ExpectedOutput      string
		VersionFileText     string
		CacheGet            CacheGet
		CacheSet            CacheSet
		ExpectedToBeEnabled bool
		VersionFileExists   bool
		HTTPResponse        HTTPResponse
	}{
		{
			Case:                "C# version cached",
			ExpectedOutput:      "\ue721 2021.3.16 C# 9",
			ExpectedToBeEnabled: true,
			VersionFileExists:   true,
			VersionFileText:     "m_EditorVersion: 2021.3.16f1\nm_EditorVersionWithRevision: 2021.3.16f1 (4016570cf34f)",
			CacheGet: CacheGet{
				key:   "2021.3",
				val:   "C# 9",
				found: true,
			},
		},
		{
			Case:                "C# version not cached",
			ExpectedOutput:      "\ue721 2021.3.16 C# 9",
			ExpectedToBeEnabled: true,
			VersionFileExists:   true,
			VersionFileText:     "m_EditorVersion: 2021.3.16f1\nm_EditorVersionWithRevision: 2021.3.16f1 (4016570cf34f)",
			CacheGet: CacheGet{
				key:   "2021.3",
				val:   "",
				found: false,
			},
			CacheSet: CacheSet{
				key: "2021.3",
				val: "C# 9",
			},
			HTTPResponse: HTTPResponse{
				body: `<a href="https://docs.microsoft.com/en-us/dotnet/csharp/whats-new/csharp-9">C# 9.0</a>`,
				err:  nil,
			},
		},
		{
			Case:                "C# version has a minor version",
			ExpectedOutput:      "\ue721 2021.3.16 C# 7.3",
			ExpectedToBeEnabled: true,
			VersionFileExists:   true,
			VersionFileText:     "m_EditorVersion: 2021.3.16f1\nm_EditorVersionWithRevision: 2021.3.16f1 (4016570cf34f)",
			CacheGet: CacheGet{
				key:   "2021.3",
				val:   "",
				found: false,
			},
			CacheSet: CacheSet{
				key: "2021.3",
				val: "C# 7.3",
			},
			HTTPResponse: HTTPResponse{
				body: `<a href="https://docs.microsoft.com/en-us/dotnet/csharp/whats-new/csharp-7-3">C# 7.3</a>`,
				err:  nil,
			},
		},
		{
			Case:                "C# version not found in webpage",
			ExpectedOutput:      "\ue721 2021.3.16",
			ExpectedToBeEnabled: true,
			VersionFileExists:   true,
			VersionFileText:     "m_EditorVersion: 2021.3.16f1\nm_EditorVersionWithRevision: 2021.3.16f1 (4016570cf34f)",
			CacheGet: CacheGet{
				key:   "2021.3",
				val:   "",
				found: false,
			},
			CacheSet: CacheSet{
				key: "2021.3",
				val: "",
			},
			HTTPResponse: HTTPResponse{
				body: `<h1>Sorry... that page seems to be missing!</h1>`,
				err:  nil,
			},
		},
		{
			Case:                "http request fails",
			ExpectedOutput:      "\ue721 2021.3.16",
			ExpectedToBeEnabled: true,
			VersionFileExists:   true,
			VersionFileText:     "m_EditorVersion: 2021.3.16f1\nm_EditorVersionWithRevision: 2021.3.16f1 (4016570cf34f)",
			CacheGet: CacheGet{
				key:   "2021.3",
				val:   "",
				found: false,
			},
			HTTPResponse: HTTPResponse{
				body: "",
				err:  errors.New("FAIL"),
			},
		},
		{
			Case:                "ProjectSettings/ProjectVersion.txt doesn't exist",
			ExpectedToBeEnabled: false,
			VersionFileExists:   false,
		},
		{
			Case:                "ProjectSettings/ProjectVersion.txt is empty",
			ExpectedToBeEnabled: false,
			VersionFileExists:   true,
			VersionFileText:     "",
		},
		{
			Case:                "ProjectSettings/ProjectVersion.txt does not have expected format",
			ExpectedToBeEnabled: false,
			VersionFileExists:   true,
			VersionFileText:     "2021.3.16f1",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Error", mock2.Anything).Return()
		env.On("Debug", mock2.Anything)

		err := errors.New("no match at root level")
		var projectDir *platform.FileInfo
		if tc.VersionFileExists {
			err = nil
			projectDir = &platform.FileInfo{
				ParentFolder: "UnityProjectRoot",
				Path:         "UnityProjectRoot/ProjectSettings",
				IsDir:        true,
			}
			env.On("HasFilesInDir", projectDir.Path, "ProjectVersion.txt").Return(tc.VersionFileExists)
			versionFilePath := filepath.Join(projectDir.Path, "ProjectVersion.txt")
			env.On("FileContent", versionFilePath).Return(tc.VersionFileText)
		}
		env.On("HasParentFilePath", "ProjectSettings").Return(projectDir, err)

		cache := &mock.MockedCache{}
		cache.On("Get", tc.CacheGet.key).Return(tc.CacheGet.val, tc.CacheGet.found)
		cache.On("Set", tc.CacheSet.key, tc.CacheSet.val, -1).Return()
		env.On("Cache").Return(cache)

		url := fmt.Sprintf("https://docs.unity3d.com/%s/Documentation/Manual/CSharpCompiler.html", tc.CacheGet.key)
		env.On("HTTPRequest", url).Return([]byte(tc.HTTPResponse.body), tc.HTTPResponse.err)

		props := properties.Map{}
		unity := &Unity{}
		unity.Init(props, env)
		assert.Equal(t, tc.ExpectedToBeEnabled, unity.Enabled())
		if tc.ExpectedToBeEnabled {
			assert.Equal(t, tc.ExpectedOutput, renderTemplate(env, unity.Template(), unity), tc.Case)
		}
	}
}
