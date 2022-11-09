package cmd

import (
	"bytes"
	"os/exec"
	"strings"
)

func Run(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
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
