package template

import (
	"testing"

	"github.com/Masterminds/sprig/v3"
)

// TestRestrictedAllowedSprigFuncsCoversAllSprigFuncs is the drift guard: a
// sprig upgrade that adds a new function must fail this test rather than
// silently expose that function to untrusted templates. Every sprig function
// must be triaged into exactly one of dangerousFuncs or
// restrictedAllowedSprigFuncs before it can be used at all — unless it's
// shadowed by an oh-my-posh local function of the same name (localFuncMap),
// which always wins over the sprig original and so is never reachable.
func TestRestrictedAllowedSprigFuncsCoversAllSprigFuncs(t *testing.T) {
	localOverrides := localFuncMap()

	for name := range sprig.TxtFuncMap() {
		if dangerousFuncs[name] || restrictedAllowedSprigFuncs[name] {
			continue
		}

		if _, ok := localOverrides[name]; ok {
			continue
		}

		t.Errorf("sprig function %q is untriaged: add it to dangerousFuncs if it does I/O, network, "+
			"or heavy-CPU work, or to restrictedAllowedSprigFuncs if it's confirmed safe for untrusted input", name)
	}
}
