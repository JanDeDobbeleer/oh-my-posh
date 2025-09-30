package cache

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

// Clear removes cache files from the cache directory.
//
// If force is true, the entire cache directory is removed.
// If force is false, only cache files older than 7 days that match certain patterns are deleted.
// The excludedFiles parameter allows you to specify file names that should not be deleted,
// even if they would otherwise be eligible for removal.
func Clear(force bool, excludedFiles ...string) error {
	defer log.Trace(time.Now())

	if force {
		return os.RemoveAll(Path())
	}

	// get all files in the cache directory that start with omp.cache and delete them
	files, err := os.ReadDir(Path())
	if err != nil {
		return err
	}

	// get all log files as well
	if logFiles, err := os.ReadDir(filepath.Join(Path(), "logs")); err == nil {
		files = append(files, logFiles...)
	}

	shouldSkip := func(fileName string) bool {
		if slices.Contains(excludedFiles, fileName) {
			return true
		}

		return strings.EqualFold(fileName, DeviceStore) || strings.HasPrefix(fileName, "init.")
	}

	if len(excludedFiles) > 0 {
		log.Debug("excluding files from deletion:", strings.Join(excludedFiles, ", "))
	}

	deleteFile := func(file string) {
		path := filepath.Join(Path(), file)
		err := os.Remove(path)
		if err != nil {
			log.Error(err)
			return
		}

		log.Debugf("removed cache file: %s", path)
	}

	cacheTTL := GetTTL()

	log.Debugf("removing cache files older than %d days", cacheTTL)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if shouldSkip(file.Name()) {
			log.Debug("skipping excluded file:", file.Name())
			continue
		}

		cacheFileInfo, err := file.Info()
		if err != nil {
			log.Debug("skipping file, cannot get info:", file.Name())
			continue
		}

		if cacheFileInfo.ModTime().After(time.Now().AddDate(0, 0, -cacheTTL)) {
			log.Debug("skipping recently used file:", file.Name())
			continue
		}

		deleteFile(file.Name())
	}

	return nil
}

func GetTTL() int {
	cacheTTL, OK := Get[int](Device, TTL)
	if !OK || cacheTTL <= 0 {
		cacheTTL = 7
	}

	return cacheTTL
}
