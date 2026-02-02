package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"
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

		segmentsToToggle := parseSegments(args)

		if ipc.SocketExists() {
			client, err := daemon.NewClient()
			if err != nil {
				fmt.Printf("daemon error: %v\n", err)
				return
			}
			defer client.Close()
			ctx, cancel := context.WithTimeout(context.Background(), renderTimeout)
			defer cancel()
			if err := client.ToggleSegment(ctx, os.Getppid(), segmentsToToggle); err != nil {
				fmt.Printf("daemon error: %v\n", err)
			}
			return
		}

		env := &runtime.Terminal{}
		env.Init(&runtime.Flags{})

		cache.Init(os.Getenv("POSH_SHELL"), cache.Persist)

		defer func() {
			cache.Close()
		}()

		// Get current toggles from cache as a map
		currentToggleSet, _ := cache.Get[map[string]bool](cache.Session, cache.TOGGLECACHE)
		if currentToggleSet == nil {
			currentToggleSet = make(map[string]bool)
		}

		// Toggle segments: remove if present, add if not present
		for _, segment := range segmentsToToggle {
			if currentToggleSet[segment] {
				delete(currentToggleSet, segment)
				continue
			}

			currentToggleSet[segment] = true
		}

		// Store the map directly in cache
		cache.Set(cache.Session, cache.TOGGLECACHE, currentToggleSet, cache.INFINITE)
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
