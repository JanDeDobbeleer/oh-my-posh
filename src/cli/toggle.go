package cli

import (
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cmdtree"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

var toggleCmd = &cmdtree.Command{
	Use:   "toggle segment1 segment2 ...",
	Short: "Toggle one or more segments on/off",
	Long:  "Toggle one or more segments on/off on the fly. Multiple segments can be specified separated by spaces.",
	Args:  cmdtree.MinimumNArgs(1),
	Run: func(cmd *cmdtree.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		env := &runtime.Terminal{}
		env.Init(&runtime.Flags{})

		cache.Init(os.Getenv("POSH_SHELL"), cache.Persist)

		defer func() {
			cache.Close()
		}()

		// Get current toggles from cache as a map
		currentToggleSet, _ := cache.Session.Get[map[string]bool](cache.TOGGLECACHE)
		if currentToggleSet == nil {
			currentToggleSet = make(map[string]bool)
		}

		segmentsToToggle := parseSegments(args)

		// Toggle segments: remove if present, add if not present
		for _, segment := range segmentsToToggle {
			if currentToggleSet[segment] {
				delete(currentToggleSet, segment)
				continue
			}

			currentToggleSet[segment] = true
		}

		// Store the map directly in cache
		cache.Session.Set(cache.TOGGLECACHE, currentToggleSet, cache.INFINITE)
	},
}

func parseSegments(args []string) []string {
	var segments []string
	for _, arg := range args {
		if segment := strings.TrimSpace(arg); segment != "" {
			segments = append(segments, segment)
		}
	}

	return segments
}

func init() {
	RootCmd.AddCommand(toggleCmd)
}
