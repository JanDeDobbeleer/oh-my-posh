package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwift(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{Case: "Swift 5.5.3", ExpectedString: "5.5.3", Version: "Swift version 5.5.3 (swift-5.5.3-RELEASE)"},
		{Case: "Swift 5.5.3 on Windows", ExpectedString: "5.5.3", Version: "compnerd.org Swift version 5.5.3 (swift-5.5.3-RELEASE)"},
		{Case: "Swift 5.5.3 on Mac", ExpectedString: "5.5.3", Version: "Apple Swift version 5.5.3 (swift-5.5.3-RELEASE)"},
		{Case: "Swift 5.5", ExpectedString: "5.5", Version: "Swift version 5.5 (swift-5.5-RELEASE)"},
		{Case: "Swift 5.6-dev", ExpectedString: "5.6-dev", Version: "Swift version 5.6-dev (LLVM 62b900d3d0d5be9, Swift ce64fe8867792d4)"},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "swift",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.swift",
		}
		env, props := getMockedLanguageEnv(params)
		s := &Swift{}
		s.Init(props, env)
		assert.True(t, s.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
