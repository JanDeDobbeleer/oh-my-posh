package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDataFormats(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		contents string
	}{
		{
			name: "json",
			file: "data.json",
			contents: `{
	"env": { "PWD": "~/dev/project", "Code": 1, "ExecutionTime": 341.0, "UserName": "jan" },
	"segments": {
		"git": { "HEAD": "main" },
		"az": { "Name": "posh-subscription", "EnvironmentName": "AzureCloud" }
	}
}`,
		},
		{
			name: "jsonc",
			file: "data.jsonc",
			contents: `{
	// deterministic render context
	"env": { "PWD": "~/dev/project", "Code": 1, "ExecutionTime": 341.0, "UserName": "jan" },
	"segments": {
		"git": { "HEAD": "main" }, // replayed, not detected
		"az": { "Name": "posh-subscription", "EnvironmentName": "AzureCloud" }
	}
}`,
		},
		{
			name: "yaml",
			file: "data.yaml",
			contents: `env:
  PWD: ~/dev/project
  Code: 1
  ExecutionTime: 341.0
  UserName: jan
segments:
  git:
    HEAD: main
  az:
    Name: posh-subscription
    EnvironmentName: AzureCloud
`,
		},
		{
			name: "toml",
			file: "data.toml",
			contents: `[env]
PWD = "~/dev/project"
Code = 1
ExecutionTime = 341.0
UserName = "jan"

[segments.git]
HEAD = "main"

[segments.az]
Name = "posh-subscription"
EnvironmentName = "AzureCloud"
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, tc.file)
			require.NoError(t, os.WriteFile(path, []byte(tc.contents), 0644))

			data, err := LoadData(path)
			require.NoError(t, err)
			require.NotNil(t, data)

			envFlags, err := data.EnvFlags()
			require.NoError(t, err)
			require.NotNil(t, envFlags.PWD)
			assert.Equal(t, "~/dev/project", *envFlags.PWD)
			require.NotNil(t, envFlags.Code)
			assert.Equal(t, 1, *envFlags.Code)
			require.NotNil(t, envFlags.ExecutionTime)
			assert.InDelta(t, 341.0, *envFlags.ExecutionTime, 0.0001)
			assert.Nil(t, envFlags.PipeStatus)

			require.Contains(t, data.Segments, "git")
			require.Contains(t, data.Segments, "az")

			var git struct {
				HEAD string
			}

			require.NoError(t, json.Unmarshal(data.Segments["git"], &git))
			assert.Equal(t, "main", git.HEAD)

			var az struct {
				Name            string
				EnvironmentName string
			}

			require.NoError(t, json.Unmarshal(data.Segments["az"], &az))
			assert.Equal(t, "posh-subscription", az.Name)
			assert.Equal(t, "AzureCloud", az.EnvironmentName)
		})
	}
}

func TestEnvFlagsPresence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	contents := `{ "env": { "Code": 0 } }`
	require.NoError(t, os.WriteFile(path, []byte(contents), 0644))

	data, err := LoadData(path)
	require.NoError(t, err)

	envFlags, err := data.EnvFlags()
	require.NoError(t, err)

	require.NotNil(t, envFlags.Code)
	assert.Equal(t, 0, *envFlags.Code)

	assert.Nil(t, envFlags.PWD)
	assert.Nil(t, envFlags.ExecutionTime)
	assert.Nil(t, envFlags.PipeStatus)
}

func TestEnvFlagsAbsent(t *testing.T) {
	data := &Data{}

	envFlags, err := data.EnvFlags()
	require.NoError(t, err)

	assert.Nil(t, envFlags.PWD)
	assert.Nil(t, envFlags.Code)
	assert.Nil(t, envFlags.ExecutionTime)
	assert.Nil(t, envFlags.PipeStatus)
}

func TestLoadDataUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	require.NoError(t, os.WriteFile(path, []byte("irrelevant"), 0644))

	data, err := LoadData(path)
	require.Error(t, err)
	assert.Nil(t, data)
}

func TestLoadDataMissingFile(t *testing.T) {
	data, err := LoadData(filepath.Join(t.TempDir(), "does-not-exist.json"))
	require.Error(t, err)
	assert.Nil(t, data)
}

func TestLoadDataInvalidContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.json")
	require.NoError(t, os.WriteFile(path, []byte("not valid json"), 0644))

	data, err := LoadData(path)
	require.Error(t, err)
	assert.Nil(t, data)
}
