package tfconfig

func cloneDocument(d Document) Document {
	c := d
	c.SourceRoots = cloneSlice(d.SourceRoots)
	c.SourceFiles = cloneSlice(d.SourceFiles)
	c.Modules = cloneModules(d.Modules)
	c.Diagnostics = cloneDiagnostics(d.Diagnostics)
	return c
}

func cloneModules(in []Module) []Module {
	out := cloneSlice(in)
	for i := range out {
		out[i].Source = cloneValuePtr(out[i].Source)
		out[i].SourceFiles = cloneSlice(out[i].SourceFiles)
		out[i].RequiredVersions = cloneValues(out[i].RequiredVersions)
		out[i].RequiredProviders = cloneProviderRequirements(out[i].RequiredProviders)
		out[i].ProviderConfigs = cloneProviderConfigs(out[i].ProviderConfigs)
		out[i].Variables = cloneVariables(out[i].Variables)
		out[i].Locals = cloneLocals(out[i].Locals)
		out[i].Outputs = cloneOutputs(out[i].Outputs)
		out[i].Resources = cloneResources(out[i].Resources)
		out[i].DataSources = cloneDataSources(out[i].DataSources)
		out[i].ModuleCalls = cloneModuleCalls(out[i].ModuleCalls)
		out[i].Moved = cloneMoved(out[i].Moved)
		out[i].Imports = cloneImports(out[i].Imports)
		out[i].Removed = cloneRemoved(out[i].Removed)
		out[i].Checks = cloneChecks(out[i].Checks)
		out[i].Tests = cloneTests(out[i].Tests)
		out[i].Diagnostics = cloneDiagnostics(out[i].Diagnostics)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneProviderRequirements(in []ProviderRequirement) []ProviderRequirement {
	out := cloneSlice(in)
	for i := range out {
		out[i].VersionConstraints = cloneSlice(out[i].VersionConstraints)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneProviderConfigs(in []ProviderConfig) []ProviderConfig {
	out := cloneSlice(in)
	for i := range out {
		out[i].Config = cloneAttributes(out[i].Config)
		out[i].References = cloneReferences(out[i].References)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneVariables(in []Variable) []Variable {
	out := cloneSlice(in)
	for i := range out {
		out[i].Type = cloneValuePtr(out[i].Type)
		out[i].Default = cloneValuePtr(out[i].Default)
		out[i].Nullable = cloneBoolPtr(out[i].Nullable)
		out[i].References = cloneReferences(out[i].References)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneLocals(in []Local) []Local {
	out := cloneSlice(in)
	for i := range out {
		out[i].Value = cloneValuePtr(out[i].Value)
		out[i].References = cloneReferences(out[i].References)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneOutputs(in []Output) []Output {
	out := cloneSlice(in)
	for i := range out {
		out[i].Value = cloneValuePtr(out[i].Value)
		out[i].DependsOn = cloneReferences(out[i].DependsOn)
		out[i].References = cloneReferences(out[i].References)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneResources(in []Resource) []Resource {
	out := cloneSlice(in)
	for i := range out {
		out[i].Provider = cloneProviderRef(out[i].Provider)
		out[i].Config = cloneAttributes(out[i].Config)
		out[i].DependsOn = cloneReferences(out[i].DependsOn)
		out[i].Count = cloneValuePtr(out[i].Count)
		out[i].ForEach = cloneValuePtr(out[i].ForEach)
		out[i].References = cloneReferences(out[i].References)
		out[i].Lifecycle = cloneLifecycle(out[i].Lifecycle)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneDataSources(in []DataSource) []DataSource {
	out := cloneSlice(in)
	for i := range out {
		out[i].Provider = cloneProviderRef(out[i].Provider)
		out[i].Config = cloneAttributes(out[i].Config)
		out[i].DependsOn = cloneReferences(out[i].DependsOn)
		out[i].Count = cloneValuePtr(out[i].Count)
		out[i].ForEach = cloneValuePtr(out[i].ForEach)
		out[i].References = cloneReferences(out[i].References)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneModuleCalls(in []ModuleCall) []ModuleCall {
	out := cloneSlice(in)
	for i := range out {
		out[i].Source = cloneValuePtr(out[i].Source)
		out[i].Inputs = cloneAttributes(out[i].Inputs)
		out[i].ProviderMappings = cloneProviderMappings(out[i].ProviderMappings)
		out[i].DependsOn = cloneReferences(out[i].DependsOn)
		out[i].Count = cloneValuePtr(out[i].Count)
		out[i].ForEach = cloneValuePtr(out[i].ForEach)
		out[i].References = cloneReferences(out[i].References)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneMoved(in []MovedBlock) []MovedBlock {
	out := cloneSlice(in)
	for i := range out {
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneImports(in []ImportBlock) []ImportBlock {
	out := cloneSlice(in)
	for i := range out {
		out[i].ID = cloneValuePtr(out[i].ID)
		out[i].Provider = cloneProviderRef(out[i].Provider)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneRemoved(in []RemovedBlock) []RemovedBlock {
	out := cloneSlice(in)
	for i := range out {
		out[i].Lifecycle = cloneLifecycle(out[i].Lifecycle)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneChecks(in []CheckBlock) []CheckBlock {
	out := cloneSlice(in)
	for i := range out {
		out[i].Assertions = cloneCheckRules(out[i].Assertions)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneTests(in []TestFile) []TestFile {
	out := cloneSlice(in)
	for i := range out {
		out[i].Runs = cloneSlice(out[i].Runs)
		out[i].Range = cloneSourceRange(out[i].Range)
		for j := range out[i].Runs {
			out[i].Runs[j].Variables = cloneAttributes(out[i].Runs[j].Variables)
			out[i].Runs[j].Assertions = cloneCheckRules(out[i].Runs[j].Assertions)
			out[i].Runs[j].Range = cloneSourceRange(out[i].Runs[j].Range)
		}
	}
	return out
}

func cloneLifecycle(in *Lifecycle) *Lifecycle {
	if in == nil {
		return nil
	}
	out := *in
	out.CreateBeforeDestroy = cloneValuePtr(in.CreateBeforeDestroy)
	out.PreventDestroy = cloneValuePtr(in.PreventDestroy)
	out.IgnoreChanges = cloneValues(in.IgnoreChanges)
	out.ReplaceTriggeredBy = cloneReferences(in.ReplaceTriggeredBy)
	out.Preconditions = cloneCheckRules(in.Preconditions)
	out.Postconditions = cloneCheckRules(in.Postconditions)
	out.Diagnostics = cloneDiagnostics(in.Diagnostics)
	out.Range = cloneSourceRange(in.Range)
	return &out
}

func cloneCheckRules(in []CheckRule) []CheckRule {
	out := cloneSlice(in)
	for i := range out {
		out[i].Condition = cloneValuePtr(out[i].Condition)
		out[i].ErrorMessage = cloneValuePtr(out[i].ErrorMessage)
		out[i].References = cloneReferences(out[i].References)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneAttributes(in []Attribute) []Attribute {
	out := cloneSlice(in)
	for i := range out {
		out[i].Value = cloneValue(out[i].Value)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneProviderMappings(in []ProviderMapping) []ProviderMapping {
	out := cloneSlice(in)
	for i := range out {
		out[i].Provider = cloneProviderRefValue(out[i].Provider)
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneProviderRef(in *ProviderRef) *ProviderRef {
	if in == nil {
		return nil
	}
	out := cloneProviderRefValue(*in)
	return &out
}

func cloneProviderRefValue(in ProviderRef) ProviderRef {
	in.Range = cloneSourceRange(in.Range)
	return in
}

func cloneValues(in []Value) []Value {
	out := cloneSlice(in)
	for i := range out {
		out[i] = cloneValue(out[i])
	}
	return out
}

func cloneValuePtr(in *Value) *Value {
	if in == nil {
		return nil
	}
	out := cloneValue(*in)
	return &out
}

func cloneValue(in Value) Value {
	in.Literal = cloneLiteral(in.Literal)
	in.References = cloneReferences(in.References)
	in.SensitiveCandidate = cloneSensitiveCandidate(in.SensitiveCandidate)
	in.Range = cloneSourceRange(in.Range)
	return in
}

func cloneLiteral(in any) any {
	switch v := in.(type) {
	case []any:
		out := make([]any, len(v))
		for i := range v {
			out[i] = cloneLiteral(v[i])
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			out[k] = cloneLiteral(item)
		}
		return out
	default:
		return in
	}
}

func cloneSensitiveCandidate(in *SensitiveCandidate) *SensitiveCandidate {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneReferences(in []Reference) []Reference {
	out := cloneSlice(in)
	for i := range out {
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneDiagnostics(in []Diagnostic) []Diagnostic {
	out := cloneSlice(in)
	for i := range out {
		out[i].Range = cloneSourceRange(out[i].Range)
	}
	return out
}

func cloneSourceRange(in *SourceRange) *SourceRange {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneBoolPtr(in *bool) *bool {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneSlice[T any](in []T) []T {
	if len(in) == 0 {
		return nil
	}
	out := make([]T, len(in))
	copy(out, in)
	return out
}
