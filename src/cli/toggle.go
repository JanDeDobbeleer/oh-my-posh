package cli

import (
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/spf13/cobra"
)

// toggleCmd represents the toggle command
var toggleCmd = &cobra.Command{
	Use:   "toggle segment1 segment2 ...",
	Short: "Toggle one or more segments on/off",
	Long:  "Toggle one or more segments on/off on the fly. Multiple segments can be specified separated by spaces.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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

		togglesCache, _ := cache.Get[string](cache.Session, cache.TOGGLECACHE)
		var currentToggles []string
		if len(togglesCache) != 0 {
			currentToggles = strings.Split(togglesCache, ",")
		}

		segmentsToToggle := parseSegments(args)

		// Get current toggles as a set for efficient operations
		currentToggleSet := make(map[string]bool)
		for _, toggle := range currentToggles {
			currentToggleSet[toggle] = true
		}

		// Toggle segments: remove if present, add if not present
		for _, segment := range segmentsToToggle {
			if currentToggleSet[segment] {
				delete(currentToggleSet, segment)
				continue
			}

			currentToggleSet[segment] = true
		}

		// Convert back to slice
		newToggles := make([]string, 0, len(currentToggleSet))
		for segment := range currentToggleSet {
			newToggles = append(newToggles, segment)
		}

		cache.Set(cache.Session, cache.TOGGLECACHE, strings.Join(newToggles, ","), cache.INFINITE)
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
