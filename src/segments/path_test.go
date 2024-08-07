package segments

import (
	"errors"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	homeDir        = "/home/someone"
	homeDirWindows = "C:\\Users\\someone"
	fooBarMan      = "\\foo\\bar\\man"
	abc            = "/abc"
	abcd           = "/a/b/c/d"
	cdefg          = "/c/d/e/f/g"
)

func renderTemplateNoTrimSpace(env *mock.Environment, segmentTemplate string, context any) string {
	found := false
	for _, call := range env.Mock.ExpectedCalls {
		if call.Method == "TemplateCache" {
			found = true
			break
		}
	}

	if !found {
		env.On("TemplateCache").Return(&cache.Template{})
	}

	env.On("Error", testify_.Anything)
	env.On("Debug", testify_.Anything)
	env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
	env.On("Shell").Return("foo")

	template.Init(env)

	tmpl := &template.Text{
		Template: segmentTemplate,
		Context:  context,
	}

	text, err := tmpl.Render()
	if err != nil {
		return err.Error()
	}

	return text
}

func renderTemplate(env *mock.Environment, segmentTemplate string, context any) string {
	return strings.TrimSpace(renderTemplateNoTrimSpace(env, segmentTemplate, context))
}

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
			GOOS:          runtime.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Home folder",
			HomePath:      homeDir,
			Pwd:           homeDir,
			GOOS:          runtime.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Home folder with a trailing separator",
			HomePath:      homeDir,
			Pwd:           homeDir + "/",
			GOOS:          runtime.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Root",
			HomePath:      homeDir,
			Pwd:           "/",
			GOOS:          runtime.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Root + 1",
			Expected:      "/",
			HomePath:      homeDir,
			Pwd:           "/usr",
			GOOS:          runtime.DARWIN,
			PathSeparator: "/",
		},
		{
			Case:          "Windows Home folder",
			HomePath:      homeDirWindows,
			Pwd:           homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows drive root",
			HomePath:      homeDirWindows,
			Pwd:           "C:",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows drive root with a trailing separator",
			HomePath:      homeDirWindows,
			Pwd:           "C:\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows drive root + 1",
			Expected:      "C:\\",
			HomePath:      homeDirWindows,
			Pwd:           "C:\\test",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "PSDrive root",
			HomePath:      homeDirWindows,
			Pwd:           "HKLM:",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		env.On("Flags").Return(&runtime.Flags{})
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
		CygpathError        error
		GOOS                string
		Shell               string
		Pswd                string
		Pwd                 string
		PathSeparator       string
		HomeIcon            string
		HomePath            string
		Style               string
		FolderSeparatorIcon string
		Cygpath             string
		Expected            string
		MaxDepth            int
		MaxWidth            int
		HideRootLocation    bool
		Cygwin              bool
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
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Unique,
			Expected:            "a",
			HomePath:            homeDir,
			Pwd:                 "/ab",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
		},

		{
			Style:               Powerlevel,
			Expected:            "t > w > o > a > v > l > p > wh > we > i > wa > th > the > d > f > u > it > c > to > a > co > stream",
			HomePath:            homeDir,
			Pwd:                 "/there/was/once/a/very/long/path/which/wended/its/way/through/the/dark/forest/until/it/came/to/a/cold/stream",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxWidth:            20,
		},
		{
			Style:               Powerlevel,
			Expected:            "t > w > o > a > v > l > p > which > wended > its > way > through > the",
			HomePath:            homeDir,
			Pwd:                 "/there/was/once/a/very/long/path/which/wended/its/way/through/the",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxWidth:            70,
		},
		{
			Style:               Powerlevel,
			Expected:            "var/cache/pacman",
			HomePath:            homeDir,
			Pwd:                 "/var/cache/pacman",
			PathSeparator:       "/",
			FolderSeparatorIcon: "/",
			MaxWidth:            50,
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
			Expected:            "C: > ",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\",
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "C > s > .w > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\something\\.whatever\\man",
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "~ > s > man",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + "\\something\\man",
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Letter,
			Expected:            "w",
			HomePath:            homeDir,
			Pwd:                 "/whatever",
			PathSeparator:       "/",
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
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Mixed,
			Expected:            "c > .. > foo > .. > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\Users\\foo\\foobar\\man",
			GOOS:                runtime.WINDOWS,
			Shell:               shell.BASH,
			Cygwin:              true,
			Cygpath:             "/c/Users/foo/foobar/man",
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               Mixed,
			Expected:            "C: > .. > foo > .. > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\Users\\foo\\foobar\\man",
			GOOS:                runtime.WINDOWS,
			Shell:               shell.BASH,
			CygpathError:        errors.New("oh no"),
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
			Expected:            "PSDRIVE: | src",
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
			Expected:            "PSDRIVE: | src",
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
			GOOS:                runtime.WINDOWS,
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
			GOOS:                runtime.WINDOWS,
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
			Pwd:                 homeDirWindows + fooBarMan,
			GOOS:                runtime.WINDOWS,
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
			Expected:            "PSDRIVE: | .. | init",
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
			Expected:            "~ > foo",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/foo",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > foo > bar",
			HomePath:            homeDir,
			Pwd:                 homeDir + "/foo/bar",
			PathSeparator:       "/",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "C: | ",
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
			Expected:            "~ | space foo",
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
			Expected:            "C: > ",
			HomePath:            homeDirWindows,
			Pwd:                 "C:",
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
		},
		{
			Style:               AgnosterShort,
			Expected:            "C: > .. > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\usr\\foo\\bar\\man",
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "C: > .. > foo > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 "C:\\usr\\foo\\bar\\man",
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > .. > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + fooBarMan,
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > foo > bar > man",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + fooBarMan,
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            3,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows,
			GOOS:                runtime.WINDOWS,
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
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            1,
			HideRootLocation:    true,
		},
		{
			Style:               AgnosterShort,
			Expected:            "~ > foo",
			HomePath:            homeDirWindows,
			Pwd:                 homeDirWindows + "\\foo",
			GOOS:                runtime.WINDOWS,
			PathSeparator:       "\\",
			FolderSeparatorIcon: " > ",
			MaxDepth:            2,
			HideRootLocation:    true,
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Home").Return(tc.HomePath)
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return(tc.GOOS)
		env.On("IsCygwin").Return(tc.Cygwin)
		env.On("StackCount").Return(0)
		env.On("IsWsl").Return(false)
		args := &runtime.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)

		if len(tc.Shell) == 0 {
			tc.Shell = shell.PWSH
		}
		env.On("Shell").Return(tc.Shell)

		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)

		displayCygpath := tc.Cygwin
		if displayCygpath {
			env.On("RunCommand", "cygpath", []string{"-u", tc.Pwd}).Return(tc.Cygpath, tc.CygpathError)
			env.On("RunCommand", "cygpath", testify_.Anything).Return("brrrr", nil)
		}

		path := &Path{
			env: env,
			props: properties.Map{
				FolderSeparatorIcon: tc.FolderSeparatorIcon,
				properties.Style:    tc.Style,
				MaxDepth:            tc.MaxDepth,
				MaxWidth:            tc.MaxWidth,
				HideRootLocation:    tc.HideRootLocation,
				DisplayCygpath:      displayCygpath,
			},
		}

		path.setPaths()
		path.setStyle()
		got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
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
		GOOS                   string
		PathSeparator          string
		Template               string
		StackCount             int
		DisableMappedLocations bool
	}{
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: homeDir, Expected: "~"},
		{Style: Full, Pwd: homeDir + abc, Expected: "~/abc"},
		{Style: Full, Pwd: homeDir + abc, Expected: homeDir + abc, DisableMappedLocations: true},
		{Style: Full, Pwd: abcd, Expected: abcd},

		{Style: Full, FolderSeparatorIcon: "|", Pwd: homeDir, Expected: "~"},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: homeDir, Expected: "/home|someone", DisableMappedLocations: true},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: homeDir + abc, Expected: "~|abc"},
		{Style: Full, FolderSeparatorIcon: "|", Pwd: abcd, Expected: "/a|b|c|d"},

		{Style: FolderType, Pwd: "/", Expected: "/"},
		{Style: FolderType, Pwd: homeDir, Expected: "~"},
		{Style: FolderType, Pwd: homeDir, Expected: "someone", DisableMappedLocations: true},
		{Style: FolderType, Pwd: homeDir + abc, Expected: "abc"},
		{Style: FolderType, Pwd: abcd, Expected: "d"},

		{Style: FolderType, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: FolderType, FolderSeparatorIcon: "|", Pwd: homeDir, Expected: "~"},
		{Style: FolderType, FolderSeparatorIcon: "|", Pwd: homeDir, Expected: "someone", DisableMappedLocations: true},
		{Style: FolderType, FolderSeparatorIcon: "|", Pwd: homeDir + abc, Expected: "abc"},
		{Style: FolderType, FolderSeparatorIcon: "|", Pwd: abcd, Expected: "d"},

		// for Windows paths
		{Style: FolderType, FolderSeparatorIcon: "\\", Pwd: "C:\\", Expected: "C:\\", PathSeparator: "\\", GOOS: runtime.WINDOWS},
		{Style: FolderType, FolderSeparatorIcon: "\\", Pwd: "\\\\localhost\\d$", Expected: "\\\\localhost\\d$", PathSeparator: "\\", GOOS: runtime.WINDOWS},
		{Style: FolderType, FolderSeparatorIcon: "\\", Pwd: homeDirWindows, Expected: "~", PathSeparator: "\\", GOOS: runtime.WINDOWS},
		{Style: Full, FolderSeparatorIcon: "\\", Pwd: homeDirWindows, Expected: "~", PathSeparator: "\\", GOOS: runtime.WINDOWS},
		{Style: Full, FolderSeparatorIcon: "\\", Pwd: homeDirWindows + "\\abc", Expected: "~\\abc", PathSeparator: "\\", GOOS: runtime.WINDOWS},
		{Style: Full, FolderSeparatorIcon: "\\", Pwd: "C:\\Users\\posh", Expected: "C:\\Users\\posh", PathSeparator: "\\", GOOS: runtime.WINDOWS},

		// StackCountEnabled=true and StackCount=2
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: "/", StackCount: 2, Expected: "2 /"},
		{Style: Full, Pwd: homeDir, StackCount: 2, Expected: "2 ~"},
		{Style: Full, Pwd: homeDir + abc, StackCount: 2, Expected: "2 ~/abc"},
		{Style: Full, Pwd: homeDir + abc, StackCount: 2, Expected: "2 " + homeDir + abc, DisableMappedLocations: true},
		{Style: Full, Pwd: abcd, StackCount: 2, Expected: "2 /a/b/c/d"},

		// StackCountEnabled=false and StackCount=2
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Template: "{{ .Path }}", StackCount: 2, Expected: "/"},
		{Style: Full, Pwd: "/", Template: "{{ .Path }}", StackCount: 2, Expected: "/"},
		{Style: Full, Pwd: homeDir, Template: "{{ .Path }}", StackCount: 2, Expected: "~"},

		{Style: Full, Pwd: homeDir + abc, Template: "{{ .Path }}", StackCount: 2, Expected: homeDir + abc, DisableMappedLocations: true},
		{Style: Full, Pwd: abcd, Template: "{{ .Path }}", StackCount: 2, Expected: abcd},

		// StackCountEnabled=true and StackCount=0
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: 0, Expected: "/"},
		{Style: Full, Pwd: "/", StackCount: 0, Expected: "/"},
		{Style: Full, Pwd: homeDir, StackCount: 0, Expected: "~"},
		{Style: Full, Pwd: homeDir + abc, StackCount: 0, Expected: "~/abc"},
		{Style: Full, Pwd: homeDir + abc, StackCount: 0, Expected: homeDir + abc, DisableMappedLocations: true},
		{Style: Full, Pwd: abcd, StackCount: 0, Expected: abcd},

		// StackCountEnabled=true and StackCount<0
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", StackCount: -1, Expected: "/"},
		{Style: Full, Pwd: "/", StackCount: -1, Expected: "/"},
		{Style: Full, Pwd: homeDir, StackCount: -1, Expected: "~"},
		{Style: Full, Pwd: homeDir + abc, StackCount: -1, Expected: "~/abc"},
		{Style: Full, Pwd: homeDir + abc, StackCount: -1, Expected: homeDir + abc, DisableMappedLocations: true},
		{Style: Full, Pwd: abcd, StackCount: -1, Expected: abcd},

		// StackCountEnabled=true and StackCount not set
		{Style: Full, FolderSeparatorIcon: "|", Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: "/", Expected: "/"},
		{Style: Full, Pwd: homeDir, Expected: "~"},
		{Style: Full, Pwd: homeDir + abc, Expected: "~/abc"},
		{Style: Full, Pwd: homeDir + abc, Expected: homeDir + abc, DisableMappedLocations: true},
		{Style: Full, Pwd: abcd, Expected: abcd},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}
		env.On("PathSeparator").Return(tc.PathSeparator)
		if tc.GOOS == runtime.WINDOWS {
			env.On("Home").Return(homeDirWindows)
		} else {
			env.On("Home").Return(homeDir)
		}
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return(tc.GOOS)
		env.On("StackCount").Return(tc.StackCount)
		env.On("IsWsl").Return(false)
		args := &runtime.Flags{
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
		got := renderTemplateNoTrimSpace(env, tc.Template, path)
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
		{Pwd: homeDir + "/d", MappedLocations: map[string]string{"{{ .Env.HOME }}/d": "#"}, Expected: "#"},
		{Pwd: abcd, MappedLocations: map[string]string{abcd: "#"}, Expected: "#"},
		{Pwd: "\\a\\b\\c\\d", MappedLocations: map[string]string{"\\a\\b": "#"}, GOOS: runtime.WINDOWS, PathSeparator: "\\", Expected: "#\\c\\d"},
		{Pwd: abcd, MappedLocations: map[string]string{"/a/b": "#"}, Expected: "#/c/d"},
		{Pwd: abcd, MappedLocations: map[string]string{"/a/b": "/e/f"}, Expected: "/e/f/c/d"},
		{Pwd: homeDir + abcd, MappedLocations: map[string]string{"~/a/b": "#"}, Expected: "#/c/d"},
		{Pwd: "/a" + homeDir + "/b/c/d", MappedLocations: map[string]string{"/a~": "#"}, Expected: "/a" + homeDir + "/b/c/d"},
		{Pwd: homeDir + abcd, MappedLocations: map[string]string{"/a/b": "#"}, Expected: homeDir + abcd},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return(homeDir)
		env.On("Pwd").Return(tc.Pwd)

		if len(tc.GOOS) == 0 {
			tc.GOOS = runtime.DARWIN
		}

		env.On("GOOS").Return(tc.GOOS)

		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}

		env.On("PathSeparator").Return(tc.PathSeparator)
		args := &runtime.Flags{
			PSWD: tc.Pwd,
		}

		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.GENERIC)
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		env.On("TemplateCache").Return(&cache.Template{})
		env.On("Getenv", "HOME").Return(homeDir)

		template.Init(env)

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

		got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got)
	}
}

