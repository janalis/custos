package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
	"custos/internal/types"
)

// suspiciousBinaryOperation flags binary operations that are almost
// certainly mistakes (see specs/SuspiciousBinaryOperation.md).
type suspiciousBinaryOperation struct{}

func init() { register(suspiciousBinaryOperation{}) }

// Semantic marks the rule as needing the project symbol index.
func (suspiciousBinaryOperation) Semantic() {}

func (suspiciousBinaryOperation) ID() string { return "SuspiciousBinaryOperation" }

func (suspiciousBinaryOperation) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KBinary, syntax.KInstanceof}
}

const sboPrecedenceMsg = "Operator precedence is unclear here; add parentheses."

func (suspiciousBinaryOperation) Check(ctx *analysis.Context, n syntax.Node) {
	if io, ok := n.(*syntax.Instanceof); ok {
		if sboTrait(ctx, io) || io.Class == nil {
			return
		}
		if util.Equivalent(ctx.File, syntax.UnwrapParens(io.Expr), syntax.UnwrapParens(io.Class)) { // D5
			ctx.ReportNode(io, "Both operands are the same.")
		}
		return
	}
	b := n.(*syntax.Binary)
	for _, check := range []func(*analysis.Context, *syntax.Binary) bool{
		sboStatement, sboArrayArrow, sboNegated, sboSame, sboMisplaced, sboCoalesce, sboConcatArray, sboConstants, sboPrecedence,
	} {
		if check(ctx, b) {
			return
		}
	}
}

func sboIsComparison(k syntax.TokenKind) bool {
	switch k {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
		syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual, syntax.TSpaceship:
		return true
	}
	return false
}

// D1
func sboTrait(ctx *analysis.Context, io *syntax.Instanceof) bool {
	nm, ok := io.Class.(*syntax.Name)
	if !ok {
		return false
	}
	switch strings.ToLower(nm.Value) {
	case "self", "static":
		return false
	}
	c := ctx.Index().Class(ctx.Names().Class(nm.Value, nm.Span().Start), ctx.PHP)
	if c == nil || c.Kind != syntax.KindTrait {
		return false
	}
	ctx.ReportNode(io, "A trait is never an instanceof target; this is always false.")
	return true
}

// D2
func sboStatement(ctx *analysis.Context, b *syntax.Binary) bool {
	if b.Op.Kind != syntax.TIsEqual {
		return false
	}
	if _, ok := b.Parent().(*syntax.ExprStmt); !ok {
		return false
	}
	ctx.Report(b.Op.Span, "Comparison result is discarded; did you mean '='?")
	return true
}

// D3
func sboArrayArrow(ctx *analysis.Context, b *syntax.Binary) bool {
	if b.Op.Kind != syntax.TIsGreaterOrEqual {
		return false
	}
	switch l := b.Left.(type) {
	case *syntax.Literal:
		if l.LitKind != syntax.LitString {
			return false
		}
	case *syntax.InterpolatedString:
		if l.Backtick {
			return false
		}
	default:
		return false
	}
	it, ok := b.Parent().(*syntax.ArrayItem)
	if !ok || it.Key != nil || it.Value != syntax.Expr(b) {
		return false
	}
	switch it.Parent().(type) {
	case *syntax.Array:
	default:
		return false
	}
	ctx.Report(b.Op.Span, "Did you mean '=>' for an array key?")
	return true
}

// D4
func sboNegated(ctx *analysis.Context, b *syntax.Binary) bool {
	if b.Op.Kind != syntax.TLess && b.Op.Kind != syntax.TIsSmallerOrEqual {
		return false
	}
	var cur syntax.Node = b
	for {
		if p, ok := cur.Parent().(*syntax.Paren); ok {
			cur = p
			continue
		}
		break
	}
	not, ok := cur.Parent().(*syntax.Unary)
	if !ok || not.Op.Kind != syntax.TExclaim {
		return false
	}
	t := ctx.TypeOf(b.Left)
	if t.IsUnknown() || !t.HasAny("null", "bool", "true", "false") {
		return false
	}
	op := " >= "
	if b.Op.Kind == syntax.TIsSmallerOrEqual {
		op = " > "
	}
	r := ctx.Text(b.Left) + op + ctx.Text(b.Right)
	ctx.ReportNode(not, "Null or false operands make this negated comparison misleading; use '"+r+"'.",
		analysis.Fix{Title: "Use '" + r + "'", Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: not.Span(), NewText: r}}
		}})
	return true
}

