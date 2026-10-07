package typecompat

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/meta"
	"custos/internal/syntax"
)

// typeUnsafeComparison reports loose equality comparisons.
type typeUnsafeComparison struct{}

func init() { register(typeUnsafeComparison{}) }

func (typeUnsafeComparison) ID() string { return "TypeUnsafeComparison" }

func (typeUnsafeComparison) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }

func (typeUnsafeComparison) Semantic() {}

// comparableClasses are classes whose instances PHP compares meaningfully.
var comparableClasses = []string{
	`DateTime`, `DateTimeImmutable`, `IntlBreakIterator`, `IntlTimeZone`,
	`PDO`, `PDOStatement`, `ArrayObject`, `SplObjectStorage`, `Closure`,
}

var tucKeywordTypes = map[string]string{
	"int": "int", "integer": "int", "string": "string", "bool": "bool", "boolean": "bool",
	"true": "true", "false": "false", "float": "float", "null": "null", "void": "void",
	"mixed": "mixed", "callable": "callable", "resource": "resource", "iterable": "iterable",
	"object": "object", "static": "static", "self": "self", "$this": "$this",
}

func (typeUnsafeComparison) Check(ctx *analysis.Context, n syntax.Node) {
	b := n.(*syntax.Binary)
	if b.Op.Kind != syntax.TIsEqual && b.Op.Kind != syntax.TIsNotEqual {
		return
	}
	if b.Left == nil || b.Right == nil || b.Left.Span().Len() == 0 || b.Right.Span().Len() == 0 { // D4
		return
	}
	strict := "!=="
	if b.Op.Kind == syntax.TIsEqual {
		strict = "==="
	}
	span := b.Span()

	// Step 1.
	lit, other := b.Right, b.Left
	content, isLit := tucStringContent(ctx, lit)
	if !isLit {
		lit, other = b.Left, b.Right
		content, isLit = tucStringContent(ctx, lit)
	}
	if isLit {
		other = syntax.UnwrapParens(other)
		if cls, objectOnly := tucObjectWithoutToString(ctx, other); objectOnly { // D2
			// A nullable object compared with '' is a null test
			// (null == '' holds), not a string comparison.
			if val, known := tucLiteralValue(lit, content); cls != "" && known && val == "" && ctx.TypeOf(other).Has("null") {
				return
			}
			if cls != "" {
				ctx.ReportSeverity(span, meta.SeverityError, cls+" has no __toString(), so it cannot be compared to a string.")
			}
			return
		}
		if val, known := tucLiteralValue(lit, content); known && val != "" && !tucNumeric(val) { // D3
			op := b.Op.Span
			ctx.ReportSeverity(span, meta.SeverityWarning, "Use '"+strict+"' here; the string is not numeric, so strict comparison is safe.", analysis.Fix{
				Title: "Use " + strict,
				Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: op, NewText: strict}} },
			})
			return
		}
	}

	// Step 2.
	if tucComparableObject(ctx, b.Left) || tucComparableObject(ctx, b.Right) { // D5
		return
	}
	ctx.ReportSeverity(span, meta.SeverityInfo, "Prefer '"+strict+"' to avoid implicit type juggling.") // D6
}

// tucStringContent returns the raw contents of a string literal operand
// (quoted, interpolated, heredoc or nowdoc; not parenthesised).
func tucStringContent(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	switch x := e.(type) {
	case *syntax.Literal:
		if x.LitKind != syntax.LitString {
			return "", false
		}
	case *syntax.InterpolatedString:
		if x.Backtick {
			return "", false
		}
	default:
		return "", false
	}
	t := ctx.Text(e)
	if len(t) > 0 && (t[0] == 'b' || t[0] == 'B') {
		t = t[1:]
	}
	if strings.HasPrefix(t, "<<<") {
		nl := strings.IndexByte(t, '\n') // a heredoc always has an opening newline
		last := strings.LastIndexByte(t, '\n')
		if last == nl { // empty body
			return "", true
		}
		return t[nl+1 : last], true
	}
	if len(t) >= 2 {
		return t[1 : len(t)-1], true
	}
	return "", true
}

