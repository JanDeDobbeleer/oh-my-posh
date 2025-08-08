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
		return "", false
	}

	path, err := scriptPath(env)
	if err != nil {
		return "", false
	}

	_, err = os.Stat(path)
	if err != nil {
		return "", false
	}

	// check if we have the same context
	if hash, _ := env.Cache().Get(cacheKey(env)); hash != scriptName(env) {
		return "", false
	}

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

	_ = f.Close()

	env.Cache().Set(cacheKey(env), scriptName(env), cache.INFINITE)

	defer func() {
		parent := filepath.Dir(path)
		purgeScripts(env, parent)
	}()

	return path, nil
}

func cacheKey(env runtime.Environment) string {
	return fmt.Sprintf("INITVERSION%s", strings.ToUpper(env.Flags().Shell))
}

func fileName(env runtime.Environment) string {
	return fmt.Sprintf("init.%s.%s", build.Version, env.Flags().ConfigHash)
}

func scriptName(env runtime.Environment) string {
	extension := env.Flags().Shell

	switch env.Flags().Shell {
	case PWSH, PWSH5:
		extension = "ps1"
	case CMD:
		extension = "lua"
	case BASH:
		extension = "sh"
	case ELVISH:
		extension = "elv"
	case XONSH:
		extension = "xsh"
	}

	return fmt.Sprintf("%s.%s", fileName(env), extension)
}

func scriptPath(env runtime.Environment) (string, error) {
	if len(scriptPathCache) != 0 {
		return scriptPathCache, nil
	}

	if env.Flags().Shell != NU {
		scriptPathCache = filepath.Join(cache.Path(), scriptName(env))
		log.Debug("init script path for non-nu shell:", scriptPathCache)
		return scriptPathCache, nil
	}

	const autoloadDir = "NUAUTOLOADDIR"

	if dir, OK := env.Cache().Get(autoloadDir); OK {
		scriptPathCache = filepath.Join(dir, scriptName(env))
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
	scriptPathCache = filepath.Join(path, scriptName(env))
	log.Debug("script path for nu:", scriptPathCache)
	return scriptPathCache, nil
}

func purgeScripts(env runtime.Environment, path string) {
	current := fileName(env)

	files, err := os.ReadDir(path)
	if err != nil {
		log.Error(err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), "init.") || strings.HasPrefix(file.Name(), current) {
			continue
		}

		if err := os.Remove(filepath.Join(path, file.Name())); err != nil {
			log.Debugf("failed to remove init script %s: %s", file.Name(), err)
			continue
		}

		log.Debug("removed init script:", file.Name())
	}
}
