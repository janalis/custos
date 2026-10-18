package senselessternaryoperator

import (
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// senselessTernaryOperator reports `$a === $b ? $a : $b` style ternaries
// that always yield the same operand.
type senselessTernaryOperator struct{}

func (senselessTernaryOperator) ID() string               { return "SenselessTernaryOperator" }
func (senselessTernaryOperator) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KTernary} }

func (senselessTernaryOperator) Check(ctx *analysis.Context, n syntax.Node) {
	t := n.(*syntax.Ternary)
	if t.Then == nil || t.Else == nil { // D1/E2
		return
	}
	cond, ok := syntax.UnwrapParens(t.Cond).(*syntax.Binary) // D2
	if !ok || cond.Left == nil || cond.Right == nil {
		return
	}
	var yes, no syntax.Expr // D3
	switch cond.Op.Kind {
	case syntax.TIsIdentical:
		yes, no = t.Then, t.Else
	case syntax.TIsNotIdentical:
		yes, no = t.Else, t.Then
	default:
		return
	}
	f := ctx.File
	eq := func(a, b syntax.Expr) bool { return astquery.EquivalentFoldNames(f, a, b) }
	if !(eq(cond.Left, yes) && eq(cond.Right, no)) && !(eq(cond.Left, no) && eq(cond.Right, yes)) { // D4
		return
	}
	if unstable(ctx, cond.Left) || unstable(ctx, cond.Right) { // E5/E6
		return
	}
	repl := ctx.Text(no) // D6
	span := t.Span()
	ctx.Report(span, "This ternary always yields '"+repl+"'; use it directly.", diagnostic.Fix{
		Title: "Replace with '" + repl + "'",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: separated(f.Src, span, repl)}}
		},
	})
}

// pureBuiltins are native functions whose result depends only on their
// arguments and which neither write through an argument nor touch global
// state, so two textually identical calls yield identical values.
var pureBuiltins = func() map[string]bool {
	m := map[string]bool{}
	for name := range strings.FieldsSeq(`
		count sizeof strlen mb_strlen strtolower strtoupper mb_strtolower mb_strtoupper
		trim ltrim rtrim ucfirst lcfirst ucwords substr mb_substr strpos stripos strrpos
		mb_strpos str_contains str_starts_with str_ends_with strrev str_repeat str_pad
		implode join explode sprintf md5 sha1 crc32 basename dirname
		abs min max floor ceil round intdiv intval floatval strval boolval gettype
		is_null is_int is_integer is_float is_string is_bool is_array is_numeric is_scalar
		array_keys array_values array_key_exists in_array array_search array_sum
		array_merge array_flip array_reverse array_slice array_unique array_key_first
		array_key_last`) {
		m[name] = true
	}
	return m
}()

// unstable reports whether e cannot stand for "the same value" when it is
// written twice: evaluating it has (or may have) a side effect, so the second
// evaluation differs and removing one changes behaviour (E5), or it contains
// a float zero literal, which is identical to but distinguishable from a
// zero of the other sign (E6). Closure bodies are not evaluated in place.
func unstable(ctx *analysis.Context, e syntax.Expr) bool {
	found := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		if found {
			return false
		}
		switch x := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.FuncCall:
			found = !pureBuiltins[semanticquery.NativeBuiltinName(ctx, x)]
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Clone,
			*syntax.Assign, *syntax.IncDec, *syntax.Include, *syntax.Eval,
			*syntax.Exit, *syntax.Print, *syntax.Throw, *syntax.Yield, *syntax.YieldFrom:
			found = true
		case *syntax.Binary:
			found = x.Op.Kind == syntax.TPipe
		case *syntax.InterpolatedString:
			found = x.Backtick
		case *syntax.Literal:
			found = x.LitKind == syntax.LitFloat && floatZero(x.Raw)
		}
		return !found
	})
	return found
}

// floatZero reports whether a float literal's value is zero (`0.0`, `.0`,
// `0e0`, `0_0.0`).
func floatZero(raw string) bool {
	v, err := strconv.ParseFloat(strings.ReplaceAll(raw, "_", ""), 64)
	return err == nil && v == 0
}

// separated pads repl with a space on each side where it would otherwise
// fuse with an adjacent word: the ternary may start with `(` or end with a
// bracket or quote right against a keyword (`return($a === null) ? ...`),
// while the replacement starts or ends with a name. A leading backslash
// fuses too (`return\FOO` is one qualified name).
func separated(src []byte, span syntax.Span, repl string) string {
	if span.Start > 0 && astquery.NameByte(src[span.Start-1]) && (astquery.NameByte(repl[0]) || repl[0] == '\\') {
		repl = " " + repl
	}
	if int(span.End) < len(src) && astquery.NameByte(src[span.End]) && astquery.NameByte(repl[len(repl)-1]) {
		repl += " "
	}
	return repl
}
