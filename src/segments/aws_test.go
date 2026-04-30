package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestAWSSegment(t *testing.T) {
	cases := []struct {
		Case              string
		ExpectedString    string
		Profile           string
		DefaultProfile    string
		Vault             string
		Region            string
		DefaultRegion     string
		ConfigFile        string
		ConfigContent     string
		CredentialsBody   string
		AccessKeyEnv      string
		Template          string
		ExpectedEnabled   bool
		DisplayDefault    bool
		DisableAccountID  bool
		EnableAccessKeyID bool
	}{
		{Case: "enabled with default user", ExpectedString: "default@eu-west", Region: "eu-west", ExpectedEnabled: true, DisplayDefault: true},
		{Case: "disabled with default user", ExpectedString: "default@eu-west", Region: "eu-west", ExpectedEnabled: false, DisplayDefault: false},
		{Case: "disabled", ExpectedString: "", ExpectedEnabled: false},
		{Case: "enabled with default user", ExpectedString: "default@eu-west", Profile: "default", Region: "eu-west", ExpectedEnabled: true, DisplayDefault: true},
		{Case: "enabled with default profile", ExpectedString: "default@eu-west", DefaultProfile: "default", Region: "eu-west", ExpectedEnabled: true, DisplayDefault: true},
		{Case: "disabled with default user", ExpectedString: "default", Profile: "default", Region: "eu-west", ExpectedEnabled: false, DisplayDefault: false},
		{Case: "enabled no region", ExpectedString: "company", ExpectedEnabled: true, Profile: "company"},
		{Case: "enabled with region", ExpectedString: "company@eu-west", ExpectedEnabled: true, Profile: "company", Region: "eu-west", DefaultRegion: "us-west"},
		{Case: "enabled with default region", ExpectedString: "company@us-west", ExpectedEnabled: true, Profile: "company", DefaultRegion: "us-west"},
		{
			Case:            "template: enabled no region",
			ExpectedString:  "profile: company",
			ExpectedEnabled: true,
			Profile:         "company",
			Template:        "profile: {{.Profile}}{{if .Region}} in {{.Region}}{{end}}",
		},
		{
			Case:            "template: enabled with region",
			ExpectedString:  "profile: company in eu-west",
			ExpectedEnabled: true,
			Profile:         "company",
			Region:          "eu-west",
			Template:        "profile: {{.Profile}}{{if .Region}} in {{.Region}}{{end}}",
		},
		{
			Case:            "template: enabled with region alias that has compound cardinal direction",
			ExpectedString:  "profile: company in apne3",
			ExpectedEnabled: true,
			Profile:         "company",
			Region:          "ap-northeast-3",
			Template:        "profile: {{.Profile}}{{if .Region}} in {{.RegionAlias}}{{end}}",
		},
		{Case: "template: invalid", ExpectedString: "{{ .Burp", ExpectedEnabled: true, Profile: "c", Template: "{{ .Burp"},
		{
			Case:            "sso account id from config",
			ExpectedString:  "aws-first@eu-west-1 (500000000007)",
			ExpectedEnabled: true,
			Profile:         "aws-first",
			ConfigContent: "[profile aws-first]\n" +
				"sso_session = xyz\n" +
				"sso_account_id = 500000000007\n" +
				"sso_role_name = SomeRoleName\n" +
				"region = eu-west-1\n",
			Template: "{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}{{ if .AccountID }} ({{ .AccountID }}){{ end }}",
		},
		{
			Case:            "sso account id disabled",
			ExpectedString:  "aws-first@eu-west-1",
			ExpectedEnabled: true,
			Profile:         "aws-first",
			ConfigContent: "[profile aws-first]\n" +
				"sso_account_id = 500000000007\n" +
				"region = eu-west-1\n",
			DisableAccountID: true,
			Template:         "{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}{{ if .AccountID }} ({{ .AccountID }}){{ end }}",
		},
		{
			Case:              "access key id from env",
			ExpectedString:    "company@eu-west [AKIA123]",
			ExpectedEnabled:   true,
			Profile:           "company",
			Region:            "eu-west",
			AccessKeyEnv:      "AKIA123",
			EnableAccessKeyID: true,
			Template:          "{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}{{ if .AccessKeyID }} [{{ .AccessKeyID }}]{{ end }}",
		},
		{
			Case:            "access key id from credentials file",
			ExpectedString:  "prof1@us-east-1 [yyy]",
			ExpectedEnabled: true,
			Profile:         "prof1",
			Region:          "us-east-1",
			CredentialsBody: "[prof1]\n" +
				"aws_access_key_id = yyy\n" +
				"aws_secret_access_key = yyyyy\n",
			EnableAccessKeyID: true,
			Template:          "{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}{{ if .AccessKeyID }} [{{ .AccessKeyID }}]{{ end }}",
		},
		{
			Case:            "access key id default hidden",
			ExpectedString:  "prof1@us-east-1",
			ExpectedEnabled: true,
			Profile:         "prof1",
			Region:          "us-east-1",
			AccessKeyEnv:    "AKIA123",
			Template:        "{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}{{ if .AccessKeyID }} [{{ .AccessKeyID }}]{{ end }}",
		},
		{
			Case:            "access key id fallback from config file",
			ExpectedString:  "prof1@us-east-1 [yyy]",
			ExpectedEnabled: true,
			Profile:         "prof1",
			ConfigContent: "[profile prof1]\n" +
				"region=us-east-1\n" +
				"output=text\n" +
				"aws_access_key_id=yyy\n" +
				"aws_secret_access_key=yyyyy\n",
			EnableAccessKeyID: true,
			Template:          "{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}{{ if .AccessKeyID }} [{{ .AccessKeyID }}]{{ end }}",
		},
		{
			Case:            "expose role_arn via Settings map",
			ExpectedString:  "company@eu-west arn:aws:iam::123456789012:role/Admin",
			ExpectedEnabled: true,
			Profile:         "company",
			ConfigContent: "[profile company]\n" +
				"region = eu-west\n" +
				"role_arn = arn:aws:iam::123456789012:role/Admin\n" +
				"source_profile = base\n",
			Template: `{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}{{ with index .Settings "role_arn" }} {{ . }}{{ end }}`,
		},
		{
			Case:            "expose multiple settings (mfa_serial, output)",
			ExpectedString:  "company@eu-west text mfa:arn:aws:iam::1:mfa/u",
			ExpectedEnabled: true,
			Profile:         "company",
			ConfigContent: "[profile company]\n" +
				"region = eu-west\n" +
				"output = text\n" +
				"mfa_serial = arn:aws:iam::1:mfa/u\n",
			Template: "{{ .Profile }}{{ if .Region }}@{{ .Region }}{{ end }}" +
				` {{ index .Settings "output" }}` +
				` mfa:{{ index .Settings "mfa_serial" }}`,
		},
		{
			Case:            "sso-session resolution",
			ExpectedString:  "aws-first https://my-sso.awsapps.com/start us-east-1",
			ExpectedEnabled: true,
			Profile:         "aws-first",
			ConfigContent: "[profile aws-first]\n" +
				"sso_session = my-sso\n" +
				"sso_account_id = 500000000007\n" +
				"sso_role_name = SomeRoleName\n" +
				"region = eu-west-1\n" +
				"\n" +
				"[sso-session my-sso]\n" +
				"sso_start_url = https://my-sso.awsapps.com/start\n" +
				"sso_region = us-east-1\n" +
				"sso_registration_scopes = sso:account:access\n",
			Template: `{{ .Profile }} {{ index .SSOSession "sso_start_url" }} {{ index .SSOSession "sso_region" }}`,
		},
		{
			Case:            "credentials file overrides config for credential keys",
			ExpectedString:  "prof1 [override-key]",
			ExpectedEnabled: true,
			Profile:         "prof1",
			ConfigContent: "[profile prof1]\n" +
				"region = us-east-1\n" +
				"aws_access_key_id = config-key\n",
			CredentialsBody: "[prof1]\n" +
				"aws_access_key_id = override-key\n",
			EnableAccessKeyID: true,
			Template:          "{{ .Profile }} [{{ .AccessKeyID }}]",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "AWS_VAULT").Return(tc.Vault)
		env.On("Getenv", "AWS_PROFILE").Return(tc.Profile)
		env.On("Getenv", "AWS_DEFAULT_PROFILE").Return(tc.DefaultProfile)
		env.On("Getenv", "AWS_REGION").Return(tc.Region)
		env.On("Getenv", "AWS_DEFAULT_REGION").Return(tc.DefaultRegion)
		env.On("Getenv", "AWS_CONFIG_FILE").Return(tc.ConfigFile)
		env.On("Getenv", "AWS_SHARED_CREDENTIALS_FILE").Return("")
		env.On("Getenv", "AWS_ACCESS_KEY_ID").Return(tc.AccessKeyEnv)
		env.On("FileContent", "/usr/home/.aws/config").Return(tc.ConfigContent)
		env.On("FileContent", "/usr/home/.aws/credentials").Return(tc.CredentialsBody)
		env.On("Home").Return("/usr/home")
		props := options.Map{
			options.DisplayDefault: tc.DisplayDefault,
		}
		if tc.DisableAccountID {
			props[DisplayAccountID] = false
		}
		if tc.EnableAccessKeyID {
			props[DisplayAccessKeyID] = true
		}
		env.On("Flags").Return(&runtime.Flags{})

		aws := &Aws{}
		aws.Init(props, env)

		if tc.Template == "" {
			tc.Template = aws.Template()
		}

		assert.Equal(t, tc.ExpectedEnabled, aws.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, aws), tc.Case)
	}
}
