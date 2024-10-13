package cache

import (
	"os"
	"path/filepath"
	"strings"
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

		if !strings.HasPrefix(file.Name(), FileName) {
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

		if info.ModTime().AddDate(0, 0, 7).After(info.ModTime()) {
			continue
		}

		deleteFile(file.Name())
	}

	return removed, nil
}
