package dsc

import (
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

func Run[T State](command, state string) error {
	env := &runtime.Terminal{}
	env.Init(&runtime.Flags{})
	defer env.Close()

	var err error

	switch command {
	case "get", "export":
		fmt.Print(export[T](env.Cache()))
	case "set":
		if state == "" {
			err = newError("please provide a state configuration to set")
			break
		}

		err = set[T](env.Cache(), state)
	case "schema":
		var state T
		fmt.Print(state.Schema())
	case "test":
		if state == "" {
			err = newError("please provide a state configuration to test")
			break
		}

		err = test[T](env.Cache(), state)
	}

	return err
}
