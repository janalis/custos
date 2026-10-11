package typescastingcanbeused

import (
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/php/syntax"
)

// typesCastingCanBeUsed reports conversion helpers, single-interpolation
// strings and explicit __toString() calls that a cast expresses directly.
type typesCastingCanBeUsed struct{}

func (typesCastingCanBeUsed) ID() string { return "TypesCastingCanBeUsed" }
func (typesCastingCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KInterpolatedString, syntax.KMethodCall}
}

var castByConversionFunc = map[string]string{
	"intval": "int", "floatval": "float", "strval": "string", "boolval": "bool",
}

var castBySettypeType = map[string]string{
	"boolean": "bool", "bool": "bool", "integer": "int", "int": "int",
	"float": "float", "double": "float", "string": "string", "array": "array",
}

func (r typesCastingCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	switch n := n.(type) {
	case *syntax.FuncCall:
		r.checkCall(ctx, n)
	case *syntax.InterpolatedString:
		if ctx.Bool("REPORT_INLINES") {
			r.checkString(ctx, n)
		}
	case *syntax.MethodCall:
		if ctx.Bool("REPORT_TO_STRING_METHOD_CALLS") {
			r.checkToString(ctx, n)
		}
	}
}

func (typesCastingCanBeUsed) checkCall(ctx *analysis.Context, call *syntax.FuncCall) {
	name := ctx.GlobalFunctionName(call) // any case, global functions only
	if name == "" {
		return
	}
	args, ok := astquery.CallArgValues(call) // spread arguments are skipped
	if !ok {
		return
	}
	if name == "settype" {
		if len(args) != 2 { // D4; D5 (shadowing) is part of GlobalFunctionName
			return
		}
		lit, ok := args[1].(*syntax.Literal)
		if !ok || lit.LitKind != syntax.LitString || len(lit.Raw) < 2 || (lit.Raw[0] != '\'' && lit.Raw[0] != '"') {
			return
		}
		typ, ok := castBySettypeType[strings.ToLower(lit.Raw[1:len(lit.Raw)-1])] // PHP compares the type name case-insensitively
		if !ok {
			return
		}
		if _, ok := call.Parent().(*syntax.ExprStmt); !ok { // D6
			return
		}
		if !reReadable(args[0]) { // D7: the target is written twice in R
			return
		}
		v := ctx.Text(args[0])
		r, tail := v+" = ("+typ+")"+v, " instead (a cast is clearer and faster)."
		if name, ok := astquery.PlainVariableName(args[0]); ok && definitelyDefined(ctx, call, name) {
			reportCast(ctx, call.Span(), r, tail)
			return
		}
		ctx.Report(call.Span(), "Use '"+r+"'"+tail) // F2: reading a missing target would warn
		return
	}
	typ, ok := castByConversionFunc[name] // D1
	if !ok {
		return
	}
	switch len(args) { // D2
	case 1:
	case 2:
		if lit, ok := args[1].(*syntax.Literal); name != "intval" || !ok || lit.LitKind != syntax.LitInt || lit.Raw != "10" {
			return
		}
	default:
		return
	}
	r := "(" + typ + ")" + castOperand(ctx, args[0]) // D3
	reportCast(ctx, call.Span(), guardCastResult(call, r), " instead (a cast is clearer and faster).")
}

func (typesCastingCanBeUsed) checkString(ctx *analysis.Context, s *syntax.InterpolatedString) {
	sp := s.Span()
	if s.Heredoc || s.Backtick || len(s.Parts) != 1 || sp.Len() < 2 || ctx.Src[sp.End-1] != '"' { // D8
		return
	}
	part := s.Parts[0] // never a StringPart: the lexer only interpolates strings with a $var, ${ or {$
	ps := part.Span()
	var r string
	switch {
	case ctx.Src[ps.Start] == '$' && ps.Len() > 1 && ctx.Src[ps.Start+1] == '{':
		return // "${name}": not reported
	case ps.Start > 0 && ctx.Src[ps.Start-1] == '{': // D9, braced
		r = "(string)(" + ctx.Text(part) + ")"
	default:
		r = "(string)" + simpleInterpolation(ctx, part)
	}
	reportCast(ctx, sp, guardCastResult(s, r), " to make the string conversion explicit.")
}

func (typesCastingCanBeUsed) checkToString(ctx *analysis.Context, m *syntax.MethodCall) {
	id, ok := m.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "__toString") || m.NullSafe { // D10; nullsafe keeps its null-safety
		return
	}
	if _, ok := astquery.ArgValues(m.Args); !ok { // `->__toString(...)` creates a closure
		return
	}
	r := "(string)" + castOperand(ctx, m.Var) // D11
	reportCast(ctx, m.Span(), guardCastResult(m, r), " instead of calling __toString() directly.")
}

