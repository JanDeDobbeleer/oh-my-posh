//go:build !windows

package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/stretchr/testify/assert"
)

var testParentCases = []testParentCase{
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
}

var testAgnosterPathStyleCases = []testAgnosterPathStyleCase{
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
}

var testAgnosterPathCases = []testAgnosterPathCase{
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

var testAgnosterLeftPathCases = []testAgnosterLeftPathCase{
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

var testFullAndFolderPathCases = []testFullAndFolderPathCase{
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

var testFullPathCustomMappedLocationsCases = []testFullPathCustomMappedLocationsCase{
	{Pwd: homeDir + "/d", MappedLocations: map[string]string{"{{ .Env.HOME }}/d": "#"}, Expected: "#"},
	{Pwd: abcd, MappedLocations: map[string]string{abcd: "#"}, Expected: "#"},
	{Pwd: abcd, MappedLocations: map[string]string{"/a/b": "#"}, Expected: "#/c/d"},
	{Pwd: abcd, MappedLocations: map[string]string{"/a/b": "/e/f"}, Expected: "/e/f/c/d"},
	{Pwd: homeDir + abcd, MappedLocations: map[string]string{"~/a/b": "#"}, Expected: "#/c/d"},
	{Pwd: "/a" + homeDir + "/b/c/d", MappedLocations: map[string]string{"/a~": "#"}, Expected: "/a" + homeDir + "/b/c/d"},
	{Pwd: homeDir + abcd, MappedLocations: map[string]string{"/a/b": "#"}, Expected: homeDir + abcd},
}

var testSplitPathCases = []testSplitPathCase{
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
}

var testNormalizePathCases = []testNormalizePathCase{
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

	template.Cache = new(cache.Template)
	template.Init(env, nil)

	props := properties.Map{
		properties.Style: FolderType,
		MappedLocations: map[string]string{
			abcd: "#",
		},
	}

	path := &Path{}
	path.Init(props, env)

	path.setPaths()
	path.setStyle()

	got := renderTemplateNoTrimSpace(env, "{{ .Path }}", path)
	assert.Equal(t, "#", got)
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

		template.Cache = new(cache.Template)
		template.Init(env, nil)

		props := properties.Map{
			MappedLocationsEnabled: tc.MappedLocationsEnabled,
			MappedLocations: map[string]string{
				abcd:       "#",
				"/f/g/h/*": "^",
				"/c/l/k/*": "",
				"~":        "@",
				"~/j/*":    "",
			},
		}

		path := &Path{}
		path.Init(props, env)

		path.setPaths()
		assert.Equal(t, tc.Expected, path.pwd)
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

		template.Cache = new(cache.Template)
		template.Init(env, nil)

		props := properties.Map{
			MappedLocationsEnabled: tc.MappedLocationsEnabled,
			MappedLocations: map[string]string{
				abcd: "#",
			},
		}

		path := &Path{}
		path.Init(props, env)

		path.setPaths()
		assert.Equal(t, tc.Expected, path.pwd)
	}
}
