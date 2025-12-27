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
		Case             string
		Output           string
		OutputError      error
		HasCommand       bool
		HasDvcDir        bool
		ExpectedStatus   string
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
			Output:         "",
			ExpectedStatus: "",
		},
		{
			Case:       "modified files",
			HasCommand: true,
			HasDvcDir:  true,
			Output: `data.xml: modified
model.pkl: modified`,
			ExpectedStatus: "~2",
		},
		{
			Case:       "new files",
			HasCommand: true,
			HasDvcDir:  true,
			Output: `data/new.csv: new
data/test.csv: new`,
			ExpectedStatus: "+2",
		},
		{
			Case:           "deleted files",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         `data.xml: deleted`,
			ExpectedStatus: "-1",
		},
		{
			Case:           "not in cache",
			HasCommand:     true,
			HasDvcDir:      true,
			Output:         `data/: not in cache`,
			ExpectedStatus: "!1",
		},
		{
			Case:       "mixed status",
			HasCommand: true,
			HasDvcDir:  true,
			Output: `data.xml: modified
model.pkl: new
old_data.csv: deleted
cache_data/: not in cache`,
			ExpectedStatus: "+1 ~1 -1 !1",
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

		env.On("RunCommand", DVCCOMMAND, []string{"status", "-q"}).Return(tc.Output, tc.OutputError)

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
