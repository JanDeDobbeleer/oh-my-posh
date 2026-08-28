package segments

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	Command options.Option = "command"
)

// terraformVersionFields lists what the version fetch populates: the single
// derived unit of this segment (see FieldRefs). Default-off historically,
// so the unanalyzable fallback narrows by the substring heuristic like the
// SCM units, unlike the language segments' fail-open version fetch.
var terraformVersionFields = []string{"Version"}

type Terraform struct {
	Base
	TerraformBlock
	WorkspaceName string
	FieldRefs
}

func (tf *Terraform) Template() string {
	return " {{ .WorkspaceName }}{{ if .Version }} {{ .Version }}{{ end }} "
}

type TerraformBlock struct {
	Version *string `json:"terraform_version"`
}

// contextConditions returns the folders and file globs whose presence puts
// the segment in context, shared between the activation gate and inContext
// so the two can never diverge.
func (tf *Terraform) contextConditions(fetchVersion bool) (folders, globs []string) {
	folders = []string{".terraform"}
	globs = []string{".tf", ".tfplan", ".tfstate"}

	if fetchVersion {
		_, tenvVersionFile := tf.tenvSources()
		globs = append(globs, "versions.tf", "main.tf", "terraform.tfstate", tenvVersionFile)
	}

	return folders, globs
}

// Activation gates on the context conditions: the segment can only activate
// when the cwd carries the .terraform folder or one of the files the
// context check reacts to.
// The reference set is delivered right after Init, before the engine
// consults the gate, so the version-file conditions join exactly when the
// version fetch is derived on.
func (tf *Terraform) Activation() Activation {
	folders, globs := tf.contextConditions(tf.fetchUnit(terraformVersionFields...))

	return Activation{
		Folders:   folders,
		FileGlobs: globs,
	}
}

// inContext re-verifies the context conditions even though a passing gate
// implies a match: Force and pinned data bypass the gate, so Enabled must
// stay standalone-correct. The re-check hits the memoized directory listing
// and stat results.
func (tf *Terraform) inContext(fetchVersion bool) bool {
	folders, globs := tf.contextConditions(fetchVersion)

	for _, folder := range folders {
		if tf.env.HasFolder(filepath.Join(tf.env.Pwd(), folder)) {
			return true
		}
	}

	return slices.ContainsFunc(globs, tf.env.HasFiles)
}

func (tf *Terraform) Enabled() bool {
	cmd := tf.options.String(Command, "terraform")
	fetchVersion := tf.fetchUnit(terraformVersionFields...)

	if !tf.env.HasCommand(cmd) || !tf.inContext(fetchVersion) {
		return false
	}

	tf.WorkspaceName, _ = tf.env.RunCommand(cmd, "workspace", "show")
	if !fetchVersion {
		return true
	}

	// tenv (https://github.com/tofuutils/tenv) pins the version through an
	// environment variable or a version file, which takes precedence over the
	// version declared in the terraform files or the state file.
	if tf.setVersionFromTenv() {
		return true
	}

	if err := tf.setVersionFromTfFiles(); err == nil {
		return true
	}

	tf.setVersionFromTfStateFile()
	return true
}

func (tf *Terraform) tenvSources() (envVar, versionFile string) {
	cmd := tf.options.String(Command, "terraform")
	if strings.Contains(cmd, "tofu") {
		return "TOFUENV_TOFU_VERSION", ".opentofu-version"
	}

	return "TFENV_TERRAFORM_VERSION", ".terraform-version"
}

func (tf *Terraform) setVersionFromTenv() bool {
	envVar, versionFile := tf.tenvSources()

	// the environment variable takes precedence over the version file
	if version := strings.TrimSpace(tf.env.Getenv(envVar)); version != "" {
		tf.Version = &version
		return true
	}

	if !tf.env.HasFiles(versionFile) {
		return false
	}

	version := strings.TrimSpace(tf.env.FileContent(versionFile))
	// a version file holds a single version reference, guard against trailing content
	if line, _, found := strings.Cut(version, "\n"); found {
		version = strings.TrimSpace(line)
	}

	if version == "" {
		return false
	}

	tf.Version = &version
	return true
}

func (tf *Terraform) setVersionFromTfFiles() error {
	files := []string{"versions.tf", "main.tf"}
	for _, file := range files {
		if !tf.env.HasFiles(file) {
			continue
		}

		content := tf.env.FileContent(file)
		version, ok := extractRequiredVersion(content)
		if !ok {
			continue
		}

		tf.Version = &version
		return nil
	}
	return errors.New("no valid terraform files found")
}

func (tf *Terraform) setVersionFromTfStateFile() {
	file := "terraform.tfstate"
	if !tf.env.HasFiles(file) {
		return
	}

	content := tf.env.FileContent(file)
	_ = json.Unmarshal([]byte(content), &tf.TerraformBlock)
}
