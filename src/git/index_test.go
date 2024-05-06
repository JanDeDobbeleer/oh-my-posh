package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIndex(t *testing.T) {
	idx, err := Index("/Users/jan/Code/oh-my-posh")
	for _, entry := range idx.Entries {
		t.Logf("%+v", entry)
	}
	assert.NoError(t, err)
}
