package environment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalHostName(t *testing.T) {
	hostName := "hello"
	assert.Equal(t, hostName, cleanHostName(hostName))
}

func TestHostNameWithLocal(t *testing.T) {
	hostName := "hello.local"
	assert.Equal(t, "hello", cleanHostName(hostName))
}

func TestHostNameWithLan(t *testing.T) {
	hostName := "hello.lan"
	cleanHostName := cleanHostName(hostName)
	assert.Equal(t, "hello", cleanHostName)
}

func TestWindowsPathWithDriveLetter(t *testing.T) {
	cases := []struct {
		Case     string
		CWD      string
		Expected string
	}{
		{Case: "C drive", CWD: `C:\Windows\`, Expected: `C:\Windows\`},
		{Case: "C drive lower case", CWD: `c:\Windows\`, Expected: `C:\Windows\`},
		{Case: "P drive lower case", CWD: `p:\some\`, Expected: `P:\some\`},
		{Case: "some drive lower case", CWD: `some:\some\`, Expected: `some:\some\`},
		{Case: "drive ending in c:", CWD: `src:\source\`, Expected: `src:\source\`},
		{Case: "registry drive", CWD: `HKLM:\SOFTWARE\magnetic:test\`, Expected: `HKLM:\SOFTWARE\magnetic:test\`},
	}
	for _, tc := range cases {
		env := &ShellEnvironment{
			CmdFlags: &Flags{
				PWD: tc.CWD,
			},
		}
		assert.Equal(t, env.Pwd(), tc.Expected)
	}
}
