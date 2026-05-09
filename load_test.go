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
	writeTestFile(t, filepath.Join(dir, "modules", "child"), "main.tf", `
variable "name" {}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	if len(doc.Diagnostics) != 0 {
		t.Fatalf("document diagnostics = %#v, want none", doc.Diagnostics)
	}
	mod := requireModule(t, doc, "")
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

func TestLoadDirRedactsNestedSensitiveCandidates(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
resource "example_resource" "main" {
  config = {
    password = "nested-secret-do-not-emit"
    child = {
      api_key = "nested-api-key-do-not-emit"
    }
  }
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	mod := requireModule(t, doc, "")
	if len(mod.Resources) != 1 || len(mod.Resources[0].Config) != 1 {
		t.Fatalf("resource config not decoded: %#v", mod.Resources)
	}
	value := mod.Resources[0].Config[0].Value
	if value.SensitiveCandidate == nil {
		t.Fatalf("nested sensitive candidate was not detected: %#v", value)
	}
	if got := value.SensitiveCandidate.AttributePath; got != "config.password" {
		t.Fatalf("sensitive candidate path = %q, want config.password", got)
	}

	out, err := doc.JSONIndent("", "  ")
	if err != nil {
		t.Fatalf("JSON projection failed: %v", err)
	}
	for _, leaked := range []string{"nested-secret-do-not-emit", "nested-api-key-do-not-emit"} {
		if strings.Contains(string(out), leaked) {
			t.Fatalf("public JSON leaked nested sensitive candidate %q:\n%s", leaked, out)
		}
	}
	if !strings.Contains(string(out), `"attribute_path": "config.password"`) {
		t.Fatalf("public JSON did not report nested sensitive candidate path:\n%s", out)
	}
}

func TestLoadDirLoadsLocalModuleTree(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
provider "aws" {
  alias = "west"
  region = "us-west-2"
}

module "child" {
  source = "./modules/child"
  providers = {
    aws = aws.west
  }
  name = var.name
  count = 1
  depends_on = [aws_instance.root]
}
`)
	writeTestFile(t, filepath.Join(dir, "modules", "child"), "main.tf", `
variable "name" {}

resource "example_child" "main" {
  name = var.name
}

module "grandchild" {
  source = "./grandchild"
  for_each = toset(["one"])
}
`)
	writeTestFile(t, filepath.Join(dir, "modules", "child", "grandchild"), "main.tf", `
output "ready" {
  value = true
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	if len(doc.Diagnostics) != 0 {
		t.Fatalf("document diagnostics = %#v, want none", doc.Diagnostics)
	}

	root := requireModule(t, doc, "")
	child := requireModule(t, doc, "module.child")
	grandchild := requireModule(t, doc, "module.child.module.grandchild")

	if child.Status != ModuleStatusLoaded || child.ParentAddress != "" || child.Dir != "modules/child" {
		t.Fatalf("child module metadata = %#v, want loaded root child in modules/child", child)
	}
	if grandchild.Status != ModuleStatusLoaded || grandchild.ParentAddress != "module.child" || grandchild.Dir != "modules/child/grandchild" {
		t.Fatalf("grandchild module metadata = %#v, want loaded nested child", grandchild)
	}
	if len(child.Resources) != 1 || child.Resources[0].Address != "example_child.main" {
		t.Fatalf("child resources not decoded: %#v", child.Resources)
	}
	if len(grandchild.Outputs) != 1 || grandchild.Outputs[0].Name != "ready" {
		t.Fatalf("grandchild outputs not decoded: %#v", grandchild.Outputs)
	}
	if len(root.ModuleCalls) != 1 {
		t.Fatalf("root module calls = %#v, want one", root.ModuleCalls)
	}
	call := root.ModuleCalls[0]
	if call.Address != "module.child" ||
		len(call.Inputs) != 1 ||
		len(call.ProviderMappings) != 1 ||
		call.Count == nil ||
		len(call.DependsOn) != 1 {
		t.Fatalf("module call facts not preserved: %#v", call)
	}
	if len(child.ModuleCalls) != 1 || child.ModuleCalls[0].Address != "module.child.module.grandchild" || child.ModuleCalls[0].ForEach == nil {
		t.Fatalf("nested module call facts not preserved with full address: %#v", child.ModuleCalls)
	}
}

func TestLoadDirDiagnosesUnavailableModuleSources(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
module "missing" {
  source = "./missing"
}

module "registry" {
  source = "hashicorp/consul/aws"
}

module "git" {
  source = "git::https://example.com/mod.git"
}

module "http" {
  source = "https://example.com/mod.zip"
}

module "s3" {
  source = "s3://bucket/key"
}

module "oci" {
  source = "oci://registry.example.com/mod"
}

module "symbolic" {
  source = var.module_source
}

module "bare_dot" {
  source = "."
}

module "bare_parent" {
  source = ".."
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}

	assertModuleDiagnostic(t, requireModule(t, doc, "module.missing"), ModuleStatusMissing, "module_source_missing")
	for _, address := range []string{"module.registry", "module.git", "module.http", "module.s3", "module.oci"} {
		assertModuleDiagnostic(t, requireModule(t, doc, address), ModuleStatusRemote, "module_source_remote")
	}
	assertModuleDiagnostic(t, requireModule(t, doc, "module.symbolic"), ModuleStatusUnsupported, "module_source_unsupported")
	assertModuleDiagnostic(t, requireModule(t, doc, "module.bare_dot"), ModuleStatusUnsupported, "module_source_unsupported")
	assertModuleDiagnostic(t, requireModule(t, doc, "module.bare_parent"), ModuleStatusUnsupported, "module_source_unsupported")
}

func TestLoadDirLoadsWindowsStyleLocalModuleSource(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
module "child" {
  source = ".\\modules\\child"
}
`)
	writeTestFile(t, filepath.Join(dir, "modules", "child"), "main.tf", `
output "ready" {
  value = true
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	child := requireModule(t, doc, "module.child")
	if child.Status != ModuleStatusLoaded || child.Dir != "modules/child" {
		t.Fatalf("windows-style local module source not loaded: %#v", child)
	}
	if len(child.Outputs) != 1 || child.Outputs[0].Name != "ready" {
		t.Fatalf("child output not decoded: %#v", child.Outputs)
	}
}

func TestLoadDirDiagnosesUnreadableLocalModuleSource(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
module "locked" {
  source = "./locked"
}
`)
	lockedDir := filepath.Join(dir, "locked")
	if err := os.MkdirAll(lockedDir, 0o755); err != nil {
		t.Fatalf("mkdir locked: %v", err)
	}
	if err := os.Chmod(lockedDir, 0); err != nil {
		t.Fatalf("chmod locked: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(lockedDir, 0o755)
	})

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	assertModuleDiagnostic(t, requireModule(t, doc, "module.locked"), ModuleStatusMissing, "module_source_missing")
}

