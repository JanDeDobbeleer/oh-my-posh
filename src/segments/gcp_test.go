package segments

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGcpSegment(t *testing.T) {
	cases := []struct {
		Case            string
		CfgData         string
		ActiveConfig    string
		EnvActiveConfig string
		ExpectedString  string
		ExpectedEnabled bool
	}{
		{
			Case:            "happy path",
			ExpectedEnabled: true,
			ActiveConfig:    "production",
			CfgData: `
			[core]
			account = test@example.com
			project = test-test-test

			[compute]
			region = europe-test1
			`,
			ExpectedString: "test-test-test :: europe-test1 :: test@example.com",
		},
		{
			Case:            "no active config",
			ExpectedEnabled: false,
		},
		{
			Case:            "empty config",
			ActiveConfig:    "production",
			ExpectedEnabled: false,
		},
		{
			Case:            "bad config",
			ActiveConfig:    "production",
			CfgData:         "{bad}",
			ExpectedEnabled: false,
		},
		{
			Case:            "use CLOUDSDK_ACTIVE_CONFIG_NAME",
			EnvActiveConfig: "myconfig",
			ExpectedEnabled: true,
			CfgData: `
			[core]
			account = user@example.com
			project = cloud-proj

			[compute]
			region = us-west1
			`,
			ExpectedString: "cloud-proj :: us-west1 :: user@example.com",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "CLOUDSDK_CONFIG").Return("config")
		env.On("Getenv", "CLOUDSDK_ACTIVE_CONFIG_NAME").Return(tc.EnvActiveConfig)

		// Only use fallback file if env var is not set
		if tc.EnvActiveConfig == "" {
			fcPath := path.Join("config", "active_config")
			env.On("FileContent", fcPath).Return(tc.ActiveConfig)
		}

		// Resolve active config name
		activeConfig := tc.EnvActiveConfig
		if activeConfig == "" {
			activeConfig = tc.ActiveConfig
		}

		cfgpath := path.Join("config", "configurations", "config_"+activeConfig)
		env.On("FileContent", cfgpath).Return(tc.CfgData)

		g := &Gcp{}
		g.Init(options.Map{}, env)

		assert.Equal(t, tc.ExpectedEnabled, g.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, "{{.Project}} :: {{.Region}} :: {{.Account}}", g), tc.Case)
		}
	}
}

func TestGcpAuthStatus(t *testing.T) {
	dbContent, err := os.ReadFile(path.Join("testdata", "gcp_access_tokens.db"))
	require.NoError(t, err)

	cases := []struct {
		Case                 string
		Account              string
		FetchAuthStatus      bool
		NoDB                 bool
		ExpectedAuthorized   bool
		ExpectedExpiryIsZero bool
	}{
		{
			Case:                 "disabled by default",
			Account:              "expired@example.com",
			ExpectedAuthorized:   false,
			ExpectedExpiryIsZero: true,
		},
		{
			Case:                 "expired token",
			Account:              "expired@example.com",
			FetchAuthStatus:      true,
			ExpectedAuthorized:   false,
			ExpectedExpiryIsZero: false,
		},
		{
			Case:                 "valid token",
			Account:              "valid@example.com",
			FetchAuthStatus:      true,
			ExpectedAuthorized:   true,
			ExpectedExpiryIsZero: false,
		},
		{
			Case:                 "account unknown to access_tokens.db",
			Account:              "unknown@example.com",
			FetchAuthStatus:      true,
			ExpectedAuthorized:   true,
			ExpectedExpiryIsZero: true,
		},
		{
			Case:                 "access_tokens.db missing",
			Account:              "expired@example.com",
			FetchAuthStatus:      true,
			NoDB:                 true,
			ExpectedAuthorized:   true,
			ExpectedExpiryIsZero: true,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "CLOUDSDK_CONFIG").Return("config")
		env.On("Getenv", "CLOUDSDK_ACTIVE_CONFIG_NAME").Return("production")

		cfgData := `
		[core]
		account = ` + tc.Account + `
		project = test-project
		`

		cfgpath := path.Join("config", "configurations", "config_production")
		env.On("FileContent", cfgpath).Return(cfgData)

		dbPath := path.Join("config", "access_tokens.db")
		if tc.NoDB {
			env.On("FileContent", dbPath).Return("")
		} else {
			env.On("FileContent", dbPath).Return(string(dbContent))
		}

		g := &Gcp{}
		opts := options.Map{}
		if tc.FetchAuthStatus {
			opts[FetchAuthStatus] = true
		}
		g.Init(opts, env)

		require.True(t, g.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedAuthorized, g.Authorized, tc.Case)
		assert.Equal(t, tc.ExpectedExpiryIsZero, g.TokenExpiresAt.IsZero(), tc.Case)
	}
}

