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
	scope := syntax.EnclosingFuncLike(call)
	if v, ok := call.Var.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == "this" { // D6
		meth, ok := scope.(*syntax.Method)
		if !ok || meth.Modifiers.Has(syntax.TStatic) {
			return
		}
		thisSpan := v.Span()
		arrow, _ := util.FindToken(ctx.File, syntax.Span{Start: thisSpan.End, End: id.Span().Start}, syntax.TObjectOperator) // not ?-> (D1)
		kw := sivtKeyword(meth, m)
		ctx.Report(thisSpan, "Static method "+m.Name+"() called through $this; use "+kw+"::"+m.Name+"().", analysis.Fix{
			Title: "Use " + kw + "::",
			Edits: func() []analysis.TextEdit {
				return []analysis.TextEdit{
					{Span: thisSpan, NewText: kw},
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
// receiver's inferred type. custos: with a union receiver every class must
// resolve the method as static (`Db|AdapterInterface` where only `Db`
// declares a static fetchAll() is an instance call on the adapter);
// otherwise the first non-static declaration is returned.
func sivtResolve(ctx *analysis.Context, recv syntax.Expr, name string) *index.Method {
	t := ctx.TypeOf(recv)
	if t.IsUnknown() {
		return nil
	}
	ix := ctx.Index()
	var first *index.Method
	for _, cls := range t.Classes() {
		m := sivtFind(ix, strings.TrimPrefix(cls, `\`), name, ctx)
		if m == nil {
			continue
		}
		if !m.Static {
			return m
		}
		if first == nil {
			first = m
		}
	}
	if first != nil && len(t.Classes()) > 1 {
		for _, cls := range t.Classes() {
			if sivtFind(ix, strings.TrimPrefix(cls, `\`), name, ctx) == nil {
				return nil // the method's kind is unknown on that member
			}
		}
	}
	return first
}

// sivtFind looks name up through the hierarchy of cls. An abstract method
// of a used trait only states a requirement (custos diverges): the
// implementation inherited from a parent decides, and when none is found
// while an ancestor does not resolve, the method stays unknown.
func sivtFind(ix *index.Index, cls, name string, ctx *analysis.Context) *index.Method {
	lname := strings.ToLower(name)
	var required *index.Method
	for _, c := range ix.Ancestors(cls, ctx.PHP) {
		m, ok := c.Methods[lname]
		if !ok || !m.Avail.In(ctx.PHP) {
			continue
		}
		if !m.Abstract || c.Kind != syntax.KindTrait {
			return m
		}
		if required == nil {
			required = m
		}
	}
	if required != nil && !util.HierarchyResolved(ix, cls, ctx.PHP) {
		return nil
	}
	return required
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
	for _, p := range syntax.FuncLikeParams(scope) {
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

// sivtKeyword picks the class keyword for the fix. `$this->m()` resolves m
// on the runtime class; `self::m()` on the lexical class, which differs
// when a subclass overrides m (custos diverges: upstream always uses self).
// self is kept only when m cannot be overridden: a private or final method,
// or a final (or enum) enclosing class.
func sivtKeyword(meth *syntax.Method, m *index.Method) string {
	if m.Visibility == index.Private || m.Final {
		return "self"
	}
	if cl, ok := meth.Parent().(*syntax.ClassLike); ok && (cl.ClassKind == syntax.KindEnum || cl.Modifiers.Has(syntax.TFinal)) {
		return "self"
	}
	return "static"
}
