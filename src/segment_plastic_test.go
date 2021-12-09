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
	p := &plastic{
		env: env,
	}
	assert.False(t, p.enabled())
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
	p := &plastic{
		env: env,
	}
	assert.True(t, p.enabled())
	assert.Equal(t, fileInfo.parentFolder, p.plasticWorkspaceFolder)
}

func setupCmStatusEnv(status, headStatus string) *plastic {
	env := new(MockedEnvironment)
	env.On("isWsl", nil).Return(false)
	env.On("runCommand", "cm", []string{"status", "--all", "--machinereadable"}).Return(status, nil)
	env.On("runCommand", "cm", []string{"status", "--head", "--machinereadable"}).Return(headStatus, nil)
	env.On("getRuntimeGOOS", nil).Return("unix")
	p := &plastic{
		env: env,
	}
	return p
}

func TestPlasticGetCmOutputForCommand(t *testing.T) {
	want := "je suis le output"
	p := setupCmStatusEnv(want, "")
	got := p.getCmCommandOutput("status", "--all", "--machinereadable")
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
		p := setupCmStatusEnv(tc.Status, tc.Head)
		p.getPlasticStatus()
		assert.Equal(t, tc.Expected, p.Behind, tc.Case)
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
		p := setupCmStatusEnv(tc.Status, "")
		p.getPlasticStatus()
		assert.Equal(t, tc.Expected, p.Changed, tc.Case)
	}
}

func TestPlasticStatusCounts(t *testing.T) {
	status := "STATUS 1 default localhost:8087" +
		"\r\nCP /some.file Merge from 321" +
		"\r\nAD /some.file" +
		"\r\nCH /some.file\r\nCH /some.file" +
		"\r\nLD /some.file\r\nLD /some.file\r\nLD /some.file" +
		"\r\nLM /some.file\r\nLM /some.file\r\nLM /some.file\r\nLM /some.file"
	p := setupCmStatusEnv(status, "")
	p.getPlasticStatus()
	assert.Equal(t, 1, p.Unmerged)
	assert.Equal(t, 1, p.Added)
	assert.Equal(t, 2, p.Modified)
	assert.Equal(t, 3, p.Deleted)
	assert.Equal(t, 4, p.Moved)
}

func TestPlasticParseIntPattern(t *testing.T) {
	cases := []struct {
		Case     string
		Expected int
		Text     string
		Pattern  string
		Name     string
		Default  int
	}{
		{
			Case:     "int found",
			Expected: 123,
			Text:     "Some number 123 in text",
			Pattern:  `\s(?P<x>[0-9]+?)\s`,
			Name:     "x",
			Default:  0,
		},
		{
			Case:     "int not found",
			Expected: 0,
			Text:     "No number in text",
			Pattern:  `\s(?P<x>[0-9]+?)\s`,
			Name:     "x",
			Default:  0,
		},
		{
			Case:     "empty text",
			Expected: 0,
			Text:     "",
			Pattern:  `\s(?P<x>[0-9]+?)\s`,
			Name:     "x",
			Default:  0,
		},
	}

	p := &plastic{}
	for _, tc := range cases {
		value := p.parseIntPattern(tc.Text, tc.Pattern, tc.Name, tc.Default)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}

func TestPlasticParseStatusChangeset(t *testing.T) {
	p := &plastic{}
	cs := p.parseStatusChangeset("STATUS 321 default localhost:8087")
	assert.Equal(t, 321, cs)
}

func TestPlasticGetHeadChangeset(t *testing.T) {
	head := "STATUS cs:321 rep:default repserver:localhost:8087"
	p := setupCmStatusEnv("", head)
	cs := p.getHeadChangeset()
	assert.Equal(t, 321, cs)
}

func TestPlasticParseChangesetSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  smartbranch \"/main\" changeset \"321\""
	p := &plastic{}
	selector := p.parseChangesetSelector(content)
	assert.Equal(t, "321", selector)
}

func TestPlasticParseLabelSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  label \"BL003\""
	p := &plastic{}
	selector := p.parseLabelSelector(content)
	assert.Equal(t, "BL003", selector)
}

func TestPlasticParseBranchSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  branch \"/main/fix-004\""
	p := &plastic{}
	selector := p.parseBranchSelector(content)
	assert.Equal(t, "/main/fix-004", selector)
}

func TestPlasticParseSmartbranchSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  smartbranch \"/main/fix-002\""
	p := &plastic{}
	selector := p.parseBranchSelector(content)
	assert.Equal(t, "/main/fix-002", selector)
}

func TestPlasticStatus(t *testing.T) {
	p := &plastic{
		Changed:  true,
		Added:    1,
		Modified: 2,
		Deleted:  3,
		Moved:    4,
		Unmerged: 5,
	}
	status := p.Status()
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
		p := &plastic{
			props: props,
			env:   env,
		}
		got := p.shouldIgnoreRootRepository(tc.Dir)
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
		p := &plastic{
			props: props,
		}
		assert.Equal(t, tc.Expected, p.truncateBranch(tc.Branch), tc.Case)
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
		p := &plastic{
			props: props,
		}
		assert.Equal(t, tc.Expected, p.truncateBranch(tc.Branch), tc.Case)
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
