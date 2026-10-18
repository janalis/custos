// Package zipentrynamecollision implements the native ZipEntryNameCollision inspection.
package zipentrynamecollision

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Make ZIP entry replacement explicit."

// invalidators are the ZipArchive methods after which the guarded entry may no longer exist under its name.
var invalidators = []string{
	"close", "open", "deleteName", "deleteIndex", "renameName", "renameIndex",
	"unchangeAll", "unchangeArchive", "unchangeName", "unchangeIndex",
}

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipEntryNameCollision" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ZipArchive", "addFromString") {
		return
	}
	name, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "name"))
	if !known || semanticquery.CallArgument(c.Args, 2, "flags") != nil {
		return
	}
	target := ctx.Flow().Value(c.Var)
	// A callback body runs when it is invoked, not where it is written.
	for p, scope := c.Parent(), syntax.EnclosingVariableScope(c); p != scope; p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(c.Span()) {
			continue
		}
		add, ok := syntax.UnwrapParens(branch.Cond).(*syntax.MethodCall)
		if !ok || !semanticquery.NativeMethod(ctx, add, "ZipArchive", "addFromString") {
			continue
		}
		previous, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(add.Args, 0, "name"))
		if !known || previous != name {
			continue
		}
		// The guard and the second insertion must act on the same archive object, not merely the same variable.
		guard := ctx.Flow().Value(add.Var)
		same := false
		switch {
		case target.Identity == 0 && guard.Identity == 0:
			same = undisturbed(ctx, branch.Body, c, add)
		case target.Complete && guard.Complete && guard.Identity == target.Identity && guard.Invalidated == target.Invalidated:
			same = guardSupportsCollision(ctx, c, add)
		}
		if same {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// undisturbed handles an archive without a flow identity, such as a property
// or a factory result. The same receiver spelling proves the same archive only
// when nothing between the guard and the second insertion can rebind it or
// change its entries: no write to the receiver or to an object it is read
// from, and no call other than a builtin function.
func undisturbed(ctx *analysis.Context, body syntax.Node, c, add *syntax.MethodCall) bool {
	if !semanticquery.ExpansionCSame(ctx, c.Var, add.Var) {
		return false
	}
	chain := []syntax.Expr{}
	for e := c.Var; e != nil; {
		chain = append(chain, e)
		switch x := e.(type) {
		case *syntax.PropertyFetch:
			e = x.Var
		case *syntax.Variable, *syntax.StaticPropertyFetch:
			e = nil
		default:
			return false
		}
	}
	clean := true
	syntax.Inspect(body, func(n syntax.Node) bool {
		if n.Span().End > c.Span().Start {
			// Ancestors of the insertion are entered; the insertion itself and later code are not.
			return n.Span().Start < c.Span().Start
		}
		switch x := n.(type) {
		case *syntax.MethodCall, *syntax.StaticCall, *syntax.New, *syntax.Closure, *syntax.ArrowFunction:
			clean = false
		case *syntax.FuncCall:
			clean = clean && semanticquery.NativeBuiltinName(ctx, x) != ""
		case *syntax.Assign:
			for _, e := range chain {
				clean = clean && !semanticquery.ExpansionCSame(ctx, x.Var, e)
			}
		}
		return clean
	})
	return clean
}

// guardSupportsCollision retains the earlier insertion as collision evidence
// unless a mandatory remover intervenes or the history is incomplete.
func guardSupportsCollision(ctx *analysis.Context, c, add *syntax.MethodCall) bool {
	history := semanticquery.NativeCallHistory(ctx, c, c.Var, invalidators...)
	if !history.Complete {
		return false
	}
	for _, event := range history.Events {
		if event.Must && event.Node.Span().Start >= add.Span().End {
			return false
		}
	}
	return true
}
