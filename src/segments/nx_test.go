package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestGetNodePackageVersion(t *testing.T) {
	cases := []struct {
		Case        string
		PackageJSON string
		Version     string
		ShouldFail  bool
		NoFiles     bool
	}{
		{Case: "14.1.5", Version: "14.1.5", PackageJSON: "{ \"name\": \"nx\",\"version\": \"14.1.5\"}"},
		{Case: "14.0.0", Version: "14.0.0", PackageJSON: "{ \"name\": \"nx\",\"version\": \"14.0.0\"}"},
		{Case: "no files", NoFiles: true, ShouldFail: true},
		{Case: "bad data", ShouldFail: true, PackageJSON: "bad data"},
	}

	for _, tc := range cases {
		var env = new(mock.MockedEnvironment)
		// mock  getVersion methods
		env.On("Pwd").Return("posh")
		path := filepath.Join("posh", "node_modules", "nx")
		env.On("HasFilesInDir", path, "package.json").Return(!tc.NoFiles)
		env.On("FileContent", filepath.Join(path, "package.json")).Return(tc.PackageJSON)
		env.On("TemplateCache").Return(&environment.TemplateCache{
			Env: make(map[string]string),
		})
		env.On("Log", mock2.Anything, mock2.Anything, mock2.Anything)
		got := getNodePackageVersion(env, "nx")
		if tc.ShouldFail {
			assert.Empty(t, got, tc.Case)
			return
		}
		assert.Equal(t, tc.Version, got, tc.Case)
	}
}
