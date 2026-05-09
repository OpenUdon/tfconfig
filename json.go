package tfconfig

import (
	"encoding/json"
	"sort"
)

// NewDocument returns a static v1 document with deterministic default metadata.
func NewDocument(rootDir string) Document {
	return Document{
		Version:  StaticV1,
		Producer: DefaultProducer,
		RootDir:  rootDir,
	}
}

// JSON returns the deterministic public JSON projection for d.
func (d Document) JSON() ([]byte, error) {
	return json.Marshal(d.Canonical())
}

// JSONIndent returns the deterministic public JSON projection for d with
// indentation suitable for review fixtures and CLI output.
func (d Document) JSONIndent(prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(d.Canonical(), prefix, indent)
}

// MarshalJSON implements the deterministic public JSON projection. It sorts
// arrays and applies safe value projection before encoding.
func (d Document) MarshalJSON() ([]byte, error) {
	type document Document
	c := d.Canonical()
	return json.Marshal(document(c))
}

// Canonical returns a deep copy sorted into deterministic review order.
func (d Document) Canonical() Document {
	c := cloneDocument(d)
	if c.Version == "" {
		c.Version = StaticV1
	}
	if c.Producer == "" {
		c.Producer = DefaultProducer
	}

	sort.Slice(c.SourceRoots, func(i, j int) bool {
		return tupleLess(
			[]string{c.SourceRoots[i].ID, c.SourceRoots[i].ModuleAddress, c.SourceRoots[i].Dir},
			[]string{c.SourceRoots[j].ID, c.SourceRoots[j].ModuleAddress, c.SourceRoots[j].Dir},
		)
	})
	sort.Slice(c.SourceFiles, func(i, j int) bool {
		return tupleLess(
			[]string{c.SourceFiles[i].ModuleAddress, c.SourceFiles[i].Path, c.SourceFiles[i].ID},
			[]string{c.SourceFiles[j].ModuleAddress, c.SourceFiles[j].Path, c.SourceFiles[j].ID},
		)
	})
	sort.Slice(c.Modules, func(i, j int) bool {
		return c.Modules[i].Address < c.Modules[j].Address
	})
	for i := range c.Modules {
		c.Modules[i].Normalize()
	}
	sortDiagnostics(c.Diagnostics)
	return c
}

// Normalize sorts module slices into deterministic review order.
func (m *Module) Normalize() {
	if m == nil {
		return
	}
	sort.Slice(m.RequiredVersions, func(i, j int) bool { return valueKey(m.RequiredVersions[i]) < valueKey(m.RequiredVersions[j]) })
	sort.Slice(m.RequiredProviders, func(i, j int) bool {
		return m.RequiredProviders[i].LocalName < m.RequiredProviders[j].LocalName
	})
	sort.Slice(m.Backends, func(i, j int) bool { return m.Backends[i].Type < m.Backends[j].Type })
	for i := range m.Backends {
		sortAttributes(m.Backends[i].Config)
		sortReferences(m.Backends[i].References)
	}
	if m.Cloud != nil {
		sortAttributes(m.Cloud.Config)
		sortReferences(m.Cloud.References)
	}
	sort.Slice(m.ProviderMetas, func(i, j int) bool { return m.ProviderMetas[i].Provider < m.ProviderMetas[j].Provider })
	for i := range m.ProviderMetas {
		sortAttributes(m.ProviderMetas[i].Config)
		sortReferences(m.ProviderMetas[i].References)
	}
	sort.Slice(m.ProviderConfigs, func(i, j int) bool {
		return m.ProviderConfigs[i].Address < m.ProviderConfigs[j].Address
	})
	for i := range m.ProviderConfigs {
		sortAttributes(m.ProviderConfigs[i].Config)
		sortReferences(m.ProviderConfigs[i].References)
	}
	sort.Slice(m.Variables, func(i, j int) bool { return m.Variables[i].Name < m.Variables[j].Name })
	for i := range m.Variables {
		sortReferences(m.Variables[i].References)
	}
	sort.Slice(m.Locals, func(i, j int) bool { return m.Locals[i].Name < m.Locals[j].Name })
	for i := range m.Locals {
		sortReferences(m.Locals[i].References)
	}
	sort.Slice(m.Outputs, func(i, j int) bool { return m.Outputs[i].Name < m.Outputs[j].Name })
	for i := range m.Outputs {
		sortReferences(m.Outputs[i].DependsOn)
		sortReferences(m.Outputs[i].References)
	}
	sort.Slice(m.Resources, func(i, j int) bool { return m.Resources[i].Address < m.Resources[j].Address })
	for i := range m.Resources {
		sortAttributes(m.Resources[i].Config)
		sortReferences(m.Resources[i].DependsOn)
		sortReferences(m.Resources[i].References)
		normalizeLifecycle(m.Resources[i].Lifecycle)
	}
	sort.Slice(m.DataSources, func(i, j int) bool { return m.DataSources[i].Address < m.DataSources[j].Address })
	for i := range m.DataSources {
		sortAttributes(m.DataSources[i].Config)
		sortReferences(m.DataSources[i].DependsOn)
		sortReferences(m.DataSources[i].References)
	}
	sort.Slice(m.EphemeralResources, func(i, j int) bool { return m.EphemeralResources[i].Address < m.EphemeralResources[j].Address })
	for i := range m.EphemeralResources {
		sortAttributes(m.EphemeralResources[i].Config)
		sortReferences(m.EphemeralResources[i].DependsOn)
		sortReferences(m.EphemeralResources[i].References)
	}
	sort.Slice(m.ModuleCalls, func(i, j int) bool { return m.ModuleCalls[i].Address < m.ModuleCalls[j].Address })
	for i := range m.ModuleCalls {
		sortAttributes(m.ModuleCalls[i].Inputs)
		sortProviderMappings(m.ModuleCalls[i].ProviderMappings)
		sortReferences(m.ModuleCalls[i].DependsOn)
		sortReferences(m.ModuleCalls[i].References)
	}
	sort.Slice(m.Moved, func(i, j int) bool {
		return tupleLess([]string{m.Moved[i].From, m.Moved[i].To}, []string{m.Moved[j].From, m.Moved[j].To})
	})
	sort.Slice(m.Imports, func(i, j int) bool { return m.Imports[i].To < m.Imports[j].To })
	sort.Slice(m.Removed, func(i, j int) bool { return m.Removed[i].From < m.Removed[j].From })
	for i := range m.Removed {
		normalizeLifecycle(m.Removed[i].Lifecycle)
	}
	sort.Slice(m.Checks, func(i, j int) bool { return m.Checks[i].Name < m.Checks[j].Name })
	for i := range m.Checks {
		normalizeCheckRules(m.Checks[i].Assertions)
	}
	sort.Slice(m.Tests, func(i, j int) bool { return m.Tests[i].Path < m.Tests[j].Path })
	for i := range m.Tests {
		sortAttributes(m.Tests[i].Variables)
		sort.Slice(m.Tests[i].Runs, func(a, b int) bool { return m.Tests[i].Runs[a].Name < m.Tests[i].Runs[b].Name })
		for j := range m.Tests[i].Runs {
			if m.Tests[i].Runs[j].Module != nil {
				sortAttributes(m.Tests[i].Runs[j].Module.Config)
				sortReferences(m.Tests[i].Runs[j].Module.References)
			}
			sortAttributes(m.Tests[i].Runs[j].Variables)
			sortAttributes(m.Tests[i].Runs[j].PlanOptions)
			normalizeCheckRules(m.Tests[i].Runs[j].Assertions)
		}
	}
	sortDiagnostics(m.Diagnostics)
}

