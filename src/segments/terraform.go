package segments

import (
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type Terraform struct {
	props properties.Properties
	env   runtime.Environment
	TerraformBlock
	WorkspaceName string
}

func (tf *Terraform) Template() string {
	return " {{ .WorkspaceName }}{{ if .Version }} {{ .Version }}{{ end }} "
}

func (tf *Terraform) Init(props properties.Properties, env runtime.Environment) {
	tf.props = props
	tf.env = env
}

type TerraFormConfig struct {
	Terraform *TerraformBlock `hcl:"terraform,block"`
}

type TerraformBlock struct {
	Version *string `hcl:"required_version" json:"terraform_version"`
}

func (tf *Terraform) Enabled() bool {
	cmd := "terraform"
	fetchVersion := tf.props.GetBool(properties.FetchVersion, false)

	if !tf.env.HasCommand(cmd) || !tf.inContext(fetchVersion) {
		return false
	}

	tf.WorkspaceName, _ = tf.env.RunCommand(cmd, "workspace", "show")
	if !fetchVersion {
		return true
	}

	if err := tf.setVersionFromTfFiles(); err == nil {
		return true
	}

	tf.setVersionFromTfStateFile()
	return true
}

func (tf *Terraform) inContext(fetchVersion bool) bool {
	terraformFolder := filepath.Join(tf.env.Pwd(), ".terraform")

	if tf.env.HasFolder(terraformFolder) {
		return true
	}

	files := []string{".tf", ".tfplan", ".tfstate"}
	for _, file := range files {
		if tf.env.HasFiles(file) {
			return true
		}
	}

	if !fetchVersion {
		return false
	}

	versionFiles := []string{"versions.tf", "main.tf", "terraform.tfstate"}
	for _, file := range versionFiles {
		if tf.env.HasFiles(file) {
			return true
		}
	}

	return false
}

func (tf *Terraform) setVersionFromTfFiles() error {
	files := []string{"versions.tf", "main.tf"}
	for _, file := range files {
		if !tf.env.HasFiles(file) {
			continue
		}

		parser := hclparse.NewParser()
		content := tf.env.FileContent(file)
		hclFile, diags := parser.ParseHCL([]byte(content), file)
		if diags != nil {
			continue
		}

		var config TerraFormConfig
		diags = gohcl.DecodeBody(hclFile.Body, nil, &config)
		if diags != nil {
			continue
		}

		tf.TerraformBlock = *config.Terraform
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
