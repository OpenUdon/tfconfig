package tfconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var configFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "terraform"},
		{Type: "provider", LabelNames: []string{"name"}},
		{Type: "variable", LabelNames: []string{"name"}},
		{Type: "locals"},
		{Type: "output", LabelNames: []string{"name"}},
		{Type: "module", LabelNames: []string{"name"}},
		{Type: "resource", LabelNames: []string{"type", "name"}},
		{Type: "data", LabelNames: []string{"type", "name"}},
		{Type: "moved"},
		{Type: "import"},
		{Type: "removed"},
		{Type: "check", LabelNames: []string{"name"}},
	},
}

var terraformBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "required_version"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "required_providers"},
	},
}

func decodeConfigFile(mod *Module, file discoveredFile, body hcl.Body, sources map[string]sourceInfo) []Diagnostic {
	content, diags := body.Content(configFileSchema)
	out := convertDiagnostics(diags, mod.Address, "", sources)

	for _, block := range content.Blocks {
		switch block.Type {
		case "terraform":
			out = append(out, decodeTerraformBlock(mod, block, sources)...)
		case "provider":
			upsertProviderConfig(mod, decodeProviderBlock(block, sources))
		case "variable":
			upsertVariable(mod, decodeVariableBlock(block, sources))
		case "locals":
			for _, local := range decodeLocalsBlock(block, sources) {
				upsertLocal(mod, local)
			}
		case "output":
			upsertOutput(mod, decodeOutputBlock(block, sources))
		case "resource":
			upsertResource(mod, decodeResourceBlock(block, sources))
		case "data":
			upsertDataSource(mod, decodeDataSourceBlock(block, sources))
		case "module":
			upsertModuleCall(mod, decodeModuleCallBlock(block, sources))
		case "moved":
			mod.Moved = append(mod.Moved, decodeMovedBlock(block, sources))
		case "import":
			mod.Imports = append(mod.Imports, decodeImportBlock(block, sources))
		case "removed":
			mod.Removed = append(mod.Removed, decodeRemovedBlock(block, sources))
		case "check":
			mod.Checks = append(mod.Checks, decodeCheckBlock(block, sources))
		}
	}
	return out
}

func decodeTerraformBlock(mod *Module, block *hcl.Block, sources map[string]sourceInfo) []Diagnostic {
	content, diags := block.Body.Content(terraformBlockSchema)
	out := convertDiagnostics(diags, mod.Address, "", sources)
	if attr, ok := content.Attributes["required_version"]; ok {
		mod.RequiredVersions = append(mod.RequiredVersions, valueFromExpr(attr.Expr, "required_version", sources))
	}
	for _, inner := range content.Blocks {
		if inner.Type == "required_providers" {
			reqs, reqDiags := decodeRequiredProvidersBlock(inner, sources)
			out = append(out, reqDiags...)
			for _, req := range reqs {
				upsertProviderRequirement(mod, req)
			}
		}
	}
	return out
}

func decodeRequiredProvidersBlock(block *hcl.Block, sources map[string]sourceInfo) ([]ProviderRequirement, []Diagnostic) {
	attrs, diags := block.Body.JustAttributes()
	out := convertDiagnostics(diags, "", "", sources)
	var reqs []ProviderRequirement
	for name, attr := range attrs {
		req := ProviderRequirement{
			LocalName: name,
			Range:     sourceRange(attr.Range, sources),
		}
		val := valueFromExpr(attr.Expr, name, sources)
		if obj, ok := val.Literal.(map[string]any); ok {
			if source, ok := obj["source"].(string); ok {
				req.Source = source
			}
			if version, ok := obj["version"].(string); ok {
				req.VersionConstraints = []string{version}
			}
		} else if version, ok := val.Literal.(string); ok {
			req.VersionConstraints = []string{version}
		}
		reqs = append(reqs, req)
	}
	return reqs, out
}

func decodeProviderBlock(block *hcl.Block, sources map[string]sourceInfo) ProviderConfig {
	name := block.Labels[0]
	attrs := justAttributes(block.Body)
	cfg := ProviderConfig{
		LocalName: name,
		Address:   providerAddress(name, ""),
		Range:     sourceRange(block.DefRange, sources),
	}
	if attr, ok := attrs["alias"]; ok {
		if alias, ok := literalString(attr.Expr, sources); ok {
			cfg.Alias = alias
			cfg.Address = providerAddress(name, alias)
		}
	}
	cfg.Config = decodeBodyAttributes(block.Body, reservedProviderAttrs(), sources)
	cfg.References = referencesFromAttributes(cfg.Config)
	return cfg
}

