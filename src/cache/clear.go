package cache

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

func Clear(cachePath string, force bool) ([]string, error) {
	// get all files in the cache directory that start with omp.cache and delete them
	files, err := os.ReadDir(cachePath)
	if err != nil {
		return []string{}, err
	}

	var removed []string

	deleteFile := func(file string) {
		path := filepath.Join(cachePath, file)
		if err := os.Remove(path); err == nil {
			removed = append(removed, path)
		}
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasPrefix(file.Name(), FileName) && !strings.HasPrefix(file.Name(), "init.") {
			continue
		}

		if force {
			deleteFile(file.Name())
			continue
		}

		// don't delete the system cache file unless forced
		if file.Name() == FileName {
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

	deletedLogs := deleteLogs(force)
	if len(deletedLogs) > 0 {
		removed = append(removed, deletedLogs...)
	}

	return removed, nil
}

func deleteLogs(force bool) []string {
	var removed []string

	home, err := os.UserHomeDir()
	if err != nil {
		log.Error(err)
		return removed
	}

	logPath := filepath.Join(home, ".oh-my-posh")

	deleteFile := func(file string) {
		path := filepath.Join(logPath, file)
		if err := os.Remove(path); err == nil {
			removed = append(removed, path)
		}
	}

	logFiles, err := os.ReadDir(logPath)
	if err != nil {
		log.Error(err)
		return removed
	}

	for _, file := range logFiles {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".log") {
			continue
		}

		if force {
			deleteFile(file.Name())
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

	return removed
}
