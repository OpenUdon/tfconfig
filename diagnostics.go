package tfconfig

import (
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

var diagnosticCodeRe = regexp.MustCompile(`[^a-z0-9]+`)

func convertDiagnostics(diags hcl.Diagnostics, moduleAddress, address string, sources map[string]sourceInfo) []Diagnostic {
	if len(diags) == 0 {
		return nil
	}
	out := make([]Diagnostic, 0, len(diags))
	for _, diag := range diags {
		if diag == nil {
			continue
		}
		converted := Diagnostic{
			Severity:      diagnosticSeverity(diag.Severity),
			Code:          diagnosticCode(diag.Summary),
			Summary:       diag.Summary,
			Detail:        diag.Detail,
			ModuleAddress: moduleAddress,
			Address:       address,
		}
		if diag.Subject != nil {
			converted.Range = sourceRange(*diag.Subject, sources)
		} else if diag.Context != nil {
			converted.Range = sourceRange(*diag.Context, sources)
		}
		out = append(out, converted)
	}
	return out
}

func diagnosticSeverity(severity hcl.DiagnosticSeverity) DiagnosticSeverity {
	switch severity {
	case hcl.DiagError:
		return DiagnosticError
	case hcl.DiagWarning:
		return DiagnosticWarning
	default:
		return DiagnosticInfo
	}
}

func diagnosticCode(summary string) string {
	code := strings.ToLower(strings.TrimSpace(summary))
	code = diagnosticCodeRe.ReplaceAllString(code, "_")
	code = strings.Trim(code, "_")
	if code == "" {
		return "hcl_diagnostic"
	}
	return code
}

func sourceRange(rng hcl.Range, sources map[string]sourceInfo) *SourceRange {
	if rng.Filename == "" && rng.Start.Line == 0 && rng.End.Line == 0 {
		return nil
	}
	out := &SourceRange{
		SourceID: rng.Filename,
		Path:     normalizePath(rng.Filename),
		Start: Position{
			Line:   rng.Start.Line,
			Column: rng.Start.Column,
			Byte:   rng.Start.Byte,
		},
		End: Position{
			Line:   rng.End.Line,
			Column: rng.End.Column,
			Byte:   rng.End.Byte,
		},
	}
	if info, ok := sources[rng.Filename]; ok {
		out.SourceID = info.id
		out.Path = info.path
	}
	return out
}

func modelDiagnostic(severity DiagnosticSeverity, code, summary, detail, moduleAddress, address string, rng *SourceRange) Diagnostic {
	return Diagnostic{
		Severity:      severity,
		Code:          code,
		Summary:       summary,
		Detail:        detail,
		ModuleAddress: moduleAddress,
		Address:       address,
		Range:         rng,
	}
}