func decodeVariableBlock(block *hcl.Block, sources map[string]sourceInfo) Variable {
	attrs := justAttributes(block.Body)
	v := Variable{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["type"]; ok {
		val := valueFromExpr(attr.Expr, "type", sources)
		val.Kind = ValueKindExpression
		v.Type = &val
	}
	if attr, ok := attrs["default"]; ok {
		val := valueFromExpr(attr.Expr, "default", sources)
		v.Default = &val
		v.References = append(v.References, val.References...)
	}
	if attr, ok := attrs["description"]; ok {
		v.Description, _ = literalString(attr.Expr, sources)
	}
	if attr, ok := attrs["sensitive"]; ok {
		v.Sensitive, _ = literalBool(attr.Expr)
	}
	if attr, ok := attrs["nullable"]; ok {
		if b, ok := literalBool(attr.Expr); ok {
			v.Nullable = &b
		}
	}
	return v
}

func decodeLocalsBlock(block *hcl.Block, sources map[string]sourceInfo) []Local {
	attrs := justAttributes(block.Body)
	locals := make([]Local, 0, len(attrs))
	for name, attr := range attrs {
		val := valueFromExpr(attr.Expr, name, sources)
		locals = append(locals, Local{
			Name:       name,
			Value:      &val,
			References: val.References,
			Range:      sourceRange(attr.Range, sources),
		})
	}
	return locals
}

func decodeOutputBlock(block *hcl.Block, sources map[string]sourceInfo) Output {
	attrs := justAttributes(block.Body)
	o := Output{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["value"]; ok {
		val := valueFromExpr(attr.Expr, "value", sources)
		o.Value = &val
		o.References = append(o.References, val.References...)
	}
	if attr, ok := attrs["description"]; ok {
		o.Description, _ = literalString(attr.Expr, sources)
	}
	if attr, ok := attrs["sensitive"]; ok {
		o.Sensitive, _ = literalBool(attr.Expr)
	}
	if attr, ok := attrs["depends_on"]; ok {
		o.DependsOn = referencesFromExpr(attr.Expr, sources)
	}
	return o
}

func decodeResourceBlock(block *hcl.Block, sources map[string]sourceInfo) Resource {
	typ, name := block.Labels[0], block.Labels[1]
	attrs := justAttributes(block.Body)
	res := Resource{
		Address: address(typ, name),
		Type:    typ,
		Name:    name,
		Range:   sourceRange(block.DefRange, sources),
	}
	if attr, ok := attrs["provider"]; ok {
		res.Provider = providerRefFromExpr(attr.Expr, sources)
	}
	if attr, ok := attrs["depends_on"]; ok {
		res.DependsOn = referencesFromExpr(attr.Expr, sources)
	}
	if attr, ok := attrs["count"]; ok {
		val := valueFromExpr(attr.Expr, "count", sources)
		res.Count = &val
		res.References = append(res.References, val.References...)
	}
	if attr, ok := attrs["for_each"]; ok {
		val := valueFromExpr(attr.Expr, "for_each", sources)
		res.ForEach = &val
		res.References = append(res.References, val.References...)
	}
	res.Config = decodeBodyAttributes(block.Body, reservedResourceAttrs(), sources)
	res.References = append(res.References, referencesFromAttributes(res.Config)...)
	res.Lifecycle = decodeLifecycle(block.Body, sources)
	return res
}

func decodeDataSourceBlock(block *hcl.Block, sources map[string]sourceInfo) DataSource {
	typ, name := block.Labels[0], block.Labels[1]
	attrs := justAttributes(block.Body)
	ds := DataSource{
		Address: "data." + address(typ, name),
		Type:    typ,
		Name:    name,
		Range:   sourceRange(block.DefRange, sources),
	}
	if attr, ok := attrs["provider"]; ok {
		ds.Provider = providerRefFromExpr(attr.Expr, sources)
	}
	if attr, ok := attrs["depends_on"]; ok {
		ds.DependsOn = referencesFromExpr(attr.Expr, sources)
	}
	if attr, ok := attrs["count"]; ok {
		val := valueFromExpr(attr.Expr, "count", sources)
		ds.Count = &val
		ds.References = append(ds.References, val.References...)
	}
	if attr, ok := attrs["for_each"]; ok {
		val := valueFromExpr(attr.Expr, "for_each", sources)
		ds.ForEach = &val
		ds.References = append(ds.References, val.References...)
	}
	ds.Config = decodeBodyAttributes(block.Body, reservedResourceAttrs(), sources)
	ds.References = append(ds.References, referencesFromAttributes(ds.Config)...)
	return ds
}

func decodeModuleCallBlock(block *hcl.Block, sources map[string]sourceInfo) ModuleCall {
	name := block.Labels[0]
	attrs := justAttributes(block.Body)
	call := ModuleCall{
		Address: "module." + name,
		Name:    name,
		Range:   sourceRange(block.DefRange, sources),
	}
	if attr, ok := attrs["source"]; ok {
		val := valueFromExpr(attr.Expr, "source", sources)
		call.Source = &val
	}
	if attr, ok := attrs["depends_on"]; ok {
		call.DependsOn = referencesFromExpr(attr.Expr, sources)
	}
	if attr, ok := attrs["count"]; ok {
		val := valueFromExpr(attr.Expr, "count", sources)
		call.Count = &val
		call.References = append(call.References, val.References...)
	}
	if attr, ok := attrs["for_each"]; ok {
		val := valueFromExpr(attr.Expr, "for_each", sources)
		call.ForEach = &val
		call.References = append(call.References, val.References...)
	}
	if attr, ok := attrs["providers"]; ok {
		call.ProviderMappings = providerMappingsFromExpr(attr.Expr, sources)
	}
	call.Inputs = decodeBodyAttributes(block.Body, reservedModuleAttrs(), sources)
	call.References = append(call.References, referencesFromAttributes(call.Inputs)...)
	return call
}

func decodeMovedBlock(block *hcl.Block, sources map[string]sourceInfo) MovedBlock {
	attrs := justAttributes(block.Body)
	moved := MovedBlock{Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["from"]; ok {
		moved.From = exprText(attr.Expr, sources)
	}
	if attr, ok := attrs["to"]; ok {
		moved.To = exprText(attr.Expr, sources)
	}
	return moved
}

func decodeImportBlock(block *hcl.Block, sources map[string]sourceInfo) ImportBlock {
	attrs := justAttributes(block.Body)
	imp := ImportBlock{Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["to"]; ok {
		imp.To = exprText(attr.Expr, sources)
	}
	if attr, ok := attrs["id"]; ok {
		val := valueFromExpr(attr.Expr, "id", sources)
		imp.ID = &val
	}
	if attr, ok := attrs["provider"]; ok {
		imp.Provider = providerRefFromExpr(attr.Expr, sources)
	}
	return imp
}

func decodeRemovedBlock(block *hcl.Block, sources map[string]sourceInfo) RemovedBlock {
	attrs := justAttributes(block.Body)
	removed := RemovedBlock{Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["from"]; ok {
		removed.From = exprText(attr.Expr, sources)
	}
	removed.Lifecycle = decodeLifecycle(block.Body, sources)
	return removed
}

func decodeCheckBlock(block *hcl.Block, sources map[string]sourceInfo) CheckBlock {
	check := CheckBlock{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	content, _ := block.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "assert"}},
	})
	for _, assert := range content.Blocks {
		check.Assertions = append(check.Assertions, decodeCheckRule(assert, sources))
	}
	return check
}

func decodeLifecycle(body hcl.Body, sources map[string]sourceInfo) *Lifecycle {
	content, _ := body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "lifecycle"},
		},
	})
	if len(content.Blocks) == 0 {
		return nil
	}
	block := content.Blocks[0]
	attrs := justAttributes(block.Body)
	l := &Lifecycle{Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["create_before_destroy"]; ok {
		val := valueFromExpr(attr.Expr, "create_before_destroy", sources)
		l.CreateBeforeDestroy = &val
	}
	if attr, ok := attrs["prevent_destroy"]; ok {
		val := valueFromExpr(attr.Expr, "prevent_destroy", sources)
		l.PreventDestroy = &val
	}
	if attr, ok := attrs["ignore_changes"]; ok {
		l.IgnoreChanges = append(l.IgnoreChanges, valueFromExpr(attr.Expr, "ignore_changes", sources))
	}
	if attr, ok := attrs["replace_triggered_by"]; ok {
		l.ReplaceTriggeredBy = referencesFromExpr(attr.Expr, sources)
	}
	content, _ = block.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "precondition"},
			{Type: "postcondition"},
		},
	})
	for _, rule := range content.Blocks {
		decoded := decodeCheckRule(rule, sources)
		switch rule.Type {
		case "precondition":
			l.Preconditions = append(l.Preconditions, decoded)
		case "postcondition":
			l.Postconditions = append(l.Postconditions, decoded)
		}
	}
	return l
}

