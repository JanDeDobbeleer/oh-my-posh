package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRust(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Rust 1.64.0", ExpectedString: "1.64.0", Version: "rustc 1.64.0"},
		{Case: "Rust 1.53.0", ExpectedString: "1.53.0", Version: "rustc 1.53.0 (4369396ce 2021-04-27)"},
		{Case: "Rust 1.66.0", ExpectedString: "1.66.0-nightly", Version: "rustc 1.66.0-nightly (01af5040f 2022-10-04)"},
		{
			Case:           "Toolchain not installed",
			ExpectedString: "1.81.0",
			Version: ` info: syncing channel updates for '1.81.0-x86_64-pc-windows-msvc'
    info: latest update on 2024-09-05, rust version 1.81.0 (eeb90cda1 2024-09-04)
    info: downloading component 'cargo'
    info: downloading component 'clippy'
    info: downloading component 'rust-analyzer'
    info: downloading component 'rust-src'
    info: downloading component 'rust-std'
    info: downloading component 'rustc'
    info: downloading component 'rustfmt'
    info: installing component 'cargo'
    info: installing component 'clippy'
    info: installing component 'rust-analyzer'
    info: installing component 'rust-src'
    info: installing component 'rust-std'
    info: installing component 'rustc'`,
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "rustc",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.rs",
		}
		env, props := getMockedLanguageEnv(params)
		r := &Rust{}
		r.Init(props, env)
		assert.True(t, r.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, r.Template(), r), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
