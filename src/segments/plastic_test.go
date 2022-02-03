package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlasticEnabledNotFound(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "cm").Return(false)
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)
	p := &Plastic{
		scm: scm{
			env:   env,
			props: properties.Map{},
		},
	}
	assert.False(t, p.Enabled())
}

func TestPlasticEnabledInWorkspaceDirectory(t *testing.T) {
	env := new(mock.MockedEnvironment)
	env.On("HasCommand", "cm").Return(true)
	env.On("GOOS").Return("")
	env.On("IsWsl").Return(false)
	env.On("FileContent", "/dir/.plastic//plastic.selector").Return("")
	fileInfo := &environment.FileInfo{
		Path:         "/dir/hello",
		ParentFolder: "/dir",
		IsDir:        true,
	}
	env.On("HasParentFilePath", ".plastic").Return(fileInfo, nil)
	p := &Plastic{
		scm: scm{
			env:   env,
			props: properties.Map{},
		},
	}
	assert.True(t, p.Enabled())
	assert.Equal(t, fileInfo.ParentFolder, p.plasticWorkspaceFolder)
}

func setupCmStatusEnv(status, headStatus string) *Plastic {
	env := new(mock.MockedEnvironment)
	env.On("RunCommand", "cm", []string{"status", "--all", "--machinereadable"}).Return(status, nil)
	env.On("RunCommand", "cm", []string{"status", "--head", "--machinereadable"}).Return(headStatus, nil)
	p := &Plastic{
		scm: scm{
			env:   env,
			props: properties.Map{},
		},
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
		p.setPlasticStatus()
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
			Status:   "STATUS 1 default localhost:8087\r\nCO /some.file NO_MERGES",
		},
	}

	for _, tc := range cases {
		p := setupCmStatusEnv(tc.Status, "")
		p.setPlasticStatus()
		assert.Equal(t, tc.Expected, p.Status.Changed(), tc.Case)
	}
}

func TestPlasticStatusCounts(t *testing.T) {
	status := "STATUS 1 default localhost:8087" +
		"\r\nCO /some.file NO_MERGES" +
		"\r\nAD /some.file" +
		"\r\nCH /some.file\r\nCH /some.file" +
		"\r\nLD /some.file\r\nLD /some.file\r\nLD /some.file" +
		"\r\nLM /some.file\r\nLM /some.file\r\nLM /some.file\r\nLM /some.file"
	p := setupCmStatusEnv(status, "")
	p.setPlasticStatus()
	s := p.Status
	assert.Equal(t, 1, s.Unmerged)
	assert.Equal(t, 1, s.Added)
	assert.Equal(t, 2, s.Modified)
	assert.Equal(t, 3, s.Deleted)
	assert.Equal(t, 4, s.Moved)
}

func TestPlasticMergePending(t *testing.T) {
	cases := []struct {
		Case     string
		Expected bool
		Status   string
	}{
		{
			Case:     "No pending merge",
			Expected: false,
			Status:   "STATUS 1 default localhost:8087",
		},
		{
			Case:     "Pending merge",
			Expected: true,
			Status:   "STATUS 1 default localhost:8087\r\nCH /some.file merge from 8",
		},
	}
	for _, tc := range cases {
		p := setupCmStatusEnv(tc.Status, "")
		p.setPlasticStatus()
		assert.Equal(t, tc.Expected, p.MergePending, tc.Case)
	}
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

	p := &Plastic{}
	for _, tc := range cases {
		value := p.parseIntPattern(tc.Text, tc.Pattern, tc.Name, tc.Default)
		assert.Equal(t, tc.Expected, value, tc.Case)
	}
}

func TestPlasticParseStatusChangeset(t *testing.T) {
	p := &Plastic{}
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
	p := &Plastic{}
	selector := p.parseChangesetSelector(content)
	assert.Equal(t, "321", selector)
}

func TestPlasticParseLabelSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  label \"BL003\""
	p := &Plastic{}
	selector := p.parseLabelSelector(content)
	assert.Equal(t, "BL003", selector)
}

func TestPlasticParseBranchSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  branch \"/main/fix-004\""
	p := &Plastic{}
	selector := p.parseBranchSelector(content)
	assert.Equal(t, "/main/fix-004", selector)
}

func TestPlasticParseSmartbranchSelector(t *testing.T) {
	content := "repository \"default\"\r\n	path \"/\"\r\n	  smartbranch \"/main/fix-002\""
	p := &Plastic{}
	selector := p.parseBranchSelector(content)
	assert.Equal(t, "/main/fix-002", selector)
}

func TestPlasticStatus(t *testing.T) {
	p := &Plastic{
		Status: &PlasticStatus{
			ScmStatus: ScmStatus{
				Added:    1,
				Modified: 2,
				Deleted:  3,
				Moved:    4,
				Unmerged: 5,
			},
		},
	}
	status := p.Status.String()
	expected := "+1 ~2 -3 >4 x5"
	assert.Equal(t, expected, status)
}

func TestPlasticTemplateString(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
		Plastic  *Plastic
	}{
		{
			Case:     "Default template",
			Expected: "/main",
			Template: "{{ .Selector }}",
			Plastic: &Plastic{
				Selector: "/main",
				Behind:   false,
			},
		},
		{
			Case:     "Workspace changes",
			Expected: "/main \uF044 +2 ~3 -1 >4",
			Template: "{{ .Selector }}{{ if .Status.Changed }} \uF044 {{ .Status.String }}{{ end }}",
			Plastic: &Plastic{
				Selector: "/main",
				Status: &PlasticStatus{
					ScmStatus: ScmStatus{
						Added:    2,
						Modified: 3,
						Deleted:  1,
						Moved:    4,
					},
				},
			},
		},
		{
			Case:     "No workspace changes",
			Expected: "/main",
			Template: "{{ .Selector }}{{ if .Status.Changed }} \uF044 {{ .Status.String }}{{ end }}",
			Plastic: &Plastic{
				Selector: "/main",
				Status:   &PlasticStatus{},
			},
		},
	}

	for _, tc := range cases {
		props := properties.Map{
			FetchStatus: true,
		}
		tc.Plastic.props = props
		env := new(mock.MockedEnvironment)
		tc.Plastic.env = env
		assert.Equal(t, tc.Expected, renderTemplate(env, tc.Template, tc.Plastic), tc.Case)
	}
}
