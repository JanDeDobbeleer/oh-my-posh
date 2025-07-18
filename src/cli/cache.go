package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCache = &cobra.Command{
	Use:   "cache [path|clear|edit|refresh-segment]",
	Short: "Interact with the oh-my-posh cache",
	Long: `Interact with the oh-my-posh cache.

You can do the following:

- path: list cache path
- clear: remove all cache values
- edit: edit cache values
- refresh-segment: refresh a specific segment cache`,
	ValidArgs: []string{
		"path",
		"clear",
		"edit",
		"refresh-segment",
	},
	Args: NoArgsOrOneValidArg,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		switch args[0] {
		case "path":
			fmt.Println(cache.Path())
		case "clear":
			deletedFiles, err := cache.Clear(cache.Path(), true)
			if err != nil {
				fmt.Println(err)
				return
			}

			for _, file := range deletedFiles {
				fmt.Println("removed cache file:", file)
			}
		case "edit":
			cacheFilePath := filepath.Join(cache.Path(), cache.FileName)
			exitcode = editFileWithEditor(cacheFilePath)
		case "refresh-segment":
			refreshSegmentCache(cmd)
		}
	},
}

// refreshSegmentCache refreshes the cache for a specific segment
func refreshSegmentCache(cmd *cobra.Command) {
	segmentName, _ := cmd.Flags().GetString("segment")
	cacheKey, _ := cmd.Flags().GetString("cache-key")
	workingDir, _ := cmd.Flags().GetString("working-dir")
	
	if segmentName == "" || cacheKey == "" || workingDir == "" {
		fmt.Println("Error: segment, cache-key, and working-dir are required")
		return
	}
	
	// Change to the working directory
	if err := os.Chdir(workingDir); err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
		return
	}
	
	// Create runtime environment
	flags := &runtime.Flags{
		SaveCache: true,
	}
	env := &runtime.Terminal{}
	env.Init(flags)
	defer env.Close()
	
	// Load configuration
	cfg, _ := config.Load(configFlag, shell.GENERIC, false)
	
	// Find the segment configuration
	var segment *config.Segment
	for _, block := range cfg.Blocks {
		for _, s := range block.Segments {
			if s.Name() == segmentName || (s.Alias != "" && s.Alias == segmentName) {
				segment = s
				break
			}
		}
		if segment != nil {
			break
		}
	}
	
	if segment == nil {
		fmt.Printf("Error: segment %s not found in configuration\n", segmentName)
		return
	}
	
	// Execute the segment to get fresh data
	segment.Execute(env)
	
	// Cache the result
	if segment.Enabled {
		asyncCache := cache.NewAsyncSegmentCache(env.Cache())
		asyncData := &cache.AsyncSegmentData{
			Text:      segment.Text(),
			Enabled:   segment.Enabled,
			Timestamp: time.Now(),
			Duration:  cache.Duration("5m"), // Default 5 minutes
		}
		asyncCache.SetSegmentData(segmentName, cacheKey, asyncData)
		fmt.Printf("Async cache refreshed for segment: %s\n", segmentName)
	}
}

func init() {
	getCache.Flags().StringP("segment", "s", "", "segment name for refresh-segment command")
	getCache.Flags().StringP("cache-key", "k", "", "cache key for refresh-segment command")
	getCache.Flags().StringP("working-dir", "w", "", "working directory for refresh-segment command")
	RootCmd.AddCommand(getCache)
}
