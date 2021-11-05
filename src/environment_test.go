package main

import (
	"flag"
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

func TestPathWithDriveLetter(t *testing.T) {
	args := &args{
		Debug: flag.Bool(
			"debug",
			false,
			"Print debug information"),
		PWD: flag.String(
			"pwd",
			"",
			"the path you are working in"),
	}

	testPath := func(t *testing.T, inputPath string, expectedResult string)  {
		env := &environment{}
		env.init(args)
		env.args.PWD = &inputPath
		assert.Equal(t, env.getcwd(), expectedResult)
	}
	testPath(t, `C:\Windows\`, `C:\Windows\`)
	testPath(t, `c:\Windows\`, `C:\Windows\`)
	testPath(t, `p:\other\`, `P:\other\`)
	testPath(t, `other:\other\`, `other:\other\`)
	testPath(t, `src:\source\`, `src:\source\`)
	testPath(t, `HKLM:\SOFTWARE\magnetic:test\`, `HKLM:\SOFTWARE\magnetic:test\`)
}

