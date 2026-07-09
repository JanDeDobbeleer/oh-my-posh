package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestDvcStatus(t *testing.T) {
	cases := []struct {
		OutputError      error
		Case             string
		Output           string
		ExpectedStatus   string
		HasCommand       bool
		HasDvcDir        bool
		ExpectedDisabled bool
	}{
		{
			Case:             "not installed",
			HasCommand:       false,
			ExpectedDisabled: true,
		},
		{
			Case:             "not in DVC repo",
			HasCommand:       true,
			HasDvcDir:        false,
			ExpectedDisabled: true,
		},
		{
			Case:             "command error",
			HasCommand:       true,
			HasDvcDir:        true,
			OutputError:      fmt.Errorf("error"),
			ExpectedDisabled: true,
		},
		{
			Case:           "clean status",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         "{}",
			ExpectedStatus: "",
		},
		{
			Case:           "invalid json is ignored",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         "not json",
			ExpectedStatus: "",
		},
		{
			Case:           "modified files",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         `{"data.xml.dvc": [{"changed outs": {"data.xml": "modified"}}], "model.pkl.dvc": [{"changed outs": {"model.pkl": "modified"}}]}`,
			ExpectedStatus: "~2",
		},
		{
			Case:           "new files",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         `{"data.dvc": [{"changed outs": {"data/new.csv": "new"}}], "test.dvc": [{"changed outs": {"data/test.csv": "new"}}]}`,
			ExpectedStatus: "+2",
		},
		{
			Case:           "deleted files",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         `{"data.xml.dvc": [{"changed outs": {"data.xml": "deleted"}}]}`,
			ExpectedStatus: "-1",
		},
		{
			Case:           "not in cache",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         `{"data.dvc": [{"changed outs": {"data": "not in cache"}}]}`,
			ExpectedStatus: "!1",
		},
		{
			Case:           "mixed status",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         `{"a.dvc": [{"changed outs": {"a": "new"}}], "b.dvc": [{"changed outs": {"b": "modified"}}], "c.dvc": [{"changed outs": {"c": "deleted"}}], "d.dvc": [{"changed outs": {"d": "not in cache"}}]}`, //nolint:lll
			ExpectedStatus: "+1 ~1 -1 !1",
		},
		{
			// real `dvc status --json` output from a pipeline: a file can show up
			// as both a changed dep and a changed out across stages.
			Case:           "pipeline with changed deps and outs",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         `{"other.txt.dvc": [{"changed outs": {"other.txt": "modified"}}], "build": [{"changed deps": {"input.txt": "modified"}}, {"changed outs": {"output.txt": "modified"}}], "input.txt.dvc": [{"changed outs": {"input.txt": "modified"}}]}`, //nolint:lll
			ExpectedStatus: "~4",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("GOOS").Return("unix")
		env.On("IsWsl").Return(false)
		env.On("InWSLSharedDrive").Return(false)
		env.On("HasCommand", DVCCOMMAND).Return(tc.HasCommand)

		if tc.HasDvcDir {
			env.On("HasParentFilePath", ".dvc", false).Return(&runtime.FileInfo{}, nil)
		} else {
			env.On("HasParentFilePath", ".dvc", false).Return(&runtime.FileInfo{}, errors.New("not found"))
		}

		env.On("RunCommand", DVCCOMMAND, []string{"status", "--json"}).Return(tc.Output, tc.OutputError)

		d := &Dvc{}
		d.Init(options.Map{}, env)

		got := d.Enabled()

		assert.Equal(t, !tc.ExpectedDisabled, got, tc.Case)
		if tc.ExpectedDisabled {
			continue
		}

		assert.Equal(t, tc.ExpectedStatus, d.Status.String(), tc.Case)
	}
}
