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

	// routedEnvDataKeys are env-section keys the CLI layer already routes into
	// runtime.Flags before Init runs (see runtime.Flags.EnvData), so by the
	// time loadCache runs, env.Pwd()/StatusCodes()/ExecutionTime() already
	// reflect the flag-precedence-resolved value (CLI flag > data file >
	// live). They must never be applied a second time here, or a raw file
	// value could clobber that resolved value.
	routedEnvDataKeys = []string{"PWD", "Code", "ExecutionTime", "PipeStatus"}
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
	tmpl.Shell = env.Shell()
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

	tmpl.UserName = env.User()
	if host, err := env.Host(); err == nil {
		tmpl.HostName = host
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

	overlayEnvData(tmpl)

	// Alias mapping must apply to a data-provided value exactly as it does to
	// a live one, so it runs after the overlay, on whichever value won.
	tmpl.Shell = aliases.GetShellName(tmpl.Shell)
	tmpl.UserName = aliases.GetUserName(tmpl.UserName)
	tmpl.HostName = aliases.GetHostName(tmpl.HostName)

	Cache = tmpl
}

// overlayEnvData merges the data file's env section onto an already-built
// template cache. It only touches keys present in the data, so anything not
// specified keeps its live value. Keys the CLI layer already routed into
// runtime.Flags (see routedEnvDataKeys) are stripped first so a raw file
// value can never clobber the flag-precedence-resolved value. Failures are
// logged, never fatal - the prompt renders with live values regardless.
func overlayEnvData(tmpl *cache.Template) {
	data := env.Flags().EnvData
	if len(data) == 0 {
		return
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		log.Error(err)
		return
	}

	for _, key := range routedEnvDataKeys {
		delete(fields, key)
	}

	// SegmentsCache/Segments are internal cache plumbing, not user-facing
	// template data; a data file must never be able to overwrite them. Var
	// (template vars) is left alone - overriding it from the data file is
	// intentional and useful.
	delete(fields, "SegmentsCache")
	delete(fields, "Segments")

	overlay, err := json.Marshal(fields)
	if err != nil {
		log.Error(err)
		return
	}

	if err := json.Unmarshal(overlay, &tmpl.SimpleTemplate); err != nil {
		log.Error(err)
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
