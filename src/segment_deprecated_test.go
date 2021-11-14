package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// GIT Segment

func TestGetStatusDetailStringDefault(t *testing.T) {
	expected := "icon +1"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverride(t *testing.T) {
	expected := "<#123456>icon +1</>"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor: "#123456",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverrideAndIconColorOverride(t *testing.T) {
	expected := "<#789123>work</> <#123456>+1</>"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor:     "#123456",
				LocalWorkingIcon: "<#789123>work</>",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverrideNoIconColorOverride(t *testing.T) {
	expected := "<#123456>work +1</>"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				WorkingColor:     "#123456",
				LocalWorkingIcon: "work",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringNoStatus(t *testing.T) {
	expected := "icon"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				DisplayStatusDetail: false,
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringNoStatusColorOverride(t *testing.T) {
	expected := "<#123456>icon</>"
	status := &GitStatus{
		Changed: true,
		Added:   1,
	}
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				DisplayStatusDetail: false,
				WorkingColor:        "#123456",
			},
			foreground: "#111111",
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusColorLocalChangesStaging(t *testing.T) {
	expected := changesColor
	g := &git{
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: expected,
			},
		},
		Staging: &GitStatus{
			Changed: true,
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorLocalChangesWorking(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{
			Changed: true,
		},
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorAheadAndBehind(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   1,
		Behind:  3,
		props: &properties{
			values: map[Property]interface{}{
				AheadAndBehindColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorAhead(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   1,
		Behind:  0,
		props: &properties{
			values: map[Property]interface{}{
				AheadColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorBehind(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   0,
		Behind:  5,
		props: &properties{
			values: map[Property]interface{}{
				BehindColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorDefault(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   0,
		Behind:  0,
		props: &properties{
			values: map[Property]interface{}{
				BehindColor: changesColor,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor(expected))
}

func TestSetStatusColorForeground(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{
			Changed: true,
		},
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: changesColor,
				ColorBackground:   false,
			},
			foreground: "#ffffff",
			background: "#111111",
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.foreground)
}

func TestSetStatusColorBackground(t *testing.T) {
	expected := changesColor
	g := &git{
		Staging: &GitStatus{
			Changed: true,
		},
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: changesColor,
				ColorBackground:   true,
			},
			foreground: "#ffffff",
			background: "#111111",
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.background)
}

func TestStatusColorsWithoutDisplayStatus(t *testing.T) {
	expected := changesColor
	context := &detachedContext{
		status: "## main...origin/main [ahead 33]\n M myfile",
	}
	g := setupHEADContextEnv(context)
	g.props = &properties{
		values: map[Property]interface{}{
			DisplayStatus:       false,
			StatusColorsEnabled: true,
			LocalChangesColor:   expected,
		},
	}
	g.string()
	assert.Equal(t, expected, g.props.background)
}

// EXIT Segement

func TestExitWriterDeprecatedString(t *testing.T) {
	cases := []struct {
		ExitCode        int
		Expected        string
		SuccessIcon     string
		ErrorIcon       string
		DisplayExitCode bool
		AlwaysNumeric   bool
	}{
		{ExitCode: 129, Expected: "SIGHUP", DisplayExitCode: true},
		{ExitCode: 5001, Expected: "5001", DisplayExitCode: true},
		{ExitCode: 147, Expected: "SIGSTOP", DisplayExitCode: true},
		{ExitCode: 147, Expected: "", DisplayExitCode: false},
		{ExitCode: 147, Expected: "147", DisplayExitCode: true, AlwaysNumeric: true},
		{ExitCode: 0, Expected: "wooopie", SuccessIcon: "wooopie"},
		{ExitCode: 129, Expected: "err SIGHUP", ErrorIcon: "err ", DisplayExitCode: true},
		{ExitCode: 129, Expected: "err", ErrorIcon: "err", DisplayExitCode: false},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("lastErrorCode", nil).Return(tc.ExitCode)
		props := &properties{
			foreground: "#111111",
			background: "#ffffff",
			values: map[Property]interface{}{
				SuccessIcon:     tc.SuccessIcon,
				ErrorIcon:       tc.ErrorIcon,
				DisplayExitCode: tc.DisplayExitCode,
				AlwaysNumeric:   tc.AlwaysNumeric,
			},
		}
		e := &exit{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.Expected, e.string())
	}
}
