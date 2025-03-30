package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/spf13/cobra"
)

var (
	configFlag   string
	shellName    string
	printVersion bool
	trace        bool
	exitcode     int

	// for internal use only
	silent bool

	// deprecated
	initialize bool
)

var RootCmd = &cobra.Command{
	Use:   "oh-my-posh",
	Short: "oh-my-posh is a tool to render your prompt",
	Long: `oh-my-posh is a cross platform tool to render your prompt.
It can use the same configuration everywhere to offer a consistent
experience, regardless of where you are. For a detailed guide
on getting started, have a look at the docs at https://ohmyposh.dev`,
	Run: func(cmd *cobra.Command, _ []string) {
		if initialize {
			runInit(strings.ToLower(shellName))
			return
		}
		if printVersion {
			fmt.Println(build.Version)
			return
		}

		_ = cmd.Help()
	},
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		traceEnv := os.Getenv("POSH_TRACE")
		if len(traceEnv) == 0 {
			return
		}

		trace = true

		log.Enable()
		log.Plain()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		defer func() {
			if exitcode != 0 {
				os.Exit(exitcode)
			}
		}()

		if !trace {
			return
		}

		timestamp := time.Now().Format("20060102T150405.000")
		cli := append([]string{cmd.Name()}, args...)
		filename := fmt.Sprintf("%s-%s.log", timestamp, strings.Join(cli, "-"))

		home, err := os.UserHomeDir()
		if err != nil {
			return
		}

		logPath := filepath.Join(home, ".oh-my-posh")
		err = os.MkdirAll(logPath, 0755)
		if err != nil {
			return
		}

		err = os.WriteFile(filepath.Join(logPath, filename), []byte(log.String()), 0644)
		if err != nil {
			return
		}
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// software error
		os.Exit(70)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "config file path")
	RootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "do not print anything")
	RootCmd.Flags().BoolVar(&printVersion, "version", false, "print the version number and exit")

	// Deprecated flags, should be kept to avoid breaking CLI integration.
	RootCmd.Flags().BoolVarP(&initialize, "init", "i", false, "init")
	RootCmd.Flags().StringVarP(&shellName, "shell", "s", "", "shell")

	// Hide flags that are deprecated or for internal use only.
	_ = RootCmd.PersistentFlags().MarkHidden("silent")

	// Disable completions
	RootCmd.CompletionOptions.DisableDefaultCmd = true
}
