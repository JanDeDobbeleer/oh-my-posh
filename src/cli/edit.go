package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func editFileWithEditor(file string) int {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		fmt.Println(`no editor specified in the environment variable "EDITOR"`)
		return 1
	}

	editor = strings.TrimSpace(editor)
	args := strings.Split(editor, " ")
	editor = args[0]
	args = append(args[1:], file)

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, editor, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println(err.Error())
		return 1
	}

	return 0
}