// D5
func sboSame(ctx *analysis.Context, b *syntax.Binary) bool {
	switch b.Op.Kind {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
		syntax.TGreater, syntax.TIsGreaterOrEqual, syntax.TLess, syntax.TIsSmallerOrEqual:
	default:
		return false
	}
	if !util.EquivalentFoldNames(ctx.File, syntax.UnwrapParens(b.Left), syntax.UnwrapParens(b.Right)) || sboMayVary(ctx, b.Left) {
		return false
	}
	ctx.ReportNode(b, "Both operands are the same.")
	return true
}

// D6
func sboMisplaced(ctx *analysis.Context, b *syntax.Binary) bool {
	switch b.Op.Kind {
	case syntax.TIsEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
		syntax.TGreater, syntax.TIsGreaterOrEqual, syntax.TLess, syntax.TIsSmallerOrEqual:
	case syntax.TIsNotEqual:
		if ctx.SpanText(b.Op.Span) == "<>" {
			return false
		}
	default:
		return false
	}
	arg, ok := b.Parent().(*syntax.Arg)
	if !ok || arg.Value != syntax.Expr(b) {
		return false
	}
	list, ok := arg.Parent().(*syntax.ArgList)
	if !ok || len(list.Args) == 0 || list.Args[len(list.Args)-1] != syntax.Expr(arg) {
		return false
	}
	call := list.Parent()
	switch call.(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
	default:
		return false
	}
	if !npeLogicalOperand(call, true) {
		return false
	}
	cal, ok := resolveCallee(ctx, call)
	if !ok || len(cal.params) < len(list.Args) {
		return false
	}
	p := cal.params[len(list.Args)-1]
	pt := p.Type
	if pt == "" {
		pt = p.DocType
	}
	if types.FromDoc(pt, nil).HasAny("bool", "true", "false") {
		return false
	}
	ret := sboCalleeReturn(ctx, call)
	rt := ctx.TypeOf(b.Right)
	if rt.IsUnknown() { // nothing known about the right operand: no report
		return false
	}
	for _, a := range rt.Atoms() {
		if !ret.Has(a) {
			return false
		}
	}
	op := ctx.SpanText(b.Op.Span)
	ctx.Report(b.Op.Span, "This comparison probably belongs outside the call parentheses.",
		analysis.Fix{Title: "Move the comparison outside the call", Edits: func() []analysis.TextEdit {
			cs, bs := call.Span(), b.Span()
			inner := ctx.SpanText(syntax.Span{Start: cs.Start, End: bs.Start}) + ctx.Text(b.Left) +
				ctx.SpanText(syntax.Span{Start: bs.End, End: cs.End})
			return []analysis.TextEdit{{Span: cs, NewText: inner + " " + op + " " + ctx.Text(b.Right)}}
		}})
	return true
}

// sboCalleeReturn returns the declared/doc return type of a call's callee.
func sboCalleeReturn(ctx *analysis.Context, call syntax.Node) types.Type {
	if c, ok := call.(*syntax.FuncCall); ok {
		f := ctx.Types().ResolveFunction(c) // non-nil: resolveCallee succeeded
		// declared and documented return types, unknown parts dropped
		decl, doc := types.FromDoc(f.Return, nil), types.FromDoc(f.DocReturn, nil)
		if decl.IsUnknown() {
			return doc
		}
		if doc.IsUnknown() {
			return decl
		}
		return types.Union(decl, doc)
	}
	return ctx.TypeOf(call.(syntax.Expr))
}

// D7
func sboCoalesce(ctx *analysis.Context, b *syntax.Binary) bool {
	if b.Op.Kind != syntax.TCoalesce {
		return false
	}
	u, ok := syntax.UnwrapParens(b.Left).(*syntax.Unary)
	if !ok {
		return false
	}
	switch u.Op.Kind {
	case syntax.TExclaim, syntax.TIntCast, syntax.TDoubleCast, syntax.TStringCast, syntax.TArrayCast,
		syntax.TObjectCast, syntax.TBoolCast, syntax.TUnsetCast:
	default:
		return false
	}
	ctx.ReportNode(u, "'"+ctx.Text(u)+"' is never null, so '??' is useless; add parentheses.")
	return true
}

// D8
func sboConcatArray(ctx *analysis.Context, b *syntax.Binary) bool {
	if b.Op.Kind != syntax.TDot {
		return false
	}
	_, l := b.Left.(*syntax.Array)
	_, r := b.Right.(*syntax.Array)
	if !l && !r {
		return false
	}
	ctx.Report(b.Op.Span, "Concatenating an array literal makes no sense.")
	return true
}

