package config

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

func TestTimeoutToCache(t *testing.T) {
	cases := []struct {
		Expected *cache.Config
		Case     string
		Timeout  int
	}{
		{Case: "No timeout set"},
		{Case: "Timeout set to 0", Timeout: 0},
		{Case: "Timeout set to 10", Timeout: 10, Expected: &cache.Config{Duration: "10m0s", Strategy: cache.Folder}},
	}

	for _, tc := range cases {
		segment := &Segment{
			Properties: properties.Map{
				cacheTimeout: tc.Timeout,
			},
		}

		segment.timeoutToCache()
		assert.Equal(t, tc.Expected, segment.Cache, tc.Case)
	}
}

func TestMigrateFolders(t *testing.T) {
	cases := []struct {
		Case    string
		Folders []string
	}{
		{Case: "No folders set"},
		{Case: "Empty folders", Folders: []string{}},
		{Case: "Folders set", Folders: []string{"/super/secret/project"}},
	}

	for _, tc := range cases {
		segment := &Segment{
			Properties: properties.Map{
				excludeFolders: tc.Folders,
			},
		}

		got := segment.migrateFolders(excludeFolders)
		assert.Equal(t, tc.Folders, got, tc.Case)
	}
}
