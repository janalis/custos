package probablebugs

import (
	"strconv"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// duplicateArrayKeys reports repeated string keys in an array literal.
type duplicateArrayKeys struct{}

func init() { register(duplicateArrayKeys{}) }

const (
	duplicateArrayPairMsg = "Same key and value already present; remove this entry."
	duplicateArrayKeyMsg  = "Key already used earlier; the earlier entry is overwritten."
)

func (duplicateArrayKeys) ID() string { return "DuplicateArrayKeys" }

func (duplicateArrayKeys) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArray} }

func (duplicateArrayKeys) Check(ctx *analysis.Context, n syntax.Node) {
	arr := n.(*syntax.Array)
	var seen map[string]syntax.Expr
	for _, it := range arr.Items {
		if it == nil || it.Key == nil || it.Unpack || it.Value == nil {
			continue
		}
		key, ok := arrayKeyValue(it.Key) // D1
		if !ok {
			continue
		}
		if seen == nil {
			seen = map[string]syntax.Expr{}
		}
		if prev, dup := seen[key]; dup { // D2
			_, isArray := it.Value.(*syntax.Array)
			if !isArray && util.EquivalentFoldNames(ctx.File, prev, it.Value) {
				ctx.Report(syntax.Span{Start: it.Key.Span().Start, End: it.Value.Span().End}, duplicateArrayPairMsg)
			} else {
				ctx.ReportNode(it.Key, duplicateArrayKeyMsg)
			}
		}
		seen[key] = it.Value // D3
	}
}

// arrayKeyValue returns the key PHP actually stores for a literal key: string
// literals are decoded, and canonical decimal integer strings ('7', '-3') as
// well as integer literals in any base become integer keys. ok is false for
// any other key expression.
func arrayKeyValue(e syntax.Expr) (string, bool) {
	e = syntax.UnwrapParens(e)
	neg := false
	if u, isUnary := e.(*syntax.Unary); isUnary && u.Op.Kind == syntax.TMinus {
		neg, e = true, syntax.UnwrapParens(u.Expr)
	}
	lit, isLit := e.(*syntax.Literal)
	if !isLit {
		return "", false
	}
	switch lit.LitKind {
	case syntax.LitInt:
		v, ok := util.ParseIntLiteral(lit.Raw)
		if !ok {
			return "", false
		}
		if neg {
			v = -v
		}
		return "i:" + strconv.FormatInt(v, 10), true
	case syntax.LitString:
		if neg {
			return "", false
		}
		val, ok := util.StringLiteralValue(lit.Raw) // not ok for nowdoc
		if !ok {
			return "", false
		}
		if canonicalIntString(val) {
			return "i:" + val, true
		}
		return "s:" + val, true
	}
	return "", false
}

// canonicalIntString reports whether PHP converts the string array key s to
// an integer key: a decimal integer without leading zeros, '+' or spaces
// that fits a 64-bit int ("-0" stays a string).
func canonicalIntString(s string) bool {
	d := s
	if len(d) > 0 && d[0] == '-' {
		d = d[1:]
		if d == "0" {
			return false
		}
	}
	if d == "" || (d[0] == '0' && len(d) > 1) {
		return false
	}
	for i := 0; i < len(d); i++ {
		if d[i] < '0' || d[i] > '9' {
			return false
		}
	}
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil
}
