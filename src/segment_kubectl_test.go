package main

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type kubectlArgs struct {
	kubectlExists       bool
	kubectlErr          bool
	kubeconfig          string
	parseKubeConfig     bool
	template            string
	displayError        bool
	kubectlOutContext   string
	kubectlOutNamespace string
	kubectlOutUser      string
	kubectlOutCluster   string
	files               map[string]string
}

const testKubectlAllInfoTemplate = "{{.Context}} :: {{.Namespace}} :: {{.User}} :: {{.Cluster}}"

func bootStrapKubectlTest(args *kubectlArgs) *kubectl {
	env := new(MockedEnvironment)
	env.On("hasCommand", "kubectl").Return(args.kubectlExists)
	kubectlOut := args.kubectlOutContext + "," + args.kubectlOutNamespace + "," + args.kubectlOutUser + "," + args.kubectlOutCluster
	var kubectlErr error
	if args.kubectlErr {
		kubectlErr = &commandError{
			err:      "oops",
			exitCode: 1,
		}
	}
	env.On("runCommand", "kubectl",
		[]string{"config", "view", "--minify", "--output", "jsonpath={..current-context},{..namespace},{..context.user},{..context.cluster}"}).Return(kubectlOut, kubectlErr)

	env.On("getenv", "KUBECONFIG").Return(args.kubeconfig)
	for path, content := range args.files {
		env.On("getFileContent", path).Return(content)
	}
	env.On("homeDir", nil).Return("testhome")

	k := &kubectl{
		env: env,
		props: map[Property]interface{}{
			SegmentTemplate: args.template,
			DisplayError:    args.displayError,
			ParseKubeConfig: args.parseKubeConfig,
		},
	}
	return k
}

func TestKubectlSegment(t *testing.T) {
	standardTemplate := "{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}"
	lsep := string(filepath.ListSeparator)

	cases := []struct {
		Case            string
		Template        string
		DisplayError    bool
		KubectlExists   bool
		Kubeconfig      string
		ParseKubeConfig bool
		Context         string
		Namespace       string
		User            string
		Cluster         string
		KubectlErr      bool
		ExpectedEnabled bool
		ExpectedString  string
		Files           map[string]string
	}{
		{Case: "disabled", Template: standardTemplate, KubectlExists: false, Context: "aaa", Namespace: "bbb", ExpectedString: "", ExpectedEnabled: false},
		{Case: "normal", Template: standardTemplate, KubectlExists: true, Context: "aaa", Namespace: "bbb", ExpectedString: "aaa :: bbb", ExpectedEnabled: true},
		{Case: "all information", Template: testKubectlAllInfoTemplate, KubectlExists: true, Context: "aaa", Namespace: "bbb", User: "ccc", Cluster: "ddd",
			ExpectedString: "aaa :: bbb :: ccc :: ddd", ExpectedEnabled: true},
		{Case: "no namespace", Template: standardTemplate, KubectlExists: true, Context: "aaa", Namespace: "", ExpectedString: "aaa", ExpectedEnabled: true},
		{Case: "kubectl error", Template: standardTemplate, DisplayError: true, KubectlExists: true, Context: "aaa", Namespace: "bbb", KubectlErr: true,
			ExpectedString: "KUBECTL ERR :: KUBECTL ERR", ExpectedEnabled: true},
		{Case: "kubectl error hidden", Template: standardTemplate, DisplayError: false, KubectlExists: true, Context: "aaa", Namespace: "bbb", KubectlErr: true, ExpectedEnabled: false},
		{Case: "kubeconfig home", Template: testKubectlAllInfoTemplate, ParseKubeConfig: true, Files: testKubeConfigFiles, ExpectedString: "aaa :: bbb :: ccc :: ddd",
			ExpectedEnabled: true},
		{Case: "kubeconfig multiple current marker first", Template: testKubectlAllInfoTemplate, ParseKubeConfig: true,
			Kubeconfig: "" + lsep + "currentcontextmarker" + lsep + "contextdefinition" + lsep + "contextredefinition",
			Files:      testKubeConfigFiles, ExpectedString: "ctx :: ns :: usr :: cl", ExpectedEnabled: true},
		{Case: "kubeconfig multiple context first", Template: testKubectlAllInfoTemplate, ParseKubeConfig: true,
			Kubeconfig: "contextdefinition" + lsep + "contextredefinition" + lsep + "currentcontextmarker" + lsep,
			Files:      testKubeConfigFiles, ExpectedString: "ctx :: ns :: usr :: cl", ExpectedEnabled: true},
		{Case: "kubeconfig error hidden", Template: testKubectlAllInfoTemplate, ParseKubeConfig: true, Kubeconfig: "invalid", Files: testKubeConfigFiles, ExpectedEnabled: false},
		{Case: "kubeconfig error", Template: testKubectlAllInfoTemplate, ParseKubeConfig: true,
			Kubeconfig: "invalid", Files: testKubeConfigFiles, DisplayError: true,
			ExpectedString: "KUBECONFIG ERR :: KUBECONFIG ERR :: KUBECONFIG ERR :: KUBECONFIG ERR", ExpectedEnabled: true},
		{Case: "kubeconfig incomplete", Template: testKubectlAllInfoTemplate, ParseKubeConfig: true,
			Kubeconfig: "currentcontextmarker" + lsep + "contextdefinitionincomplete",
			Files:      testKubeConfigFiles, ExpectedString: "ctx ::  ::  :: ", ExpectedEnabled: true},
	}

	for _, tc := range cases {
		args := &kubectlArgs{
			kubectlExists:       tc.KubectlExists,
			template:            tc.Template,
			displayError:        tc.DisplayError,
			kubectlOutContext:   tc.Context,
			kubectlOutNamespace: tc.Namespace,
			kubectlOutUser:      tc.User,
			kubectlOutCluster:   tc.Cluster,
			kubectlErr:          tc.KubectlErr,
			parseKubeConfig:     tc.ParseKubeConfig,
			files:               tc.Files,
			kubeconfig:          tc.Kubeconfig,
		}
		kubectl := bootStrapKubectlTest(args)
		assert.Equal(t, tc.ExpectedEnabled, kubectl.enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, kubectl.string(), tc.Case)
		}
	}
}

var testKubeConfigFiles = map[string]string{
	filepath.Join("testhome", ".kube/config"): `
apiVersion: v1
contexts:
  - context:
      cluster: ddd
      user: ccc
      namespace: bbb
    name: aaa
current-context: aaa
`,
	"contextdefinition": `
apiVersion: v1
contexts:
  - context:
      cluster: cl
      user: usr
      namespace: ns
    name: ctx
`,
	"currentcontextmarker": `
apiVersion: v1
current-context: ctx
`,
	"invalid": "this is not yaml",
	"contextdefinitionincomplete": `
apiVersion: v1
contexts:
  - name: ctx
`,
	"contextredefinition": `
apiVersion: v1
contexts:
  - context:
      cluster: wrongcl
      user: wrongu
      namespace: wrongns
    name: ctx
`,
}