func TestLoadDirSetsChildDiagnosticModuleAddress(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
module "child" {
  source = "./child"
}
`)
	writeTestFile(t, filepath.Join(dir, "child"), "main.tf", `
variable "name" {}
variable "name" {}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	child := requireModule(t, doc, "module.child")
	if len(child.Diagnostics) != 1 || child.Diagnostics[0].Code != "duplicate_declaration" {
		t.Fatalf("child diagnostics = %#v, want duplicate declaration", child.Diagnostics)
	}
	if got := child.Diagnostics[0].ModuleAddress; got != "module.child" {
		t.Fatalf("child diagnostic module address = %q, want module.child", got)
	}
}

func TestLoadDirModuleCycleGuard(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
module "a" {
  source = "./a"
}
`)
	writeTestFile(t, filepath.Join(dir, "a"), "main.tf", `
module "back" {
  source = "../"
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	assertModuleDiagnostic(t, requireModule(t, doc, "module.a.module.back"), ModuleStatusUnsupported, "module_source_cycle")
}

func TestLoadDirDoesNotReadModuleManifest(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
module "remote" {
  source = "hashicorp/consul/aws"
}
`)
	writeTestFile(t, filepath.Join(dir, ".terraform", "modules"), "modules.json", `{
  "Modules": [
    {"Key": "remote", "Source": "hashicorp/consul/aws", "Dir": "../../cached"}
  ]
}`)
	writeTestFile(t, filepath.Join(dir, "cached"), "main.tf", `
resource "should_not" "load" {}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	remote := requireModule(t, doc, "module.remote")
	assertModuleDiagnostic(t, remote, ModuleStatusRemote, "module_source_remote")
	if len(remote.Resources) != 0 {
		t.Fatalf("manifest-backed remote module was loaded: %#v", remote.Resources)
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

func TestLoadDirDiagnosesDuplicateDeclarations(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "a.tf", `variable "name" { default = "first" }`)
	writeTestFile(t, dir, "b.tf", `variable "name" { default = "second" }`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	mod := doc.Modules[0]
	if len(mod.Diagnostics) != 1 || mod.Diagnostics[0].Code != "duplicate_declaration" {
		t.Fatalf("duplicate declaration diagnostics = %#v, want one duplicate_declaration", mod.Diagnostics)
	}
	if len(mod.Variables) != 1 || mod.Variables[0].Default == nil || mod.Variables[0].Default.Literal != "first" {
		t.Fatalf("duplicate declaration did not preserve first variable: %#v", mod.Variables)
	}
}

func TestLoadDirMergesOverrideResourceAttributes(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
resource "example_resource" "main" {
  a = "base-a"
  b = "base-b"
}
`)
	writeTestFile(t, dir, "override.tf", `
resource "example_resource" "main" {
  b = "override-b"
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	mod := doc.Modules[0]
	if len(mod.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", mod.Diagnostics)
	}
	if len(mod.Resources) != 1 {
		t.Fatalf("resources = %d, want 1", len(mod.Resources))
	}
	got := map[string]any{}
	for _, attr := range mod.Resources[0].Config {
		got[attr.Path] = attr.Value.Literal
	}
	if got["a"] != "base-a" || got["b"] != "override-b" {
		t.Fatalf("merged resource attrs = %#v, want a=base-a b=override-b", got)
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

func TestLoadDirPropagatesSensitiveDeclarationsToValues(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
variable "api_token" {
  sensitive = true
  default = "plain-token"
}

output "token" {
  sensitive = true
  value = "plain-output"
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	mod := doc.Modules[0]
	if len(mod.Variables) != 1 || mod.Variables[0].Default == nil || !mod.Variables[0].Default.Sensitive {
		t.Fatalf("sensitive variable default not marked sensitive: %#v", mod.Variables)
	}
	if len(mod.Outputs) != 1 || mod.Outputs[0].Value == nil || !mod.Outputs[0].Value.Sensitive {
		t.Fatalf("sensitive output value not marked sensitive: %#v", mod.Outputs)
	}

	out, err := doc.JSONIndent("", "  ")
	if err != nil {
		t.Fatalf("JSON projection failed: %v", err)
	}
	if strings.Contains(string(out), "plain-token") || strings.Contains(string(out), "plain-output") {
		t.Fatalf("sensitive declaration value leaked literal:\n%s", out)
	}
}

func TestLoadDirPreservesNestedProviderBlocks(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
resource "example_resource" "main" {
  nested {
    value = "hidden"
  }
  repeated {
    id = "one"
  }
  repeated {
    id = "two"
  }
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	mod := doc.Modules[0]
	if len(mod.Diagnostics) != 0 {
		t.Fatalf("nested provider blocks should not produce diagnostics: %#v", mod.Diagnostics)
	}
	got := map[string]string{}
	for _, attr := range mod.Resources[0].Config {
		if text, ok := attr.Value.Literal.(string); ok {
			got[attr.Path] = text
		}
	}
	for path, want := range map[string]string{
		"nested.value":   "hidden",
		"repeated[0].id": "one",
		"repeated[1].id": "two",
	} {
		if got[path] != want {
			t.Fatalf("nested provider block attribute %s = %q, want %q; config=%#v", path, got[path], want, mod.Resources[0].Config)
		}
	}
}

func TestLoadDirDecodesBareTestRunCommand(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `variable "name" { default = "x" }`)
	writeTestFile(t, filepath.Join(dir, "tests"), "main.tftest.hcl", `
run "basic" {
  command = plan
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	mod := doc.Modules[0]
	if len(mod.Tests) != 1 || len(mod.Tests[0].Runs) != 1 {
		t.Fatalf("test runs not decoded: %#v", mod.Tests)
	}
	if got := mod.Tests[0].Runs[0].Command; got != "plan" {
		t.Fatalf("test run command = %q, want plan", got)
	}
}

func TestLoadDirSurfacesNestedTestDiagnostics(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `variable "name" { default = "x" }`)
	writeTestFile(t, filepath.Join(dir, "tests"), "main.tftest.hcl", `
run "basic" {
  unsupported {
    value = "hidden"
  }
}
`)

	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir failed: %v", err)
	}
	if len(doc.Modules[0].Diagnostics) == 0 {
		t.Fatalf("expected test file nested diagnostics")
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

func requireModule(t *testing.T, doc Document, address string) Module {
	t.Helper()
	for _, mod := range doc.Modules {
		if mod.Address == address {
			return mod
		}
	}
	t.Fatalf("module %q not found in %#v", address, doc.Modules)
	return Module{}
}

func assertModuleDiagnostic(t *testing.T, mod Module, status ModuleStatus, code string) {
	t.Helper()
	if mod.Status != status {
		t.Fatalf("module %s status = %s, want %s", mod.Address, mod.Status, status)
	}
	if len(mod.Diagnostics) != 1 || mod.Diagnostics[0].Code != code {
		t.Fatalf("module %s diagnostics = %#v, want one %s", mod.Address, mod.Diagnostics, code)
	}
}
