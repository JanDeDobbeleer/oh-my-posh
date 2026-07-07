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

	// Build fully before assigning: a long-lived process (serve) rebuilds
	// this per prompt while goroutines from an abandoned render cycle may
	// still read the global - they must only ever observe a complete object.
	tmpl := new(cache.Template)

	tmpl.Root = env.Root()
	tmpl.Shell = aliases.GetShellName(env.Shell())
	tmpl.ShellVersion = env.Flags().ShellVersion
	tmpl.Code, _ = env.StatusCodes()
	tmpl.WSL = env.IsWsl()
	tmpl.Segments = maps.NewConcurrent[any]()
	tmpl.PromptCount = env.Flags().PromptCount
	tmpl.Var = make(map[string]any)
	tmpl.Jobs = env.Flags().JobCount
	tmpl.Version = build.Version

	if vars != nil {
		tmpl.Var = vars
	}

	pwd := env.Pwd()
	tmpl.PWD = path.ReplaceHomeDirPrefixWithTilde(pwd)

	tmpl.AbsolutePWD = pwd
	if env.IsWsl() {
		tmpl.AbsolutePWD, _ = env.RunCommand("wslpath", "-m", pwd)
	}

	env.Flags().AbsolutePWD = tmpl.AbsolutePWD
	tmpl.PSWD = env.Flags().PSWD

	tmpl.Folder = path.Base(pwd)
	if env.GOOS() == runtime.WINDOWS && strings.HasSuffix(tmpl.Folder, ":") {
		tmpl.Folder += `\`
	}

	tmpl.UserName = aliases.GetUserName(env.User())
	if host, err := env.Host(); err == nil {
		tmpl.HostName = aliases.GetHostName(host)
	}

	goos := env.GOOS()
	tmpl.OS = goos
	if goos == runtime.LINUX {
		tmpl.OS = env.Platform()
	}

	val := env.Getenv("SHLVL")
	if shlvl, err := strconv.Atoi(val); err == nil {
		tmpl.SHLVL = shlvl
	}

	Cache = tmpl
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
