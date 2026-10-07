package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/syntax"
)

// staticInvocationViaThis flags static methods called through an object
// (`$this->make()`, `$obj->make()`).
type staticInvocationViaThis struct{}

func init() { register(staticInvocationViaThis{}) }

func (staticInvocationViaThis) ID() string { return "StaticInvocationViaThis" }

func (staticInvocationViaThis) Semantic() {}

func (staticInvocationViaThis) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall}
}

func (staticInvocationViaThis) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.MethodCall)
	if call.NullSafe || call.Span().Len() == 0 { // D1
		return
	}
	id, ok := call.Name.(*syntax.Identifier)
	if !ok || id.Value == "" || strings.HasPrefix(id.Value, "static") { // D2
		return
	}
	switch call.Var.(type) { // D3
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		return
	}
	m := sivtResolve(ctx, call.Var, id.Value) // D4
	if m == nil || !m.Static {
		return
	}
	if sivtExcluded(ctx, m) { // D5
		return
	}
	scope := util.EnclosingFuncLike(call)
	if v, ok := call.Var.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == "this" { // D6
		meth, ok := scope.(*syntax.Method)
		if !ok || meth.Modifiers.Has(syntax.TStatic) {
			return
		}
		thisSpan := v.Span()
		arrow, ok := util.FindToken(ctx.File, syntax.Span{Start: thisSpan.End, End: id.Span().Start}, syntax.TObjectOperator)
		if !ok {
			return
		}
		ctx.Report(thisSpan, "Static method "+m.Name+"() called through $this; use self::"+m.Name+"().", analysis.Fix{
			Title: "Use self::",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{
					{Span: thisSpan, NewText: "self"},
					{Span: syntax.Span{Start: arrow.Start, End: arrow.End}, NewText: "::"},
				}
			},
		})
		return
	}
	if v, ok := call.Var.(*syntax.Variable); ok && v.NameExpr == nil && scope != nil && sivtScopeInput(scope, v.Name) { // D7 / E6
		return
	}
	ctx.Report(call.Span(), "Static method "+id.Value+"() called on an instance; call it with ::.")
}

// sivtResolve resolves `recv->name()` to a method declaration via the
// receiver's inferred type (first class declaring it).
func sivtResolve(ctx *analysis.Context, recv syntax.Expr, name string) *index.Method {
	t := ctx.TypeOf(recv)
	if t.IsUnknown() {
		return nil
	}
	ix := ctx.Index()
	for _, cls := range t.Classes() {
		if m := ix.FindMethod(strings.TrimPrefix(cls, `\`), name, ctx.PHP); m != nil {
			return m
		}
	}
	return nil
}

// sivtExcluded applies EXCEPT_PHPUNIT_ASSERTIONS / EXCEPT_ELOQUENT_MODELS.
func sivtExcluded(ctx *analysis.Context, m *index.Method) bool {
	fqn := `\` + strings.TrimPrefix(m.Class, `\`)
	if ctx.Bool("EXCEPT_PHPUNIT_ASSERTIONS") && strings.HasPrefix(fqn, `\PHPUnit`) &&
		strings.HasPrefix(strings.ReplaceAll(fqn, "_", `\`), `\PHPUnit\Framework\`) {
		return true
	}
	// Symfony's PHPUnit assertion traits (WebTestAssertionsTrait,
	// MailerAssertionsTrait, …) extend the same convention.
	if ctx.Bool("EXCEPT_PHPUNIT_ASSERTIONS") && strings.HasPrefix(fqn, `\Symfony\Bundle\FrameworkBundle\Test\`) {
		return true
	}
	if ctx.Bool("EXCEPT_ELOQUENT_MODELS") && strings.EqualFold(fqn, `\Illuminate\Database\Eloquent\Model`) {
		return true
	}
	return false
}

// sivtScopeInput reports whether name is a parameter or closure `use`
// variable of the function-like scope.
func sivtScopeInput(scope syntax.Node, name string) bool {
	for _, p := range util.FuncLikeParams(scope) {
		if p.Var != nil && p.Var.Name == name {
			return true
		}
	}
	if c, ok := scope.(*syntax.Closure); ok {
		for _, u := range c.Uses {
			if u.Var != nil && u.Var.Name == name {
				return true
			}
		}
	}
	return false
}
