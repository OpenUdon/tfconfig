package tfconfig

import (
	"encoding/json"
	"math/big"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type sourceInfo struct {
	id   string
	path string
	data []byte
}

func valueFromExpr(expr hcl.Expression, attrPath string, sources map[string]sourceInfo) Value {
	refs := referencesFromExpr(expr, sources)
	text := exprText(expr, sources)
	rng := sourceRange(expr.Range(), sources)

	val, diags := expr.Value(nil)
	if !diags.HasErrors() && val.IsWhollyKnown() {
		kind, literal, ok := literalFromCty(val)
		if ok {
			out := Value{
				Kind:       kind,
				Literal:    literal,
				References: refs,
				Range:      rng,
			}
			if candidatePath, ok := sensitiveCandidatePath(attrPath, out.Literal); ok {
				out.SensitiveCandidate = &SensitiveCandidate{
					Reason:        "attribute name suggests secret material",
					AttributePath: candidatePath,
				}
			}
			return out
		}
	}

	if text == "" {
		text = expr.Range().String()
	}
	return Value{
		Kind:       ValueKindExpression,
		Expression: text,
		References: refs,
		Range:      rng,
	}
}

func referencesFromExpr(expr hcl.Expression, sources map[string]sourceInfo) []Reference {
	var refs []Reference
	for _, traversal := range expr.Variables() {
		text := traversalText(traversal)
		refs = append(refs, Reference{
			Traversal: text,
			Subject:   referenceSubject(traversal, text),
			Range:     sourceRange(traversal.SourceRange(), sources),
		})
	}
	sortReferences(refs)
	return refs
}

func exprText(expr hcl.Expression, sources map[string]sourceInfo) string {
	rng := expr.Range()
	info, ok := sources[rng.Filename]
	if !ok {
		return strings.TrimSpace(rng.String())
	}
	if len(info.data) == 0 {
		return strings.TrimSpace(rng.String())
	}
	if rng.Start.Byte < 0 || rng.End.Byte > len(info.data) || rng.Start.Byte >= rng.End.Byte {
		return strings.TrimSpace(rng.String())
	}
	return strings.TrimSpace(string(info.data[rng.Start.Byte:rng.End.Byte]))
}

func traversalText(traversal hcl.Traversal) string {
	return strings.TrimSpace(string(hclwrite.TokensForTraversal(traversal).Bytes()))
}

func referenceSubject(traversal hcl.Traversal, fallback string) string {
	if len(traversal) == 0 {
		return fallback
	}
	switch root := traversal[0].(type) {
	case hcl.TraverseRoot:
		if len(traversal) > 1 {
			if attr, ok := traversal[1].(hcl.TraverseAttr); ok {
				return root.Name + "." + attr.Name
			}
		}
		return root.Name
	default:
		return fallback
	}
}

func literalFromCty(val cty.Value) (ValueKind, any, bool) {
	if val.IsNull() {
		return ValueKindNull, nil, true
	}
	typ := val.Type()
	switch {
	case typ == cty.Bool:
		return ValueKindBool, val.True(), true
	case typ == cty.String:
		return ValueKindString, val.AsString(), true
	case typ == cty.Number:
		return ValueKindNumber, numberLiteral(val), true
	case typ.IsObjectType() || typ.IsMapType():
		obj := map[string]any{}
		for key, child := range val.AsValueMap() {
			_, literal, ok := literalFromCty(child)
			if !ok {
				return ValueKindExpression, nil, false
			}
			obj[key] = literal
		}
		return ValueKindCollection, obj, true
	case typ.IsTupleType() || typ.IsListType() || typ.IsSetType():
		values := val.AsValueSlice()
		arr := make([]any, 0, len(values))
		for _, child := range values {
			_, literal, ok := literalFromCty(child)
			if !ok {
				return ValueKindExpression, nil, false
			}
			arr = append(arr, literal)
		}
		return ValueKindCollection, arr, true
	default:
		return ValueKindExpression, nil, false
	}
}

func numberLiteral(val cty.Value) any {
	f := val.AsBigFloat()
	if i, acc := f.Int64(); acc == big.Exact {
		return i
	}
	return json.Number(f.Text('g', -1))
}

func sensitiveCandidatePath(attrPath string, literal any) (string, bool) {
	if hasSensitivePathMarker(attrPath) {
		return attrPath, true
	}
	switch typed := literal.(type) {
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			childPath := joinAttrPath(attrPath, key)
			if hasSensitivePathMarker(key) || hasSensitivePathMarker(childPath) {
				return childPath, true
			}
		}
		for _, key := range keys {
			childPath := joinAttrPath(attrPath, key)
			if candidatePath, ok := sensitiveCandidatePath(childPath, typed[key]); ok {
				return candidatePath, true
			}
		}
	case []any:
		for _, child := range typed {
			if candidatePath, ok := sensitiveCandidatePath(attrPath, child); ok {
				return candidatePath, true
			}
		}
	}
	return "", false
}

func hasSensitivePathMarker(path string) bool {
	lower := strings.ToLower(path)
	for _, marker := range []string{"password", "passwd", "secret", "token", "api_key", "apikey", "access_key", "private_key"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func joinAttrPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}
