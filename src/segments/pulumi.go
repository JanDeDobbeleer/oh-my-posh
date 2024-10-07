package segments

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"gopkg.in/yaml.v3"
)

const (
	FetchStack properties.Property = "fetch_stack"
	FetchAbout properties.Property = "fetch_about"

	JSON string = "json"
	YAML string = "yaml"

	pulumiJSON string = "Pulumi.json"
	pulumiYAML string = "Pulumi.yaml"
)

type Pulumi struct {
	props properties.Properties
	env   runtime.Environment

	Stack string
	Name  string

	workspaceSHA1 string

	backend
}

type backend struct {
	URL  string `json:"url"`
	User string `json:"user"`
}

type pulumiFileSpec struct {
	Name string `yaml:"name" json:"name"`
}

type pulumiWorkSpaceFileSpec struct {
	Stack string `json:"stack"`
}

func (p *Pulumi) Template() string {
	return "\U000f0d46 {{ .Stack }}{{if .User }} :: {{ .User }}@{{ end }}{{ if .URL }}{{ .URL }}{{ end }}"
}

func (p *Pulumi) Init(props properties.Properties, env runtime.Environment) {
	p.props = props
	p.env = env
}

func (p *Pulumi) Enabled() bool {
	if !p.env.HasCommand("pulumi") {
		return false
	}

	err := p.getProjectName()
	if err != nil {
		p.env.Error(err)
		return false
	}

	if p.props.GetBool(FetchStack, false) {
		p.getPulumiStackName()
	}

	if p.props.GetBool(FetchAbout, false) {
		p.getPulumiAbout()
	}

	return true
}

func (p *Pulumi) getPulumiStackName() {
	if len(p.Name) == 0 || len(p.workspaceSHA1) == 0 {
		p.env.Debug("pulumi project name or workspace sha1 is empty")
		return
	}

	stackNameFile := p.Name + "-" + p.workspaceSHA1 + "-" + "workspace.json"

	homedir := p.env.Home()

	workspaceCacheDir := filepath.Join(homedir, ".pulumi", "workspaces")
	if !p.env.HasFolder(workspaceCacheDir) || !p.env.HasFilesInDir(workspaceCacheDir, stackNameFile) {
		return
	}

	workspaceCacheFile := filepath.Join(workspaceCacheDir, stackNameFile)
	workspaceCacheFileContent := p.env.FileContent(workspaceCacheFile)

	var pulumiWorkspaceSpec pulumiWorkSpaceFileSpec
	err := json.Unmarshal([]byte(workspaceCacheFileContent), &pulumiWorkspaceSpec)
	if err != nil {
		p.env.Error(fmt.Errorf("pulumi workspace file decode error"))
		return
	}

	p.env.DebugF("pulumi stack name: %s", pulumiWorkspaceSpec.Stack)
	p.Stack = pulumiWorkspaceSpec.Stack
}

func (p *Pulumi) getProjectName() error {
	var kind, fileName string
	for _, file := range []string{pulumiYAML, pulumiJSON} {
		if p.env.HasFiles(file) {
			fileName = file
			kind = filepath.Ext(file)[1:]
		}
	}

	if len(kind) == 0 {
		return fmt.Errorf("no pulumi spec file found")
	}

	var pulumiFileSpec pulumiFileSpec
	var err error

	pulumiFile := p.env.FileContent(fileName)

	switch kind {
	case YAML:
		err = yaml.Unmarshal([]byte(pulumiFile), &pulumiFileSpec)
	case JSON:
		err = json.Unmarshal([]byte(pulumiFile), &pulumiFileSpec)
	default:
		err = fmt.Errorf("unknown pulumi spec file format")
	}

	if err != nil {
		p.env.Error(err)
		return nil
	}

	p.Name = pulumiFileSpec.Name

	sha1HexString := func(value string) string {
		h := sha1.New()

		_, err := h.Write([]byte(value))
		if err != nil {
			p.env.Error(err)
			return ""
		}

		return hex.EncodeToString(h.Sum(nil))
	}

	p.workspaceSHA1 = sha1HexString(p.env.Pwd() + p.env.PathSeparator() + fileName)

	return nil
}

func (p *Pulumi) getPulumiAbout() {
	if len(p.Stack) == 0 {
		p.env.Error(fmt.Errorf("pulumi stack name is empty, use `fetch_stack` property to enable stack fetching"))
		return
	}

	aboutOutput, err := p.env.RunCommand("pulumi", "about", "--json")

	if err != nil {
		p.env.Error(fmt.Errorf("unable to get pulumi about output"))
		return
	}

	var about struct {
		Backend *backend `json:"backend"`
	}

	err = json.Unmarshal([]byte(aboutOutput), &about)
	if err != nil {
		p.env.Error(fmt.Errorf("pulumi about output decode error"))
		return
	}

	if about.Backend == nil {
		p.env.Debug("pulumi about backend is not set")
		return
	}

	p.backend = *about.Backend
}
