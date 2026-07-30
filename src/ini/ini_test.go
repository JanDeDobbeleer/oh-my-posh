package ini

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	src := `# aws style comment
; another comment
[default]
region = us-east-1
output = json ; trailing comment

[profile dev]
region = "eu-west-1"
sso_start_url = https://example.awsapps.com/start

[core]
project = smoke-test
account = someone@example.com
`

	file, err := Load(src)
	assert.NoError(t, err)

	section, err := file.GetSection("default")
	assert.NoError(t, err)
	assert.Len(t, section.Keys(), 2)
	assert.Equal(t, "us-east-1", section.Key("region").Value())
	assert.Equal(t, "json", section.Key("output").Value())

	assert.Equal(t, "smoke-test", file.Section("core").Key("project").String())
	assert.Equal(t, "eu-west-1", file.Section("profile dev").Key("region").Value())

	_, err = file.GetSection("missing")
	assert.Error(t, err)
	assert.Equal(t, "", file.Section("missing").Key("anything").String())
}

func TestLoadGitConfig(t *testing.T) {
	src := `[core]
	repositoryformatversion = 0
	bare = false
[remote "origin"]
	url = https://github.com/jandedobbeleer/oh-my-posh.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
	pushDefault = origin
`

	file, err := Load(src)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/jandedobbeleer/oh-my-posh.git", file.Section(`remote "origin"`).Key("url").String())
	assert.Equal(t, "origin", file.Section(`branch "main"`).Key("remote").String())
	assert.Equal(t, "false", file.Section("core").Key("bare").String())
}

func TestLoadVerbatim(t *testing.T) {
	file, err := LoadVerbatim("[s]\nkey = value ; kept\n")
	assert.NoError(t, err)
	assert.Equal(t, "value ; kept", file.Section("s").Key("key").Value())
}

func TestLoadErrors(t *testing.T) {
	_, err := Load("[s]\nkey without delimiter\n")
	assert.Error(t, err)

	_, err = Load("[unclosed\n")
	assert.Error(t, err)
}

func TestLoadFirstKeyWins(t *testing.T) {
	file, err := Load("[s]\nkey = first\nkey = second\n")
	assert.NoError(t, err)
	assert.Equal(t, "first", file.Section("s").Key("key").Value())
	assert.Len(t, file.Section("s").Keys(), 1)
}

func TestLoadBOMAndColonDelimiter(t *testing.T) {
	file, err := Load("\ufeff[s]\nkey: value\n")
	assert.NoError(t, err)
	assert.Equal(t, "value", file.Section("s").Key("key").Value())
}

func TestSections(t *testing.T) {
	file, err := Load("[remote \"origin\"]\nurl = a\n[remote \"upstream\"]\nurl = b\n")
	assert.NoError(t, err)

	var names []string
	for _, section := range file.Sections() {
		names = append(names, section.Name())
	}

	assert.Equal(t, []string{"", `remote "origin"`, `remote "upstream"`}, names)
}
