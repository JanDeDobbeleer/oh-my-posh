package template

import (
	"os"
	"path/filepath"
)

func glob(pattern string) (bool, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}

func readFile(path string) string {
	content, _ := os.ReadFile(path)
	return string(content)
}
