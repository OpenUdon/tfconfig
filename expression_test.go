package tfconfig

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestLiteralFromCtyPreservesCollectionShape(t *testing.T) {
	tests := []struct {
		name string
		in   cty.Value
		want CollectionKind
	}{
		{name: "object", in: cty.ObjectVal(map[string]cty.Value{"a": cty.StringVal("one")}), want: CollectionKindObject},
		{name: "map", in: cty.MapVal(map[string]cty.Value{"a": cty.StringVal("one")}), want: CollectionKindMap},
		{name: "tuple", in: cty.TupleVal([]cty.Value{cty.StringVal("one")}), want: CollectionKindTuple},
		{name: "list", in: cty.ListVal([]cty.Value{cty.StringVal("one")}), want: CollectionKindList},
		{name: "set", in: cty.SetVal([]cty.Value{cty.StringVal("one")}), want: CollectionKindSet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, collectionKind, literal, ok := literalFromCty(tt.in)
			if !ok || kind != ValueKindCollection || collectionKind != tt.want || literal == nil {
				t.Fatalf("literalFromCty() = %q/%q/%#v/%t, want collection/%q/literal/true", kind, collectionKind, literal, ok, tt.want)
			}
		})
	}
}

func TestLoadDirCollectionShapeIsDeterministicJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.tf", `
resource "example_resource" "main" {
  for_each = { blue = "primary", green = "secondary" }
  values   = ["one", "two"]
}
`)
	doc, err := LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	resource := requireModule(t, doc, "").Resources[0]
	if resource.ForEach == nil || resource.ForEach.CollectionKind != CollectionKindObject {
		t.Fatalf("for_each collection = %#v, want object", resource.ForEach)
	}
	if len(resource.Config) != 1 || resource.Config[0].Value.CollectionKind != CollectionKindTuple {
		t.Fatalf("values collection = %#v, want tuple", resource.Config)
	}
	keys := make([]string, 0, len(resource.ForEach.Literal.(map[string]any)))
	for key := range resource.ForEach.Literal.(map[string]any) {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	if !slices.Equal(keys, []string{"blue", "green"}) {
		t.Fatalf("for_each keys = %v", keys)
	}
	data, err := json.Marshal(resource.ForEach)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"kind":"collection","collection_kind":"object","literal":{"blue":"primary","green":"secondary"},"range":{"source_id":"main.tf","path":"main.tf","start":{"line":2,"column":14,"byte":50},"end":{"line":2,"column":55,"byte":91}}}` {
		t.Fatalf("collection JSON drifted: %s", data)
	}
}
