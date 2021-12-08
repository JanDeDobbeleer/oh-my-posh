package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlasticEnabledNotFound(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "cm").Return(false)
	env.On("getRuntimeGOOS", nil).Return("")
	env.On("isWsl", nil).Return(false)
	g := &plastic{
		env: env,
	}
	assert.False(t, g.enabled())
}

func TestPlasticEnabledInWorkspaceDirectory(t *testing.T) {
	env := new(MockedEnvironment)
	env.On("hasCommand", "cm").Return(true)
	env.On("getRuntimeGOOS", nil).Return("")
	env.On("isWsl", nil).Return(false)
	fileInfo := &fileInfo{
		path:         "/dir/hello",
		parentFolder: "/dir",
		isDir:        true,
	}
	env.On("hasParentFilePath", ".plastic").Return(fileInfo, nil)
	g := &plastic{
		env: env,
	}
	assert.True(t, g.enabled())
	assert.Equal(t, fileInfo.parentFolder, g.plasticWorkspaceFolder)
}

func setupCmStatusEnv(status, headStatus string) *plastic {
	env := new(MockedEnvironment)
	env.On("isWsl", nil).Return(false)
	env.On("runCommand", "cm", []string{"status", "--all", "--machinereadable"}).Return(status, nil)
	env.On("runCommand", "cm", []string{"status", "--head", "--machinereadable"}).Return(headStatus, nil)
	env.On("getRuntimeGOOS", nil).Return("unix")
	g := &plastic{
		env: env,
	}
	return g
}

func TestPlasticGetCmOutputForCommand(t *testing.T) {
	want := "je suis le output"
	g := setupCmStatusEnv(want, "")
	got := g.getCmCommandOutput("status", "--all", "--machinereadable")
	assert.Equal(t, want, got)
}

func TestPlasticStatusBehind(t *testing.T) {
	cases := []struct {
		Case     string
		Expected bool
		Status   string
		Head     string
	}{
		{
			Case:     "Not behind",
			Expected: false,
			Status:   "STATUS 20 default localhost:8087",
			Head:     "STATUS cs:20 rep:default repserver:localhost:8087",
		},
		{
			Case:     "Behind",
			Expected: true,
			Status:   "STATUS 2 default localhost:8087",
			Head:     "STATUS cs:20 rep:default repserver:localhost:8087",
		},
	}

	for _, tc := range cases {
		g := setupCmStatusEnv(tc.Status, tc.Head)
		g.getPlasticStatus()
		assert.Equal(t, tc.Expected, g.Behind, tc.Case)
	}
}

func TestPlasticStatusChanged(t *testing.T) {
	cases := []struct {
		Case     string
		Expected bool
		Status   string
	}{
		{
			Case:     "No changes",
			Expected: false,
			Status:   "STATUS 1 default localhost:8087",
		},
		{
			Case:     "Changed file",
			Expected: true,
			Status:   "STATUS 1 default localhost:8087\r\nCH /some.file",
		},
		{
			Case:     "Added file",
			Expected: true,
			Status:   "STATUS 1 default localhost:8087\r\nAD /some.file",
		},
		{
			Case:     "Added (pivate) file",
			Expected: true,
			Status:   "STATUS 1 default localhost:8087\r\nPR /some.file",
		},
		{
			Case:     "Moved file",
			Expected: true,
			Status:   "STATUS 1 default localhost:8087\r\nLM /some.file",
		},
		{
			Case:     "Deleted file",
			Expected: true,
			Status:   "STATUS 1 default localhost:8087\r\nLD /some.file",
		},
		{
			Case:     "Unmerged file",
			Expected: true,
			Status:   "STATUS 1 default localhost:8087\r\nCP /some.file Merge from 321",
		},
	}

	for _, tc := range cases {
		g := setupCmStatusEnv(tc.Status, "")
		g.getPlasticStatus()
		assert.Equal(t, tc.Expected, g.Changed, tc.Case)
	}
}

func TestPlasticStatusCounts(t *testing.T) {
	status := "STATUS 1 default localhost:8087" +
		"\r\nCP /some.file Merge from 321" +
		"\r\nAD /some.file" +
		"\r\nCH /some.file\r\nCH /some.file" +
		"\r\nLD /some.file\r\nLD /some.file\r\nLD /some.file" +
		"\r\nLM /some.file\r\nLM /some.file\r\nLM /some.file\r\nLM /some.file"
	g := setupCmStatusEnv(status, "")
	g.getPlasticStatus()
	assert.Equal(t, 1, g.Unmerged)
	assert.Equal(t, 1, g.Added)
	assert.Equal(t, 2, g.Modified)
	assert.Equal(t, 3, g.Deleted)
	assert.Equal(t, 4, g.Moved)
}

