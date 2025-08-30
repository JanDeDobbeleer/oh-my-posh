package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

var scriptPathCache string

func hasScript(env runtime.Environment) (string, bool) {
	if env.Flags().Debug || env.Flags().Eval {
		log.Debug("in debug or eval mode, no script path will be used")
		return "", false
	}

	path, err := scriptPath(env)
	if err != nil {
		log.Debug("failed to get script path")
		return "", false
	}

	_, err = os.Stat(path)
	if err != nil {
		log.Debug("script path does not exist")
		return "", false
	}

	// check if we have the same context
	if val, _ := env.Cache().Get(cacheKey(env.Flags().Shell)); val != cacheValue(env) {
		log.Debug("script context has changed")
		return "", false
	}

	log.Debug("script context is unchanged")
	return path, true
}

func writeScript(env runtime.Environment, script string) (string, error) {
	path, err := scriptPath(env)
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		log.Error(err)
		return "", err
	}

	defer func() {
		_ = f.Close()
	}()

	_, err = f.WriteString(script)
	if err != nil {
		log.Error(err)
		return "", err
	}

	defer func() {
		_ = f.Sync()
		_ = f.Close()
	}()

	env.Cache().Set(cacheKey(env.Flags().Shell), cacheValue(env), cache.INFINITE)

	return path, nil
}

func cacheKey(sh string) string {
	return fmt.Sprintf("INITVERSION%s", strings.ToUpper(sh))
}

func cacheValue(env runtime.Environment) string {
	return env.Flags().ConfigHash + build.Version
}

func scriptName(sh string) string {
	switch sh {
	case PWSH, PWSH5:
		sh = "ps1"
	case CMD:
		sh = "lua"
	case BASH:
		sh = "sh"
	case ELVISH:
		sh = "elv"
	case XONSH:
		sh = "xsh"
	}

	return fmt.Sprintf("init.%s", sh)
}

func scriptPath(env runtime.Environment) (string, error) {
	if len(scriptPathCache) != 0 {
		return scriptPathCache, nil
	}

	if env.Flags().Shell != NU {
		scriptPathCache = filepath.Join(cache.Path(), scriptName(env.Flags().Shell))
		log.Debug("init script path for non-nu shell:", scriptPathCache)
		return scriptPathCache, nil
	}

	const autoloadDir = "NUAUTOLOADDIR"

	if dir, OK := env.Cache().Get(autoloadDir); OK {
		scriptPathCache = filepath.Join(dir, scriptName(env.Flags().Shell))
		log.Debug("autoload path for nu from cache:", dir)
		return scriptPathCache, nil
	}

	path, err := env.RunCommand("nu", "-c", "$nu.data-dir | path join vendor autoload")
	if err != nil || path == "" {
		log.Error(err)
		return "", err
	}

	log.Debug("autoload path for nu:", path)

	// create the path if non-existent
	_, err = os.Stat(path)
	if err != nil {
		log.Debug("autoload path does not exist, creating")
		err = os.MkdirAll(path, 0o700)
	}

	if err != nil {
		log.Debugf("failed to create autoload dir %s: %s", path, err)
		return "", err
	}

	env.Cache().Set(autoloadDir, path, cache.INFINITE)
	scriptPathCache = filepath.Join(path, scriptName(env.Flags().Shell))
	log.Debug("script path for nu:", scriptPathCache)
	return scriptPathCache, nil
}
