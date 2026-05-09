package tfconfig

import "github.com/hashicorp/hcl/v2"

var testFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "run", LabelNames: []string{"name"}},
	},
}

func decodeTestFile(file discoveredFile, body hcl.Body, sources map[string]sourceInfo) []TestFile {
	tf := TestFile{Path: file.RelPath}
	content, _ := body.Content(testFileSchema)
	for _, block := range content.Blocks {
		tf.Runs = append(tf.Runs, decodeTestRun(block, sources))
	}
	if len(tf.Runs) == 0 {
		return nil
	}
	return []TestFile{tf}
}

func decodeTestRun(block *hcl.Block, sources map[string]sourceInfo) TestRun {
	attrs := justAttributes(block.Body)
	run := TestRun{Name: block.Labels[0], Range: sourceRange(block.DefRange, sources)}
	if attr, ok := attrs["command"]; ok {
		run.Command, _ = literalString(attr.Expr, sources)
	}
	content, _ := block.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variables"},
			{Type: "assert"},
		},
	})
	for _, inner := range content.Blocks {
		switch inner.Type {
		case "variables":
			run.Variables = append(run.Variables, decodeBodyAttributes(inner.Body, nil, sources)...)
		case "assert":
			run.Assertions = append(run.Assertions, decodeCheckRule(inner, sources))
		}
	}
	return run
}
