package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/distatus/battery"
	"github.com/stretchr/testify/assert"
)

// GIT Segment

func TestGetStatusDetailStringDefault(t *testing.T) {
	expected := "icon +1"
	status := &GitStatus{
		ScmStatus: ScmStatus{
			Added: 1,
		},
	}
	g := &git{
		scm: scm{
			props: properties{},
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverride(t *testing.T) {
	expected := "<#123456>icon +1</>"
	status := &GitStatus{
		ScmStatus: ScmStatus{
			Added: 1,
		},
	}
	props := properties{
		WorkingColor: "#123456",
	}
	g := &git{
		scm: scm{
			props: props,
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverrideAndIconColorOverride(t *testing.T) {
	expected := "<#789123>work</> <#123456>+1</>"
	status := &GitStatus{
		ScmStatus: ScmStatus{
			Added: 1,
		},
	}
	props := properties{
		WorkingColor:     "#123456",
		LocalWorkingIcon: "<#789123>work</>",
	}
	g := &git{
		scm: scm{
			props: props,
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringDefaultColorOverrideNoIconColorOverride(t *testing.T) {
	expected := "<#123456>work +1</>"
	status := &GitStatus{
		ScmStatus: ScmStatus{
			Added: 1,
		},
	}
	props := properties{
		WorkingColor:     "#123456",
		LocalWorkingIcon: "work",
	}
	g := &git{
		scm: scm{
			props: props,
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringNoStatus(t *testing.T) {
	expected := "icon"
	status := &GitStatus{
		ScmStatus: ScmStatus{
			Added: 1,
		},
	}
	props := properties{
		DisplayStatusDetail: false,
	}
	g := &git{
		scm: scm{
			props: props,
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusDetailStringNoStatusColorOverride(t *testing.T) {
	expected := "<#123456>icon</>"
	status := &GitStatus{
		ScmStatus: ScmStatus{
			Added: 1,
		},
	}
	props := properties{
		DisplayStatusDetail: false,
		WorkingColor:        "#123456",
	}
	g := &git{
		scm: scm{
			props: props,
		},
	}
	assert.Equal(t, expected, g.getStatusDetailString(status, WorkingColor, LocalWorkingIcon, "icon"))
}

func TestGetStatusColorLocalChangesStaging(t *testing.T) {
	expected := changesColor
	props := properties{
		LocalChangesColor: expected,
	}
	g := &git{
		scm: scm{
			props: props,
		},
		Staging: &GitStatus{
			ScmStatus: ScmStatus{
				Modified: 1,
			},
		},
		Working: &GitStatus{},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorLocalChangesWorking(t *testing.T) {
	expected := changesColor
	props := properties{
		LocalChangesColor: expected,
	}
	g := &git{
		scm: scm{
			props: props,
		},
		Staging: &GitStatus{},
		Working: &GitStatus{
			ScmStatus: ScmStatus{
				Modified: 1,
			},
		},
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorAheadAndBehind(t *testing.T) {
	expected := changesColor
	props := properties{
		AheadAndBehindColor: expected,
	}
	g := &git{
		scm: scm{
			props: props,
		},
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   1,
		Behind:  3,
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorAhead(t *testing.T) {
	expected := changesColor
	props := properties{
		AheadColor: expected,
	}
	g := &git{
		scm: scm{
			props: props,
		},
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   1,
		Behind:  0,
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorBehind(t *testing.T) {
	expected := changesColor
	props := properties{
		BehindColor: expected,
	}
	g := &git{
		scm: scm{
			props: props,
		},
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   0,
		Behind:  5,
	}
	assert.Equal(t, expected, g.getStatusColor("#fg1111"))
}

func TestGetStatusColorDefault(t *testing.T) {
	expected := changesColor
	props := properties{
		BehindColor: changesColor,
	}
	g := &git{
		scm: scm{
			props: props,
		},
		Staging: &GitStatus{},
		Working: &GitStatus{},
		Ahead:   0,
		Behind:  0,
	}
	assert.Equal(t, expected, g.getStatusColor(expected))
}

func TestSetStatusColorForeground(t *testing.T) {
	expected := changesColor
	props := properties{
		LocalChangesColor: changesColor,
		ColorBackground:   false,
	}
	g := &git{
		scm: scm{
			props: props,
		},
		Staging: &GitStatus{
			ScmStatus: ScmStatus{
				Added: 1,
			},
		},
		Working: &GitStatus{},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.getColor(ForegroundOverride, ""))
}

func TestSetStatusColorBackground(t *testing.T) {
	expected := changesColor
	props := properties{
		LocalChangesColor: changesColor,
		ColorBackground:   true,
	}
	g := &git{
		scm: scm{
			props: props,
		},
		Staging: &GitStatus{},
		Working: &GitStatus{
			ScmStatus: ScmStatus{
				Modified: 1,
			},
		},
	}
	g.SetStatusColor()
	assert.Equal(t, expected, g.props.getColor(BackgroundOverride, ""))
}

func TestStatusColorsWithoutDisplayStatus(t *testing.T) {
	expected := changesColor
	status := "## main...origin/main [ahead 33]\n M myfile"
	env := new(MockedEnvironment)
	env.On("isWsl").Return(false)
	env.On("inWSLSharedDrive").Return(false)
	env.On("getRuntimeGOOS").Return("unix")
	env.On("hasCommand", "git").Return(true)
	fileInfo := &fileInfo{
		path:         "/dir/hello",
		parentFolder: "/dir",
		isDir:        true,
	}
	env.On("hasParentFilePath", ".git").Return(fileInfo, nil)
	env.On("getFileContent", fmt.Sprintf("%s/HEAD", fileInfo.path)).Return("")
	env.mockGitCommand(fileInfo.path, "", "describe", "--tags", "--exact-match")
	env.mockGitCommand(fileInfo.path, status, "status", "-unormal", "--branch", "--porcelain=2")
	env.On("hasFolder", fmt.Sprintf("%s/rebase-merge", fileInfo.path)).Return(false)
	env.On("hasFolder", fmt.Sprintf("%s/rebase-apply", fileInfo.path)).Return(false)
	env.On("hasFilesInDir", fileInfo.path, "CHERRY_PICK_HEAD").Return(false)
	env.On("hasFilesInDir", fileInfo.path, "REVERT_HEAD").Return(false)
	env.On("hasFilesInDir", fileInfo.path, "MERGE_MSG").Return(false)
	env.On("hasFilesInDir", fileInfo.path, "MERGE_HEAD").Return(false)
	env.On("hasFilesInDir", fileInfo.path, "sequencer/todo").Return(false)

	props := properties{
		DisplayStatus:       false,
		StatusColorsEnabled: true,
		LocalChangesColor:   expected,
	}
	g := &git{
		scm: scm{
			env:   env,
			props: props,
		},
		gitWorkingFolder: "",
	}
	g.Working = &GitStatus{}
	g.Staging = &GitStatus{}
	_ = g.enabled()
	g.string()
	assert.Equal(t, expected, g.props.getColor(BackgroundOverride, ""))
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
		env.On("lastErrorCode").Return(tc.ExitCode)
		props := properties{
			SuccessIcon:     tc.SuccessIcon,
			ErrorIcon:       tc.ErrorIcon,
			DisplayExitCode: tc.DisplayExitCode,
			AlwaysNumeric:   tc.AlwaysNumeric,
		}
		e := &exit{
			env:   env,
			props: props,
		}
		_ = e.enabled()
		assert.Equal(t, tc.Expected, e.deprecatedString())
	}
}

// Battery Segment

func TestBatterySegmentSingle(t *testing.T) {
	cases := []struct {
		Case            string
		Batteries       []*battery.Battery
		ExpectedString  string
		ExpectedEnabled bool
		ExpectedColor   string
		ColorBackground bool
		DisplayError    bool
		Error           error
		DisableCharging bool
		DisableCharged  bool
	}{
		{Case: "80% charging", Batteries: []*battery.Battery{{Full: 100, State: battery.Charging, Current: 80}}, ExpectedString: "charging 80", ExpectedEnabled: true},
		{Case: "battery full", Batteries: []*battery.Battery{{Full: 100, State: battery.Full, Current: 100}}, ExpectedString: "charged 100", ExpectedEnabled: true},
		{Case: "70% discharging", Batteries: []*battery.Battery{{Full: 100, State: battery.Discharging, Current: 70}}, ExpectedString: "going down 70", ExpectedEnabled: true},
		{
			Case:            "discharging background color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Discharging, Current: 70}},
			ExpectedString:  "going down 70",
			ExpectedEnabled: true,
			ColorBackground: true,
			ExpectedColor:   dischargingColor,
		},
		{
			Case:            "charging background color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Charging, Current: 70}},
			ExpectedString:  "charging 70",
			ExpectedEnabled: true,
			ColorBackground: true,
			ExpectedColor:   chargingColor,
		},
		{
			Case:            "charged background color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Full, Current: 70}},
			ExpectedString:  "charged 70",
			ExpectedEnabled: true,
			ColorBackground: true,
			ExpectedColor:   chargedColor,
		},
		{
			Case:            "discharging foreground color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Discharging, Current: 70}},
			ExpectedString:  "going down 70",
			ExpectedEnabled: true,
			ExpectedColor:   dischargingColor,
		},
		{
			Case:            "charging foreground color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Charging, Current: 70}},
			ExpectedString:  "charging 70",
			ExpectedEnabled: true,
			ExpectedColor:   chargingColor,
		},
		{
			Case:            "charged foreground color",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Full, Current: 70}},
			ExpectedString:  "charged 70",
			ExpectedEnabled: true,
			ExpectedColor:   chargedColor,
		},
		{Case: "battery error", DisplayError: true, Error: errors.New("oh snap"), ExpectedString: "oh snap", ExpectedEnabled: true},
		{Case: "battery error disabled", Error: errors.New("oh snap")},
		{Case: "no batteries", DisplayError: true, Error: &noBatteryError{}},
		{Case: "no batteries without error"},
		{Case: "display charging disabled: charging", Batteries: []*battery.Battery{{Full: 100, State: battery.Charging}}, DisableCharging: true},
		{Case: "display charged disabled: charged", Batteries: []*battery.Battery{{Full: 100, State: battery.Full}}, DisableCharged: true},
		{
			Case:            "display charging disabled/display charged enabled: charging",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Charging}},
			DisableCharging: true,
			DisableCharged:  false},
		{
			Case:            "display charged disabled/display charging enabled: charged",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Full}},
			DisableCharged:  true,
			DisableCharging: false},
		{
			Case:            "display charging disabled: discharging",
			Batteries:       []*battery.Battery{{Full: 100, State: battery.Discharging, Current: 70}},
			ExpectedString:  "going down 70",
			ExpectedEnabled: true,
			DisableCharging: true,
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := properties{
			ChargingIcon:     "charging ",
			ChargedIcon:      "charged ",
			DischargingIcon:  "going down ",
			DischargingColor: dischargingColor,
			ChargedColor:     chargedColor,
			ChargingColor:    chargingColor,
			ColorBackground:  tc.ColorBackground,
			DisplayError:     tc.DisplayError,
		}
		// default values
		if tc.DisableCharging {
			props[DisplayCharging] = false
		}
		if tc.DisableCharged {
			props[DisplayCharged] = false
		}
		env.On("getBatteryInfo").Return(tc.Batteries, tc.Error)
		env.onTemplate()
		b := &batt{
			props: props,
			env:   env,
		}
		enabled := b.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}
		assert.Equal(t, tc.ExpectedString, b.string(), tc.Case)
		if len(tc.ExpectedColor) == 0 {
			continue
		}
		actualColor := b.props.getColor(ForegroundOverride, "")
		if tc.ColorBackground {
			actualColor = b.props.getColor(BackgroundOverride, "")
		}
		assert.Equal(t, tc.ExpectedColor, actualColor, tc.Case)
	}
}

// Session

func TestPropertySessionSegment(t *testing.T) {
	cases := []struct {
		Case               string
		ExpectedEnabled    bool
		ExpectedString     string
		UserName           string
		Host               string
		DefaultUserName    string
		DefaultUserNameEnv string
		SSHSession         bool
		SSHClient          bool
		Root               bool
		DisplayUser        bool
		DisplayHost        bool
		DisplayDefault     bool
		HostColor          string
		UserColor          string
		GOOS               string
		HostError          bool
	}{
		{
			Case:            "user and computer",
			ExpectedString:  "john at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			ExpectedEnabled: true,
		},
		{
			Case:            "user and computer with host color",
			ExpectedString:  "john at <yellow>company-laptop</>",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			HostColor:       "yellow",
			ExpectedEnabled: true,
		},
		{
			Case:            "user and computer with user color",
			ExpectedString:  "<yellow>john</> at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			UserColor:       "yellow",
			ExpectedEnabled: true,
		},
		{
			Case:            "user and computer with both colors",
			ExpectedString:  "<yellow>john</> at <green>company-laptop</>",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			UserColor:       "yellow",
			HostColor:       "green",
			ExpectedEnabled: true,
		},
		{
			Case:            "SSH Session",
			ExpectedString:  "ssh john at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			SSHSession:      true,
			ExpectedEnabled: true,
		},
		{
			Case:            "SSH Client",
			ExpectedString:  "ssh john at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			SSHClient:       true,
			ExpectedEnabled: true,
		},
		{
			Case:            "SSH Client",
			ExpectedString:  "ssh john at company-laptop",
			Host:            "company-laptop",
			DisplayUser:     true,
			DisplayHost:     true,
			UserName:        "john",
			SSHClient:       true,
			ExpectedEnabled: true,
		},
		{
			Case:            "only user name",
			ExpectedString:  "john",
			Host:            "company-laptop",
			UserName:        "john",
			DisplayUser:     true,
			ExpectedEnabled: true,
		},
		{
			Case:            "windows user name",
			ExpectedString:  "john at company-laptop",
			Host:            "company-laptop",
			UserName:        "surface\\john",
			DisplayHost:     true,
			DisplayUser:     true,
			ExpectedEnabled: true,
			GOOS:            string(Windows),
		},
		{
			Case:            "only host name",
			ExpectedString:  "company-laptop",
			Host:            "company-laptop",
			UserName:        "john",
			DisplayDefault:  true,
			DisplayHost:     true,
			ExpectedEnabled: true,
		},
		{
			Case:            "display default - hidden",
			Host:            "company-laptop",
			UserName:        "john",
			DefaultUserName: "john",
			DisplayDefault:  false,
			DisplayHost:     true,
			DisplayUser:     true,
			ExpectedEnabled: false,
		},
		{
			Case:               "display default with env var - hidden",
			Host:               "company-laptop",
			UserName:           "john",
			DefaultUserNameEnv: "john",
			DefaultUserName:    "jake",
			DisplayDefault:     false,
			DisplayHost:        true,
			DisplayUser:        true,
			ExpectedEnabled:    false,
		},
		{
			Case:            "host error",
			ExpectedString:  "john at unknown",
			Host:            "company-laptop",
			HostError:       true,
			UserName:        "john",
			DisplayHost:     true,
			DisplayUser:     true,
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getCurrentUser").Return(tc.UserName)
		env.On("getRuntimeGOOS").Return(tc.GOOS)
		if tc.HostError {
			env.On("getHostName").Return(tc.Host, errors.New("oh snap"))
		} else {
			env.On("getHostName").Return(tc.Host, nil)
		}
		var SSHSession string
		if tc.SSHSession {
			SSHSession = "zezzion"
		}
		var SSHClient string
		if tc.SSHClient {
			SSHClient = "clientz"
		}
		env.On("getenv", "SSH_CONNECTION").Return(SSHSession)
		env.On("getenv", "SSH_CLIENT").Return(SSHClient)
		env.On("getenv", "SSH_CLIENT").Return(SSHSession)
		env.On("getenv", defaultUserEnvVar).Return(tc.DefaultUserNameEnv)
		env.On("isRunningAsRoot").Return(tc.Root)
		props := properties{
			UserInfoSeparator: " at ",
			SSHIcon:           "ssh ",
			DefaultUserName:   tc.DefaultUserName,
			DisplayDefault:    tc.DisplayDefault,
			DisplayUser:       tc.DisplayUser,
			DisplayHost:       tc.DisplayHost,
			HostColor:         tc.HostColor,
			UserColor:         tc.UserColor,
		}
		session := &session{
			env:   env,
			props: props,
		}
		assert.Equal(t, tc.ExpectedEnabled, session.enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, session.string(), tc.Case)
		}
	}
}

// Language

func TestLanguageVersionMismatch(t *testing.T) {
	cases := []struct {
		Case            string
		Enabled         bool
		Mismatch        bool
		ExpectedColor   string
		ColorBackground bool
	}{
		{Case: "Mismatch - Foreground color", Enabled: true, Mismatch: true, ExpectedColor: "#566777"},
		{Case: "Mismatch - Background color", Enabled: true, Mismatch: true, ExpectedColor: "#566777", ColorBackground: true},
		{Case: "Disabled", Enabled: false},
		{Case: "No mismatch", Enabled: true, Mismatch: false},
	}
	for _, tc := range cases {
		props := properties{
			EnableVersionMismatch: tc.Enabled,
			VersionMismatchColor:  tc.ExpectedColor,
			ColorBackground:       tc.ColorBackground,
		}
		var matchesVersionFile func() bool
		switch tc.Mismatch {
		case true:
			matchesVersionFile = func() bool {
				return false
			}
		default:
			matchesVersionFile = func() bool {
				return true
			}
		}
		args := &languageArgs{
			commands: []*cmd{
				{
					executable: "unicorn",
					args:       []string{"--version"},
					regex:      "(?P<version>.*)",
				},
			},
			extensions:         []string{uni, corn},
			enabledExtensions:  []string{uni, corn},
			enabledCommands:    []string{"unicorn"},
			version:            universion,
			properties:         props,
			matchesVersionFile: matchesVersionFile,
		}
		lang := bootStrapLanguageTest(args)
		assert.True(t, lang.enabled(), tc.Case)
		assert.Equal(t, universion, lang.string(), tc.Case)
		if tc.ColorBackground {
			assert.Equal(t, tc.ExpectedColor, lang.props.getColor(BackgroundOverride, ""), tc.Case)
			return
		}
		assert.Equal(t, tc.ExpectedColor, lang.props.getColor(ForegroundOverride, ""), tc.Case)
	}
}

// Python

func TestPythonVirtualEnv(t *testing.T) {
	cases := []struct {
		Case                string
		Expected            string
		ExpectedDisabled    bool
		VirtualEnvName      string
		CondaEnvName        string
		CondaDefaultEnvName string
		PyEnvName           string
		FetchVersion        bool
		DisplayDefault      bool
	}{
		{Case: "VENV", Expected: "VENV", VirtualEnvName: "VENV"},
		{Case: "CONDA", Expected: "CONDA", CondaEnvName: "CONDA"},
		{Case: "CONDA default", Expected: "CONDA", CondaDefaultEnvName: "CONDA"},
		{Case: "Display Base", Expected: "base", CondaDefaultEnvName: "base", DisplayDefault: true},
		{Case: "Hide base", Expected: "", CondaDefaultEnvName: "base", ExpectedDisabled: true},
		{Case: "PYENV", Expected: "PYENV", PyEnvName: "PYENV"},
		{Case: "PYENV Version", Expected: "PYENV 3.8.4", PyEnvName: "PYENV", FetchVersion: true},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("hasCommand", "python").Return(true)
		env.On("runCommand", "python", []string{"--version"}).Return("Python 3.8.4", nil)
		env.On("hasFiles", "*.py").Return(true)
		env.On("getenv", "VIRTUAL_ENV").Return(tc.VirtualEnvName)
		env.On("getenv", "CONDA_ENV_PATH").Return(tc.CondaEnvName)
		env.On("getenv", "CONDA_DEFAULT_ENV").Return(tc.CondaDefaultEnvName)
		env.On("getenv", "PYENV_VERSION").Return(tc.PyEnvName)
		env.On("getPathSeperator").Return("")
		env.On("pwd").Return("/usr/home/project")
		env.On("homeDir").Return("/usr/home")
		env.onTemplate()
		props := properties{
			FetchVersion:      tc.FetchVersion,
			DisplayVirtualEnv: true,
			DisplayDefault:    tc.DisplayDefault,
		}
		python := &python{}
		python.init(props, env)
		assert.Equal(t, !tc.ExpectedDisabled, python.enabled(), tc.Case)
		assert.Equal(t, tc.Expected, python.string(), tc.Case)
	}
}

// Environment Variable

func TestEnvvarAvailable(t *testing.T) {
	name := "HERP"
	expected := "derp"
	env := new(MockedEnvironment)
	env.On("getenv", name).Return(expected)
	e := &envvar{
		env: env,
		props: properties{
			VarName: name,
		},
	}
	assert.True(t, e.enabled())
	assert.Equal(t, expected, e.string())
}

func TestEnvvarNotAvailable(t *testing.T) {
	name := "HERP"
	expected := ""
	env := new(MockedEnvironment)
	env.On("getenv", name).Return(expected)
	e := &envvar{
		env: env,
		props: properties{
			VarName: name,
		},
	}
	assert.False(t, e.enabled())
}
