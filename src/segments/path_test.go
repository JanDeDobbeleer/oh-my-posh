package segments

import (
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func renderTemplate(env *mock.MockedEnvironment, segmentTemplate string, context interface{}) string {
	found := false
	for _, call := range env.Mock.ExpectedCalls {
		if call.Method == "TemplateCache" {
			found = true
			break
		}
	}
	if !found {
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env: make(map[string]string),
		})
	}
	env.On("Error", mock2.Anything, mock2.Anything)
	env.On("Debug", mock2.Anything, mock2.Anything)
	tmpl := &template.Text{
		Template: segmentTemplate,
		Context:  context,
		Env:      env,
	}
	text, err := tmpl.Render()
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(text)
}

const (
	homeDir        = "/home/someone"
	homeDirWindows = "C:\\Users\\someone"
)

func TestParent(t *testing.T) {
	cases := []struct {
		Case                string
		Expected            string
		HomePath            string
		Pwd                 string
		GOOS                string
		PathSeparator       string
		FolderSeparatorIcon string
	}{
		{
			Case:          "Inside Home folder",
			Expected:      "~/",
			HomePath:      homeDir,
			Pwd:           homeDir + "/test",
			GOOS:          platform.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Home folder",
			HomePath:      homeDir,
			Pwd:           homeDir,
			GOOS:          platform.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Home folder with a trailing separator",
			HomePath:      homeDir,
			Pwd:           homeDir + "/",
			GOOS:          platform.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Root",
			HomePath:      homeDir,
			Pwd:           "/",
			GOOS:          platform.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Root + 1",
			Expected:      "/",
			HomePath:      homeDir,
			Pwd:           "/usr",
			GOOS:          platform.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Windows Home folder",
			HomePath:      homeDirWindows,
			Pwd:           homeDirWindows,
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows drive root",
			HomePath:      homeDirWindows,
			Pwd:           "C:",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows drive root with a trailing separator",
			HomePath:      homeDirWindows,
			Pwd:           "C:\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows drive root + 1",
			Expected:      "C:\\",
			HomePath:      homeDirWindows,
			Pwd:           "C:\\test",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "PSDrive root",
			HomePath:      homeDirWindows,
			Pwd:           "HKLM:",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		env.On("Flags").Return(&platform.Flags{})
		env.On("Shell").Return(shell.GENERIC)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("GOOS").Return(tc.GOOS)
		path := &Path{
			env: env,
			props: properties.Map{
				FolderSeparatorIcon: tc.FolderSeparatorIcon,
			},
		}
		path.setPaths()
		got := path.Parent()
		assert.EqualValues(t, tc.Expected, got, tc.Case)
	}
}

func TestAgnosterPathStyles(t *testing.T) {
	cases := []struct {
		Style               string
		Expected            string
		HomePath            string
		Pswd                string
		Pwd                 string
		PathSeparator       string
		HomeIcon            string
		FolderSeparatorIcon string
		GOOS                string
		MaxDepth            int
		HideRootLocation    bool
	}{
		{
			Style:               Unique,
			Expected:            "~ > a > ab > abcd",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/ab/abc/abcd",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Unique,
			Expected:            "~ > a > .a > abcd",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/ab/.abc/abcd",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Unique,
			Expected:            "~ > a > ab > abcd",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/ab/ab/abcd",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Unique,
			Expected:            "C > a > ab > abcd",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\ab\\ab\\abcd",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},

		{
			Style:               Letter,
			Expected:            "~",
			HomePath:            homeDir,
			Pwd:                 homeDir,
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "~ > a > w > man",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/ab/whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "u > b > a > w > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/burp/ab/whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "u > .b > a > w > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/.burp/ab/whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "u > .b > a > .w > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/.burp/ab/.whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "u > .b > a > ._w > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/.burp/ab/._whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "u > .ä > ū > .w > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/.äufbau/ūmgebung/.whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "u > .b > 1 > .w > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/.burp/12345/.whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "u > .b > 1 > .w > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/.burp/12345abc/.whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "u > .b > __p > .w > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/.burp/__pycache__/.whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "➼ > .w > man",
			HomePath:            homeDir,
			Pwd:                 "/➼/.whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "➼ s > .w > man",
			HomePath:            homeDir,
			Pwd:                 "/➼ something/.whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "C:\\",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "C > s > .w > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\something\\.whatever\\man",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "~ > s > man",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + "\\something\\man",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},

		{
			Style:               Mixed,
			Expected:            "~ > .. > man",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Mixed,
			Expected:            "~ > ab > .. > man",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/ab/whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Mixed,
			Expected:            "usr > foo > bar > .. > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/foo/bar/foobar/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Mixed,
			Expected:            "whatever > .. > foo > bar",
			HomePath:            homeDir,
			Pwd:                 "/whatever/foobar/foo/bar",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Mixed,
			Expected:            "C: > .. > foo > .. > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\Users\\foo\\foobar\\man",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},

		{
			Style:               AgnosterFull,
			Expected:            "usr > location > whatever",
			HomePath:            homeDir,
			Pwd:                 "/usr/location/whatever",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               AgnosterFull,
			Expected:            "PSDRIVE:/ | src",
			HomePath:            homeDir,
			Pwd:                 "/foo",
			Pswd:                "PSDRIVE:/src",
			PathSeparator:       "/",
			FolderSeparatorIcon: " | ",
		},

		{
			Style:               AgnosterShort,
			Expected:            ".. | src | init",
			HomePath:            homeDir,
			Pwd:                 "/foo",
			Pswd:                "PSDRIVE:/src/init",
			PathSeparator:       "/",
			FolderSeparatorIcon: " | ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "usr > foo > bar > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/foo/bar/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. | src",
			HomePath:            homeDir,
			Pwd:                 "/foo",
			Pswd:                "PSDRIVE:/src",
			PathSeparator:       "/",
			FolderSeparatorIcon: " | ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > projects",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/projects",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               AgnosterShort,
			Expected:            "\\\\localhost\\c$ > some",
			HomePath:            homeDirWindows,
			Pwd:                 "\\\\localhost\\c$\\some",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},

		{
			Style:               AgnosterShort,
			Expected:            "~",
			HomePath:            homeDir,
			Pwd:                 homeDir,
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            1,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "\\\\localhost\\c$",
			HomePath:            homeDirWindows,
			Pwd:                 "\\\\localhost\\c$",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               AgnosterShort,
			Expected:            "usr > .. > bar > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/foo/bar/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > .. > man",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               AgnosterShort,
			Expected:            "usr > .. > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/location/whatever/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + "\\foo\\bar\\man",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > .. > bar > man",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/foo/bar/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > foo > bar > man",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/foo/bar/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},
		{
			Style:               AgnosterShort,
			Expected:            "PSDRIVE:/ | .. | init",
			HomePath:            homeDir,
			Pwd:                 "/foo",
			Pswd:                "PSDRIVE:/src/init",
			PathSeparator:       "/",
			FolderSeparatorIcon: " | ",
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. > foo",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/foo",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            1,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. > bar > man",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/bar/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. > foo > bar > man",
			HomePath:            homeDir,
			Pwd:                 "/usr/foo/bar/man",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. > foo",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/foo",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. > foo > bar",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/foo/bar",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "C:/",
			HomePath:            homeDir,
			Pwd:                 "/mnt/c",
			Pswd:                "C:",
			PathSeparator:       "/",
			FolderSeparatorIcon: " | ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. | space foo",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/space foo",
			PathSeparator:       "/",
			FolderSeparatorIcon: " | ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. | space foo",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/space foo",
			PathSeparator:       "/",
			FolderSeparatorIcon: " | ",
			MaxDepth:            1,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "C:\\",
			HomePath:            homeDirWindows,
			Pwd:                 "C:",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               AgnosterShort,
			Expected:            "C: > .. > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\usr\\foo\\bar\\man",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "C: > .. > foo > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\usr\\foo\\bar\\man",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > .. > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + "\\foo\\bar\\man",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > foo > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + "\\foo\\bar\\man",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows,
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            1,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. > foo",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + "\\foo",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            1,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            ".. > foo",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + "\\foo",
			GOOS:                platform.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return(tc.GOOS)
		env.On("StackCount").Return(0)
		env.On("IsWsl").Return(false)
		args := &platform.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.PWSH)
		path := &Path{
			env: env,
			props: properties.Map{
				FolderSeparatorIcon: tc.FolderSeparatorIcon,
				properties.Style:    tc.Style,
				MaxDepth:            tc.MaxDepth,
				HideRootLocation:    tc.HideRootLocation,
			},
		}
		path.setPaths()
		path.setStyle()
		got := renderTemplate(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestFullAndFolderPath(t *testing.T) {
	cases := []struct {
		Style                  string
		HomePath               string
		FolderSeparatorIcon    string
		Pwd                    string
		Pswd                   string
		Expected               string
		DisableMappedLocations bool
		GOOS                   string
		PathSeparator          string
		StackCount             int
		Template               string
	}{
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: homeDir, Expected: "~"},
		{Style: Full, Pwd: homeDir + "/abc", Expected: "~/abc"},
		{Style: Full, Pwd: homeDir + "/abc", Expected: homeDir + "/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", Expected: "/a/b/c/d"},

		{Style: Full, FolderSeparatorIcon: "|", Pwd: homeDir, Expected: "~"},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: homeDir, Expected: "/home|someone", DisableMappedLocations: true},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: homeDir + "/abc", Expected: "~|abc"},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/a/b/c/d", Expected: "/a|b|c|d"},

		{Style: Folder, Pwd: "/", Expected: "/"},
		{Style: Folder, Pwd: homeDir, Expected: "~"},
		{Style: Folder, Pwd: homeDir, Expected: "someone", DisableMappedLocations: true},
		{Style: Folder, Pwd: homeDir + "/abc", Expected: "abc"},
		{Style: Folder, Pwd: "/a/b/c/d", Expected: "d"},

		{Style: Folder, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: homeDir, Expected: "~"},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: homeDir, Expected: "someone", DisableMappedLocations: true},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: homeDir + "/abc", Expected: "abc"},
		{Style: Folder, FolderSeparatorIcon: "|", Pwd: "/a/b/c/d", Expected: "d"},

		// for Windows paths
		{Style: Folder, FolderSeparatorIcon: "\\", Pwd: "C:\\", Expected: "C:\\", PathSeparator: "\\", GOOS: platform.WINDOWS},
		{Style: Folder, FolderSeparatorIcon: "\\", Pwd: homeDirWindows, Expected: "~", PathSeparator: "\\", GOOS: platform.WINDOWS},
		{Style: Full, FolderSeparatorIcon: "\\", Pwd: homeDirWindows, Expected: "~", PathSeparator: "\\", GOOS: platform.WINDOWS},
		{Style: Full, FolderSeparatorIcon: "\\", Pwd: homeDirWindows + "\\abc", Expected: "~\\abc", PathSeparator: "\\", GOOS: platform.WINDOWS},
		{Style: Full, FolderSeparatorIcon: "\\", Pwd: "C:\\Users\\posh", Expected: "C:\\Users\\posh", PathSeparator: "\\", GOOS: platform.WINDOWS},

		// StackCountEnabled=true and StackCount=2
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: "/", StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: homeDir, StackCount: 2, Expected: "2 ~"},
		{Style: Full, Pwd: homeDir + "/abc", StackCount: 2, Expected: "2 ~/abc"},
		{Style: Full, Pwd: homeDir + "/abc", StackCount: 2, Expected: "2 " + homeDir + "/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCount: 2, Expected: "2 /a/b/c/d"},

		// StackCountEnabled=false and StackCount=2
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Template: "{{ .Path }}", StackCount: 2, Expected: "/"},
		{Style: Full, Pwd: "/", Template: "{{ .Path }}", StackCount: 2, Expected: "/"},
		{Style: Full, Pwd: homeDir, Template: "{{ .Path }}", StackCount: 2, Expected: "~"},

		{Style: Full, Pwd: homeDir + "/abc", Template: "{{ .Path }}", StackCount: 2, Expected: homeDir + "/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", Template: "{{ .Path }}", StackCount: 2, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount=0
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: 0, Expected: "/"},
		{Style: Full, Pwd: "/", StackCount: 0, Expected: "/"},
		{Style: Full, Pwd: homeDir, StackCount: 0, Expected: "~"},
		{Style: Full, Pwd: homeDir + "/abc", StackCount: 0, Expected: "~/abc"},
		{Style: Full, Pwd: homeDir + "/abc", StackCount: 0, Expected: homeDir + "/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCount: 0, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount<0
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: -1, Expected: "/"},
		{Style: Full, Pwd: "/", StackCount: -1, Expected: "/"},
		{Style: Full, Pwd: homeDir, StackCount: -1, Expected: "~"},
		{Style: Full, Pwd: homeDir + "/abc", StackCount: -1, Expected: "~/abc"},
		{Style: Full, Pwd: homeDir + "/abc", StackCount: -1, Expected: homeDir + "/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", StackCount: -1, Expected: "/a/b/c/d"},

		// StackCountEnabled=true and StackCount not set
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: homeDir, Expected: "~"},
		{Style: Full, Pwd: homeDir + "/abc", Expected: "~/abc"},
		{Style: Full, Pwd: homeDir + "/abc", Expected: homeDir + "/abc", DisableMappedLocations: true},
		{Style: Full, Pwd: "/a/b/c/d", Expected: "/a/b/c/d"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}
		env.On("PathSeparator").Return(tc.PathSeparator)
		if tc.GOOS == platform.WINDOWS {
			env.On("Home").Return(homeDirWindows)
		} else {
			env.On("Home").Return(homeDir)
		}
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return(tc.GOOS)
		env.On("StackCount").Return(tc.StackCount)
		env.On("IsWsl").Return(false)
		args := &platform.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.GENERIC)
		if len(tc.Template) == 0 {
			tc.Template = "{{ if gt .StackCount 0 }}{{ .StackCount }} {{ end }}{{ .Path }}"
		}
		props := properties.Map{
			properties.Style: tc.Style,
		}
		if tc.FolderSeparatorIcon != "" {
			props[FolderSeparatorIcon] = tc.FolderSeparatorIcon
		}
		if tc.DisableMappedLocations {
			props[MappedLocationsEnabled] = false
		}
		path := &Path{
			env:        env,
			props:      props,
			StackCount: env.StackCount(),
		}
		path.setPaths()
		path.setStyle()
		got := renderTemplate(env, tc.Template, path)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestFullPathCustomMappedLocations(t *testing.T) {
	cases := []struct {
		Pwd             string
		MappedLocations map[string]string
		GOOS            string
		PathSeparator   string
		Expected        string
	}{
		{Pwd: "/a/b/c/d", MappedLocations: map[string]string{"{{ .Env.HOME }}/d": "#"}, Expected: "#"},
		{Pwd: "/a/b/c/d", MappedLocations: map[string]string{"/a/b/c/d": "#"}, Expected: "#"},
		{Pwd: "\\a\\b\\c\\d", MappedLocations: map[string]string{"\\a\\b": "#"}, GOOS: platform.WINDOWS, PathSeparator: "\\", Expected: "#\\c\\d"},
		{Pwd: "/a/b/c/d", MappedLocations: map[string]string{"/a/b": "#"}, Expected: "#/c/d"},
		{Pwd: "/a/b/c/d", MappedLocations: map[string]string{"/a/b": "/e/f"}, Expected: "/e/f/c/d"},
		{Pwd: homeDir + "/a/b/c/d", MappedLocations: map[string]string{"~/a/b": "#"}, Expected: "#/c/d"},
		{Pwd: "/a" + homeDir + "/b/c/d", MappedLocations: map[string]string{"/a~": "#"}, Expected: "/a" + homeDir + "/b/c/d"},
		{Pwd: homeDir + "/a/b/c/d", MappedLocations: map[string]string{"/a/b": "#"}, Expected: homeDir + "/a/b/c/d"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(homeDir)
		env.On("Pwd").Return(tc.Pwd)
		if tc.GOOS == "" {
			tc.GOOS = platform.DARWIN
		}
		env.On("GOOS").Return(tc.GOOS)
		if tc.PathSeparator == "" {
			tc.PathSeparator = "/"
		}
		env.On("PathSeparator").Return(tc.PathSeparator)
		args := &platform.Flags{
			PSWD: tc.Pwd,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.GENERIC)
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env: map[string]string{
				"HOME": "/a/b/c",
			},
		})
		path := &Path{
			env: env,
			props: properties.Map{
				properties.Style:       Full,
				MappedLocationsEnabled: false,
				MappedLocations:        tc.MappedLocations,
			},
		}
		path.setPaths()
		path.setStyle()
		got := renderTemplate(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestFolderPathCustomMappedLocations(t *testing.T) {
	pwd := "/a/b/c/d"
	env := new(mock.MockedEnvironment)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return(homeDir)
	env.On("Pwd").Return(pwd)
	env.On("GOOS").Return("")
	args := &platform.Flags{
		PSWD: pwd,
	}
	env.On("Flags").Return(args)
	env.On("Shell").Return(shell.GENERIC)
	path := &Path{
		env: env,
		props: properties.Map{
			properties.Style: Folder,
			MappedLocations: map[string]string{
				"/a/b/c/d": "#",
			},
		},
	}
	path.setPaths()
	path.setStyle()
	got := renderTemplate(env, "{{ .Path }}", path)
	assert.Equal(t, "#", got)
}

func TestAgnosterPath(t *testing.T) {
	cases := []struct {
		Case          string
		Expected      string
		Home          string
		PWD           string
		GOOS          string
		PathSeparator string
	}{
		{
			Case:          "Windows registry drive case sensitive",
			Expected:      "\uf013 > f > magnetic:TOAST",
			Home:          homeDirWindows,
			PWD:           "HKLM:\\SOFTWARE\\magnetic:TOAST\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows outside home",
			Expected:      "C: > f > f > location",
			Home:          homeDirWindows,
			PWD:           "C:\\Program Files\\Go\\location",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows oustide home",
			Expected:      "~ > f > f > location",
			Home:          homeDirWindows,
			PWD:           homeDirWindows + "\\Documents\\Bill\\location",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows inside home zero levels",
			Expected:      "C: > location",
			Home:          homeDirWindows,
			PWD:           "C:\\location",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows inside home one level",
			Expected:      "C: > f > location",
			Home:          homeDirWindows,
			PWD:           "C:\\Program Files\\location",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower case drive letter",
			Expected:      "C: > Windows",
			Home:          homeDirWindows,
			PWD:           "C:\\Windows\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower case drive letter (other)",
			Expected:      "P: > Other",
			Home:          homeDirWindows,
			PWD:           "P:\\Other\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive",
			Expected:      "some: > some",
			Home:          homeDirWindows,
			PWD:           "some:\\some\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive (ending with c)",
			Expected:      "src: > source",
			Home:          homeDirWindows,
			PWD:           "src:\\source\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive (arbitrary cases)",
			Expected:      "sRc: > source",
			Home:          homeDirWindows,
			PWD:           "sRc:\\source\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows registry drive",
			Expected:      "\uf013 > f > magnetic:test",
			Home:          homeDirWindows,
			PWD:           "HKLM:\\SOFTWARE\\magnetic:test\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Unix outside home",
			Expected:      "mnt > f > f > location",
			Home:          homeDir,
			PWD:           "/mnt/go/test/location",
			PathSeparator: "/",
		},
		{
			Case:          "Unix inside home",
			Expected:      "~ > f > f > location",
			Home:          homeDir,
			PWD:           homeDir + "/docs/jan/location",
			PathSeparator: "/",
		},
		{
			Case:          "Unix outside home zero levels",
			Expected:      "mnt > location",
			Home:          homeDir,
			PWD:           "/mnt/location",
			PathSeparator: "/",
		},
		{
			Case:          "Unix outside home one level",
			Expected:      "mnt > f > location",
			Home:          homeDir,
			PWD:           "/mnt/folder/location",
			PathSeparator: "/",
		},
	}

	for _, tc := range cases { //nolint:dupl
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(tc.Home)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Pwd").Return(tc.PWD)
		env.On("GOOS").Return(tc.GOOS)
		args := &platform.Flags{
			PSWD: tc.PWD,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.PWSH)
		path := &Path{
			env: env,
			props: properties.Map{
				properties.Style:    Agnoster,
				FolderSeparatorIcon: " > ",
				FolderIcon:          "f",
				HomeIcon:            "~",
			},
		}
		path.setPaths()
		path.setStyle()
		got := renderTemplate(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestAgnosterLeftPath(t *testing.T) {
	cases := []struct {
		Case          string
		Expected      string
		Home          string
		PWD           string
		GOOS          string
		PathSeparator string
	}{
		{
			Case:          "Windows inside home",
			Expected:      "~ > Documents > f > f",
			Home:          homeDirWindows,
			PWD:           homeDirWindows + "\\Documents\\Bill\\location",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows outside home",
			Expected:      "C: > Program Files > f > f",
			Home:          homeDirWindows,
			PWD:           "C:\\Program Files\\Go\\location",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows inside home zero levels",
			Expected:      "C: > location",
			Home:          homeDirWindows,
			PWD:           "C:\\location",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows inside home one level",
			Expected:      "C: > Program Files > f",
			Home:          homeDirWindows,
			PWD:           "C:\\Program Files\\location",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower case drive letter",
			Expected:      "C: > Windows",
			Home:          homeDirWindows,
			PWD:           "C:\\Windows\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower case drive letter (other)",
			Expected:      "P: > Other",
			Home:          homeDirWindows,
			PWD:           "P:\\Other\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive",
			Expected:      "some: > some",
			Home:          homeDirWindows,
			PWD:           "some:\\some\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive (ending with c)",
			Expected:      "src: > source",
			Home:          homeDirWindows,
			PWD:           "src:\\source\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive (arbitrary cases)",
			Expected:      "sRc: > source",
			Home:          homeDirWindows,
			PWD:           "sRc:\\source\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows registry drive",
			Expected:      "\uf013 > SOFTWARE > f",
			Home:          homeDirWindows,
			PWD:           "HKLM:\\SOFTWARE\\magnetic:test\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows registry drive case sensitive",
			Expected:      "\uf013 > SOFTWARE > f",
			Home:          homeDirWindows,
			PWD:           "HKLM:\\SOFTWARE\\magnetic:TOAST\\",
			GOOS:          platform.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Unix outside home",
			Expected:      "mnt > go > f > f",
			Home:          homeDir,
			PWD:           "/mnt/go/test/location",
			PathSeparator: "/",
		},
		{
			Case:          "Unix inside home",
			Expected:      "~ > docs > f > f",
			Home:          homeDir,
			PWD:           homeDir + "/docs/jan/location",
			PathSeparator: "/",
		},
		{
			Case:          "Unix outside home zero levels",
			Expected:      "mnt > location",
			Home:          homeDir,
			PWD:           "/mnt/location",
			PathSeparator: "/",
		},
		{
			Case:          "Unix outside home one level",
			Expected:      "mnt > folder > f",
			Home:          homeDir,
			PWD:           "/mnt/folder/location",
			PathSeparator: "/",
		},
	}

	for _, tc := range cases { //nolint:dupl
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(tc.Home)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Pwd").Return(tc.PWD)
		env.On("GOOS").Return(tc.GOOS)
		args := &platform.Flags{
			PSWD: tc.PWD,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.PWSH)
		path := &Path{
			env: env,
			props: properties.Map{
				properties.Style:    AgnosterLeft,
				FolderSeparatorIcon: " > ",
				FolderIcon:          "f",
				HomeIcon:            "~",
			},
		}
		path.setPaths()
		path.setStyle()
		got := renderTemplate(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetPwd(t *testing.T) {
	cases := []struct {
		MappedLocationsEnabled bool
		Pwd                    string
		Pswd                   string
		Expected               string
	}{
		{MappedLocationsEnabled: true, Pwd: homeDir, Expected: "~"},
		{MappedLocationsEnabled: true, Pwd: homeDir + "-test", Expected: homeDir + "-test"},
		{MappedLocationsEnabled: true},
		{MappedLocationsEnabled: true, Pwd: "/usr", Expected: "/usr"},
		{MappedLocationsEnabled: true, Pwd: homeDir + "/abc", Expected: "~/abc"},
		{MappedLocationsEnabled: true, Pwd: "/a/b/c/d", Expected: "#"},
		{MappedLocationsEnabled: true, Pwd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
		{MappedLocationsEnabled: true, Pwd: "/z/y/x/w", Expected: "/z/y/x/w"},

		{MappedLocationsEnabled: false},
		{MappedLocationsEnabled: false, Pwd: homeDir + "/abc", Expected: homeDir + "/abc"},
		{MappedLocationsEnabled: false, Pwd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
		{MappedLocationsEnabled: false, Pwd: homeDir + "/c/d/e/f/g", Expected: homeDir + "/c/d/e/f/g"},
		{MappedLocationsEnabled: true, Pwd: homeDir + "/c/d/e/f/g", Expected: "~/c/d/e/f/g"},

		{MappedLocationsEnabled: true, Pwd: "/w/d/x/w", Pswd: "/z/y/x/w", Expected: "/z/y/x/w"},
		{MappedLocationsEnabled: false, Pwd: "/f/g/k/d/e/f/g", Pswd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return(homeDir)
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return("")
		args := &platform.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.PWSH)
		path := &Path{
			env: env,
			props: properties.Map{
				MappedLocationsEnabled: tc.MappedLocationsEnabled,
				MappedLocations: map[string]string{
					"/a/b/c/d": "#",
				},
			},
		}
		path.setPaths()
		assert.Equal(t, tc.Expected, path.pwd)
	}
}

func TestGetFolderSeparator(t *testing.T) {
	cases := []struct {
		Case                    string
		FolderSeparatorIcon     string
		FolderSeparatorTemplate string
		Expected                string
	}{
		{Case: "default", Expected: "/"},
		{Case: "icon - no template", FolderSeparatorIcon: "\ue5fe", Expected: "\ue5fe"},
		{Case: "template", FolderSeparatorTemplate: "{{ if eq .Shell \"bash\" }}\\{{ end }}", Expected: "\\"},
		{Case: "template empty", FolderSeparatorTemplate: "{{ if eq .Shell \"pwsh\" }}\\{{ end }}", Expected: "/"},
		{Case: "invalid template", FolderSeparatorTemplate: "{{ if eq .Shell \"pwsh\" }}", Expected: "/"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("PathSeparator").Return("/")
		env.On("Error", mock2.Anything, mock2.Anything)
		env.On("Debug", mock2.Anything, mock2.Anything)
		path := &Path{
			env: env,
		}
		props := properties.Map{}
		if len(tc.FolderSeparatorTemplate) > 0 {
			props[FolderSeparatorTemplate] = tc.FolderSeparatorTemplate
		}
		if len(tc.FolderSeparatorIcon) > 0 {
			props[FolderSeparatorIcon] = tc.FolderSeparatorIcon
		}
		env.On("TemplateCache").Return(&platform.TemplateCache{
			Env:   make(map[string]string),
			Shell: "bash",
		})
		path.props = props
		got := path.getFolderSeparator()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		Input    string
		HomeDir  string
		GOOS     string
		Expected string
	}{
		{Input: "/foo/~/bar", HomeDir: homeDirWindows, GOOS: platform.WINDOWS, Expected: "\\foo\\~\\bar"},
		{Input: homeDirWindows + "\\Foo", HomeDir: homeDirWindows, GOOS: platform.WINDOWS, Expected: "c:\\users\\someone\\foo"},
		{Input: "~/Bob\\Foo", HomeDir: homeDir, GOOS: platform.LINUX, Expected: homeDir + "/Bob\\Foo"},
		{Input: "~/Bob\\Foo", HomeDir: homeDir, GOOS: platform.DARWIN, Expected: homeDir + "/bob\\foo"},
		{Input: "~\\Bob\\Foo", HomeDir: homeDirWindows, GOOS: platform.WINDOWS, Expected: "c:\\users\\someone\\bob\\foo"},
		{Input: "/foo/~/bar", HomeDir: homeDir, GOOS: platform.LINUX, Expected: "/foo/~/bar"},
		{Input: "~/baz", HomeDir: homeDir, GOOS: platform.LINUX, Expected: homeDir + "/baz"},
		{Input: "~/baz", HomeDir: homeDirWindows, GOOS: platform.WINDOWS, Expected: "c:\\users\\someone\\baz"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Home").Return(tc.HomeDir)
		env.On("GOOS").Return(tc.GOOS)
		pt := &Path{
			env: env,
		}
		got := pt.normalize(tc.Input)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestReplaceMappedLocations(t *testing.T) {
	cases := []struct {
		Case     string
		Pwd      string
		Expected string
	}{
		{Pwd: "/c/l/k/f", Expected: "f"},
		{Pwd: "/f/g/h", Expected: "/f/g/h"},
		{Pwd: "/f/g/h/e", Expected: "^/e"},
		{Pwd: "/a/b/c/d", Expected: "#"},
		{Pwd: "/a/b/c/d/e", Expected: "#/e"},
		{Pwd: "/a/b/c/d/e", Expected: "#/e"},
		{Pwd: "/a/b/k/j/e", Expected: "e"},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("PathSeparator").Return("/")
		env.On("Pwd").Return(tc.Pwd)
		env.On("Shell").Return(shell.FISH)
		env.On("GOOS").Return(platform.DARWIN)
		env.On("Home").Return("/a/b/k")
		path := &Path{
			env: env,
			props: properties.Map{
				MappedLocationsEnabled: false,
				MappedLocations: map[string]string{
					"/a/b/c/d": "#",
					"/f/g/h/*": "^",
					"/c/l/k/*": "",
					"~/j/*":    "",
				},
			},
		}
		path.setPaths()
		assert.Equal(t, tc.Expected, path.pwd)
	}
}
