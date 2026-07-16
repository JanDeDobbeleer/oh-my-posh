package shell

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/stretchr/testify/assert"
)

func TestStrictUsesSeparateInitScript(t *testing.T) {
	defaultFlags := &runtime.Flags{Shell: BASH, ConfigPath: "default.omp.json"}
	strictFlags := &runtime.Flags{Shell: BASH, ConfigPath: "default.omp.json", Strict: true}

	assert.NotEqual(t, cacheKey(defaultFlags), cacheKey(strictFlags))
	assert.NotEqual(t, InitScriptName(defaultFlags), InitScriptName(strictFlags))
	assert.Contains(t, InitScriptName(strictFlags), ".strict.sh")
}
