package template

import (
	"strconv"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

var (
	Cache *cache.Template
)

func loadCache(vars maps.Simple[any], aliases *maps.Config) {
	if !env.Flags().IsPrimary {
		// Load the template cache for a non-primary prompt before rendering any templates.
		if OK := restoreCache(); OK {
			return
		}
	}

	Cache = new(cache.Template)

	Cache.Root = env.Root()
	Cache.Shell = aliases.GetShellName(env.Shell())
	Cache.ShellVersion = env.Flags().ShellVersion
	Cache.Code, _ = env.StatusCodes()
	Cache.WSL = env.IsWsl()
	Cache.Segments = maps.NewConcurrent[any]()
	Cache.PromptCount = env.Flags().PromptCount
	Cache.Var = make(map[string]any)
	Cache.Jobs = env.Flags().JobCount
	Cache.Version = build.Version

	if vars != nil {
		Cache.Var = vars
	}

	pwd := env.Pwd()
	Cache.PWD = path.ReplaceHomeDirPrefixWithTilde(pwd)

	Cache.AbsolutePWD = pwd
	if env.IsWsl() {
		Cache.AbsolutePWD, _ = env.RunCommand("wslpath", "-m", pwd)
	}

	env.Flags().AbsolutePWD = Cache.AbsolutePWD
	Cache.PSWD = env.Flags().PSWD

	Cache.Folder = path.Base(pwd)
	if env.GOOS() == runtime.WINDOWS && strings.HasSuffix(Cache.Folder, ":") {
		Cache.Folder += `\`
	}

	Cache.UserName = aliases.GetUserName(env.User())
	if host, err := env.Host(); err == nil {
		Cache.HostName = aliases.GetHostName(host)
	}

	goos := env.GOOS()
	Cache.OS = goos
	if goos == runtime.LINUX {
		Cache.OS = env.Platform()
	}

	val := env.Getenv("SHLVL")
	if shlvl, err := strconv.Atoi(val); err == nil {
		Cache.SHLVL = shlvl
	}
}

func restoreCache() bool {
	defer log.Trace(time.Now())

	val, OK := cache.Get[cache.SimpleTemplate](cache.Session, cache.TEMPLATECACHE)
	if !OK {
		return false
	}

	Cache = new(cache.Template)
	Cache.SimpleTemplate = val
	Cache.Segments = Cache.SegmentsCache.ToConcurrent()

	return true
}

func SaveCache() {
	// only store this when in a primary prompt
	// and when we have any extra prompt in the config
	canSave := env.Flags().IsPrimary && env.Flags().HasExtra
	if !canSave {
		return
	}

	Cache.SegmentsCache = Cache.Segments.ToSimple()

	cache.Set(cache.Session, cache.TEMPLATECACHE, &Cache.SimpleTemplate, cache.ONEDAY)
}
