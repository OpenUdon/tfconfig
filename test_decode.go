package tfconfig

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
)

var testFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "run", LabelNames: []string{"name"}},
	},
}

func decodeTestFile(file discoveredFile, body hcl.Body, sources map[string]sourceInfo) ([]TestFile, []Diagnostic) {
	tf := TestFile{Path: file.RelPath}
	content, contentDiags := body.Content(testFileSchema)
	diags := convertDiagnostics(contentDiags, "", "", sources)
	for _, block := range content.Blocks {
		run, runDiags := decodeTestRun(block, sources)
		diags = append(diags, runDiags...)
		tf.Runs = append(tf.Runs, run)
	}
	if len(tf.Runs) == 0 {
		return nil, diags
	}
	return []TestFile{tf}, diags
}

func decodeTestRun(block *hcl.Block, sources map[string]sourceInfo) (TestRun, []Diagnostic) {
	attrs, diags := attributesWithAllowedBlocks(block.Body, []hcl.BlockHeaderSchema{{Type: "variables"}, {Type: "assert"}}, sources)
	run := TestRun{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["command"]; ok {
		run.Command = strings.Trim(exprText(attr.Expr, sources), `"`)
	}
	content, _, contentDiags := block.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variables"},
			{Type: "assert"},
		},
	})
	diags = append(diags, convertDiagnostics(contentDiags, "", "", sources)...)
	for _, inner := range content.Blocks {
		switch inner.Type {
		case "variables":
			vars, varDiags := decodeBodyAttributes(inner.Body, nil, nil, sources)
			diags = append(diags, varDiags...)
			run.Variables = append(run.Variables, vars...)
		case "assert":
			rule, ruleDiags := decodeCheckRule(inner, sources)
			diags = append(diags, ruleDiags...)
			run.Assertions = append(run.Assertions, rule)
		}
	}
	return run, diags
}