// D9
func sboConstants(ctx *analysis.Context, b *syntax.Binary) bool {
	var and bool
	switch b.Op.Kind {
	case syntax.TBooleanAnd, syntax.TAnd:
		and = true
	case syntax.TBooleanOr, syntax.TOr:
	default:
		return false
	}
	if !ctx.Bool("VERIFY_CONSTANTS_IN_CONDITIONS") {
		return false
	}
	for _, op := range []syntax.Expr{b.Left, b.Right} {
		c, ok := op.(*syntax.ConstFetch)
		if !ok || c.Name == nil {
			continue
		}
		v := strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`))
		var decides bool
		switch {
		case v == "true":
			decides = !and
		case v == "false" || v == "null":
			decides = and
		default:
			continue
		}
		if decides {
			ctx.ReportNode(op, "This constant decides the whole condition.")
		} else {
			ctx.ReportNode(op, "This constant has no effect in the condition.")
		}
		return true
	}
	return false
}

// D10
func sboPrecedence(ctx *analysis.Context, b *syntax.Binary) bool {
	if !ctx.Bool("VERIFY_UNCLEAR_OPERATIONS_PRIORITIES") {
		return false
	}
	wrap := func(n syntax.Node) analysis.Fix {
		return analysis.Fix{Title: "Add parentheses", Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: n.Span(), NewText: "(" + ctx.Text(n) + ")"}}
		}}
	}
	k := b.Op.Kind
	if sboIsWordLogical(k) { // D10a, keyword forms (custos diverges)
		if pb, ok := b.Parent().(*syntax.Binary); ok && sboIsWordLogical(pb.Op.Kind) && pb.Op.Kind != k {
			ctx.ReportNode(b, sboPrecedenceMsg, wrap(b))
			return true
		}
		return false
	}
	if k == syntax.TBooleanAnd || k == syntax.TBooleanOr {
		other := syntax.TBooleanOr
		if k == syntax.TBooleanOr {
			other = syntax.TBooleanAnd
		}
		if pb, ok := b.Parent().(*syntax.Binary); ok && pb.Op.Kind == other { // D10a
			ctx.ReportNode(b, sboPrecedenceMsg, wrap(b))
			return true
		}
		if a, ok := b.Parent().(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && a.Value == syntax.Expr(b) { // D10b
			if _, stmt := a.Parent().(*syntax.ExprStmt); !stmt {
				ctx.ReportNode(b, sboPrecedenceMsg, wrap(b))
				return true
			}
		}
		return false
	}
	if sboIsComparison(k) { // D10c
		if a, ok := b.Parent().(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && a.Value == syntax.Expr(b) {
			if ifs, ok := a.Parent().(*syntax.If); ok && ifs.Cond == syntax.Expr(a) {
				ctx.ReportNode(a, sboPrecedenceMsg, wrap(b))
				return true
			}
		}
	}
	switch k { // D10d
	case syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
	default:
		return false
	}
	not, ok := b.Left.(*syntax.Unary)
	if !ok || not.Op.Kind != syntax.TExclaim || not.Expr == nil {
		return false
	}
	ctx.ReportNode(b, sboPrecedenceMsg, analysis.Fix{Title: "Add parentheses", Edits: func() []analysis.TextEdit {
		return []analysis.TextEdit{{Span: not.Span(), NewText: "(!" + ctx.SpanText(syntax.Span{Start: not.Expr.Span().Start, End: not.Span().End}) + ")"}}
	}})
	return true
}

// sboIsWordLogical reports the keyword logical operators `and`, `or`, `xor`.
func sboIsWordLogical(k syntax.TokenKind) bool {
	return k == syntax.TAnd || k == syntax.TOr || k == syntax.TXor
}

// sboVarying lists built-ins whose result differs between two identical
// calls (randomness, clocks, cursors and streams).
var sboVarying = map[string]bool{
	"rand": true, "mt_rand": true, "random_int": true, "random_bytes": true, "lcg_value": true,
	"uniqid": true, "microtime": true, "hrtime": true, "time": true, "array_rand": true,
	"next": true, "prev": true, "each": true, "array_shift": true, "array_pop": true,
	"fgets": true, "fgetc": true, "fread": true, "fgetcsv": true, "fscanf": true,
	"readdir": true, "openssl_random_pseudo_bytes": true,
}

// sboMayVary reports whether evaluating e twice may give two different
// values (custos): method, static and nullsafe calls, `new`, calls to user
// or unresolved functions, `++`/`--`, assignments and the built-ins above.
// `count($tags) > count($tags)` stays reported.
func sboMayVary(ctx *analysis.Context, e syntax.Expr) bool {
	vary := false
	syntax.Inspect(e, func(x syntax.Node) bool {
		switch c := x.(type) {
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.IncDec, *syntax.Assign:
			vary = true
		case *syntax.FuncCall:
			f := ctx.Types().ResolveFunction(c)
			vary = f == nil || !f.Builtin || sboVarying[ctx.GlobalFunctionName(c)]
		}
		return !vary
	})
	return vary
}
