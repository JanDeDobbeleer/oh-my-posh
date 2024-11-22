package segments

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFortran(t *testing.T) {
	cases := []struct {
		Case           string
		ExpectedString string
		Version        string
	}{
		{
			Case:           "GNU Fortran 10.2.1 Debian",
			ExpectedString: "10.2.1",
			Version: `GNU Fortran (Debian 10.2.1-6) 10.2.1 20210110
			Copyright (C) 2020 Free Software Foundation, Inc.
			This is free software; see the source for copying conditions.  There is NO
			warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.`,
		},
		{
			Case:           "GNU Fortran 11.4.0 Ubuntu",
			ExpectedString: "11.4.0",
			Version: `GNU Fortran (Ubuntu 11.4.0-1ubuntu1~22.04) 11.4.0
			Copyright (C) 2021 Free Software Foundation, Inc.
			This is free software; see the source for copying conditions.  There is NO
			warranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.`,
		},
	}
	for _, tc := range cases {
		params := &mockedLanguageParams{
			cmd:           "gfortran",
			versionParam:  "--version",
			versionOutput: tc.Version,
			extension:     "*.f",
		}
		env, props := getMockedLanguageEnv(params)
		f := &Fortran{}
		f.Init(props, env)
		assert.True(t, f.Enabled(), fmt.Sprintf("Failed in case: %s", tc.Case))
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, f.Template(), f), fmt.Sprintf("Failed in case: %s", tc.Case))
	}
}
