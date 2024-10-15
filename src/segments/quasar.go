package segments

import (
	"encoding/json"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

const (
	FetchDependencies properties.Property = "fetch_dependencies"
)

type Package struct {
	Version string `json:"version"`
	Dev     bool   `json:"dev"`
}

type Quasar struct {
	Vite    *Package
	AppVite *Package
	language
	HasVite bool
}

func (q *Quasar) Enabled() bool {
	if !q.language.Enabled() {
		return false
	}

	if q.language.props.GetBool(FetchDependencies, false) {
		q.fetchDependencies()
	}

	return true
}

func (q *Quasar) Template() string {
	return " \uea6a {{.Full}}{{ if .HasVite }} \ueb29 {{ .Vite.Version }}{{ end }} "
}

func (q *Quasar) Init(props properties.Properties, env runtime.Environment) {
	q.language = language{
		env:          env,
		props:        props,
		projectFiles: []string{"quasar.config", "quasar.config.js"},
		commands: []*cmd{
			{
				executable: "quasar",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://github.com/quasarframework/quasar/releases/tag/quasar-v{{ .Full }}",
	}
}

func (q *Quasar) fetchDependencies() {
	if !q.language.env.HasFilesInDir(q.projectRoot.ParentFolder, "package-lock.json") {
		return
	}

	packageFilePath := filepath.Join(q.projectRoot.ParentFolder, "package-lock.json")
	content := q.language.env.FileContent(packageFilePath)

	var objmap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(content), &objmap); err != nil {
		return
	}

	var dependencies map[string]*Package
	if err := json.Unmarshal(objmap["dependencies"], &dependencies); err != nil {
		return
	}

	if p, ok := dependencies["vite"]; ok {
		q.HasVite = true
		q.Vite = p
	}

	if p, ok := dependencies["@quasar/app-vite"]; ok {
		q.AppVite = p
	}
}
