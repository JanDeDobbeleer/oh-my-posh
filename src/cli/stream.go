package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/spf13/cobra"
)

const (
	partialTimeout = 100 * time.Millisecond
)

// StreamRequest represents a request to render a prompt
type StreamRequest struct {
	ID    string       `json:"id"`
	Flags RequestFlags `json:"flags"`
}

// RequestFlags contains all the flags needed to render a prompt
type RequestFlags struct {
	ConfigPath    string  `json:"config"`
	Shell         string  `json:"shell"`
	ShellVersion  string  `json:"shell_version"`
	PWD           string  `json:"pwd"`
	PSWD          string  `json:"pswd"`
	Status        int     `json:"status"`
	NoStatus      bool    `json:"no_status"`
	PipeStatus    string  `json:"pipestatus"`
	ExecutionTime float64 `json:"execution_time"`
	TerminalWidth int     `json:"terminal_width"`
	StackCount    int     `json:"stack_count"`
	JobCount      int     `json:"job_count"`
	Cleared       bool    `json:"cleared"`
	Column        int     `json:"column"`
}

// StreamResponse represents a response with rendered prompts
type StreamResponse struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"` // "update" or "complete"
	Prompts map[string]string `json:"prompts"`
	Error   string            `json:"error,omitempty"`
}

var streamCmd = createStreamCmd()

func init() {
	RootCmd.AddCommand(streamCmd)
}

func createStreamCmd() *cobra.Command {
	streamCmd := &cobra.Command{
		Use:    "stream",
		Short:  "Run in streaming mode for async prompt rendering",
		Long:   "Run oh-my-posh in streaming mode, reading JSON requests from stdin and writing JSON responses to stdout. Used for async prompt rendering in shells like NuShell.",
		Hidden: true, // Hidden as it's an internal command
		Run: func(cmd *cobra.Command, args []string) {
			runStream()
		},
	}

	return streamCmd
}

func runStream() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		var req StreamRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(req.ID, fmt.Sprintf("failed to parse request: %s", err.Error()))
			continue
		}

		processRequest(&req)
	}

	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
}

func processRequest(req *StreamRequest) {
	// Set up cache for this request
	cache.Init(req.Flags.Shell, cache.Persist)
	defer cache.Close()

	// Build runtime flags from request
	flags := &runtime.Flags{
		ConfigPath:    req.Flags.ConfigPath,
		Shell:         req.Flags.Shell,
		ShellVersion:  req.Flags.ShellVersion,
		PWD:           req.Flags.PWD,
		PSWD:          req.Flags.PSWD,
		ErrorCode:     req.Flags.Status,
		NoExitCode:    req.Flags.NoStatus,
		PipeStatus:    req.Flags.PipeStatus,
		ExecutionTime: req.Flags.ExecutionTime,
		TerminalWidth: req.Flags.TerminalWidth,
		StackCount:    req.Flags.StackCount,
		JobCount:      req.Flags.JobCount,
		Cleared:       req.Flags.Cleared,
		Column:        req.Flags.Column,
		IsPrimary:     true,
		Type:          prompt.PRIMARY,
	}

	// Create context with timeout for partial rendering
	ctx, cancel := context.WithTimeout(context.Background(), partialTimeout)
	defer cancel()

	// Start rendering in background
	resultChan := make(chan *renderResult, 1)
	go func() {
		resultChan <- renderPrompts(flags)
	}()

	// Wait for either timeout or completion
	select {
	case <-ctx.Done():
		// Timeout reached, send partial update
		// For now, we'll wait for completion since we don't have partial rendering yet
		result := <-resultChan
		sendResponse(req.ID, "complete", result)
	case result := <-resultChan:
		// Rendering completed quickly
		sendResponse(req.ID, "complete", result)
	}

	template.SaveCache()
}

type renderResult struct {
	primary string
	right   string
	err     error
}

func renderPrompts(flags *runtime.Flags) *renderResult {
	result := &renderResult{}

	defer func() {
		if r := recover(); r != nil {
			result.err = fmt.Errorf("panic during rendering: %v", r)
		}
	}()

	// Create engine for primary prompt
	flags.Type = prompt.PRIMARY
	flags.IsPrimary = true
	eng := prompt.New(flags)
	result.primary = eng.Primary()

	// Create engine for right prompt
	flags.Type = prompt.RIGHT
	flags.IsPrimary = false
	eng = prompt.New(flags)
	result.right = eng.RPrompt()

	return result
}

func sendResponse(id, responseType string, result *renderResult) {
	resp := StreamResponse{
		ID:      id,
		Type:    responseType,
		Prompts: make(map[string]string),
	}

	if result.err != nil {
		resp.Error = result.err.Error()
		sendJSON(&resp)
		return
	}

	resp.Prompts["primary"] = result.primary
	resp.Prompts["right"] = result.right

	sendJSON(&resp)
}

func sendError(id, message string) {
	resp := StreamResponse{
		ID:      id,
		Type:    "error",
		Error:   message,
		Prompts: make(map[string]string),
	}
	sendJSON(&resp)
}

func sendJSON(resp *StreamResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Error(err)
		return
	}

	fmt.Println(string(data))
}
