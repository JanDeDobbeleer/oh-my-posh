//go:build !windows

package cache

import "os"

func platformCachePath() (string, bool) {
	if cachePath, OK := returnOrBuildCachePath(os.Getenv("XDG_CACHE_HOME")); OK {
		return cachePath, true
	}

	return "", false
}

func PackageFamilyName() (string, bool) {
	return "", false
}
