package cmd

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

// Run is used to correctly run a command with a timeout.
func Run(command string, args ...string) (string, error) {
	// set a timeout of 4 seconds
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	cmdErr := cmd.Run()
	if cmdErr != nil {
		output := err.String()
		return output, cmdErr
	}
	// some silly commands return 0 and the output is in stderr instead of stdout
	result := out.String()
	if len(result) == 0 {
		result = err.String()
	}
	output := strings.TrimSpace(result)
	return output, nil
}