// MarshalJSON redacts likely-secret literals from the public JSON projection.
func (v Value) MarshalJSON() ([]byte, error) {
	type value Value
	c := v
	c.References = cloneSlice(v.References)
	sortReferences(c.References)
	if c.Redacted || c.Sensitive || c.SensitiveCandidate != nil {
		c.Kind = ValueKindRedacted
		c.Literal = nil
		c.Redacted = true
	}
	return json.Marshal(value(c))
}

func normalizeLifecycle(l *Lifecycle) {
	if l == nil {
		return
	}
	sort.Slice(l.IgnoreChanges, func(i, j int) bool { return valueKey(l.IgnoreChanges[i]) < valueKey(l.IgnoreChanges[j]) })
	sortReferences(l.ReplaceTriggeredBy)
	normalizeCheckRules(l.Preconditions)
	normalizeCheckRules(l.Postconditions)
	sortDiagnostics(l.Diagnostics)
}

func normalizeCheckRules(rules []CheckRule) {
	sort.SliceStable(rules, func(i, j int) bool {
		return tupleLess(
			[]string{valuePtrKey(rules[i].Condition), valuePtrKey(rules[i].ErrorMessage)},
			[]string{valuePtrKey(rules[j].Condition), valuePtrKey(rules[j].ErrorMessage)},
		)
	})
	for i := range rules {
		sortReferences(rules[i].References)
	}
}

func sortAttributes(attrs []Attribute) {
	sort.Slice(attrs, func(i, j int) bool { return attrs[i].Path < attrs[j].Path })
}

func sortProviderMappings(mappings []ProviderMapping) {
	sort.Slice(mappings, func(i, j int) bool {
		return tupleLess(
			[]string{mappings[i].ChildName, mappings[i].Provider.Address, mappings[i].Provider.LocalName, mappings[i].Provider.Alias},
			[]string{mappings[j].ChildName, mappings[j].Provider.Address, mappings[j].Provider.LocalName, mappings[j].Provider.Alias},
		)
	})
}

func sortReferences(refs []Reference) {
	sort.Slice(refs, func(i, j int) bool {
		return tupleLess(
			[]string{refs[i].Traversal, refs[i].Subject},
			[]string{refs[j].Traversal, refs[j].Subject},
		)
	})
}

func sortDiagnostics(diags []Diagnostic) {
	sort.Slice(diags, func(i, j int) bool {
		return tupleLess(
			[]string{string(diags[i].Severity), diags[i].Code, diags[i].ModuleAddress, diags[i].Address, diags[i].Summary},
			[]string{string(diags[j].Severity), diags[j].Code, diags[j].ModuleAddress, diags[j].Address, diags[j].Summary},
		)
	})
}

func tupleLess(a, b []string) bool {
	for i := range a {
		if i >= len(b) {
			return false
		}
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return len(a) < len(b)
}

func valuePtrKey(v *Value) string {
	if v == nil {
		return ""
	}
	return valueKey(*v)
}

func valueKey(v Value) string {
	if v.Expression != "" {
		return string(v.Kind) + ":" + v.Expression
	}
	if v.UnknownReason != "" {
		return string(v.Kind) + ":" + v.UnknownReason
	}
	if v.SensitiveCandidate != nil {
		return string(ValueKindRedacted) + ":" + v.SensitiveCandidate.AttributePath + ":" + v.SensitiveCandidate.Reason
	}
	b, err := json.Marshal(v.Literal)
	if err != nil {
		return string(v.Kind)
	}
	return string(v.Kind) + ":" + string(b)
}