func TestFolderPathCustomMappedLocations(t *testing.T) {
	pwd := abcd
	env := new(mock.Environment)
	env.On("PathSeparator").Return("/")
	env.On("Home").Return(homeDir)
	env.On("Pwd").Return(pwd)
	env.On("GOOS").Return("")
	args := &runtime.Flags{
		PSWD: pwd,
	}
	env.On("Flags").Return(args)
	env.On("Shell").Return(shell.GENERIC)
	env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)

	template.Init(env)

	path := &Path{
		env: env,
		props: properties.Map{
			properties.Style: FolderType,
			MappedLocations: map[string]string{
				abcd: "#",
			},
		},
	}

	path.setPaths()
	path.setStyle()

	got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
	assert.Equal(t, "#", got)
}

func TestAgnosterPath(t *testing.T) {
	cases := []struct {
		Case           string
		Expected       string
		Home           string
		PWD            string
		GOOS           string
		PathSeparator  string
		Cycle          []string
		ColorSeparator bool
	}{
		{
			Case:          "Windows registry drive case sensitive",
			Expected:      "\uf013 > f > magnetic:TOAST",
			Home:          homeDirWindows,
			PWD:           "HKLM:\\SOFTWARE\\magnetic:TOAST\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows outside home",
			Expected:      "C: > f > f > location",
			Home:          homeDirWindows,
			PWD:           "C:\\Program Files\\Go\\location",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows oustide home",
			Expected:      "~ > f > f > location",
			Home:          homeDirWindows,
			PWD:           homeDirWindows + "\\Documents\\Bill\\location",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows inside home zero levels",
			Expected:      "C: > location",
			Home:          homeDirWindows,
			PWD:           "C:\\location",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows inside home one level",
			Expected:      "C: > f > location",
			Home:          homeDirWindows,
			PWD:           "C:\\Program Files\\location",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower case drive letter",
			Expected:      "C: > Windows",
			Home:          homeDirWindows,
			PWD:           "C:\\Windows\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower case drive letter (other)",
			Expected:      "P: > Other",
			Home:          homeDirWindows,
			PWD:           "P:\\Other\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive",
			Expected:      "some: > some",
			Home:          homeDirWindows,
			PWD:           "some:\\some\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive (ending with c)",
			Expected:      "src: > source",
			Home:          homeDirWindows,
			PWD:           "src:\\source\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive (arbitrary cases)",
			Expected:      "sRc: > source",
			Home:          homeDirWindows,
			PWD:           "sRc:\\source\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows registry drive",
			Expected:      "\uf013 > f > magnetic:test",
			Home:          homeDirWindows,
			PWD:           "HKLM:\\SOFTWARE\\magnetic:test\\",
			GOOS:          runtime.WINDOWS,
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
		{
			Case:          "Unix, colorize",
			Expected:      "<blue>mnt</> > <yellow>f</> > <blue>location</>",
			Home:          homeDir,
			PWD:           "/mnt/folder/location",
			PathSeparator: "/",
			Cycle:         []string{"blue", "yellow"},
		},
		{
			Case:           "Unix, colorize with folder separator",
			Expected:       "<blue>mnt</><yellow> > </><yellow>f</><blue> > </><blue>location</>",
			Home:           homeDir,
			PWD:            "/mnt/folder/location",
			PathSeparator:  "/",
			Cycle:          []string{"blue", "yellow"},
			ColorSeparator: true,
		},
		{
			Case:          "Unix one level",
			Expected:      "mnt",
			Home:          homeDir,
			PWD:           "/mnt",
			PathSeparator: "/",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.Home)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Pwd").Return(tc.PWD)
		env.On("GOOS").Return(tc.GOOS)
		args := &runtime.Flags{
			PSWD: tc.PWD,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.PWSH)
		path := &Path{
			env: env,
			props: properties.Map{
				properties.Style:     Agnoster,
				FolderSeparatorIcon:  " > ",
				FolderIcon:           "f",
				HomeIcon:             "~",
				Cycle:                tc.Cycle,
				CycleFolderSeparator: tc.ColorSeparator,
			},
		}
		path.setPaths()
		path.setStyle()
		got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
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
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows outside home",
			Expected:      "C: > Program Files > f > f",
			Home:          homeDirWindows,
			PWD:           "C:\\Program Files\\Go\\location",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows inside home zero levels",
			Expected:      "C: > location",
			Home:          homeDirWindows,
			PWD:           "C:\\location",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows inside home one level",
			Expected:      "C: > Program Files > f",
			Home:          homeDirWindows,
			PWD:           "C:\\Program Files\\location",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower case drive letter",
			Expected:      "C: > Windows",
			Home:          homeDirWindows,
			PWD:           "C:\\Windows\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower case drive letter (other)",
			Expected:      "P: > Other",
			Home:          homeDirWindows,
			PWD:           "P:\\Other\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive",
			Expected:      "some: > some",
			Home:          homeDirWindows,
			PWD:           "some:\\some\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive (ending with c)",
			Expected:      "src: > source",
			Home:          homeDirWindows,
			PWD:           "src:\\source\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows lower word drive (arbitrary cases)",
			Expected:      "sRc: > source",
			Home:          homeDirWindows,
			PWD:           "sRc:\\source\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows registry drive",
			Expected:      "\uf013 > SOFTWARE > f",
			Home:          homeDirWindows,
			PWD:           "HKLM:\\SOFTWARE\\magnetic:test\\",
			GOOS:          runtime.WINDOWS,
			PathSeparator: "\\",
		},
		{
			Case:          "Windows registry drive case sensitive",
			Expected:      "\uf013 > SOFTWARE > f",
			Home:          homeDirWindows,
			PWD:           "HKLM:\\SOFTWARE\\magnetic:TOAST\\",
			GOOS:          runtime.WINDOWS,
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

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.Home)
		env.On("PathSeparator").Return(tc.PathSeparator)
		env.On("Pwd").Return(tc.PWD)
		env.On("GOOS").Return(tc.GOOS)
		args := &runtime.Flags{
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
		got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetPwd(t *testing.T) {
	cases := []struct {
		Pwd                    string
		Pswd                   string
		Expected               string
		MappedLocationsEnabled bool
	}{
		{MappedLocationsEnabled: true, Pwd: homeDir, Expected: "~"},
		{MappedLocationsEnabled: true, Pwd: homeDir + "-test", Expected: homeDir + "-test"},
		{MappedLocationsEnabled: true},
		{MappedLocationsEnabled: true, Pwd: "/usr", Expected: "/usr"},
		{MappedLocationsEnabled: true, Pwd: homeDir + abc, Expected: "~/abc"},
		{MappedLocationsEnabled: true, Pwd: abcd, Expected: "#"},
		{MappedLocationsEnabled: true, Pwd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
		{MappedLocationsEnabled: true, Pwd: "/z/y/x/w", Expected: "/z/y/x/w"},

		{MappedLocationsEnabled: false},
		{MappedLocationsEnabled: false, Pwd: homeDir + abc, Expected: homeDir + abc},
		{MappedLocationsEnabled: false, Pwd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
		{MappedLocationsEnabled: false, Pwd: homeDir + cdefg, Expected: homeDir + cdefg},
		{MappedLocationsEnabled: true, Pwd: homeDir + cdefg, Expected: "~/c/d/e/f/g"},

		{MappedLocationsEnabled: true, Pwd: "/w/d/x/w", Pswd: "/z/y/x/w", Expected: "/z/y/x/w"},
		{MappedLocationsEnabled: false, Pwd: "/f/g/k/d/e/f/g", Pswd: "/a/b/c/d/e/f/g", Expected: "#/e/f/g"},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return(homeDir)
		env.On("Pwd").Return(tc.Pwd)
		env.On("GOOS").Return("")
		args := &runtime.Flags{
			PSWD: tc.Pswd,
		}
		env.On("Flags").Return(args)
		env.On("Shell").Return(shell.PWSH)
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)

		template.Init(env)

		path := &Path{
			env: env,
			props: properties.Map{
				MappedLocationsEnabled: tc.MappedLocationsEnabled,
				MappedLocations: map[string]string{
					abcd: "#",
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
		env := new(mock.Environment)
		env.On("Error", testify_.Anything)
		env.On("Debug", testify_.Anything)
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		env.On("Shell").Return(shell.GENERIC)

		template.Init(env)

		path := &Path{
			env:           env,
			pathSeparator: "/",
		}

		props := properties.Map{}

		if len(tc.FolderSeparatorTemplate) > 0 {
			props[FolderSeparatorTemplate] = tc.FolderSeparatorTemplate
		}

		if len(tc.FolderSeparatorIcon) > 0 {
			props[FolderSeparatorIcon] = tc.FolderSeparatorIcon
		}

		env.On("TemplateCache").Return(&cache.Template{
			Shell: "bash",
		})

		path.props = props
		got := path.getFolderSeparator()
		assert.Equal(t, tc.Expected, got)
	}
}

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		Case          string
		Input         string
		HomeDir       string
		GOOS          string
		PathSeparator string
		Expected      string
	}{
		{
			Case:          "Windows: absolute w/o drive letter, forward slash included",
			Input:         "/foo/~/bar",
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "\\foo\\~\\bar",
		},
		{
			Case:          "Windows: absolute",
			Input:         homeDirWindows + "\\Foo",
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "c:\\users\\someone\\foo",
		},
		{
			Case:     "Linux: home prefix, backslash included",
			Input:    "~/Bob\\Foo",
			HomeDir:  homeDir,
			GOOS:     runtime.LINUX,
			Expected: homeDir + "/Bob\\Foo",
		},
		{
			Case:     "macOS: home prefix, backslash included",
			Input:    "~/Bob\\Foo",
			HomeDir:  homeDir,
			GOOS:     runtime.DARWIN,
			Expected: homeDir + "/bob\\foo",
		},
		{
			Case:          "Windows: home prefix",
			Input:         "~\\Bob\\Foo",
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "c:\\users\\someone\\bob\\foo",
		},
		{
			Case:     "Linux: absolute",
			Input:    "/foo/~/bar",
			HomeDir:  homeDir,
			GOOS:     runtime.LINUX,
			Expected: "/foo/~/bar",
		},
		{
			Case:     "Linux: home prefix",
			Input:    "~/baz",
			HomeDir:  homeDir,
			GOOS:     runtime.LINUX,
			Expected: homeDir + "/baz",
		},
		{
			Case:          "Windows: home prefix",
			Input:         "~/baz",
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "c:\\users\\someone\\baz",
		},
		{
			Case:          "Windows: UNC root w/ prefix",
			Input:         `\\.\UNC\localhost\c$`,
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "\\\\localhost\\c$",
		},
		{
			Case:          "Windows: UNC root w/ prefix, forward slash included",
			Input:         "//./UNC/localhost/c$",
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "\\\\localhost\\c$",
		},
		{
			Case:          "Windows: UNC root",
			Input:         `\\localhost\c$\`,
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "\\\\localhost\\c$",
		},
		{
			Case:          "Windows: UNC root, forward slash included",
			Input:         "//localhost/c$",
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "\\\\localhost\\c$",
		},
		{
			Case:          "Windows: UNC",
			Input:         `\\localhost\c$\some`,
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "\\\\localhost\\c$\\some",
		},
		{
			Case:          "Windows: UNC, forward slash included",
			Input:         "//localhost/c$/some",
			HomeDir:       homeDirWindows,
			GOOS:          runtime.WINDOWS,
			PathSeparator: `\`,
			Expected:      "\\\\localhost\\c$\\some",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Home").Return(tc.HomeDir)
		env.On("GOOS").Return(tc.GOOS)

		if len(tc.PathSeparator) == 0 {
			tc.PathSeparator = "/"
		}

		env.On("PathSeparator").Return(tc.PathSeparator)
		pt := &Path{env: env}
		got := pt.normalize(tc.Input)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestReplaceMappedLocations(t *testing.T) {
	cases := []struct {
		Case                   string
		Pwd                    string
		Expected               string
		MappedLocationsEnabled bool
	}{
		{Pwd: "/c/l/k/f", Expected: "f"},
		{Pwd: "/f/g/h", Expected: "/f/g/h"},
		{Pwd: "/f/g/h/e", Expected: "^/e"},
		{Pwd: abcd, Expected: "#"},
		{Pwd: "/a/b/c/d/e", Expected: "#/e"},
		{Pwd: "/a/b/c/D/e", Expected: "#/e"},
		{Pwd: "/a/b/k/j/e", Expected: "e"},
		{Pwd: "/a/b/k/l", Expected: "@/l"},
		{Pwd: "/a/b/k/l", MappedLocationsEnabled: true, Expected: "~/l"},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("PathSeparator").Return("/")
		env.On("Pwd").Return(tc.Pwd)
		env.On("Shell").Return(shell.FISH)
		env.On("GOOS").Return(runtime.DARWIN)
		env.On("Home").Return("/a/b/k")
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)

		template.Init(env)

		path := &Path{
			env: env,
			props: properties.Map{
				MappedLocationsEnabled: tc.MappedLocationsEnabled,
				MappedLocations: map[string]string{
					abcd:       "#",
					"/f/g/h/*": "^",
					"/c/l/k/*": "",
					"~":        "@",
					"~/j/*":    "",
				},
			},
		}

		path.setPaths()
		assert.Equal(t, tc.Expected, path.pwd)
	}
}

func TestSplitPath(t *testing.T) {
	cases := []struct {
		Case         string
		GOOS         string
		Relative     string
		Root         string
		GitDir       *runtime.FileInfo
		GitDirFormat string
		Expected     Folders
	}{
		{Case: "Root directory", Root: "/", Expected: Folders{}},
		{
			Case:     "Regular directory",
			Root:     "/",
			Relative: "c/d",
			GOOS:     runtime.DARWIN,
			Expected: Folders{
				{Name: "c", Path: "/c"},
				{Name: "d", Path: "/c/d"},
			},
		},
		{
			Case:         "Home directory - git folder",
			Root:         "~",
			Relative:     "c/d",
			GOOS:         runtime.DARWIN,
			GitDir:       &runtime.FileInfo{IsDir: true, ParentFolder: "/a/b/c"},
			GitDirFormat: "<b>%s</b>",
			Expected: Folders{
				{Name: "<b>c</b>", Path: "~/c", Display: true},
				{Name: "d", Path: "~/c/d"},
			},
		},
		{
			Case:         "Home directory - git folder on Windows",
			Root:         "C:",
			Relative:     "a/b/c/d",
			GOOS:         runtime.WINDOWS,
			GitDir:       &runtime.FileInfo{IsDir: true, ParentFolder: "C:/a/b/c"},
			GitDirFormat: "<b>%s</b>",
			Expected: Folders{
				{Name: "a", Path: "C:/a"},
				{Name: "b", Path: "C:/a/b"},
				{Name: "<b>c</b>", Path: "C:/a/b/c", Display: true},
				{Name: "d", Path: "C:/a/b/c/d"},
			},
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("PathSeparator").Return("/")
		env.On("Home").Return("/a/b")
		env.On("HasParentFilePath", ".git", false).Return(tc.GitDir, nil)
		env.On("GOOS").Return(tc.GOOS)

		path := &Path{
			env: env,
			props: properties.Map{
				GitDirFormat: tc.GitDirFormat,
			},
			root:          tc.Root,
			relative:      tc.Relative,
			pathSeparator: "/",
			windowsPath:   tc.GOOS == runtime.WINDOWS,
		}

		got := path.splitPath()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestGetMaxWidth(t *testing.T) {
	cases := []struct {
		MaxWidth any
		Case     string
		Expected int
	}{
		{
			Case:     "Nil",
			Expected: 0,
		},
		{
			Case:     "Empty string",
			MaxWidth: "",
			Expected: 0,
		},
		{
			Case:     "Invalid template",
			MaxWidth: "{{ .Unknown }}",
			Expected: 0,
		},
		{
			Case:     "Environment variable",
			MaxWidth: "{{ .Env.MAX_WIDTH }}",
			Expected: 120,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
		env.On("Error", testify_.Anything).Return(nil)
		env.On("TemplateCache").Return(&cache.Template{})
		env.On("Getenv", "MAX_WIDTH").Return("120")
		env.On("Shell").Return(shell.BASH)

		template.Init(env)

		path := &Path{
			env: env,
			props: properties.Map{
				MaxWidth: tc.MaxWidth,
			},
		}

		got := path.getMaxWidth()
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}
