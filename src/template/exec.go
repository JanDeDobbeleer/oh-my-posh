package template

import "strings"

func cmd(command string, args ...string) (string, error) {
	output, err := env.RunCommand(command, args...)
	return strings.TrimSpace(output), err
}
