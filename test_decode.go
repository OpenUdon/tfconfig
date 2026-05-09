package tfconfig

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

var testFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "variables"},
		{Type: "run", LabelNames: []string{"name"}},
	},
}

func decodeTestFile(file discoveredFile, body hcl.Body, sources map[string]sourceInfo) ([]TestFile, []Diagnostic) {
	tf := TestFile{Path: file.RelPath}
	content, contentDiags := body.Content(testFileSchema)
	var diags []Diagnostic
	if syntaxBody, ok := body.(*hclsyntax.Body); ok {
		diags = unexpectedTestFileBlockDiagnostics(syntaxBody.Blocks, sources)
	} else {
		diags = convertDiagnostics(contentDiags, "", "", sources)
	}
	for _, block := range content.Blocks {
		switch block.Type {
		case "variables":
			vars, varDiags := decodeBodyAttributes(block.Body, nil, nil, sources)
			diags = append(diags, varDiags...)
			tf.Variables = append(tf.Variables, vars...)
		case "run":
			run, runDiags := decodeTestRun(block, sources)
			diags = append(diags, runDiags...)
			tf.Runs = append(tf.Runs, run)
		}
	}
	if len(tf.Runs) == 0 && len(tf.Variables) == 0 {
		return nil, diags
	}
	return []TestFile{tf}, diags
}

func unexpectedTestFileBlockDiagnostics(blocks hclsyntax.Blocks, sources map[string]sourceInfo) []Diagnostic {
	var out []Diagnostic
	for _, block := range blocks {
		switch block.Type {
		case "variables", "run":
			continue
		}
		out = append(out, modelDiagnostic(
			DiagnosticError,
			"unexpected_"+strings.ToLower(block.Type)+"_block",
			fmt.Sprintf("Unexpected %q block", block.Type),
			"Blocks are not allowed here.",
			"",
			"",
			sourceRange(block.TypeRange, sources),
		))
	}
	return out
}

func decodeTestRun(block *hcl.Block, sources map[string]sourceInfo) (TestRun, []Diagnostic) {
	allowedBlocks := []hcl.BlockHeaderSchema{{Type: "module"}, {Type: "variables"}, {Type: "plan_options"}, {Type: "assert"}}
	attrs, diags := testRunAttributes(block.Body, allowedBlocks, sources)
	run := TestRun{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["command"]; ok {
		run.Command = strings.Trim(exprText(attr.Expr, sources), `"`)
	}
	content, _, contentDiags := block.Body.PartialContent(&hcl.BodySchema{
		Blocks: allowedBlocks,
	})
	if _, ok := block.Body.(*hclsyntax.Body); !ok {
		diags = append(diags, convertDiagnostics(contentDiags, "", "", sources)...)
	}
	for _, inner := range content.Blocks {
		switch inner.Type {
		case "module":
			module, moduleDiags := decodeTestModule(inner, sources)
			diags = append(diags, moduleDiags...)
			run.Module = module
		case "variables":
			vars, varDiags := decodeBodyAttributes(inner.Body, nil, nil, sources)
			diags = append(diags, varDiags...)
			run.Variables = append(run.Variables, vars...)
		case "plan_options":
			options, optionDiags := decodeBodyAttributes(inner.Body, nil, nil, sources)
			diags = append(diags, optionDiags...)
			run.PlanOptions = append(run.PlanOptions, options...)
		case "assert":
			rule, ruleDiags := decodeCheckRule(inner, sources)
			diags = append(diags, ruleDiags...)
			run.Assertions = append(run.Assertions, rule)
		}
	}
	return run, diags
}

func testRunAttributes(body hcl.Body, allowedBlocks []hcl.BlockHeaderSchema, sources map[string]sourceInfo) (hcl.Attributes, []Diagnostic) {
	syntaxBody, ok := body.(*hclsyntax.Body)
	if !ok {
		return attributesWithAllowedBlocks(body, allowedBlocks, sources)
	}
	attrs := make(hcl.Attributes, len(syntaxBody.Attributes))
	for name, attr := range syntaxBody.Attributes {
		attrs[name] = attr.AsHCLAttribute()
	}
	return attrs, unexpectedTestRunBlockDiagnostics(syntaxBody.Blocks, sources)
}

func unexpectedTestRunBlockDiagnostics(blocks hclsyntax.Blocks, sources map[string]sourceInfo) []Diagnostic {
	var out []Diagnostic
	for _, block := range blocks {
		if allowedTestRunBlockType(block.Type) {
			continue
		}
		out = append(out, modelDiagnostic(
			DiagnosticError,
			"unexpected_"+strings.ToLower(block.Type)+"_block",
			fmt.Sprintf("Unexpected %q block", block.Type),
			"Blocks are not allowed here.",
			"",
			"",
			sourceRange(block.TypeRange, sources),
		))
	}
	return out
}

func allowedTestRunBlockType(blockType string) bool {
	switch blockType {
	case "module", "variables", "plan_options", "assert":
		return true
	default:
		return false
	}
}

func decodeTestModule(block *hcl.Block, sources map[string]sourceInfo) (*TestModule, []Diagnostic) {
	attrs, diags := attributesWithAllowedBlocks(block.Body, nil, sources)
	module := &TestModule{Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["source"]; ok {
		val := valueFromExpr(attr.Expr, "source", sources)
		module.Source = &val
		module.References = append(module.References, val.References...)
	}
	module.Config = attributesFromMap(attrs, map[string]bool{"source": true}, sources)
	module.References = append(module.References, referencesFromAttributes(module.Config)...)
	return module, diags
}
