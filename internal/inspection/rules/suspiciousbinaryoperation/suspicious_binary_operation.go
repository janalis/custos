package suspiciousbinaryoperation

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// suspiciousBinaryOperation flags binary operations that are almost
// certainly mistakes (see specs/SuspiciousBinaryOperation.md).
type suspiciousBinaryOperation struct{}

// Semantic marks the rule as needing the project symbol index.
func (suspiciousBinaryOperation) Semantic()  {}
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
		if astquery.Equivalent(ctx.File, syntax.UnwrapParens(io.Expr), syntax.UnwrapParens(io.Class)) { // D5
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
	msg := "Null or false operands make this negated comparison misleading; use '" + r + "'."
	// `!(a < b)` and `a >= b` only agree when the two operands always
	// compare one way or the other: NAN and uncomparable arrays or objects
	// make both `<` and `>=` false.
	if !t.OnlyOf(sboOrdered...) || !sboOrderedRight(ctx, b.Right) {
		ctx.ReportNode(not, msg)
		return true
	}
	// The `!(...)` group is replaced by a bare comparison, which binds
	// looser than `!`: keep a pair of parentheses where that would regroup.
	text := r
	if astquery.NeedsParensAsComparison(not) || sboOpenEnded(b.Right) {
		text = "(" + r + ")"
	}
	ctx.ReportNode(not, msg, diagnostic.Fix{Title: "Use '" + r + "'", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: not.Span(), NewText: text}}
	}})
	return true
}

// sboOrdered lists the type atoms whose values always compare one way or
// the other with each other.
var sboOrdered = []string{"int", "null", "bool", "true", "false"}

// sboOrderedRight reports whether the right operand of a D4 comparison is a
// number literal or has only ordered types.
func sboOrderedRight(ctx *analysis.Context, e syntax.Expr) bool {
	if l, ok := e.(*syntax.Literal); ok && (l.LitKind == syntax.LitInt || l.LitKind == syntax.LitFloat) {
		return true
	}
	return ctx.TypeOf(e).OnlyOf(sboOrdered...)
}

// sboOpenEnded reports whether e, written last in an unparenthesised
// expression, would absorb the operators that follow it.
func sboOpenEnded(e syntax.Expr) bool {
	switch e.(type) {
	case *syntax.Assign, *syntax.Print, *syntax.Yield, *syntax.YieldFrom, *syntax.Include,
		*syntax.Throw, *syntax.ArrowFunction:
		return true
	}
	return false
}

// D5
func sboSame(ctx *analysis.Context, b *syntax.Binary) bool {
	switch b.Op.Kind {
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
		syntax.TGreater, syntax.TIsGreaterOrEqual, syntax.TLess, syntax.TIsSmallerOrEqual:
	default:
		return false
	}
	if !astquery.EquivalentFoldNames(ctx.File, syntax.UnwrapParens(b.Left), syntax.UnwrapParens(b.Right)) || sboMayVary(ctx, b.Left) {
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
	if !astquery.LogicalOperand(call, true) {
		return false
	}
	cal, ok := semanticquery.ResolveCallee(ctx, call)
	if !ok || len(cal.Params) < len(list.Args) {
		return false
	}
	p := cal.Params[len(list.Args)-1]
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
		diagnostic.Fix{Title: "Move the comparison outside the call", Edits: func() []diagnostic.TextEdit {
			cs, bs := call.Span(), b.Span()
			inner := ctx.SpanText(syntax.Span{Start: cs.Start, End: bs.Start}) + ctx.Text(b.Left) +
				ctx.SpanText(syntax.Span{Start: bs.End, End: cs.End})
			text := inner + " " + op + " " + ctx.Text(b.Right)
			// directly under `!` the comparison would apply to the negation
			if _, neg := call.Parent().(*syntax.Unary); neg || sboOpenEnded(b.Right) {
				text = "(" + text + ")"
			}
			return []diagnostic.TextEdit{{Span: cs, NewText: text}}
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
		switch v {
		case "true":
			decides = !and
		case "false", "null":
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
	wrap := func(n syntax.Node) diagnostic.Fix {
		return diagnostic.Fix{Title: "Add parentheses", Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: n.Span(), NewText: "(" + ctx.Text(n) + ")"}}
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
	ctx.ReportNode(b, sboPrecedenceMsg, diagnostic.Fix{Title: "Add parentheses", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: not.Span(), NewText: "(!" + ctx.SpanText(syntax.Span{Start: not.Expr.Span().Start, End: not.Span().End}) + ")"}}
	}})
	return true
}

// sboIsWordLogical reports the keyword logical operators `and`, `or`, `xor`.
func sboIsWordLogical(k syntax.TokenKind) bool {
	return k == syntax.TAnd || k == syntax.TOr || k == syntax.TXor
}

// sboVarying lists built-ins whose result differs between two identical
// calls (randomness, clocks, cursors, streams, buffers and processes).
var sboVarying = map[string]bool{
	"rand": true, "mt_rand": true, "random_int": true, "random_bytes": true, "lcg_value": true,
	"uniqid": true, "microtime": true, "hrtime": true, "time": true, "array_rand": true,
	"gettimeofday": true, "date": true, "gmdate": true, "idate": true, "getdate": true, "localtime": true,
	"str_shuffle": true, "password_hash": true, "tempnam": true, "tmpfile": true,
	"next": true, "prev": true, "each": true, "array_shift": true, "array_pop": true,
	"fgets": true, "fgetc": true, "fread": true, "fgetcsv": true, "fscanf": true, "fgetss": true,
	"fpassthru": true, "readdir": true, "openssl_random_pseudo_bytes": true,
	"readline": true, "stream_get_contents": true, "stream_get_line": true,
	"stream_socket_recvfrom": true, "stream_socket_accept": true,
	"socket_read": true, "socket_recv": true, "socket_recvfrom": true, "socket_accept": true,
	"gzgets": true, "gzgetc": true, "gzread": true, "gzpassthru": true, "bzread": true,
	"ob_get_clean": true, "ob_get_flush": true, "curl_exec": true, "curl_multi_getcontent": true,
	"shell_exec": true, "exec": true, "system": true, "passthru": true,
}

// sboVaries reports whether the built-in called name returns something else
// on each call: the list above and the row-fetching database families
// (`mysqli_fetch_assoc`, `pg_fetch_row`, `oci_fetch`, ...).
func sboVaries(name string) bool {
	return sboVarying[name] || strings.Contains(name, "_fetch")
}

// sboMayVary reports whether evaluating e twice may give two different
// values (custos): method, static and nullsafe calls, `new`, calls to user
// or unresolved functions, `++`/`--`, assignments, `yield`, `include`,
// `eval`, backtick commands and the built-ins above.
// `count($tags) > count($tags)` stays reported.
func sboMayVary(ctx *analysis.Context, e syntax.Expr) bool {
	vary := false
	syntax.Inspect(e, func(x syntax.Node) bool {
		if vary {
			return false
		}
		switch c := x.(type) {
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.IncDec, *syntax.Assign,
			*syntax.Yield, *syntax.YieldFrom, *syntax.Include, *syntax.Eval:
			vary = true
		case *syntax.InterpolatedString:
			vary = c.Backtick
		case *syntax.FuncCall:
			f := ctx.Types().ResolveFunction(c)
			vary = f == nil || !f.Builtin || sboVaries(ctx.GlobalFunctionName(c))
		}
		return !vary
	})
	return vary
}
