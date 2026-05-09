package tfconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDirDecodesM4StaticFacts(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "versions.tf", `
terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
`)
	writeTestFile(t, dir, "providers.tf", `
provider "aws" {
  region = var.region
}

provider "aws" {
  alias = "west"
  region = "us-west-2"
}
`)
	writeTestFile(t, dir, "variables.tf", `
variable "region" {
  type = string
  default = "us-east-1"
}

locals {
  name = "${var.region}-web"
}
`)
	writeTestFile(t, dir, "main.tf", `
resource "aws_instance" "web" {
  provider = aws.west
  ami = var.ami
  token = "do-not-emit"
  count = 2
  depends_on = [data.aws_ami.base]

  lifecycle {
    prevent_destroy = true
    ignore_changes = [tags]
    replace_triggered_by = [aws_security_group.web]
    precondition {
      condition = var.region != ""
      error_message = "region required"
    }
  }
}

data "aws_ami" "base" {
  most_recent = true
}

module "child" {
  source = "./modules/child"
  providers = {
    aws = aws.west
  }
  name = local.name
}

moved {
  from = aws_instance.old
  to = aws_instance.web
}

import {
  to = aws_instance.web
  id = "i-123"
}

removed {
  from = aws_instance.gone
}

check "health" {
  assert {
    condition = data.aws_ami.base.id != ""
    error_message = "AMI missing"
  }
}
`)
	writeTestFile(t, dir, "outputs.tofu", `
output "instance_name" {
  value = local.name
  depends_on = [aws_instance.web]
}
`)
	writeTestFile(t, filepath.Join(dir, "tests"), "main.tftest.hcl", `
run "basic" {
  command = plan
  variables {
    region = "us-west-2"
  }
  assert {
    condition = output.instance_name != ""
    error_message = "missing name"
  }
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	if len(doc.Diagnostics) != 0 {
		t.Fatalf("document diagnostics = %#v, want none", doc.Diagnostics)
	}
	if len(doc.Modules) != 1 {
		t.Fatalf("modules = %d, want 1", len(doc.Modules))
	}
	mod := doc.Modules[0]
	if len(mod.RequiredVersions) != 1 {
		t.Fatalf("required versions = %d, want 1", len(mod.RequiredVersions))
	}
	if got := mod.RequiredProviders[0].Source; got != "hashicorp/aws" {
		t.Fatalf("required provider source = %q, want hashicorp/aws", got)
	}
	if len(mod.ProviderConfigs) != 2 {
		t.Fatalf("provider configs = %d, want 2", len(mod.ProviderConfigs))
	}
	if len(mod.Resources) != 1 || mod.Resources[0].Provider == nil || mod.Resources[0].Provider.Address != "provider.aws.west" {
		t.Fatalf("resource provider ref not decoded: %#v", mod.Resources)
	}
	if len(mod.Resources[0].Lifecycle.ReplaceTriggeredBy) != 1 {
		t.Fatalf("lifecycle replace_triggered_by not decoded: %#v", mod.Resources[0].Lifecycle)
	}
	if len(mod.DataSources) != 1 || mod.DataSources[0].Address != "data.aws_ami.base" {
		t.Fatalf("data source not decoded: %#v", mod.DataSources)
	}
	if len(mod.ModuleCalls) != 1 || len(mod.ModuleCalls[0].ProviderMappings) != 1 {
		t.Fatalf("module call provider mappings not decoded: %#v", mod.ModuleCalls)
	}
	if len(mod.Moved) != 1 || len(mod.Imports) != 1 || len(mod.Removed) != 1 || len(mod.Checks) != 1 {
		t.Fatalf("structural blocks not decoded: moved=%d import=%d removed=%d checks=%d", len(mod.Moved), len(mod.Imports), len(mod.Removed), len(mod.Checks))
	}
	if len(mod.Tests) != 1 || len(mod.Tests[0].Runs) != 1 {
		t.Fatalf("test facts not decoded: %#v", mod.Tests)
	}

	out, err := doc.JSONIndent("", "  ")
	if err != nil {
		t.Fatalf("JSON projection failed: %v", err)
	}
	if strings.Contains(string(out), "do-not-emit") {
		t.Fatalf("public JSON leaked sensitive candidate:\n%s", out)
	}
	if !strings.Contains(string(out), `"kind": "redacted"`) {
		t.Fatalf("public JSON did not include redacted sensitive candidate:\n%s", out)
	}
}

func TestLoadDirUsesTofuAlternativeAndOverrideOrdering(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "z.tf", `variable "name" { default = "tf" }`)
	writeTestFile(t, dir, "z.tofu", `variable "name" { default = "tofu" }`)
	writeTestFile(t, dir, "a_override.tf", `variable "name" { default = "override" }`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	mod := doc.Modules[0]
	if len(mod.SourceFiles) != 2 {
		t.Fatalf("source files = %#v, want tofu primary plus override", mod.SourceFiles)
	}
	if mod.SourceFiles[0] != "z.tofu" || mod.SourceFiles[1] != "a_override.tf" {
		t.Fatalf("source file ordering = %#v, want z.tofu then a_override.tf", mod.SourceFiles)
	}
	if len(mod.Variables) != 1 || mod.Variables[0].Default == nil || mod.Variables[0].Default.Literal != "override" {
		t.Fatalf("override variable not applied: %#v", mod.Variables)
	}
}

func TestLoadDirDecodesJSONConfigFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf.json", `{
  "variable": {
    "json_name": {
      "default": "json"
    }
  },
  "output": {
    "json_name": {
      "value": "${var.json_name}"
    }
  }
}`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	mod := doc.Modules[0]
	if len(mod.Variables) != 1 || mod.Variables[0].Name != "json_name" {
		t.Fatalf("JSON variable not decoded: %#v", mod.Variables)
	}
	if len(mod.Outputs) != 1 || len(mod.Outputs[0].References) != 1 {
		t.Fatalf("JSON output references not decoded: %#v", mod.Outputs)
	}
	if len(doc.SourceFiles) != 1 || doc.SourceFiles[0].Format != FileFormatJSON {
		t.Fatalf("JSON source file not recorded: %#v", doc.SourceFiles)
	}
}

func TestLoadDirReturnsParseDiagnostics(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `resource "bad" {`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	if len(doc.Diagnostics) == 0 {
		t.Fatalf("expected parse diagnostics")
	}
	if doc.Diagnostics[0].Range == nil || doc.Diagnostics[0].Range.Path != "main.tf" {
		t.Fatalf("diagnostic range = %#v, want main.tf range", doc.Diagnostics[0].Range)
	}
}

func TestLoadDirJSONIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "b.tf", `output "b" { value = var.b }`)
	writeTestFile(t, dir, "a.tf", `variable "a" { default = "a" }`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	first, err := doc.JSON()
	if err != nil {
		t.Fatalf("first JSON failed: %v", err)
	}
	second, err := doc.JSON()
	if err != nil {
		t.Fatalf("second JSON failed: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("JSON not deterministic\nfirst: %s\nsecond: %s", first, second)
	}
	var decoded Document
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatalf("decode deterministic JSON: %v", err)
	}
}

func writeTestFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(strings.TrimSpace(body)+"\n"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
