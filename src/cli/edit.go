package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func editFileWithEditor(file string) int {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if len(editor) == 0 {
		fmt.Println(`no editor specified in the environment variable "EDITOR"`)
		return 1
	}

	var args []string
	if strings.Contains(editor, " ") {
		strs := strings.Split(editor, " ")
		editor = strs[0]
		args = strs[1:]
	}

	args = append(args, file)
	cmd := exec.Command(editor, args...)

	err := cmd.Run()
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	return cmd.ProcessState.ExitCode()
}