// tucLiteralValue returns the runtime value of the string literal lit
// (raw contents given); known is false when the value cannot be determined
// statically (interpolation, escapes in a heredoc).
func tucLiteralValue(lit syntax.Expr, content string) (string, bool) {
	l, ok := lit.(*syntax.Literal)
	if !ok {
		return "", false // interpolated
	}
	raw := l.Raw
	if len(raw) > 0 && (raw[0] == 'b' || raw[0] == 'B') {
		raw = raw[1:]
	}
	if v, ok := util.StringLiteralValue(raw); ok {
		return v, true
	}
	if strings.HasPrefix(raw, "<<<") {
		head := strings.TrimLeft(raw[3:], " \t")
		if strings.HasPrefix(head, "'") || !strings.Contains(content, "\\") {
			return content, true // nowdoc, or heredoc without escapes
		}
	}
	return "", false
}

// tucNumeric reports whether s is a numeric string per PHP's grammar:
// optional surrounding whitespace, an optional sign, then an integer,
// a decimal (`1.`, `.5`, `1.5`) or either with an exponent (`1e3`).
func tucNumeric(s string) bool {
	isWS := func(c byte) bool {
		return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
	}
	for len(s) > 0 && isWS(s[0]) {
		s = s[1:]
	}
	for len(s) > 0 && isWS(s[len(s)-1]) {
		s = s[:len(s)-1]
	}
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	digits := func() int {
		j := i
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
		return i - j
	}
	n := digits()
	if i < len(s) && s[i] == '.' {
		i++
		n += digits()
	}
	if n == 0 {
		return false
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		if digits() == 0 {
			return false
		}
	}
	return i == len(s)
}

// tucNormalizedParts returns the normalised type parts of e (nil when unknown).
func tucNormalizedParts(ctx *analysis.Context, e syntax.Expr) []string {
	t := ctx.TypeOf(e)
	if t.IsUnknown() {
		return nil
	}
	var out []string
	for _, a := range t.Atoms() {
		if strings.HasSuffix(a, "[]") {
			out = append(out, "array")
			continue
		}
		if k, ok := tucKeywordTypes[strings.ToLower(strings.TrimPrefix(a, `\`))]; ok {
			out = append(out, k)
			continue
		}
		out = append(out, a)
	}
	return out
}

// tucObjectWithoutToString implements D2. objectOnly reports whether every
// non-null part is a class; cls is the first resolved class lacking
// __toString() ("" when all have one).
func tucObjectWithoutToString(ctx *analysis.Context, e syntax.Expr) (cls string, objectOnly bool) {
	parts := tucNormalizedParts(ctx, e)
	var classes []string
	for _, p := range parts {
		if p == "null" {
			continue
		}
		if !strings.HasPrefix(p, `\`) {
			return "", false
		}
		classes = append(classes, p)
	}
	if len(classes) == 0 {
		return "", false
	}
	ix := ctx.Index()
	for _, c := range classes {
		decl := ix.Class(c, ctx.PHP)
		if decl == nil || !tucHierarchyKnown(ctx, decl.FQN) {
			continue
		}
		if ix.FindMethod(decl.FQN, "__toString", ctx.PHP) == nil {
			return `\` + strings.TrimPrefix(decl.FQN, `\`), true
		}
	}
	return "", true
}

// tucComparableObject implements D5.
func tucComparableObject(ctx *analysis.Context, e syntax.Expr) bool {
	ix := ctx.Index()
	for _, p := range tucNormalizedParts(ctx, e) {
		if !strings.HasPrefix(p, `\`) || ix.Class(p, ctx.PHP) == nil {
			continue
		}
		for _, c := range comparableClasses {
			if ix.IsSubtype(p, c, ctx.PHP) {
				return true
			}
		}
	}
	return false
}

// tucHierarchyKnown reports whether every parent, interface and trait of
// class (transitively) resolves: otherwise an unresolved ancestor may
// declare __toString() and D2 cannot decide.
func tucHierarchyKnown(ctx *analysis.Context, class string) bool {
	ix := ctx.Index()
	for _, c := range ix.Ancestors(class, ctx.PHP) {
		refs := append(append([]string{c.Parent}, c.Interfaces...), c.Traits...)
		for _, r := range refs {
			if r != "" && ix.Class(r, ctx.PHP) == nil {
				return false
			}
		}
	}
	return true
}
