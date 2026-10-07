package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/syntax"
	"custos/internal/types"
)

// unnecessaryCasting reports casts whose operand already has the target type
// and string casts used as concatenation operands.
type unnecessaryCasting struct{}

func init() { register(unnecessaryCasting{}) }

func (unnecessaryCasting) ID() string { return "UnnecessaryCasting" }

func (unnecessaryCasting) Semantic() {}

func (unnecessaryCasting) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KUnary} }

const (
	unnecessaryCastingTypeMsg   = "Operand already has the target type; remove the cast."
	unnecessaryCastingConcatMsg = "Concatenation converts to string anyway; remove the cast."
)

var castTargets = map[syntax.TokenKind]string{
	syntax.TIntCast: "int", syntax.TDoubleCast: "float", syntax.TBoolCast: "bool",
	syntax.TStringCast: "string", syntax.TArrayCast: "array",
}

func (r unnecessaryCasting) Check(ctx *analysis.Context, n syntax.Node) {
	u := n.(*syntax.Unary)
	target, ok := castTargets[u.Op.Kind]
	if !ok || u.Expr == nil || u.Op.Span.Len() == 0 || u.Expr.Span().Len() == 0 {
		return
	}
	if target == "string" { // D1 / D2
		switch p := u.Parent().(type) {
		case *syntax.Binary:
			if p.Op.Kind == syntax.TDot {
				r.report(ctx, u, unnecessaryCastingConcatMsg)
				return
			}
		case *syntax.Assign:
			if p.Op.Kind == syntax.TConcatEqual {
				r.report(ctx, u, unnecessaryCastingConcatMsg)
				return
			}
		}
	}
	a := syntax.UnwrapParens(u.Expr)
	tr := infer.NewTRules(ctx.Types())
	tr.DivisionIntOrFloat = true // `int / int` may be a float
	tr.SoundArithmetic = true    // `"4" * 100` is an int, `$unknown * 2` unknown
	ts := castStrictTypes(ctx, tr, a)
	if len(ts) != 1 || ts[0] != target { // D3 / E1
		return
	}
	if v, ok := a.(*syntax.Variable); ok && v.NameExpr == nil { // D3a
		for _, p := range syntax.FuncLikeParams(syntax.EnclosingFuncLike(a)) {
			if p.Var != nil && p.Var.Name == v.Name && p.Type == nil {
				return
			}
		}
	}
	vals, known := util.PossibleValuesKnown(ctx.File, a)
	if !known { // D3b: unknown discovery result (unstable variable)
		return
	}
	if len(vals) == 1 { // D3b
		parent, _ := util.ParentSkipParens(vals[0])
		if b, ok := parent.(*syntax.Binary); ok && b.Op.Kind == syntax.TCoalesce {
			return
		}
	}
	r.report(ctx, u, unnecessaryCastingTypeMsg)
}

func (unnecessaryCasting) report(ctx *analysis.Context, u *syntax.Unary, msg string) {
	del := syntax.Span{Start: u.Span().Start, End: u.Expr.Span().Start}
	ctx.Report(u.Op.Span, msg, analysis.Fix{
		Title: "Remove the cast",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: del}} },
	})
}

// castStrictTypes computes the normalised strict type of a (T1–T4).
func castStrictTypes(ctx *analysis.Context, tr *infer.TRules, a syntax.Expr) []string {
	env := tr.Env
	var t types.Type
	nullsafe := false
	switch x := a.(type) {
	case *syntax.PropertyFetch: // T1
		nullsafe = x.NullSafe
		t = castPrivatePropertyType(ctx, env, x)
	case *syntax.FuncCall: // T2
		if f := env.ResolveFunction(x); f != nil && f.Return != "" {
			t = tr.TypeOf(x)
		}
	case *syntax.MethodCall:
		nullsafe = x.NullSafe
		if m := castMethod(env, x.Var, x.Name, false); m != nil && m.Return != "" {
			t = tr.TypeOf(x)
		}
	case *syntax.StaticCall:
		if m := castStaticMethod(ctx, env, x); m != nil && m.Return != "" {
			t = tr.TypeOf(x)
		}
	default: // T3
		t = tr.TypeOf(a)
	}
	if t.IsUnknown() {
		return nil
	}
	if nullsafe && !t.Has("null") {
		t = infer.KnownUnion(t, types.Null)
	}
	seen := map[string]bool{}
	var out []string
	for _, at := range t.Atoms() {
		// Atoms are already normalised (integer→int, boolean→bool); a
		// remaining `\Integer` / `\Boolean` is a class, not a scalar alias.
		switch {
		case strings.Contains(at, "[]"):
			at = "array"
		case at == "true" || at == "false":
			at = "bool"
		default:
			at = strings.ToLower(at)
		}
		if !seen[at] {
			seen[at] = true
			out = append(out, at)
		}
	}
	return out
}

func castMethod(env *infer.Env, recv, name syntax.Expr, _ bool) *index.Method {
	id, ok := name.(*syntax.Identifier)
	if !ok {
		return nil
	}
	cs := env.TypeOf(recv).Classes()
	if len(cs) != 1 {
		return nil
	}
	return env.Index.FindMethod(strings.TrimPrefix(cs[0], `\`), id.Value, env.PHP)
}

func castStaticMethod(ctx *analysis.Context, env *infer.Env, x *syntax.StaticCall) *index.Method {
	id, ok := x.Name.(*syntax.Identifier)
	if !ok {
		return nil
	}
	nm, ok := x.Class.(*syntax.Name)
	if !ok {
		return nil
	}
	var cls string
	switch strings.ToLower(nm.Value) {
	case "self", "static":
		cls = env.ClassFQN(syntax.EnclosingClass(x))
	case "parent":
		if c := syntax.EnclosingClass(x); c != nil && len(c.Extends) > 0 {
			cls = ctx.Names().Class(c.Extends[0].Value, c.Span().Start)
		}
	default:
		cls = ctx.Names().Class(nm.Value, nm.Span().Start)
	}
	if cls == "" {
		return nil
	}
	return env.Index.FindMethod(cls, id.Value, env.PHP)
}

// castPrivatePropertyType: only private properties yield their own type
// (declared, else @var types united with the default value's type).
func castPrivatePropertyType(ctx *analysis.Context, env *infer.Env, x *syntax.PropertyFetch) types.Type {
	id, ok := x.Name.(*syntax.Identifier)
	if !ok {
		return types.Unknown
	}
	cs := env.TypeOf(x.Var).Classes()
	if len(cs) != 1 {
		return types.Unknown
	}
	p := env.Index.FindProperty(strings.TrimPrefix(cs[0], `\`), id.Value, env.PHP)
	if p == nil || p.Visibility != index.Private || p.Magic {
		return types.Unknown
	}
	if p.Type != "" {
		return types.FromDoc(p.Type, nil)
	}
	var def types.Type
	switch {
	case p.HasDefault:
		def = infer.LiteralTextType(p.Default)
	case !p.Promoted && !util.CtorAssignsProperty(syntax.EnclosingClass(x), id.Value):
		// No default: an untyped property holds null until written.
		def = types.Null
	}
	return infer.KnownUnion(types.FromDoc(p.DocType, nil), def)
}
