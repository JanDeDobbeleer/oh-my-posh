package segments

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestExtractRequiredVersion(t *testing.T) {
	cases := []struct {
		Case     string
		Source   string
		Expected string
		OK       bool
	}{
		{
			Case:     "simple block",
			Source:   "terraform {\n  required_version = \">= 1.0.10\"\n}\n",
			Expected: ">= 1.0.10",
			OK:       true,
		},
		{
			Case: "nested required_providers before the attribute",
			Source: `terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
  required_version = ">= 1.3.0, < 2.0.0"
}`,
			Expected: ">= 1.3.0, < 2.0.0",
			OK:       true,
		},
		{
			Case: "required_version inside required_providers is ignored",
			Source: `terraform {
  required_providers {
    fake = {
      required_version = "9.9.9"
    }
  }
}`,
			OK: false,
		},
		{
			Case: "other top-level blocks before terraform",
			Source: `provider "aws" {
  region = "eu-west-1"
}

resource "aws_instance" "web" {
  tags = {
    Name = "terraform"
  }
}

terraform {
  required_version = "~> 1.5"
}`,
			Expected: "~> 1.5",
			OK:       true,
		},
		{
			Case: "comments of all flavors",
			Source: `# leading comment with terraform {
// another one with required_version = "0.0.0"
/* block comment
   terraform { required_version = "0.0.1" }
*/
terraform { # trailing
  /* inline */ required_version = "1.6.2" // done
}`,
			Expected: "1.6.2",
			OK:       true,
		},
		{
			Case: "string containing braces does not break depth tracking",
			Source: `locals {
  tpl = "prefix {  nested "
}

terraform {
  required_version = ">= 1.1"
}`,
			Expected: ">= 1.1",
			OK:       true,
		},
		{
			Case: "interpolation with nested string and braces",
			Source: `locals {
  greeting = "${var.enabled ? "y{es" : "no}"} done"
}

terraform {
  required_version = "1.7.0"
}`,
			Expected: "1.7.0",
			OK:       true,
		},
		{
			Case: "heredoc containing a fake terraform block",
			Source: `resource "local_file" "f" {
  content = <<EOF
terraform {
  required_version = "0.0.0"
}
EOF
}

terraform {
  required_version = "1.8.1"
}`,
			Expected: "1.8.1",
			OK:       true,
		},
		{
			Case: "indented heredoc",
			Source: `resource "local_file" "f" {
  content = <<-DOC
    terraform { required_version = "0.0.0" }
    DOC
}

terraform {
  required_version = "1.8.2"
}`,
			Expected: "1.8.2",
			OK:       true,
		},
		{
			Case:   "no terraform block",
			Source: "provider \"aws\" {\n  region = \"eu-west-1\"\n}\n",
			OK:     false,
		},
		{
			Case:   "terraform block without required_version",
			Source: "terraform {\n  required_providers {\n  }\n}\n",
			OK:     false,
		},
		{
			Case:   "empty source",
			Source: "",
			OK:     false,
		},
		{
			Case:   "unterminated string fails closed",
			Source: "terraform {\n  required_version = \">= 1.0",
			OK:     false,
		},
		{
			Case:     "escaped quote in value",
			Source:   `terraform { required_version = "a\"b" }`,
			Expected: `a"b`,
			OK:       true,
		},
		{
			Case:     "windows line endings",
			Source:   "terraform {\r\n  required_version = \">= 1.2.3\"\r\n}\r\n",
			Expected: ">= 1.2.3",
			OK:       true,
		},
		{
			Case: "identifier merely containing terraform is not a block",
			Source: `terraform_version = "0.0.0"

terraform {
  required_version = "1.9.9"
}`,
			Expected: "1.9.9",
			OK:       true,
		},
		{
			Case: "second terraform block wins when first has no version",
			Source: `terraform {
  backend "s3" {}
}

terraform {
  required_version = "1.4.4"
}`,
			Expected: "1.4.4",
			OK:       true,
		},
	}

	for _, tc := range cases {
		version, ok := extractRequiredVersion(tc.Source)
		assert.Equal(t, tc.OK, ok, tc.Case)
		assert.Equal(t, tc.Expected, version, tc.Case)
	}
}
