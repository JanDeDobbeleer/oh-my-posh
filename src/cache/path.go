package cache

import (
	"os"
	"path/filepath"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
)

var cachePath string

func realPath(filePath string) string {
	canonicalPath, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		// If we can't resolve symlinks, return the original path
		return filePath
	}

	return canonicalPath
}

func Path() string {
	defer log.Trace(time.Now())

	if cachePath != "" {
		return cachePath
	}

	returnOrBuildCachePath := func(input string) (string, bool) {
		// validate root path
		if _, err := os.Stat(input); err != nil {
			return "", false
		}

		// validate oh-my-posh folder, if non existent, create it
		cachePath := filepath.Join(input, "oh-my-posh")
		if _, err := os.Stat(cachePath); err == nil {
			return realPath(cachePath), true
		}

		if err := os.Mkdir(cachePath, 0o755); err != nil {
			return "", false
		}

		return realPath(cachePath), true
	}

	var OK bool

	// allow the user to set the cache path using OMP_CACHE_DIR
	if cachePath, OK = returnOrBuildCachePath(os.Getenv("OMP_CACHE_DIR")); OK {
		return cachePath
	}

	// WINDOWS cache folder, should not exist elsewhere
	if cachePath, OK = returnOrBuildCachePath(os.Getenv("LOCALAPPDATA")); OK {
		return cachePath
	}

	// get XDG_CACHE_HOME if present
	if cachePath, OK = returnOrBuildCachePath(os.Getenv("XDG_CACHE_HOME")); OK {
		return cachePath
	}

	// try to create the cache folder in the user's home directory if non-existent
	dotCache := filepath.Join(path.Home(), ".cache")
	if _, err := os.Stat(dotCache); err != nil {
		_ = os.Mkdir(dotCache, 0o755)
	}

	// HOME cache folder
	if cachePath, OK = returnOrBuildCachePath(dotCache); OK {
		return cachePath
	}

	cachePath = realPath(path.Home())
	return cachePath
}
