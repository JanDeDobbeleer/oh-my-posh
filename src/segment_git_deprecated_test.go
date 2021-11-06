package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	repo := &Repo{
		Staging: &GitStatus{
			Changed: true,
		},
	}
	g := &git{
		Repo: repo,
		props: &properties{
			values: map[Property]interface{}{
				LocalChangesColor: expected,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorLocalChangesWorking(t *testing.T) {
	expected := changesColor
	repo := &Repo{
		Staging: &GitStatus{},
		Working: &GitStatus{
			Changed: true,
		},
	}
	g := &git{
		Repo: repo,
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
	repo := &Repo{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   1,
		Behind:  3,
	}
	g := &git{
		Repo: repo,
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
	repo := &Repo{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   1,
		Behind:  0,
	}
	g := &git{
		Repo: repo,
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
	repo := &Repo{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   0,
		Behind:  5,
	}
	g := &git{
		Repo: repo,
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
	repo := &Repo{
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   0,
		Behind:  0,
	}
	g := &git{
		Repo: repo,
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
	repo := &Repo{
		Staging: &GitStatus{
			Changed: true,
		},
	}
	g := &git{
		Repo: repo,
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
	repo := &Repo{
		Staging: &GitStatus{
			Changed: true,
		},
	}
	g := &git{
		Repo: repo,
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
