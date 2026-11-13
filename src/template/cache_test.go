package template

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
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
		EnvData: json.RawMessage(`{"PWD":"/bogus/path","Code":99,"ExecutionTime":42.5,"PipeStatus":"1 0","Interrupted":true,"Executed":false}`),
	}

	env = newLoadCacheMockEnv(flags)

	loadCache(nil, &maps.Config{})

	// The routed keys must be ignored by the overlay: PWD/Code/Interrupted/Executed
	// keep the value already resolved by the CLI layer (here, the live mock
	// values), never the raw file value.
	assert.Equal(t, "/tmp/omp-test/project", Cache.PWD)
	assert.Equal(t, 0, Cache.Code)
	assert.False(t, Cache.Interrupted)
	assert.True(t, Cache.Executed)
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

func TestLoadCacheExecutedReflectsNoExitCode(t *testing.T) {
	flags := &runtime.Flags{NoExitCode: true}

	env = newLoadCacheMockEnv(flags)

	loadCache(nil, &maps.Config{})

	assert.False(t, Cache.Executed)
}

func TestLoadCacheExecutedTrueWhenExitCodeValid(t *testing.T) {
	flags := &runtime.Flags{NoExitCode: false}

	env = newLoadCacheMockEnv(flags)

	loadCache(nil, &maps.Config{})

	assert.True(t, Cache.Executed)
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

// TestLoadCacheVarTrustsMarkup guards the collateral damage from escaping
// every action result: var section strings are user configuration, so they
// must reach the renderer as Markup and keep their anchors intact.
func TestLoadCacheVarTrustsMarkup(t *testing.T) {
	cases := []struct {
		Case     string
		Value    string
		Expected string
	}{
		{Case: "chevrons render as markup anchors", Value: "<red>hi</>", Expected: "<red>hi</>"},
		{Case: "a plain value without chevrons is unchanged", Value: "hi", Expected: "hi"},
	}

	origCache := Cache
	t.Cleanup(func() { Cache = origCache })

	for _, tc := range cases {
		env = newLoadCacheMockEnv(&runtime.Flags{})

		Cache = nil
		Init(env, maps.Simple[any]{"Hello": tc.Value}, &maps.Config{})

		text, err := RenderTrusted(`{{ .Var.Hello }}`, nil)
		assert.NoError(t, err, tc.Case)
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
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

func TestLoadCachePromptArrival(t *testing.T) {
	cases := []struct {
		Case              string
		Waited            time.Duration
		Saved             bool
		KeepTemplateCache bool
		Primary           bool
	}{
		{Case: "the primary prompt arrives now", Primary: true},
		{Case: "restored with the template cache", Saved: true, Waited: time.Minute, KeepTemplateCache: true},
		{Case: "survives an expired template cache", Saved: true, Waited: 25 * time.Hour},
		{Case: "unknown when nothing was saved"},
	}

	origCache := Cache
	t.Cleanup(func() {
		Cache = origCache
		cache.Session.DeleteAll()
	})

	for _, tc := range cases {
		cache.Session.DeleteAll()

		var arrivedAt time.Time
		if tc.Saved {
			env = newLoadCacheMockEnv(&runtime.Flags{HasExtra: true})
			loadCache(nil, &maps.Config{})
			Cache.PromptArrivedAt = Cache.PromptArrivedAt.Add(-tc.Waited)
			arrivedAt = Cache.PromptArrivedAt
			SaveCache()
		}

		if !tc.KeepTemplateCache {
			cache.Session.Delete(cache.TEMPLATECACHE)
		}

		flags := &runtime.Flags{}
		env = newLoadCacheMockEnv(flags)
		flags.IsPrimary = tc.Primary
		before := time.Now()
		loadCache(nil, &maps.Config{})

		if tc.Primary {
			assert.WithinDuration(t, before, Cache.PromptArrivedAt, time.Second, tc.Case)
			assert.Zero(t, Cache.IdleTime, tc.Case)
			continue
		}

		assert.True(t, arrivedAt.Equal(Cache.PromptArrivedAt), tc.Case)
		assert.InDelta(t, tc.Waited.Milliseconds(), Cache.IdleTime, float64(time.Second.Milliseconds()), tc.Case)
	}
}
