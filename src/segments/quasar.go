package segments

import (
	"encoding/json"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	FetchDependencies options.Option = "fetch_dependencies"
)

type Package struct {
	Version string `json:"version"`
	Dev     bool   `json:"dev"`
}

type Quasar struct {
	Vite    *Package
	AppVite *Package
	Language
	HasVite bool
}

func (q *Quasar) Enabled() bool {
	q.loadSpec()

	if !q.Language.Enabled() {
		return false
	}

	if q.options.Bool(FetchDependencies, false) {
		q.fetchDependencies()
	}

	return true
}

// Activation implements the activation gate; see Language.activation.
func (q *Quasar) Activation() Activation {
	q.loadSpec()

	return q.activation()
}

func (q *Quasar) loadSpec() {
	const quasarToolName = "quasar"

	q.projectFiles = []string{"quasar.config", "quasar.config.js"}
	q.tooling = map[string]*cmd{
		// quasar --version reports the Quasar CLI's own version, not the
		// project's quasar/vite dependency versions - those are read
		// separately from package-lock.json by fetchDependencies below.
		quasarToolName: {
			executable:       quasarToolName,
			args:             []string{versionFlagArg},
			regex:            versionRegex,
			versionCacheable: true,
		},
	}
	q.defaultTooling = []string{quasarToolName}
	q.versionURLTemplate = "https://github.com/quasarframework/quasar/releases/tag/quasar-v{{ .Full }}"
}

func (q *Quasar) Template() string {
	return " \ue87f {{.Full}}{{ if .HasVite }} \ueb29 {{ .Vite.Version }}{{ end }} "
}

func (q *Quasar) fetchDependencies() {
	if !q.env.HasFilesInDir(q.projectRoot.ParentFolder, "package-lock.json") {
		return
	}

	packageFilePath := filepath.Join(q.projectRoot.ParentFolder, "package-lock.json")
	content := q.env.FileContent(packageFilePath)

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
