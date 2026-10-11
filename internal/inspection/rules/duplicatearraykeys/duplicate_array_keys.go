package duplicatearraykeys

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// duplicateArrayKeys reports repeated string keys in an array literal.
type duplicateArrayKeys struct{}

const (
	duplicateArrayPairMsg = "Same key and value already present; remove this entry."
	duplicateArrayKeyMsg  = "Key already used earlier; the earlier entry is overwritten."
)

func (duplicateArrayKeys) ID() string               { return "DuplicateArrayKeys" }
func (duplicateArrayKeys) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArray} }
func (duplicateArrayKeys) Check(ctx *analysis.Context, n syntax.Node) {
	arr := n.(*syntax.Array)
	var seen map[string]syntax.Expr
	for _, it := range arr.Items {
		if it == nil || it.Key == nil || it.Unpack || it.Value == nil {
			continue
		}
		key, ok := arrayKeyValue(it.Key, ctx.PHP) // D1
		if !ok {
			continue
		}
		if seen == nil {
			seen = map[string]syntax.Expr{}
		}
		if prev, dup := seen[key]; dup { // D2
			_, isArray := it.Value.(*syntax.Array)
			if !isArray && astquery.EquivalentFoldNames(ctx.File, prev, it.Value) {
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
// any other key expression. Strings are decoded with the escapes the target
// PHP version knows.
func arrayKeyValue(e syntax.Expr, php phpversion.Version) (string, bool) {
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
		v, ok := astquery.ParseIntLiteral(lit.Raw)
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
		val, ok := stringKeyValue(lit.Raw, php) // not ok for nowdoc
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

// stringKeyValue decodes a quoted string literal like
// astquery.StringLiteralValue, except that escapes the target version does
// not know stay as written: `\u{...}` before PHP 7.0 and `\e` before PHP 5.4
// are a backslash followed by ordinary characters there.
func stringKeyValue(raw string, php phpversion.Version) (string, bool) {
	val, ok := astquery.StringLiteralValue(raw)
	// Only a double-quoted literal has these escapes: the quote opens the
	// literal, possibly after a binary prefix.
	quote := strings.IndexByte(raw, '"')
	if !ok || php >= phpversion.PHP70 || quote < 0 || quote > 1 || raw[0] == '\'' {
		return val, ok
	}
	body := raw[quote+1 : len(raw)-1]
	var b strings.Builder
	start := 0
	for i := 0; i+1 < len(body); i++ {
		if body[i] != '\\' {
			continue
		}
		if n := body[i+1]; n == 'u' || (n == 'e' && php < phpversion.PHP54) {
			// Decode what precedes the escape, then keep its two bytes.
			piece, _ := astquery.StringLiteralValue(`"` + body[start:i] + `"`)
			b.WriteString(piece)
			b.WriteString(body[i : i+2])
			start = i + 2
		}
		i++ // the escaped byte never starts another escape
	}
	if start == 0 {
		return val, true
	}
	piece, _ := astquery.StringLiteralValue(`"` + body[start:] + `"`)
	b.WriteString(piece)
	return b.String(), true
}

// canonicalIntString reports whether PHP converts the string array key s to
// an integer key: a decimal integer without leading zeros, '+' or spaces
// that fits a 64-bit int ("-0" stays a string).
func canonicalIntString(s string) bool {
	d := s
	if d != "" && d[0] == '-' {
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
