package config

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestTimeoutToCache(t *testing.T) {
	cases := []struct {
		Expected *Cache
		Case     string
		Timeout  int
	}{
		{Case: "No timeout set"},
		{Case: "Timeout set to 0", Timeout: 0},
		{Case: "Timeout set to 10", Timeout: 10, Expected: &Cache{Duration: "10m0s", Strategy: Folder}},
	}

	for _, tc := range cases {
		segment := &Segment{
			Options: options.Map{
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
			Options: options.Map{
				excludeFolders: tc.Folders,
			},
		}

		got := segment.migrateFolders(excludeFolders)
		assert.Equal(t, tc.Folders, got, tc.Case)
	}
}
