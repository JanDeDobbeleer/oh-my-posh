package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type kubectlArgs struct {
	kubectlExists bool
	kubectlErr    bool
	template      string
	displayError  bool
	context       string
	namespace     string
}

func bootStrapKubectlTest(args *kubectlArgs) *kubectl {
	env := new(MockedEnvironment)
	env.On("hasCommand", "kubectl").Return(args.kubectlExists)
	kubectlOut := args.context + "," + args.namespace
	var kubectlErr error
	if args.kubectlErr {
		kubectlErr = &commandError{
			err:      "oops",
			exitCode: 1,
		}
	}
	env.On("runCommand", "kubectl", []string{"config", "view", "--minify", "--output", "jsonpath={..current-context},{..namespace}"}).Return(kubectlOut, kubectlErr)
	k := &kubectl{
		env: env,
		props: map[Property]interface{}{
			SegmentTemplate: args.template,
			DisplayError:    args.displayError,
		},
	}
	return k
}

func TestKubectlSegment(t *testing.T) {
	standardTemplate := "{{.Context}}{{if .Namespace}} :: {{.Namespace}}{{end}}"
	cases := []struct {
		Case            string
		Template        string
		DisplayError    bool
		KubectlExists   bool
		Context         string
		Namespace       string
		KubectlErr      bool
		ExpectedEnabled bool
		ExpectedString  string
	}{
		{Case: "disabled", Template: standardTemplate, KubectlExists: false, Context: "aaa", Namespace: "bbb", ExpectedString: "", ExpectedEnabled: false},
		{Case: "normal", Template: standardTemplate, KubectlExists: true, Context: "aaa", Namespace: "bbb", ExpectedString: "aaa :: bbb", ExpectedEnabled: true},
		{Case: "no namespace", Template: standardTemplate, KubectlExists: true, Context: "aaa", Namespace: "", ExpectedString: "aaa", ExpectedEnabled: true},
		{Case: "kubectl error", Template: standardTemplate, DisplayError: true, KubectlExists: true, Context: "aaa", Namespace: "bbb", KubectlErr: true,
			ExpectedString: "KUBECTL ERR :: KUBECTL ERR", ExpectedEnabled: true},
		{Case: "kubectl error hidden", Template: standardTemplate, DisplayError: false, KubectlExists: true, Context: "aaa", Namespace: "bbb", KubectlErr: true,
			ExpectedString: "", ExpectedEnabled: false},
	}

	for _, tc := range cases {
		args := &kubectlArgs{
			kubectlExists: tc.KubectlExists,
			template:      tc.Template,
			displayError:  tc.DisplayError,
			context:       tc.Context,
			namespace:     tc.Namespace,
			kubectlErr:    tc.KubectlErr,
		}
		kubectl := bootStrapKubectlTest(args)
		assert.Equal(t, tc.ExpectedEnabled, kubectl.enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, kubectl.string(), tc.Case)
	}
}
