package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/daemon"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"

	"github.com/spf13/cobra"
)

var (
	foreground bool

	daemonCmd = &cobra.Command{
		Use:   "daemon [start|status|serve|log]",
		Short: "Manage the oh-my-posh daemon",
		Long: `Manage the oh-my-posh daemon for faster prompt rendering.

The daemon runs in the background and renders prompt segments asynchronously.
It automatically shuts down after being idle (no connections) for 5 minutes.

  - start:  Start the daemon (detached)
  - status: Check if the daemon is running
  - serve:  Run the daemon server (foreground, silent by default)
  - log:    Enable/disable daemon logging (log <path> to enable, log off to disable)`,
		ValidArgs: []string{
			"start",
			"status",
			"serve",
			"log",
		},
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				_ = cmd.Help()
				return
			}

			switch args[0] {
			case "start":
				startDaemon()
			case "status":
				daemonStatus()
			case "serve":
				silent = true
				runDaemonServe()
			case "log":
				daemonLog(args[1:])
			default:
				_ = cmd.Help()
			}
		},
	}
)

func init() {
	daemonCmd.Flags().BoolVar(&foreground, "foreground", false, "run daemon in foreground (for debugging)")
	RootCmd.AddCommand(daemonCmd)
}

func startDaemon() {
	// Check if already running
	if daemon.IsRunning() {
		fmt.Println("daemon is already running")
		return
	}

	if foreground {
		// Enable logging to stderr for debugging
		log.Enable(false)
		runDaemonServe()
		return
	}

	if err := startDetachedDaemon(); err != nil {
		log.Error(err)
		fmt.Fprintln(os.Stderr, "failed to start daemon:", err)
		exitcode = 1
	}
}

func runDaemonServe() {
	if configFlag != "" {
		configFlag = path.ReplaceTildePrefixWithHomeDir(configFlag)
		if abs, err := filepath.Abs(configFlag); err == nil {
			configFlag = abs
		}
	}

	d, err := daemon.New(configFlag)
	if err != nil {
		log.Error(err)
		fmt.Fprintln(os.Stderr, "failed to start daemon:", err)
		exitcode = 1
		return
	}

	log.Debug("daemon started")

	if err := d.Start(); err != nil {
		log.Error(err)
		fmt.Fprintln(os.Stderr, "daemon error:", err)
		exitcode = 1
	}
}

func daemonLog(args []string) {
	if len(args) == 0 {
		fmt.Println("usage: daemon log <file_path>  (enable logging)")
		fmt.Println("       daemon log off          (disable logging)")
		return
	}

	client, err := daemon.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "daemon is not running")
		exitcode = 1
		return
	}
	defer client.Close()

	ctx := context.Background()
	logPath := args[0]

	if logPath == "off" {
		logPath = ""
	}

	if err := client.SetLogging(ctx, logPath); err != nil {
		fmt.Fprintln(os.Stderr, "failed to set logging:", err)
		exitcode = 1
		return
	}

	if logPath == "" {
		fmt.Println("daemon logging disabled")
	} else {
		fmt.Println("daemon logging to", logPath)
	}
}

func daemonStatus() {
	if daemon.IsRunning() {
		fmt.Println("daemon is running")
	} else {
		fmt.Println("daemon is not running")
	}
}
