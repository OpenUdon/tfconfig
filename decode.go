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
	override := file.Role == FileRoleOverride

	for _, block := range content.Blocks {
		switch block.Type {
		case "terraform":
			out = append(out, decodeTerraformBlock(mod, block, sources)...)
		case "provider":
			cfg, cfgDiags := decodeProviderBlock(block, sources)
			out = append(out, cfgDiags...)
			out = append(out, upsertProviderConfig(mod, cfg, override)...)
		case "variable":
			v, vDiags := decodeVariableBlock(block, sources)
			out = append(out, vDiags...)
			out = append(out, upsertVariable(mod, v, override)...)
		case "locals":
			locals, localDiags := decodeLocalsBlock(block, sources)
			out = append(out, localDiags...)
			for _, local := range locals {
				out = append(out, upsertLocal(mod, local, override)...)
			}
		case "output":
			output, outputDiags := decodeOutputBlock(block, sources)
			out = append(out, outputDiags...)
			out = append(out, upsertOutput(mod, output, override)...)
		case "resource":
			res, resDiags := decodeResourceBlock(block, sources)
			out = append(out, resDiags...)
			out = append(out, upsertResource(mod, res, override)...)
		case "data":
			ds, dsDiags := decodeDataSourceBlock(block, sources)
			out = append(out, dsDiags...)
			out = append(out, upsertDataSource(mod, ds, override)...)
		case "module":
			call, callDiags := decodeModuleCallBlock(block, sources)
			out = append(out, callDiags...)
			out = append(out, upsertModuleCall(mod, call, override)...)
		case "moved":
			if override {
				out = append(out, unsupportedOverrideDiag("'moved' blocks", block.DefRange, sources))
				continue
			}
			mod.Moved = append(mod.Moved, decodeMovedBlock(block, sources))
		case "import":
			if override {
				out = append(out, unsupportedOverrideDiag("'import' blocks", block.DefRange, sources))
				continue
			}
			mod.Imports = append(mod.Imports, decodeImportBlock(block, sources))
		case "removed":
			if override {
				out = append(out, unsupportedOverrideDiag("'removed' blocks", block.DefRange, sources))
				continue
			}
			mod.Removed = append(mod.Removed, decodeRemovedBlock(block, sources))
		case "check":
			if override {
				out = append(out, unsupportedOverrideDiag("'check' blocks", block.DefRange, sources))
				continue
			}
			check, checkDiags := decodeCheckBlock(block, sources)
			out = append(out, checkDiags...)
			mod.Checks = append(mod.Checks, check)
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
				out = append(out, upsertProviderRequirement(mod, req)...)
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

func decodeProviderBlock(block *hcl.Block, sources map[string]sourceInfo) (ProviderConfig, []Diagnostic) {
	name := block.Labels[0]
	attrs, diags := attributesWithAllowedBlocks(block.Body, nil, sources)
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
	cfg.Config = attributesFromMap(attrs, reservedProviderAttrs(), sources)
	cfg.References = referencesFromAttributes(cfg.Config)
	return cfg, diags
}

func decodeVariableBlock(block *hcl.Block, sources map[string]sourceInfo) (Variable, []Diagnostic) {
	attrs, diags := attributesWithAllowedBlocks(block.Body, nil, sources)
	v := Variable{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["sensitive"]; ok {
		v.Sensitive, _ = literalBool(attr.Expr)
	}
	if attr, ok := attrs["type"]; ok {
		val := valueFromExpr(attr.Expr, "type", sources)
		val.Kind = ValueKindExpression
		v.Type = &val
	}
	if attr, ok := attrs["default"]; ok {
		val := valueFromExpr(attr.Expr, "default", sources)
		val.Sensitive = v.Sensitive
		v.Default = &val
		v.References = append(v.References, val.References...)
	}
	if attr, ok := attrs["description"]; ok {
		v.Description, _ = literalString(attr.Expr, sources)
	}
	if attr, ok := attrs["nullable"]; ok {
		if b, ok := literalBool(attr.Expr); ok {
			v.Nullable = &b
		}
	}
	return v, diags
}

func decodeLocalsBlock(block *hcl.Block, sources map[string]sourceInfo) ([]Local, []Diagnostic) {
	attrs, diags := attributesWithAllowedBlocks(block.Body, nil, sources)
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
	return locals, diags
}

func decodeOutputBlock(block *hcl.Block, sources map[string]sourceInfo) (Output, []Diagnostic) {
	attrs, diags := attributesWithAllowedBlocks(block.Body, nil, sources)
	o := Output{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["sensitive"]; ok {
		o.Sensitive, _ = literalBool(attr.Expr)
	}
	if attr, ok := attrs["value"]; ok {
		val := valueFromExpr(attr.Expr, "value", sources)
		val.Sensitive = o.Sensitive
		o.Value = &val
		o.References = append(o.References, val.References...)
	}
	if attr, ok := attrs["description"]; ok {
		o.Description, _ = literalString(attr.Expr, sources)
	}
	if attr, ok := attrs["depends_on"]; ok {
		o.DependsOn = referencesFromExpr(attr.Expr, sources)
	}
	return o, diags
}

func decodeResourceBlock(block *hcl.Block, sources map[string]sourceInfo) (Resource, []Diagnostic) {
	typ, name := block.Labels[0], block.Labels[1]
	attrs, diags := attributesWithAllowedBlocks(block.Body, []hcl.BlockHeaderSchema{{Type: "lifecycle"}}, sources)
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
	res.Config = attributesFromMap(attrs, reservedResourceAttrs(), sources)
	res.References = append(res.References, referencesFromAttributes(res.Config)...)
	lifecycle, lifecycleDiags := decodeLifecycle(block.Body, sources)
	diags = append(diags, lifecycleDiags...)
	res.Lifecycle = lifecycle
	return res, diags
}

func decodeDataSourceBlock(block *hcl.Block, sources map[string]sourceInfo) (DataSource, []Diagnostic) {
	typ, name := block.Labels[0], block.Labels[1]
	attrs, diags := attributesWithAllowedBlocks(block.Body, nil, sources)
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
	ds.Config = attributesFromMap(attrs, reservedResourceAttrs(), sources)
	ds.References = append(ds.References, referencesFromAttributes(ds.Config)...)
	return ds, diags
}

func decodeModuleCallBlock(block *hcl.Block, sources map[string]sourceInfo) (ModuleCall, []Diagnostic) {
	name := block.Labels[0]
	attrs, diags := attributesWithAllowedBlocks(block.Body, nil, sources)
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
	call.Inputs = attributesFromMap(attrs, reservedModuleAttrs(), sources)
	call.References = append(call.References, referencesFromAttributes(call.Inputs)...)
	return call, diags
}

func decodeMovedBlock(block *hcl.Block, sources map[string]sourceInfo) MovedBlock {
	attrs, _ := attributesWithAllowedBlocks(block.Body, nil, sources)
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
	attrs, _ := attributesWithAllowedBlocks(block.Body, nil, sources)
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
	attrs, _ := attributesWithAllowedBlocks(block.Body, []hcl.BlockHeaderSchema{{Type: "lifecycle"}}, sources)
	removed := RemovedBlock{Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["from"]; ok {
		removed.From = exprText(attr.Expr, sources)
	}
	lifecycle, _ := decodeLifecycle(block.Body, sources)
	removed.Lifecycle = lifecycle
	return removed
}

func decodeCheckBlock(block *hcl.Block, sources map[string]sourceInfo) (CheckBlock, []Diagnostic) {
	check := CheckBlock{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	content, contentDiags := block.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{{Type: "assert"}},
	})
	diags := convertDiagnostics(contentDiags, "", "", sources)
	for _, assert := range content.Blocks {
		rule, ruleDiags := decodeCheckRule(assert, sources)
		diags = append(diags, ruleDiags...)
		check.Assertions = append(check.Assertions, rule)
	}
	return check, diags
}

func decodeLifecycle(body hcl.Body, sources map[string]sourceInfo) (*Lifecycle, []Diagnostic) {
	content, _, contentDiags := body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "lifecycle"},
		},
	})
	diags := convertDiagnostics(contentDiags, "", "", sources)
	if len(content.Blocks) == 0 {
		return nil, diags
	}
	block := content.Blocks[0]
	attrs, attrDiags := attributesWithAllowedBlocks(block.Body, []hcl.BlockHeaderSchema{{Type: "precondition"}, {Type: "postcondition"}}, sources)
	diags = append(diags, attrDiags...)
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
	content, _, contentDiags = block.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "precondition"},
			{Type: "postcondition"},
		},
	})
	diags = append(diags, convertDiagnostics(contentDiags, "", "", sources)...)
	for _, rule := range content.Blocks {
		decoded, ruleDiags := decodeCheckRule(rule, sources)
		diags = append(diags, ruleDiags...)
		switch rule.Type {
		case "precondition":
			l.Preconditions = append(l.Preconditions, decoded)
		case "postcondition":
			l.Postconditions = append(l.Postconditions, decoded)
		}
	}
	return l, diags
}

