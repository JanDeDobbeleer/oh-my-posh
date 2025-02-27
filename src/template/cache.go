package template

import (
	"encoding/json"
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

func loadCache(vars maps.Simple) {
	if !env.Flags().IsPrimary {
		// Load the template cache for a non-primary prompt before rendering any templates.
		if OK := restoreCache(env); OK {
			return
		}
	}

	Cache = new(cache.Template)

	Cache.Root = env.Root()
	Cache.Shell = env.Shell()
	Cache.ShellVersion = env.Flags().ShellVersion
	Cache.Code, _ = env.StatusCodes()
	Cache.WSL = env.IsWsl()
	Cache.Segments = maps.NewConcurrent()
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

	Cache.UserName = env.User()
	if host, err := env.Host(); err == nil {
		Cache.HostName = host
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

func restoreCache(env runtime.Environment) bool {
	defer log.Trace(time.Now())

	val, OK := env.Session().Get(cache.TEMPLATECACHE)
	if !OK {
		return false
	}

	var tmplCache cache.Template
	err := json.Unmarshal([]byte(val), &tmplCache)
	if err != nil {
		log.Error(err)
		return false
	}

	Cache = &tmplCache
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

	templateCache, err := json.Marshal(Cache)
	if err == nil {
		env.Session().Set(cache.TEMPLATECACHE, string(templateCache), cache.ONEDAY)
	}
}
