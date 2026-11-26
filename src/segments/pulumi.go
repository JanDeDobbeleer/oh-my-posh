package segments

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/path"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	yaml "go.yaml.in/yaml/v3"
)

const (
	FetchStack options.Option = "fetch_stack"
	FetchAbout options.Option = "fetch_about"

	JSON string = "json"
	YAML string = "yaml"

	pulumiJSON string = "Pulumi.json"
	pulumiYAML string = "Pulumi.yaml"
)

type Pulumi struct {
	Base

	Stack string
	Name  string

	workspaceSHA1 string
	projectBackendURL string

	Backend
}

type Backend struct {
	URL  string `json:"url"`
	User string `json:"user"`
}

type pulumiFileSpec struct {
	Name    string `yaml:"name" json:"name"`
	Backend struct {
		URL string `yaml:"url" json:"url"`
	} `yaml:"backend" json:"backend"`
}

type pulumiWorkSpaceFileSpec struct {
	Stack  string            `json:"stack"`
	Stacks map[string]string `json:"stacks"`
}

type pulumiCredentialsSpec struct {
	Current string `json:"current"`
}

func (p *Pulumi) Template() string {
	return "\ue873 {{ .Stack }}{{if .User }} :: {{ .User }}@{{ end }}{{ if .URL }}{{ .URL }}{{ end }}"
}

func (p *Pulumi) Enabled() bool {
	if !p.env.HasCommand("pulumi") {
		return false
	}

	err := p.getProjectName()
	if err != nil {
		log.Error(err)
		return false
	}

	if p.options.Bool(FetchStack, false) {
		p.getPulumiStackName()
	}

	if p.options.Bool(FetchAbout, false) {
		p.getPulumiAbout()
	}

	return true
}

func (p *Pulumi) cleanStackName(stack string) string {
	// If fully qualified (e.g. organization/project/stack or owner/project/stack),
	// extract the leaf stack name.
	parts := strings.Split(stack, "/")
	return parts[len(parts)-1]
}

func (p *Pulumi) getActiveBackendURL(pulumiHome string) string {
	// 1. Check PULUMI_BACKEND_URL env var
	if envBackend := p.env.Getenv("PULUMI_BACKEND_URL"); envBackend != "" {
		return envBackend
	}

	// 2. Check project backend specified in Pulumi.yaml / Pulumi.json
	if p.projectBackendURL != "" {
		return p.projectBackendURL
	}

	// 3. Check credentials.json "current" logged-in backend
	if p.env.HasFilesInDir(pulumiHome, "credentials.json") {
		credsFile := filepath.Join(pulumiHome, "credentials.json")
		content := p.env.FileContent(credsFile)
		var creds pulumiCredentialsSpec
		if err := json.Unmarshal([]byte(content), &creds); err == nil && creds.Current != "" {
			return creds.Current
		}
	}

	return ""
}

func (p *Pulumi) getPulumiStackName() {
	// 0. High priority: PULUMI_STACK environment variable
	if envStack := p.env.Getenv("PULUMI_STACK"); envStack != "" {
		p.Stack = p.cleanStackName(envStack)
		return
	}

	if p.Name == "" || p.workspaceSHA1 == "" {
		log.Debug("pulumi project name or workspace sha1 is empty")
		return
	}

	stackNameFile := p.Name + "-" + p.workspaceSHA1 + "-" + "workspace.json"

	pulumiHome := p.env.Getenv("PULUMI_HOME")
	if pulumiHome == "" {
		pulumiHome = filepath.Join(p.env.Home(), ".pulumi")
	}

	workspaceCacheDir := filepath.Join(pulumiHome, "workspaces")
	if !p.env.HasFolder(workspaceCacheDir) || !p.env.HasFilesInDir(workspaceCacheDir, stackNameFile) {
		return
	}

	workspaceCacheFile := filepath.Join(workspaceCacheDir, stackNameFile)
	workspaceCacheFileContent := p.env.FileContent(workspaceCacheFile)

	var pulumiWorkspaceSpec pulumiWorkSpaceFileSpec
	err := json.Unmarshal([]byte(workspaceCacheFileContent), &pulumiWorkspaceSpec)
	if err != nil {
		log.Error(fmt.Errorf("pulumi workspace file decode error"))
		return
	}

	// 1. If backend URL is known, resolve the matching stack from the Stacks map
	backendURL := p.getActiveBackendURL(pulumiHome)
	if backendURL != "" && len(pulumiWorkspaceSpec.Stacks) > 0 {
		if stack, ok := pulumiWorkspaceSpec.Stacks[backendURL]; ok && stack != "" {
			p.Stack = p.cleanStackName(stack)
			log.Debugf("pulumi stack name from backend %s: %s", backendURL, p.Stack)
			return
		}
	}

	// 2. If only one stack exists in the Stacks map, use it
	if len(pulumiWorkspaceSpec.Stacks) == 1 {
		for _, stack := range pulumiWorkspaceSpec.Stacks {
			p.Stack = p.cleanStackName(stack)
			log.Debugf("pulumi stack name from single stack entry: %s", p.Stack)
			return
		}
	}

	// 3. Fall back to legacy single stack field if present
	if pulumiWorkspaceSpec.Stack != "" {
		p.Stack = p.cleanStackName(pulumiWorkspaceSpec.Stack)
		log.Debugf("pulumi stack name: %s", p.Stack)
		return
	}

	// 4. Fallback for multi-stack when backend URL cannot be resolved: pick any entry
	if len(pulumiWorkspaceSpec.Stacks) > 0 {
		for _, stack := range pulumiWorkspaceSpec.Stacks {
			p.Stack = p.cleanStackName(stack)
			log.Debugf("pulumi stack name fallback: %s", p.Stack)
			return
		}
	}
}

func (p *Pulumi) getProjectName() error {
	var kind, fileName string
	for _, file := range []string{pulumiYAML, pulumiJSON} {
		if p.env.HasFiles(file) {
			fileName = file
			kind = filepath.Ext(file)[1:]
		}
	}

	if kind == "" {
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
		log.Error(err)
		return nil
	}

	p.Name = pulumiFileSpec.Name
	p.projectBackendURL = pulumiFileSpec.Backend.URL

	p.workspaceSHA1 = p.sha1HexString(p.env.Pwd() + path.Separator() + fileName)

	return nil
}

func (p *Pulumi) sha1HexString(s string) string {
	h := sha1.New()

	_, err := h.Write([]byte(s))
	if err != nil {
		log.Error(err)
		return ""
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (p *Pulumi) getPulumiAbout() {
	aboutOutput, err := p.env.RunCommand("pulumi", "about", "--json")
	if err != nil {
		log.Error(fmt.Errorf("unable to get pulumi about output"))
		return
	}

	var about struct {
		Backend      *Backend `json:"backend"`
		CurrentStack struct {
			Name string `json:"name"`
		} `json:"currentStack"`
	}

	err = json.Unmarshal([]byte(aboutOutput), &about)
	if err != nil {
		log.Error(fmt.Errorf("pulumi about output decode error"))
		return
	}

	// If fetch_stack failed or was disabled, use stack name from `pulumi about`
	if p.Stack == "" && about.CurrentStack.Name != "" {
		p.Stack = about.CurrentStack.Name
	}

	if about.Backend == nil {
		log.Debug("pulumi about backend is not set")
		return
	}

	p.Backend = *about.Backend
}
