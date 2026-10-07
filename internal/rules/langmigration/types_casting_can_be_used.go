package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// typesCastingCanBeUsed reports conversion helpers, single-interpolation
// strings and explicit __toString() calls that a cast expresses directly.
type typesCastingCanBeUsed struct{}

func init() { register(typesCastingCanBeUsed{}) }

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
	args, ok := util.CallArgValues(call) // spread arguments are skipped
	if !ok {
		return
	}
	if name == "settype" {
		if len(args) != 2 || !isBuiltinSettype(ctx, call) { // D4, D5
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
		v := ctx.Text(args[0])
		reportCast(ctx, call.Span(), v+" = ("+typ+")"+v, " instead (a cast is clearer and faster).") // D7
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
	part := s.Parts[0]
	if _, ok := part.(*syntax.StringPart); ok {
		return
	}
	ps := part.Span()
	var r string
	switch {
	case ctx.Src[ps.Start] == '$' && ps.Len() > 1 && ctx.Src[ps.Start+1] == '{':
		return // "${name}": not reported
	case ps.Start > 0 && ctx.Src[ps.Start-1] == '{': // D9, braced
		r = "(string)(" + ctx.Text(part) + ")"
	default:
		r = "(string)" + ctx.Text(part)
	}
	reportCast(ctx, sp, guardCastResult(s, r), " to make the string conversion explicit.")
}

func (typesCastingCanBeUsed) checkToString(ctx *analysis.Context, m *syntax.MethodCall) {
	id, ok := m.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "__toString") || m.NullSafe { // D10; nullsafe keeps its null-safety
		return
	}
	r := "(string)" + castOperand(ctx, m.Var) // D11
	reportCast(ctx, m.Span(), guardCastResult(m, r), " instead of calling __toString() directly.")
}

func reportCast(ctx *analysis.Context, span syntax.Span, r, tail string) {
	ctx.Report(span, "Use '"+r+"'"+tail, analysis.Fix{
		Title: "Use a type cast",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: r}}
		},
	})
}

// castOperand renders e as the operand of a cast.
func castOperand(ctx *analysis.Context, e syntax.Expr) string {
	if util.NeedsParensAsUnaryOperand(e) {
		return "(" + ctx.Text(e) + ")"
	}
	return ctx.Text(e)
}

// guardCastResult parenthesises a cast replacing n when n is the base of an
// array/member access or an operand of `**`, which bind tighter than a cast.
func guardCastResult(n syntax.Expr, r string) string {
	switch p := n.Parent().(type) {
	case *syntax.ArrayDimFetch:
		if p.Var == n {
			return "(" + r + ")"
		}
	case *syntax.PropertyFetch:
		if p.Var == n {
			return "(" + r + ")"
		}
	case *syntax.MethodCall:
		if p.Var == n {
			return "(" + r + ")"
		}
	case *syntax.Binary:
		if p.Op.Kind == syntax.TPow {
			return "(" + r + ")"
		}
	}
	return r
}

// isBuiltinSettype reports whether call targets the global settype(): not
// qualified into another namespace, not imported from one, and not shadowed
// by a settype() declared in the call's namespace.
func isBuiltinSettype(ctx *analysis.Context, call *syntax.FuncCall) bool {
	if !ctx.IsGlobalFunctionCall(call, "settype") {
		return false
	}
	name := call.Name.(*syntax.Name)
	if name.NameKind != syntax.NameUnqualified {
		return true
	}
	ns := ctx.Names().Namespace(name.Span().Start)
	if ns == "" {
		return true
	}
	shadowed := false
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		if shadowed {
			return false
		}
		if fn, ok := n.(*syntax.Function); ok && fn.Name != nil && strings.EqualFold(fn.Name.Value, "settype") &&
			strings.EqualFold(ctx.Names().Namespace(fn.Span().Start), ns) {
			shadowed = true
		}
		return true
	})
	return !shadowed
}