func TestPlasticParseStatusChangeset(t *testing.T) {
	g := &plastic{}
	cs := g.parseStatusChangeset("STATUS 321 default localhost:8087")
	assert.Equal(t, 321, cs)
}

func TestPlasticGetHeadChangeset(t *testing.T) {
	head := "STATUS cs:321 rep:default repserver:localhost:8087"
	g := setupCmStatusEnv("", head)
	cs := g.getHeadChangeset()
	assert.Equal(t, 321, cs)
}

func TestPlasticParseChangesetSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  smartbranch \"/main\" changeset \"321\""
	g := &plastic{}
	selector := g.parseChangesetSelector(content)
	assert.Equal(t, "321", selector)
}

func TestPlasticParseLabelSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  label \"BL003\""
	g := &plastic{}
	selector := g.parseLabelSelector(content)
	assert.Equal(t, "BL003", selector)
}

func TestPlasticParseBranchSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  branch \"/main/fix-004\""
	g := &plastic{}
	selector := g.parseBranchSelector(content)
	assert.Equal(t, "/main/fix-004", selector)
}

func TestPlasticParseSmartbranchSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  smartbranch \"/main/fix-002\""
	g := &plastic{}
	selector := g.parseBranchSelector(content)
	assert.Equal(t, "/main/fix-002", selector)
}

func TestPlasticStatus(t *testing.T) {
	g := &plastic{
		Changed:  true,
		Added:    1,
		Modified: 2,
		Deleted:  3,
		Moved:    4,
		Unmerged: 5,
	}
	status := g.Status()
	expected := "+1 ~2 -3 >4 x5"
	assert.Equal(t, expected, status)
}

func TestPlasticShouldIgnoreRootRepository(t *testing.T) {
	cases := []struct {
		Case     string
		Dir      string
		Expected bool
	}{
		{Case: "inside excluded", Dir: "/home/bill/repo"},
		{Case: "oustide excluded", Dir: "/home/melinda"},
		{Case: "excluded exact match", Dir: "/home/gates", Expected: true},
		{Case: "excluded inside match", Dir: "/home/gates/bill", Expected: true},
	}

	for _, tc := range cases {
		var props properties = map[Property]interface{}{
			ExcludeFolders: []string{
				"/home/bill",
				"/home/gates.*",
			},
		}
		env := new(MockedEnvironment)
		env.On("homeDir", nil).Return("/home/bill")
		env.On("getRuntimeGOOS", nil).Return(windowsPlatform)
		plastic := &plastic{
			props: props,
			env:   env,
		}
		got := plastic.shouldIgnoreRootRepository(tc.Dir)
		assert.Equal(t, tc.Expected, got, tc.Case)
	}
}

func TestPlasticTruncateBranch(t *testing.T) {
	cases := []struct {
		Case       string
		Expected   string
		Branch     string
		FullBranch bool
		MaxLength  interface{}
	}{
		{Case: "No limit", Expected: "are-belong-to-us", Branch: "/all-your-base/are-belong-to-us", FullBranch: false},
		{Case: "No limit - larger", Expected: "are-belong", Branch: "/all-your-base/are-belong-to-us", FullBranch: false, MaxLength: 10.0},
		{Case: "No limit - smaller", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: 13.0},
		{Case: "Invalid setting", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: "burp"},
		{Case: "Lower than limit", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: 20.0},

		{Case: "No limit - full branch", Expected: "/all-your-base/are-belong-to-us", Branch: "/all-your-base/are-belong-to-us", FullBranch: true},
		{Case: "No limit - larger - full branch", Expected: "/all-your-base", Branch: "/all-your-base/are-belong-to-us", FullBranch: true, MaxLength: 14.0},
		{Case: "No limit - smaller - full branch ", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: 14.0},
		{Case: "Invalid setting - full branch", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: "burp"},
		{Case: "Lower than limit - full branch", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: 20.0},
	}

	for _, tc := range cases {
		var props properties = map[Property]interface{}{
			BranchMaxLength: tc.MaxLength,
			FullBranchPath:  tc.FullBranch,
		}
		g := &plastic{
			props: props,
		}
		assert.Equal(t, tc.Expected, g.truncateBranch(tc.Branch), tc.Case)
	}
}

