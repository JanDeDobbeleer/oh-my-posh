package shell

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

var scriptPathCache string

func hasScript(env runtime.Environment) (string, bool) {
	if env.Flags().Debug || env.Flags().Eval || env.Flags().Shell == NU {
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
	if val, _ := cache.Get[string](cache.Device, cacheKey(env.Flags().Shell)); val != cacheValue(env) {
		log.Debug("script context has changed")
		return "", false
	}

	log.Debug("script context is unchanged")
	return path, true
}

func setFile(name string, data []byte, perm os.FileMode) error {
	f, err := os.ReadFile(name)

	if err == nil && bytes.Equal(data, f) {
		return nil
	}

	return os.WriteFile(name, data, perm)
}

func writeScript(env runtime.Environment, script string) (string, error) {
	path, err := scriptPath(env)
	if err != nil {
		return "", err
	}

	deadline := time.Now().Add(10 * time.Second)
	waitTime := 50 * time.Millisecond

	var firstErr error

	for time.Now().Before(deadline) {
		err = setFile(path, []byte(script), 0o644)
		if err == nil {
			firstErr = nil
			break
		}

		log.Error(err)
		if firstErr == nil {
			firstErr = err
		}

		time.Sleep(waitTime)

		waitTime *= 2

		if waitTime > time.Second {
			waitTime = time.Second
		}
	}

	if firstErr != nil {
		return "", firstErr
	}


	log.Debug("init script written successfully")
	cache.Set(cache.Device, cacheKey(env.Flags().Shell), cacheValue(env), cache.INFINITE)

	return path, nil
}

func cacheKey(sh string) string {
	return fmt.Sprintf("INITVERSION%s", strings.ToUpper(sh))
}

func cacheValue(env runtime.Environment) string {
	return fmt.Sprintf("%d%s", env.Flags().ConfigHash, build.Version)
}

func InitScriptName(flags *runtime.Flags) string {
	sh := flags.Shell
	switch flags.Shell {
	case PWSH:
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

	// to avoid a single init scripts for different configs
	// we hash the config path as part of the script name
	// that way we have a single init script per config
	// avoiding conflicts
	h := fnv.New64a()
	h.Write([]byte(flags.ConfigPath))
	hash := h.Sum64()

	return fmt.Sprintf("init.%d.%s", hash, sh)
}

func scriptPath(env runtime.Environment) (string, error) {
	if len(scriptPathCache) != 0 {
		return scriptPathCache, nil
	}

	if env.Flags().Shell != NU {
		scriptPathCache = filepath.Join(cache.Path(), InitScriptName(env.Flags()))
		log.Debug("init script path for non-nu shell:", scriptPathCache)
		return scriptPathCache, nil
	}

	const autoloadDir = "NUAUTOLOADDIR"
	const fileName = "oh-my-posh.nu"

	if dir, OK := cache.Get[string](cache.Device, autoloadDir); OK {
		scriptPathCache = filepath.Join(dir, fileName)
		log.Debug("autoload path for nu from cache:", dir)
		return scriptPathCache, nil
	}

	autoloadPath, err := env.RunCommand("nu", "-c", "$nu.data-dir | path join vendor autoload")
	if err != nil || autoloadPath == "" {
		log.Error(err)
		return "", err
	}

	log.Debug("autoload path for nu:", autoloadPath)

	// create the path if non-existent
	_, err = os.Stat(autoloadPath)
	if err != nil {
		log.Debug("autoload path does not exist, creating")
		err = os.MkdirAll(autoloadPath, 0o700)
	}

	if err != nil {
		log.Debugf("failed to create autoload dir %s: %s", autoloadPath, err)
		return "", err
	}

	cache.Set(cache.Device, autoloadDir, autoloadPath, cache.INFINITE)
	scriptPathCache = filepath.Join(autoloadPath, fileName)
	log.Debug("script path for nu:", scriptPathCache)
	return scriptPathCache, nil
}