func reportCast(ctx *analysis.Context, span syntax.Span, r, tail string) {
	ctx.Report(span, "Use '"+r+"'"+tail, diagnostic.Fix{
		Title: "Use a type cast",
		Edits: func() []diagnostic.TextEdit {
			return []diagnostic.TextEdit{{Span: span, NewText: r}}
		},
	})
}

// castOperand renders e as the operand of a cast.
func castOperand(ctx *analysis.Context, e syntax.Expr) string {
	if astquery.NeedsParensAsUnaryOperand(e) {
		return "(" + ctx.Text(e) + ")"
	}
	return ctx.Text(e)
}

// simpleInterpolation renders an unbraced interpolation as code. An
// unquoted in-string key is a string unless it is a canonical decimal
// integer; written as is it would name a constant or another number.
func simpleInterpolation(ctx *analysis.Context, part syntax.Expr) string {
	d, ok := part.(*syntax.ArrayDimFetch)
	if !ok {
		return ctx.Text(part)
	}
	key, ok := d.Dim.(*syntax.Literal)
	if !ok || canonicalInt(key.Raw) {
		return ctx.Text(part)
	}
	return ctx.Text(d.Var) + "['" + key.Raw + "']"
}

// canonicalInt reports whether an in-string key is an integer key: decimal,
// within the int range, without a leading zero and not `-0`.
func canonicalInt(raw string) bool {
	n, err := strconv.ParseInt(raw, 10, 64)
	return err == nil && strconv.FormatInt(n, 10) == raw
}

// reReadable reports whether a settype() target can be evaluated a second
// time and read: no side effects and no `[]` append.
func reReadable(e syntax.Expr) bool {
	appends := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		if d, ok := n.(*syntax.ArrayDimFetch); ok && d.Dim == nil {
			appends = true
		}
		return !appends
	})
	return !appends && !flowquery.MayHaveSideEffects(e)
}

// definitelyDefined reports whether the local variable $name holds a value
// at the statement at: it is a parameter of the enclosing scope or assigned
// by an earlier statement of a statement list enclosing at, and the scope
// never unsets it.
func definitelyDefined(ctx *analysis.Context, at syntax.Node, name string) bool {
	scope := syntax.EnclosingVariableScope(at)
	defined := false
	for _, p := range syntax.VariableScopeParams(scope) {
		defined = defined || p.Var.Name == name
	}
	for n := at.Parent(); !defined && n != nil && n != scope; n = n.Parent() {
		if s, ok := n.(syntax.Stmt); ok {
			list, idx, _ := astquery.StmtList(ctx.File, s)
			for _, prev := range list[:max(idx, 0)] {
				if es, ok := prev.(*syntax.ExprStmt); ok {
					if a, ok := es.Expr.(*syntax.Assign); ok && isVariableNamed(a.Var, name) {
						defined = true
					}
				}
			}
		}
	}
	if !defined {
		return false
	}
	unsets := false
	find := func(n syntax.Node) bool {
		if u, ok := n.(*syntax.Unset); ok {
			for _, v := range u.Vars {
				unsets = unsets || isVariableNamed(v, name)
			}
		}
		return !unsets
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, find)
	} else {
		syntax.Inspect(scope, find)
	}
	return !unsets
}

func isVariableNamed(e syntax.Expr, name string) bool {
	got, ok := astquery.PlainVariableName(e)
	return ok && got == name
}

// guardCastResult parenthesises a cast replacing n when n is the base of an
// array/member access, a callee, the class operand of `::` or an operand of
// `**`, which bind tighter than a cast.
func guardCastResult(n syntax.Expr, r string) string {
	tighter := false
	switch p := n.Parent().(type) {
	case *syntax.ArrayDimFetch:
		tighter = p.Var == n
	case *syntax.PropertyFetch:
		tighter = p.Var == n
	case *syntax.MethodCall:
		tighter = p.Var == n
	case *syntax.FuncCall:
		tighter = p.Name == n
	case *syntax.StaticCall:
		tighter = p.Class == n
	case *syntax.ClassConstFetch:
		tighter = p.Class == n
	case *syntax.StaticPropertyFetch:
		tighter = p.Class == n
	case *syntax.Binary:
		tighter = p.Op.Kind == syntax.TPow
	}
	if tighter {
		return "(" + r + ")"
	}
	return r
}