func decodeCheckRule(block *hcl.Block, sources map[string]sourceInfo) CheckRule {
	attrs := justAttributes(block.Body)
	rule := CheckRule{Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["condition"]; ok {
		val := valueFromExpr(attr.Expr, "condition", sources)
		rule.Condition = &val
		rule.References = append(rule.References, val.References...)
	}
	if attr, ok := attrs["error_message"]; ok {
		val := valueFromExpr(attr.Expr, "error_message", sources)
		rule.ErrorMessage = &val
		rule.References = append(rule.References, val.References...)
	}
	return rule
}

func decodeBodyAttributes(body hcl.Body, reserved map[string]bool, sources map[string]sourceInfo) []Attribute {
	attrs := justAttributes(body)
	out := make([]Attribute, 0, len(attrs))
	for name, attr := range attrs {
		if reserved[name] {
			continue
		}
		out = append(out, Attribute{
			Path:  name,
			Value: valueFromExpr(attr.Expr, name, sources),
			Range: sourceRange(attr.Range, sources),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func justAttributes(body hcl.Body) hcl.Attributes {
	attrs, _ := body.JustAttributes()
	return attrs
}

func reservedProviderAttrs() map[string]bool {
	return map[string]bool{"alias": true}
}

func reservedResourceAttrs() map[string]bool {
	return map[string]bool{
		"provider":   true,
		"depends_on": true,
		"count":      true,
		"for_each":   true,
	}
}

func reservedModuleAttrs() map[string]bool {
	return map[string]bool{
		"source":     true,
		"providers":  true,
		"depends_on": true,
		"count":      true,
		"for_each":   true,
	}
}

func providerRefFromExpr(expr hcl.Expression, sources map[string]sourceInfo) *ProviderRef {
	text := exprText(expr, sources)
	parts := strings.Split(text, ".")
	ref := &ProviderRef{
		LocalName: text,
		Address:   providerAddress(text, ""),
		Range:     sourceRange(expr.Range(), sources),
	}
	if len(parts) > 0 {
		ref.LocalName = parts[0]
		ref.Address = providerAddress(parts[0], "")
	}
	if len(parts) > 1 {
		ref.Alias = parts[1]
		ref.Address = providerAddress(parts[0], parts[1])
	}
	return ref
}

func providerMappingsFromExpr(expr hcl.Expression, sources map[string]sourceInfo) []ProviderMapping {
	var mappings []ProviderMapping
	if obj, ok := expr.(*hclsyntax.ObjectConsExpr); ok {
		for _, item := range obj.Items {
			childName := strings.Trim(exprText(item.KeyExpr, sources), `"`)
			provider := exprText(item.ValueExpr, sources)
			mappings = append(mappings, ProviderMapping{
				ChildName: childName,
				Provider:  *providerRefFromText(provider, sourceRange(item.ValueExpr.Range(), sources)),
				Range:     sourceRange(item.KeyExpr.Range(), sources),
			})
		}
		if len(mappings) > 0 {
			return mappings
		}
	}
	value := valueFromExpr(expr, "providers", sources)
	if obj, ok := value.Literal.(map[string]any); ok {
		for child, raw := range obj {
			if provider, ok := raw.(string); ok {
				mappings = append(mappings, ProviderMapping{
					ChildName: child,
					Provider:  *providerRefFromText(provider, sourceRange(expr.Range(), sources)),
					Range:     sourceRange(expr.Range(), sources),
				})
			}
		}
	}
	if len(mappings) > 0 {
		return mappings
	}

	for _, ref := range referencesFromExpr(expr, sources) {
		mappings = append(mappings, ProviderMapping{
			ChildName: ref.Subject,
			Provider:  *providerRefFromText(ref.Traversal, ref.Range),
			Range:     ref.Range,
		})
	}
	return mappings
}

func providerRefFromText(text string, rng *SourceRange) *ProviderRef {
	parts := strings.Split(text, ".")
	ref := &ProviderRef{
		LocalName: text,
		Address:   providerAddress(text, ""),
		Range:     rng,
	}
	if len(parts) > 0 {
		ref.LocalName = parts[0]
		ref.Address = providerAddress(parts[0], "")
	}
	if len(parts) > 1 {
		ref.Alias = parts[1]
		ref.Address = providerAddress(parts[0], parts[1])
	}
	return ref
}

func providerAddress(localName, alias string) string {
	if alias == "" {
		return "provider." + localName
	}
	return "provider." + localName + "." + alias
}

func address(typ, name string) string {
	return typ + "." + name
}

func literalString(expr hcl.Expression, sources map[string]sourceInfo) (string, bool) {
	val := valueFromExpr(expr, "", sources)
	s, ok := val.Literal.(string)
	return s, ok
}

func literalBool(expr hcl.Expression) (bool, bool) {
	val, diags := expr.Value(nil)
	if diags.HasErrors() || !val.IsKnown() || val.IsNull() || val.Type().FriendlyName() != "bool" {
		return false, false
	}
	return val.True(), true
}

func referencesFromAttributes(attrs []Attribute) []Reference {
	var refs []Reference
	for _, attr := range attrs {
		refs = append(refs, attr.Value.References...)
	}
	return refs
}

func upsertProviderRequirement(mod *Module, req ProviderRequirement) {
	for i := range mod.RequiredProviders {
		if mod.RequiredProviders[i].LocalName == req.LocalName {
			mod.RequiredProviders[i] = req
			return
		}
	}
	mod.RequiredProviders = append(mod.RequiredProviders, req)
}

func upsertProviderConfig(mod *Module, cfg ProviderConfig) {
	for i := range mod.ProviderConfigs {
		if mod.ProviderConfigs[i].Address == cfg.Address {
			mod.ProviderConfigs[i] = cfg
			return
		}
	}
	mod.ProviderConfigs = append(mod.ProviderConfigs, cfg)
}

func upsertVariable(mod *Module, v Variable) {
	for i := range mod.Variables {
		if mod.Variables[i].Name == v.Name {
			mod.Variables[i] = v
			return
		}
	}
	mod.Variables = append(mod.Variables, v)
}

func upsertLocal(mod *Module, local Local) {
	for i := range mod.Locals {
		if mod.Locals[i].Name == local.Name {
			mod.Locals[i] = local
			return
		}
	}
	mod.Locals = append(mod.Locals, local)
}

func upsertOutput(mod *Module, o Output) {
	for i := range mod.Outputs {
		if mod.Outputs[i].Name == o.Name {
			mod.Outputs[i] = o
			return
		}
	}
	mod.Outputs = append(mod.Outputs, o)
}

func upsertResource(mod *Module, r Resource) {
	for i := range mod.Resources {
		if mod.Resources[i].Address == r.Address {
			mod.Resources[i] = r
			return
		}
	}
	mod.Resources = append(mod.Resources, r)
}

func upsertDataSource(mod *Module, ds DataSource) {
	for i := range mod.DataSources {
		if mod.DataSources[i].Address == ds.Address {
			mod.DataSources[i] = ds
			return
		}
	}
	mod.DataSources = append(mod.DataSources, ds)
}

func upsertModuleCall(mod *Module, call ModuleCall) {
	for i := range mod.ModuleCalls {
		if mod.ModuleCalls[i].Address == call.Address {
			mod.ModuleCalls[i] = call
			return
		}
	}
	mod.ModuleCalls = append(mod.ModuleCalls, call)
}

func missingAttributeDiag(block *hcl.Block, attr string, sources map[string]sourceInfo) Diagnostic {
	return Diagnostic{
		Severity: DiagnosticError,
		Code:     "missing_attribute",
		Summary:  "Missing required attribute",
		Detail:   fmt.Sprintf("Block %q is missing required attribute %q.", block.Type, attr),
		Range:    sourceRange(block.DefRange, sources),
	}
}
