package segments

import (
    "testing"

    "github.com/jandedobbeleer/oh-my-posh/src/properties"
    "github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
    "github.com/stretchr/testify/assert"
)

func TestDistrobox(t *testing.T) {
    cases := []struct {
        Case            string
        ContainerID     string
        ExpectedEnabled bool
        ExpectedString  string
        Template        string
    }{
        {
            Case:            "in distrobox container",
            ContainerID:     "fedora-toolbox",
            ExpectedEnabled: true,
            ExpectedString:  "fedora-toolbox",
            Template:        "{{ .ContainerID }}",
        },
        {
            Case:            "not in container",
            ContainerID:     "",
            ExpectedEnabled: false,
        },
        {
            Case:            "with icon template",
            ContainerID:     "ubuntu-dev",
            ExpectedEnabled: true,
            ExpectedString:  "\uED95 ubuntu-dev",
            Template:        "{{ .Icon }} {{ .ContainerID }}",
        },
    }

    for _, tc := range cases {
        distrobox := &Distrobox{}
        env := new(mock.Environment)

        env.On("Getenv", "CONTAINER_ID").Return(tc.ContainerID)

        distrobox.Init(properties.Map{}, env)

        assert.Equal(t, tc.ExpectedEnabled, distrobox.Enabled(), tc.Case)
        if tc.ExpectedEnabled {
            assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, distrobox), tc.Case)
        }
    }
}