func decodeCheckRule(block *hcl.Block, sources map[string]sourceInfo) (CheckRule, []Diagnostic) {
	attrs, diags := attributesWithAllowedBlocks(block.Body, nil, sources)
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
	return rule, diags
}

func decodeBodyAttributes(body hcl.Body, reserved map[string]bool, allowedBlocks []hcl.BlockHeaderSchema, sources map[string]sourceInfo) ([]Attribute, []Diagnostic) {
	attrs, diags := attributesWithAllowedBlocks(body, allowedBlocks, sources)
	return attributesFromMap(attrs, reserved, sources), diags
}

func attributesFromMap(attrs hcl.Attributes, reserved map[string]bool, sources map[string]sourceInfo) []Attribute {
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

func attributesWithAllowedBlocks(body hcl.Body, allowedBlocks []hcl.BlockHeaderSchema, sources map[string]sourceInfo) (hcl.Attributes, []Diagnostic) {
	remaining := body
	var out []Diagnostic
	if len(allowedBlocks) > 0 {
		_, remain, diags := body.PartialContent(&hcl.BodySchema{Blocks: allowedBlocks})
		out = append(out, convertDiagnostics(diags, "", "", sources)...)
		remaining = remain
	}
	attrs, diags := remaining.JustAttributes()
	out = append(out, convertDiagnostics(diags, "", "", sources)...)
	return attrs, out
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

func upsertProviderRequirement(mod *Module, req ProviderRequirement) []Diagnostic {
	for i := range mod.RequiredProviders {
		if mod.RequiredProviders[i].LocalName == req.LocalName {
			return []Diagnostic{duplicateDiag("provider requirement", req.LocalName, req.Range)}
		}
	}
	mod.RequiredProviders = append(mod.RequiredProviders, req)
	return nil
}

func upsertProviderConfig(mod *Module, cfg ProviderConfig, override bool) []Diagnostic {
	for i := range mod.ProviderConfigs {
		if mod.ProviderConfigs[i].Address == cfg.Address {
			if !override {
				return []Diagnostic{duplicateDiag("provider configuration", cfg.Address, cfg.Range)}
			}
			mergeProviderConfig(&mod.ProviderConfigs[i], cfg)
			return nil
		}
	}
	if override {
		return []Diagnostic{missingBaseOverrideDiag("provider configuration", cfg.Address, cfg.Range)}
	}
	mod.ProviderConfigs = append(mod.ProviderConfigs, cfg)
	return nil
}

func upsertVariable(mod *Module, v Variable, override bool) []Diagnostic {
	for i := range mod.Variables {
		if mod.Variables[i].Name == v.Name {
			if !override {
				return []Diagnostic{duplicateDiag("variable", v.Name, v.Range)}
			}
			mergeVariable(&mod.Variables[i], v)
			return nil
		}
	}
	if override {
		return []Diagnostic{missingBaseOverrideDiag("variable", v.Name, v.Range)}
	}
	mod.Variables = append(mod.Variables, v)
	return nil
}

func upsertLocal(mod *Module, local Local, override bool) []Diagnostic {
	for i := range mod.Locals {
		if mod.Locals[i].Name == local.Name {
			if !override {
				return []Diagnostic{duplicateDiag("local value", local.Name, local.Range)}
			}
			mergeLocal(&mod.Locals[i], local)
			return nil
		}
	}
	if override {
		return []Diagnostic{missingBaseOverrideDiag("local value", local.Name, local.Range)}
	}
	mod.Locals = append(mod.Locals, local)
	return nil
}

func upsertOutput(mod *Module, o Output, override bool) []Diagnostic {
	for i := range mod.Outputs {
		if mod.Outputs[i].Name == o.Name {
			if !override {
				return []Diagnostic{duplicateDiag("output", o.Name, o.Range)}
			}
			mergeOutput(&mod.Outputs[i], o)
			return nil
		}
	}
	if override {
		return []Diagnostic{missingBaseOverrideDiag("output", o.Name, o.Range)}
	}
	mod.Outputs = append(mod.Outputs, o)
	return nil
}

func upsertResource(mod *Module, r Resource, override bool) []Diagnostic {
	for i := range mod.Resources {
		if mod.Resources[i].Address == r.Address {
			if !override {
				return []Diagnostic{duplicateDiag("resource", r.Address, r.Range)}
			}
			mergeResource(&mod.Resources[i], r)
			return nil
		}
	}
	if override {
		return []Diagnostic{missingBaseOverrideDiag("resource", r.Address, r.Range)}
	}
	mod.Resources = append(mod.Resources, r)
	return nil
}

func upsertDataSource(mod *Module, ds DataSource, override bool) []Diagnostic {
	for i := range mod.DataSources {
		if mod.DataSources[i].Address == ds.Address {
			if !override {
				return []Diagnostic{duplicateDiag("data source", ds.Address, ds.Range)}
			}
			mergeDataSource(&mod.DataSources[i], ds)
			return nil
		}
	}
	if override {
		return []Diagnostic{missingBaseOverrideDiag("data source", ds.Address, ds.Range)}
	}
	mod.DataSources = append(mod.DataSources, ds)
	return nil
}

func upsertModuleCall(mod *Module, call ModuleCall, override bool) []Diagnostic {
	for i := range mod.ModuleCalls {
		if mod.ModuleCalls[i].Address == call.Address {
			if !override {
				return []Diagnostic{duplicateDiag("module call", call.Address, call.Range)}
			}
			mergeModuleCall(&mod.ModuleCalls[i], call)
			return nil
		}
	}
	if override {
		return []Diagnostic{missingBaseOverrideDiag("module call", call.Address, call.Range)}
	}
	mod.ModuleCalls = append(mod.ModuleCalls, call)
	return nil
}

func duplicateDiag(kind, name string, rng *SourceRange) Diagnostic {
	return modelDiagnostic(DiagnosticError, "duplicate_declaration", "Duplicate declaration", fmt.Sprintf("Duplicate %s %q was ignored.", kind, name), "", name, rng)
}

func missingBaseOverrideDiag(kind, name string, rng *SourceRange) Diagnostic {
	return modelDiagnostic(DiagnosticError, "missing_base_override", "Missing base declaration for override", fmt.Sprintf("There is no %s %q for this override file declaration to override.", kind, name), "", name, rng)
}

func unsupportedOverrideDiag(kind string, rng hcl.Range, sources map[string]sourceInfo) Diagnostic {
	return modelDiagnostic(DiagnosticError, "unsupported_override", "Unsupported override", fmt.Sprintf("Override files cannot override %s.", kind), "", "", sourceRange(rng, sources))
}

func mergeProviderConfig(dst *ProviderConfig, src ProviderConfig) {
	if src.Alias != "" {
		dst.Alias = src.Alias
	}
	dst.Config = mergeAttributes(dst.Config, src.Config)
	dst.References = referencesFromAttributes(dst.Config)
	dst.Range = src.Range
}

func mergeVariable(dst *Variable, src Variable) {
	if src.Type != nil {
		dst.Type = src.Type
	}
	if src.Default != nil {
		dst.Default = src.Default
	}
	if src.Description != "" {
		dst.Description = src.Description
	}
	if src.Sensitive {
		dst.Sensitive = true
		if dst.Default != nil {
			dst.Default.Sensitive = true
		}
	}
	if src.Nullable != nil {
		dst.Nullable = src.Nullable
	}
	dst.References = src.References
	dst.Range = src.Range
}

func mergeLocal(dst *Local, src Local) {
	if src.Value != nil {
		dst.Value = src.Value
	}
	dst.References = src.References
	dst.Range = src.Range
}

func mergeOutput(dst *Output, src Output) {
	if src.Value != nil {
		dst.Value = src.Value
	}
	if src.Description != "" {
		dst.Description = src.Description
	}
	if src.Sensitive {
		dst.Sensitive = true
		if dst.Value != nil {
			dst.Value.Sensitive = true
		}
	}
	if len(src.DependsOn) > 0 {
		dst.DependsOn = src.DependsOn
	}
	dst.References = src.References
	dst.Range = src.Range
}

func mergeResource(dst *Resource, src Resource) {
	if src.Provider != nil {
		dst.Provider = src.Provider
	}
	dst.Config = mergeAttributes(dst.Config, src.Config)
	if src.Lifecycle != nil {
		dst.Lifecycle = src.Lifecycle
	}
	if len(src.DependsOn) > 0 {
		dst.DependsOn = src.DependsOn
	}
	if src.Count != nil {
		dst.Count = src.Count
	}
	if src.ForEach != nil {
		dst.ForEach = src.ForEach
	}
	dst.References = append(referencesFromAttributes(dst.Config), valueRefs(dst.Count)...)
	dst.References = append(dst.References, valueRefs(dst.ForEach)...)
	dst.Range = src.Range
}

func mergeDataSource(dst *DataSource, src DataSource) {
	if src.Provider != nil {
		dst.Provider = src.Provider
	}
	dst.Config = mergeAttributes(dst.Config, src.Config)
	if len(src.DependsOn) > 0 {
		dst.DependsOn = src.DependsOn
	}
	if src.Count != nil {
		dst.Count = src.Count
	}
	if src.ForEach != nil {
		dst.ForEach = src.ForEach
	}
	dst.References = append(referencesFromAttributes(dst.Config), valueRefs(dst.Count)...)
	dst.References = append(dst.References, valueRefs(dst.ForEach)...)
	dst.Range = src.Range
}

func mergeModuleCall(dst *ModuleCall, src ModuleCall) {
	if src.Source != nil {
		dst.Source = src.Source
	}
	dst.Inputs = mergeAttributes(dst.Inputs, src.Inputs)
	if len(src.ProviderMappings) > 0 {
		dst.ProviderMappings = src.ProviderMappings
	}
	if len(src.DependsOn) > 0 {
		dst.DependsOn = src.DependsOn
	}
	if src.Count != nil {
		dst.Count = src.Count
	}
	if src.ForEach != nil {
		dst.ForEach = src.ForEach
	}
	dst.References = append(referencesFromAttributes(dst.Inputs), valueRefs(dst.Count)...)
	dst.References = append(dst.References, valueRefs(dst.ForEach)...)
	dst.Range = src.Range
}

func mergeAttributes(dst, src []Attribute) []Attribute {
	for _, incoming := range src {
		replaced := false
		for i := range dst {
			if dst[i].Path == incoming.Path {
				dst[i] = incoming
				replaced = true
				break
			}
		}
		if !replaced {
			dst = append(dst, incoming)
		}
	}
	sort.Slice(dst, func(i, j int) bool { return dst[i].Path < dst[j].Path })
	return dst
}

func valueRefs(v *Value) []Reference {
	if v == nil {
		return nil
	}
	return v.References
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
