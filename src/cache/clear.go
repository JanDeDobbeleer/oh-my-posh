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

	canDelete := func(fileName string) bool {
		if slices.Contains(excludedFiles, fileName) {
			return false
		}

		return strings.EqualFold(fileName, FileName) || strings.HasPrefix(fileName, "init.")
	}

	deleteFile := func(file string) {
		path := filepath.Join(Path(), file)
		if err := os.Remove(path); err == nil {
			log.Debugf("removed cache file: %s", path)
		}
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !canDelete(file.Name()) {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(time.Now().AddDate(0, 0, -7)) {
			continue
		}

		deleteFile(file.Name())
	}

	return nil
}
