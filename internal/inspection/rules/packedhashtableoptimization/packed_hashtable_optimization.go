package packedhashtableoptimization

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// packedHashtableOptimization reports array literals with integer-like keys
// written out of order or as numeric strings (PHP 7 packed arrays).
type packedHashtableOptimization struct{}

func (packedHashtableOptimization) ID() string { return "PackedHashtableOptimization" }
func (packedHashtableOptimization) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KArray}
}

func (packedHashtableOptimization) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP70 { // D1
		return
	}
	arr := n.(*syntax.Array)
	if len(arr.Items) < 3 || isDestructuringTarget(arr) { // D2
		return
	}
	if semanticquery.InTestContext(ctx, arr) { // D3
		return
	}
	ascending, hasString := true, false
	var prev int64
	for i, it := range arr.Items {
		if it == nil || it.Key == nil || it.Unpack { // D4
			return
		}
		var (
			key int64
			ok  bool
		)
		if content, _, isStr := astquery.QuotedStringRaw(it.Key); isStr {
			hasString = true
			key, ok = pkStringKey(content)
		} else {
			key, ok = pkNumberKey(ctx, it.Key)
		}
		if !ok { // D5
			return
		}
		if i > 0 && key < prev {
			ascending = false
		}
		prev = key
	}
	var msg string
	switch {
	case !ascending: // D6a
		msg = "Sort the integer keys ascending so the array can be stored packed."
	case hasString: // D6b
		msg = "Write the keys as integers so the array can be stored packed."
	default:
		return
	}
	start := arr.Span().Start
	end := start + 1
	if !arr.Short {
		end = start + uint32(len("array"))
	}
	ctx.Report(syntax.Span{Start: start, End: end}, msg)
}

// pkStringKey parses a string key's raw content as a canonical decimal
// integer (no sign other than `-`, no leading zero, no `-0`; see spec
// Divergences) within the 64-bit integer range.
func pkStringKey(s string) (int64, bool) {
	if len(s) > 1 && s[0] == '0' {
		return 0, false
	}
	digits := strings.TrimPrefix(s, "-")
	if digits == "" || digits[0] == '0' && len(s) != 1 {
		return 0, false // "", "-", "-0", "-01"
	}
	for i := 0; i < len(digits); i++ {
		if digits[i] < '0' || digits[i] > '9' {
			return 0, false
		}
	}
	v, err := strconv.ParseInt(s, 10, 64)
	return v, err == nil
}

// pkNumberKey evaluates an integer literal key (decimal, hex, octal,
// binary, with `_` separators; optionally negated) as PHP does on 64-bit
// platforms. Floats and overflowing literals fail.
func pkNumberKey(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	lit, neg := e, false
	if u, ok := e.(*syntax.Unary); ok && u.Op.Kind == syntax.TMinus {
		lit, neg = u.Expr, true
	}
	l, ok := lit.(*syntax.Literal)
	if !ok || l.LitKind == syntax.LitString {
		return 0, false
	}
	v, ok := astquery.ParseIntLiteral(ctx.Text(l))
	if !ok {
		return 0, false
	}
	if neg {
		v = -v
	}
	return v, true
}

// isDestructuringTarget reports whether arr is a `[...] = …` / foreach
// destructuring pattern rather than an array creation.
func isDestructuringTarget(arr *syntax.Array) bool {
	var n syntax.Node = arr
	for {
		switch p := n.Parent().(type) {
		case *syntax.Assign:
			return p.Var == n
		case *syntax.Foreach:
			return p.Value == n || p.Key == n
		case *syntax.ArrayItem:
			outer, ok := p.Parent().(*syntax.Array)
			if !ok {
				return true // an item of a list(...) pattern
			}
			n = outer
		default:
			return false
		}
	}
}
