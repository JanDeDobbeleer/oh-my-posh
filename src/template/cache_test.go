package template

import (
	"encoding/json"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

// newLoadCacheMockEnv builds a mock environment stubbed with sane live values
// for every call loadCache makes, so each test only needs to layer on top of
// it (EnvData, aliases, ...) rather than reconstructing the whole surface.
func newLoadCacheMockEnv(flags *runtime.Flags) *mock.Environment {
	flags.IsPrimary = true

	env := new(mock.Environment)
	env.On("Flags").Return(flags)
	env.On("Root").Return(false)
	env.On("Shell").Return("pwsh")
	env.On("StatusCodes").Return(0, "")
	env.On("IsWsl").Return(false)
	env.On("Pwd").Return("/tmp/omp-test/project")
	env.On("User").Return("jan")
	env.On("Host").Return("live-host", nil)
	env.On("GOOS").Return(runtime.DARWIN)
	env.On("Platform").Return("darwin")
	env.On("Getenv", "SHLVL").Return("1")

	return env
}

func TestLoadCacheEnvDataOverlay(t *testing.T) {
	flags := &runtime.Flags{
		EnvData: json.RawMessage(`{"UserName":"data-user","HostName":"data-host","Shell":"zsh","Root":true}`),
	}

	env = newLoadCacheMockEnv(flags)

	loadCache(nil, &maps.Config{})

	assert.Equal(t, "data-user", Cache.UserName)
	assert.Equal(t, "data-host", Cache.HostName)
	assert.Equal(t, "zsh", Cache.Shell)
	assert.True(t, Cache.Root)

	// Untouched fields keep their live values.
	assert.False(t, Cache.WSL)
	assert.Equal(t, 1, Cache.SHLVL)
	assert.Equal(t, 0, Cache.Code)
}

func TestLoadCacheEnvDataIgnoresRoutedKeys(t *testing.T) {
	flags := &runtime.Flags{
		EnvData: json.RawMessage(`{"PWD":"/bogus/path","Code":99,"ExecutionTime":42.5,"PipeStatus":"1 0"}`),
	}

	env = newLoadCacheMockEnv(flags)

	loadCache(nil, &maps.Config{})

	// The routed keys must be ignored by the overlay: PWD/Code keep the
	// value already resolved by the CLI layer (here, the live mock values),
	// never the raw file value.
	assert.Equal(t, "/tmp/omp-test/project", Cache.PWD)
	assert.Equal(t, 0, Cache.Code)
}

func TestLoadCacheEnvDataAliasAppliesToOverlaidValue(t *testing.T) {
	flags := &runtime.Flags{
		EnvData: json.RawMessage(`{"UserName":"data-user"}`),
	}

	env = newLoadCacheMockEnv(flags)

	aliases := &maps.Config{
		UserName: &maps.Map{"data-user": "aliased-user"},
	}

	loadCache(nil, aliases)

	assert.Equal(t, "aliased-user", Cache.UserName)
}

func TestLoadCacheEnvDataFolderOverride(t *testing.T) {
	flags := &runtime.Flags{
		EnvData: json.RawMessage(`{"Folder":"custom-folder"}`),
	}

	env = newLoadCacheMockEnv(flags)

	loadCache(nil, &maps.Config{})

	// Derived from pwd it would be "project"; the explicit override wins.
	assert.Equal(t, "custom-folder", Cache.Folder)
}

func TestLoadCacheNoEnvData(t *testing.T) {
	flags := &runtime.Flags{}

	env = newLoadCacheMockEnv(flags)

	loadCache(nil, &maps.Config{})

	assert.Equal(t, "jan", Cache.UserName)
	assert.Equal(t, "live-host", Cache.HostName)
	assert.Equal(t, "pwsh", Cache.Shell)
	assert.False(t, Cache.Root)
	assert.Equal(t, "project", Cache.Folder)
	assert.Equal(t, "/tmp/omp-test/project", Cache.PWD)
}

func TestLoadCacheInvalidEnvData(t *testing.T) {
	flags := &runtime.Flags{
		EnvData: json.RawMessage(`{invalid`),
	}

	env = newLoadCacheMockEnv(flags)

	assert.NotPanics(t, func() {
		loadCache(nil, &maps.Config{})
	})

	// The prompt still renders with the live values.
	assert.Equal(t, "jan", Cache.UserName)
	assert.Equal(t, "live-host", Cache.HostName)
	assert.Equal(t, "pwsh", Cache.Shell)
}