func TestParseTokenExpiry(t *testing.T) {
	cases := []struct {
		Expected time.Time
		Case     string
		Value    string
		HasError bool
	}{
		{
			Case:     "with microseconds",
			Value:    "2024-01-01 12:00:00.123456",
			Expected: time.Date(2024, 1, 1, 12, 0, 0, 123456000, time.UTC),
		},
		{
			Case:     "without microseconds",
			Value:    "2024-01-01 12:00:00",
			Expected: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			Case:     "empty",
			Value:    "",
			HasError: true,
		},
		{
			Case:     "garbage",
			Value:    "not-a-date",
			HasError: true,
		},
	}

	for _, tc := range cases {
		got, err := parseTokenExpiry(tc.Value)
		if tc.HasError {
			assert.Error(t, err, tc.Case)
			continue
		}

		require.NoError(t, err, tc.Case)
		assert.True(t, tc.Expected.Equal(got), tc.Case)
	}
}

func TestGetConfigDirectory(t *testing.T) {
	cases := []struct {
		Case           string
		GOOS           string
		Home           string
		AppData        string
		CloudSDKConfig string
		Expected       string
	}{
		{
			Case:           "CLOUDSDK_CONFIG",
			CloudSDKConfig: "/Users/posh/.config/gcloud",
			Expected:       "/Users/posh/.config/gcloud",
		},
		{
			Case:     "Windows",
			GOOS:     runtime.WINDOWS,
			AppData:  "/Users/posh/.config",
			Expected: "/Users/posh/.config/gcloud",
		},
		{
			Case:     "default",
			Home:     "/Users/posh2/",
			Expected: "/Users/posh2/.config/gcloud",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "CLOUDSDK_CONFIG").Return(tc.CloudSDKConfig)
		env.On("Getenv", "APPDATA").Return(tc.AppData)
		env.On("Home").Return(tc.Home)
		env.On("GOOS").Return(tc.GOOS)

		g := &Gcp{}
		g.Init(options.Map{}, env)

		assert.Equal(t, tc.Expected, g.getConfigDirectory(), tc.Case)
	}
}

func TestGetActiveConfig(t *testing.T) {
	cases := []struct {
		Case                    string
		EnvActiveConfigName     string
		FileActiveConfigContent string
		ExpectedString          string
		ExpectedError           string
	}{
		{
			Case:                "CLOUDSDK_ACTIVE_CONFIG_NAME set",
			EnvActiveConfigName: "envconfig",
			ExpectedString:      "envconfig",
		},
		{
			Case:                    "Fallback to file content",
			FileActiveConfigContent: "fileconfig",
			ExpectedString:          "fileconfig",
		},
		{
			Case:          "No config anywhere",
			ExpectedError: GCPNOACTIVECONFIG,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "CLOUDSDK_ACTIVE_CONFIG_NAME").Return(tc.EnvActiveConfigName)

		// If env var not set, mock file fallback
		if tc.EnvActiveConfigName == "" {
			env.On("FileContent", path.Join("", "active_config")).Return(tc.FileActiveConfigContent)
		}

		g := &Gcp{}
		g.Init(options.Map{}, env)

		got, err := g.getActiveConfig("")
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
		if len(tc.ExpectedError) > 0 {
			assert.EqualError(t, err, tc.ExpectedError, tc.Case)
		} else {
			assert.NoError(t, err, tc.Case)
		}
	}
}
