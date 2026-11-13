package runtime

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// cache.Path resolves once per process, so the cache directory must be set
	// before any test runs, or a test that persists the cache writes to the real one.
	dir, err := os.MkdirTemp("", "omp-runtime-test")
	if err != nil {
		panic(err)
	}

	os.Setenv("OMP_CACHE_DIR", dir)

	code := m.Run()

	os.RemoveAll(dir)
	os.Exit(code)
}
