package tfconfig

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDocumentJSONIsDeterministicAndSafe(t *testing.T) {
	doc := NewDocument("./tf")
	doc.SourceFiles = []SourceFile{
		{ID: "root/main.tf", ModuleAddress: "", Path: "main.tf", Format: FileFormatHCL, Role: FileRolePrimary},
		{ID: "root/versions.tf", ModuleAddress: "", Path: "versions.tf", Format: FileFormatHCL, Role: FileRolePrimary},
	}
	doc.Modules = []Module{
		{
			Address: "module.app",
			Dir:     "modules/app",
			Status:  ModuleStatusLoaded,
		},
		{
			Address:     "",
			Dir:         ".",
			Status:      ModuleStatusRoot,
			SourceFiles: []string{"versions.tf", "main.tf"},
			ProviderConfigs: []ProviderConfig{
				{LocalName: "aws", Alias: "west", Address: "provider.aws.west"},
				{LocalName: "aws", Address: "provider.aws"},
			},
			Resources: []Resource{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Name:    "web",
					Config: []Attribute{
						{
							Path: "token",
							Value: Value{
								Kind:    ValueKindString,
								Literal: "super-secret",
								SensitiveCandidate: &SensitiveCandidate{
									Reason:        "attribute name suggests secret material",
									AttributePath: "token",
								},
							},
						},
						{
							Path: "name",
							Value: Value{
								Kind:       ValueKindExpression,
								Expression: "var.instance_name",
								References: []Reference{
									{Traversal: "var.instance_name", Subject: "var.instance_name"},
								},
							},
						},
					},
				},
			},
			Variables: []Variable{
				{Name: "z_name"},
				{Name: "a_name"},
			},
		},
	}

	first, err := doc.JSONIndent("", "  ")
	if err != nil {
		t.Fatalf("first JSON projection failed: %v", err)
	}
	second, err := doc.JSONIndent("", "  ")
	if err != nil {
		t.Fatalf("second JSON projection failed: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("JSON projection is not deterministic\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if strings.Contains(string(first), "super-secret") {
		t.Fatalf("public JSON projection leaked likely-secret literal:\n%s", first)
	}
	if !strings.Contains(string(first), `"kind": "redacted"`) {
		t.Fatalf("public JSON projection did not mark likely-secret value as redacted:\n%s", first)
	}
	if strings.Index(string(first), `"address": ""`) > strings.Index(string(first), `"address": "module.app"`) {
		t.Fatalf("modules are not sorted by address:\n%s", first)
	}
	if strings.Index(string(first), `"name": "a_name"`) > strings.Index(string(first), `"name": "z_name"`) {
		t.Fatalf("variables are not sorted by name:\n%s", first)
	}
}

func TestDocumentJSONDefaultsVersionAndProducer(t *testing.T) {
	doc := Document{RootDir: "."}
	got, err := doc.JSON()
	if err != nil {
		t.Fatalf("JSON projection failed: %v", err)
	}
	var decoded Document
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("decode JSON projection: %v", err)
	}
	if decoded.Version != StaticV1 {
		t.Fatalf("Version = %q, want %q", decoded.Version, StaticV1)
	}
	if decoded.Producer != DefaultProducer {
		t.Fatalf("Producer = %q, want %q", decoded.Producer, DefaultProducer)
	}
}

func TestValueJSONDoesNotMutateReferences(t *testing.T) {
	v := Value{
		Kind: ValueKindExpression,
		References: []Reference{
			{Traversal: "var.z"},
			{Traversal: "var.a"},
		},
	}

	if _, err := json.Marshal(v); err != nil {
		t.Fatalf("marshal value: %v", err)
	}

	if got := v.References[0].Traversal; got != "var.z" {
		t.Fatalf("Value.MarshalJSON mutated references[0] = %q, want var.z", got)
	}
}
