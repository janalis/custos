// Package recursivetraversalomitsdirectories implements the native RecursiveTraversalOmitsDirectories inspection.
package recursivetraversalomitsdirectories

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Include parent nodes when traversing directories."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RecursiveTraversalOmitsDirectories" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.New)
	if semanticquery.NativeConstruction(ctx, c, "RecursiveIteratorIterator") == nil {
		return
	}
	source := semanticquery.CallArgument(c.Args, 0, "iterator")
	if wrapper := semanticquery.NativeConstruction(ctx, source, "RecursiveCallbackFilterIterator"); wrapper != nil {
		source = semanticquery.CallArgument(wrapper.Args, 0, "iterator")
	}
	if semanticquery.NativeConstruction(ctx, source, "RecursiveDirectoryIterator") == nil {
		return
	}
	// An explicit mode, LEAVES_ONLY included, states which nodes the code
	// wants (E3).
	if semanticquery.CallArgument(c.Args, 1, "mode") != nil {
		return
	}
	if semanticquery.NativeConstruction(ctx, wrapperOf(c), "RegexIterator") != nil || keepsFilesOnly(ctx.File, c) { // E4
		return
	}
	ctx.ReportNode(c, message)
}

// wrapperOf returns the `new` expression that takes e as its first
// argument, or nil.
func wrapperOf(e syntax.Expr) syntax.Expr {
	arg, ok := e.Parent().(*syntax.Arg)
	if !ok {
		return nil
	}
	list := arg.Parent().(*syntax.ArgList)
	w, ok := list.Parent().(*syntax.New)
	if !ok || len(list.Args) == 0 || list.Args[0] != syntax.Expr(arg) {
		return nil
	}
	return w
}

// keepsFilesOnly reports whether the iterator is only consumed by foreach
// loops that test each entry for being a file (`->isFile()`) or for its
// extension (`->getExtension()`): directories are filtered out anyway (E4).
// The iterator is the loop's source itself or a local assigned once and
// read only as a loop source.
func keepsFilesOnly(f *syntax.File, c *syntax.New) bool {
	walk := func(visit func(syntax.Node) bool) { syntax.InspectFile(f, visit) }
	if fn := syntax.EnclosingFuncLike(c); fn != nil {
		body := syntax.FuncLikeBody(fn)
		if body == nil { // an arrow function holds no loop
			return false
		}
		walk = func(visit func(syntax.Node) bool) { syntax.Inspect(body, visit) }
	}
	name := ""
	if a, ok := c.Parent().(*syntax.Assign); ok && a.Value == syntax.Expr(c) && !a.ByRef {
		v, ok := a.Var.(*syntax.Variable)
		if !ok || v.NameExpr != nil {
			return false
		}
		name = v.Name
	}
	loops, others := 0, 0
	filtered := true
	walk(func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return false // another variable scope
		case *syntax.Foreach:
			src := syntax.UnwrapParens(x.Expr)
			v, isVar := src.(*syntax.Variable)
			if src == syntax.Expr(c) || name != "" && isVar && v.NameExpr == nil && v.Name == name {
				loops++
				filtered = filtered && filtersFiles(x)
			}
		case *syntax.Variable:
			if name != "" && x.NameExpr == nil && x.Name == name {
				others++
			}
		}
		return true
	})
	if name != "" && others != loops+1 { // the assignment plus one read per loop
		return false
	}
	return loops > 0 && filtered
}

// filtersFiles reports whether the loop body calls isFile() or
// getExtension() on the loop's value variable.
func filtersFiles(loop *syntax.Foreach) bool {
	value, ok := loop.Value.(*syntax.Variable)
	if !ok || value.NameExpr != nil {
		return false
	}
	found := false
	syntax.Inspect(loop.Body, func(n syntax.Node) bool {
		m, ok := n.(*syntax.MethodCall)
		if !ok {
			return !found
		}
		v, onValue := m.Var.(*syntax.Variable)
		id, named := m.Name.(*syntax.Identifier)
		if onValue && named && v.NameExpr == nil && v.Name == value.Name &&
			(strings.EqualFold(id.Value, "isFile") || strings.EqualFold(id.Value, "getExtension")) {
			found = true
		}
		return !found
	})
	return found
}
