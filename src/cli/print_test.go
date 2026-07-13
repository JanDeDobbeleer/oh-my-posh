package cli

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/prompt"

	"github.com/stretchr/testify/assert"
)

func TestPrintCommandTransientRight(t *testing.T) {
	cmd := createPrintCmd()

	assert.Contains(t, cmd.ValidArgs, prompt.TRANSIENT_RIGHT)
}
