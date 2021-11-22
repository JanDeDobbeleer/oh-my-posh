//go:build windows

package main

import "testing"

//	"testing"

func TestRegQueryGetRegistryFunction(t *testing.T) {

	// cases := []struct {
	// 	CaseDescription     string
	// 	Path                string
	// 	Key                 string
	// 	ExpectedSuccess     bool
	// 	OutputShouldContain string
	// }{
	// 	// calls should fail
	// 	{
	// 		CaseDescription:     "Incorrect HK name",
	// 		Path:                "HKLLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
	// 		Key:                 "ProductName",
	// 		ExpectedSuccess:     false,
	// 		OutputShouldContain: "Error",
	// 	},
	// 	{
	// 		CaseDescription:     "Fwd not backslashes",
	// 		Path:                "HKLM/Software/Microsoft/Windows NT/CurrentVersion",
	// 		Key:                 "ProductName",
	// 		ExpectedSuccess:     false,
	// 		OutputShouldContain: "Error",
	// 	},
	// 	{
	// 		CaseDescription:     "Non-existent path",
	// 		Path:                "HKLM\\Software\\Microsoftoftoft\\Windows NT\\CurrentVersion",
	// 		Key:                 "ProductName",
	// 		ExpectedSuccess:     false,
	// 		OutputShouldContain: "Error",
	// 	},
	// 	{
	// 		CaseDescription:     "Non-existent key",
	// 		Path:                "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
	// 		Key:                 "nonexistentkey",
	// 		ExpectedSuccess:     false,
	// 		OutputShouldContain: "Error",
	// 	},
	// 	{
	// 		CaseDescription:     "QWORD key",
	// 		Path:                "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
	// 		Key:                 "InstallTime",
	// 		ExpectedSuccess:     false,
	// 		OutputShouldContain: "no formatter",
	// 	},
	// 	{
	// 		CaseDescription:     "Binary key",
	// 		Path:                "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
	// 		Key:                 "DigitalProductId",
	// 		ExpectedSuccess:     false,
	// 		OutputShouldContain: "no formatter",
	// 	},
	// 	// calls should succeed
	// 	{
	// 		CaseDescription:     "String key",
	// 		Path:                "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
	// 		Key:                 "ProductName",
	// 		ExpectedSuccess:     true,
	// 		OutputShouldContain: "Windows", // "Windows" should be in the reg key - hopefully this doesn't depend on the machine config that runs the test
	// 	},
	// 	{
	// 		CaseDescription:     "String key long-form HK",
	// 		Path:                "HKEY_LOCAL_MACHINE\\Software\\Microsoft\\Windows NT\\CurrentVersion",
	// 		Key:                 "ProductName",
	// 		ExpectedSuccess:     true,
	// 		OutputShouldContain: "Windows", // note as above on "Windows" being in the string.
	// 	},
	// 	{
	// 		CaseDescription:     "DWORD key",
	// 		Path:                "HKLM\\Software\\Microsoft\\Windows NT\\CurrentVersion",
	// 		Key:                 "PendingInstall",
	// 		ExpectedSuccess:     true,
	// 		OutputShouldContain: "0x", // should have a hex number formatted into the string
	// 	},
	// }

	// for _, tc := range cases {
	// 	r := regquery{}
	// 	success, outputString := r.getRegistryKeyValue(tc.Path, tc.Key)

	// 	assert.EqualValues(t, success, tc.ExpectedSuccess)
	// 	assert.Contains(t, outputString, tc.OutputShouldContain)
	// }
}