func TestPlasticTruncateBranchWithSymbol(t *testing.T) {
	cases := []struct {
		Case           string
		Expected       string
		Branch         string
		FullBranch     bool
		MaxLength      interface{}
		TruncateSymbol interface{}
	}{
		{Case: "No limit", Expected: "are-belong-to-us", Branch: "/all-your-base/are-belong-to-us", FullBranch: false, TruncateSymbol: "..."},
		{Case: "No limit - larger", Expected: "are-belong...", Branch: "/all-your-base/are-belong-to-us", FullBranch: false, MaxLength: 10.0, TruncateSymbol: "..."},
		{Case: "No limit - smaller", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: 13.0, TruncateSymbol: "..."},
		{Case: "Invalid setting", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: "burp", TruncateSymbol: "..."},
		{Case: "Lower than limit", Expected: "all-your-base", Branch: "/all-your-base", FullBranch: false, MaxLength: 20.0, TruncateSymbol: "..."},

		{Case: "No limit - full branch", Expected: "/all-your-base/are-belong-to-us", Branch: "/all-your-base/are-belong-to-us", FullBranch: true, TruncateSymbol: "..."},
		{Case: "No limit - larger - full branch", Expected: "/all-your-base...", Branch: "/all-your-base/are-belong-to-us", FullBranch: true, MaxLength: 14.0, TruncateSymbol: "..."},
		{Case: "No limit - smaller - full branch ", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: 14.0, TruncateSymbol: "..."},
		{Case: "Invalid setting - full branch", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: "burp", TruncateSymbol: "..."},
		{Case: "Lower than limit - full branch", Expected: "/all-your-base", Branch: "/all-your-base", FullBranch: true, MaxLength: 20.0, TruncateSymbol: "..."},
	}

	for _, tc := range cases {
		var props properties = map[Property]interface{}{
			BranchMaxLength: tc.MaxLength,
			TruncateSymbol:  tc.TruncateSymbol,
			FullBranchPath:  tc.FullBranch,
		}
		g := &plastic{
			props: props,
		}
		assert.Equal(t, tc.Expected, g.truncateBranch(tc.Branch), tc.Case)
	}
}

func TestGetCmCommand(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		IsWSL    bool
		IsWSL1   bool
		GOOS     string
		CWD      string
	}{
		{Case: "On Windows", Expected: "cm.exe", GOOS: windowsPlatform},
		{Case: "Non Windows", Expected: "cm"},
		{Case: "Iside WSL2, non shared", IsWSL: true, Expected: "cm"},
		{Case: "Iside WSL2, shared", Expected: "cm.exe", IsWSL: true, CWD: "/mnt/bill"},
		{Case: "Iside WSL1, shared", Expected: "cm", IsWSL: true, IsWSL1: true, CWD: "/mnt/bill"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("isWsl", nil).Return(tc.IsWSL)
		env.On("getRuntimeGOOS", nil).Return(tc.GOOS)
		env.On("getcwd", nil).Return(tc.CWD)
		wslUname := "5.10.60.1-microsoft-standard-WSL2"
		if tc.IsWSL1 {
			wslUname = "4.4.0-19041-Microsoft"
		}
		env.On("runCommand", "uname", []string{"-r"}).Return(wslUname, nil)
		g := &plastic{
			env: env,
		}
		assert.Equal(t, tc.Expected, g.getCmCommand(), tc.Case)
	}
}

func TestPlasticTemplateString(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
		Plastic  *plastic
	}{
		{
			Case:     "Only Selector name",
			Expected: "/main",
			Template: "{{ .Selector }}",
			Plastic: &plastic{
				Selector: "/main",
				Behind:   false,
			},
		},
		{
			Case:     "Workspace changes",
			Expected: "/main \uF044 +2 ~3 -1 >4",
			Template: "{{ .Selector }}{{ if .Changed }} \uF044 {{ .Status }}{{ end }}",
			Plastic: &plastic{
				Selector: "/main",
				Changed:  true,
				Added:    2,
				Modified: 3,
				Deleted:  1,
				Moved:    4,
			},
		},
		{
			Case:     "No workspace changes",
			Expected: "/main",
			Template: "{{ .Selector }}{{ if .Changed }} \uF044 {{ .Status }}{{ end }}",
			Plastic: &plastic{
				Selector: "/main",
				Changed:  false,
			},
		},
	}

	for _, tc := range cases {
		var props properties = map[Property]interface{}{
			FetchStatus: true,
		}
		tc.Plastic.props = props
		assert.Equal(t, tc.Expected, tc.Plastic.templateString(tc.Template), tc.Case)
	}
}
