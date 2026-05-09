package tfconfig

func cloneDocument(d Document) Document {
	c := d
	c.SourceRoots = cloneSlice(d.SourceRoots)
	c.SourceFiles = cloneSlice(d.SourceFiles)
	c.Modules = cloneModules(d.Modules)
	c.Diagnostics = cloneSlice(d.Diagnostics)
	return c
}

func cloneModules(in []Module) []Module {
	out := cloneSlice(in)
	for i := range out {
		out[i].SourceFiles = cloneSlice(out[i].SourceFiles)
		out[i].RequiredProviders = cloneSlice(out[i].RequiredProviders)
		out[i].ProviderConfigs = cloneProviderConfigs(out[i].ProviderConfigs)
		out[i].Variables = cloneVariables(out[i].Variables)
		out[i].Locals = cloneLocals(out[i].Locals)
		out[i].Outputs = cloneOutputs(out[i].Outputs)
		out[i].Resources = cloneResources(out[i].Resources)
		out[i].DataSources = cloneDataSources(out[i].DataSources)
		out[i].ModuleCalls = cloneModuleCalls(out[i].ModuleCalls)
		out[i].Moved = cloneSlice(out[i].Moved)
		out[i].Imports = cloneImports(out[i].Imports)
		out[i].Removed = cloneRemoved(out[i].Removed)
		out[i].Checks = cloneChecks(out[i].Checks)
		out[i].Tests = cloneTests(out[i].Tests)
		out[i].Diagnostics = cloneSlice(out[i].Diagnostics)
	}
	return out
}

func cloneProviderConfigs(in []ProviderConfig) []ProviderConfig {
	out := cloneSlice(in)
	for i := range out {
		out[i].Config = cloneSlice(out[i].Config)
		out[i].References = cloneSlice(out[i].References)
	}
	return out
}

func cloneVariables(in []Variable) []Variable {
	out := cloneSlice(in)
	for i := range out {
		out[i].References = cloneSlice(out[i].References)
	}
	return out
}

func cloneLocals(in []Local) []Local {
	out := cloneSlice(in)
	for i := range out {
		out[i].References = cloneSlice(out[i].References)
	}
	return out
}

func cloneOutputs(in []Output) []Output {
	out := cloneSlice(in)
	for i := range out {
		out[i].DependsOn = cloneSlice(out[i].DependsOn)
		out[i].References = cloneSlice(out[i].References)
	}
	return out
}

func cloneResources(in []Resource) []Resource {
	out := cloneSlice(in)
	for i := range out {
		out[i].Config = cloneSlice(out[i].Config)
		out[i].DependsOn = cloneSlice(out[i].DependsOn)
		out[i].References = cloneSlice(out[i].References)
		out[i].Lifecycle = cloneLifecycle(out[i].Lifecycle)
	}
	return out
}

func cloneDataSources(in []DataSource) []DataSource {
	out := cloneSlice(in)
	for i := range out {
		out[i].Config = cloneSlice(out[i].Config)
		out[i].DependsOn = cloneSlice(out[i].DependsOn)
		out[i].References = cloneSlice(out[i].References)
	}
	return out
}

func cloneModuleCalls(in []ModuleCall) []ModuleCall {
	out := cloneSlice(in)
	for i := range out {
		out[i].Inputs = cloneSlice(out[i].Inputs)
		out[i].ProviderMappings = cloneSlice(out[i].ProviderMappings)
		out[i].DependsOn = cloneSlice(out[i].DependsOn)
		out[i].References = cloneSlice(out[i].References)
	}
	return out
}

func cloneImports(in []ImportBlock) []ImportBlock {
	return cloneSlice(in)
}

func cloneRemoved(in []RemovedBlock) []RemovedBlock {
	out := cloneSlice(in)
	for i := range out {
		out[i].Lifecycle = cloneLifecycle(out[i].Lifecycle)
	}
	return out
}

func cloneChecks(in []CheckBlock) []CheckBlock {
	out := cloneSlice(in)
	for i := range out {
		out[i].Assertions = cloneCheckRules(out[i].Assertions)
	}
	return out
}

func cloneTests(in []TestFile) []TestFile {
	out := cloneSlice(in)
	for i := range out {
		out[i].Runs = cloneSlice(out[i].Runs)
		for j := range out[i].Runs {
			out[i].Runs[j].Variables = cloneSlice(out[i].Runs[j].Variables)
			out[i].Runs[j].Assertions = cloneCheckRules(out[i].Runs[j].Assertions)
		}
	}
	return out
}

func cloneLifecycle(in *Lifecycle) *Lifecycle {
	if in == nil {
		return nil
	}
	out := *in
	out.IgnoreChanges = cloneSlice(in.IgnoreChanges)
	out.ReplaceTriggeredBy = cloneSlice(in.ReplaceTriggeredBy)
	out.Preconditions = cloneCheckRules(in.Preconditions)
	out.Postconditions = cloneCheckRules(in.Postconditions)
	out.Diagnostics = cloneSlice(in.Diagnostics)
	return &out
}

func cloneCheckRules(in []CheckRule) []CheckRule {
	out := cloneSlice(in)
	for i := range out {
		out[i].References = cloneSlice(out[i].References)
	}
	return out
}

func cloneSlice[T any](in []T) []T {
	if len(in) == 0 {
		return nil
	}
	out := make([]T, len(in))
	copy(out, in)
	return out
}
