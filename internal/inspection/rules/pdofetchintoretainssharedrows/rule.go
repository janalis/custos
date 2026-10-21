// Package pdofetchintoretainssharedrows implements PdoFetchIntoRetainsSharedRows.
package pdofetchintoretainssharedrows

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Clone fetched objects before retaining snapshots."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoFetchIntoRetainsSharedRows" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	assign := n.(*syntax.Assign)
	target, ok := assign.Var.(*syntax.ArrayDimFetch)
	if !ok || target.Dim != nil || assign.Op.Kind != syntax.TEqual {
		return
	}
	loop := assign.Parent()
	for loop != nil {
		if _, ok := loop.(*syntax.While); ok {
			break
		}
		if syntax.IsVariableScope(loop) {
			return
		}
		loop = loop.Parent()
	}
	if loop == nil {
		return
	}
	var source *syntax.MethodCall
	var produced syntax.Expr
	while := loop.(*syntax.While)
	syntax.Inspect(while.Cond, func(n syntax.Node) bool {
		if a, ok := n.(*syntax.Assign); ok {
			if v, ok := a.Var.(*syntax.Variable); ok {
				if use, ok := assign.Value.(*syntax.Variable); ok && use.Name == v.Name {
					source, _ = a.Value.(*syntax.MethodCall)
					produced = a.Var
				}
			}
		}
		return true
	})
	if source == nil || !pdoFetchIntoMethod(ctx, source, "fetch") || !semanticquery.SameStableLocal(ctx, produced, assign.Value) {
		return
	}
	if override, state := semanticquery.CallArgumentBinding(source.Args, 0, "mode"); state != semanticquery.ArgumentMissing {
		mode, known := semanticquery.NativeContractInt(ctx, override)
		if !known || mode != 0 {
			return
		}
	}
	latest := pdoFetchIntoMode(ctx, source, while)
	if latest == nil || pdoFetchIntoSingleRow(ctx, assign, while) {
		return
	}
	mode, known := semanticquery.NativeContractInt(ctx, semanticquery.CallArgument(latest.Args, 0, "mode"))
	object := semanticquery.CallArgument(latest.Args, 1, "args")
	if known && mode == 9 && object != nil && len(ctx.TypeOf(object).Classes()) > 0 {
		ctx.ReportNode(assign, message)
	}
}

// pdoFetchIntoMethod proves a PDOStatement method call. Targets below PHP 8.0
// declare no native return type for PDO::query(), so a receiver whose latest
// local write is a dominating assignment of that builtin call is accepted too.
func pdoFetchIntoMethod(ctx *analysis.Context, c *syntax.MethodCall, name string) bool {
	if semanticquery.NativeMethod(ctx, c, "PDOStatement", name) {
		return true
	}
	id, ok := c.Name.(*syntax.Identifier)
	receiver, simple := c.Var.(*syntax.Variable)
	if !ok || !simple || !strings.EqualFold(id.Value, name) || !semanticquery.ExpansionDUnaliased(ctx, receiver, c) {
		return false
	}
	var origin *syntax.MethodCall
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		write, ok := event.(*syntax.Assign)
		if !ok || write.Span().Start >= c.Span().Start || !astquery.MentionsVariable(write.Var, receiver.Name) {
			continue
		}
		origin = nil
		if target, ok := write.Var.(*syntax.Variable); ok && target.Name == receiver.Name && !write.ByRef && write.Op.Kind == syntax.TEqual && semanticquery.NativeDominates(write, c) {
			origin, _ = write.Value.(*syntax.MethodCall)
		}
	}
	if origin == nil {
		return false
	}
	// The stub index lists PDO::query() from PHP 8.0 only: look the factory up
	// without a target version and still reject user overrides.
	factory, ok := origin.Name.(*syntax.Identifier)
	classes := ctx.Types().Native().TypeOf(origin.Var).Classes()
	if !ok || !strings.EqualFold(factory.Value, "query") || len(classes) != 1 || !ctx.Index().IsSubtype(classes[0], "PDO", ctx.PHP) {
		return false
	}
	method := ctx.Index().FindMethod(classes[0], "query", 0)
	return method != nil && method.Builtin
}

// pdoFetchIntoMode returns the setFetchMode call that governs every fetch of
// the loop. Another call on the same statement that is conditional, or that
// runs inside a loop around the fetch, leaves the mode and its target unproven.
func pdoFetchIntoMode(ctx *analysis.Context, fetch *syntax.MethodCall, loop syntax.Node) *syntax.MethodCall {
	end := loop.Span().End
	for p := loop.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach:
			end = p.Span().End
		}
	}
	var latest *syntax.MethodCall
	for _, event := range semanticquery.ScopeCallEvents(ctx, fetch) {
		c, ok := event.(*syntax.MethodCall)
		if !ok || c.Span().Start >= end {
			continue
		}
		id, ok := c.Name.(*syntax.Identifier)
		if !ok || !strings.EqualFold(id.Value, "setFetchMode") || !astquery.Equivalent(ctx.File, c.Var, fetch.Var) {
			continue
		}
		latest = nil
		if c.Span().End < fetch.Span().Start && (semanticquery.NativeDominates(c, fetch) || semanticquery.NativeCallSucceeded(ctx, c, fetch)) && pdoFetchIntoMethod(ctx, c, "setFetchMode") && semanticquery.SameStableLocal(ctx, c.Var, fetch.Var) {
			latest = c
		}
	}
	return latest
}

// pdoFetchIntoSingleRow reports whether the loop is left unconditionally after
// the append, so at most one fetched row is ever retained.
func pdoFetchIntoSingleRow(ctx *analysis.Context, assign syntax.Node, loop *syntax.While) bool {
	for n := assign.Parent(); n != loop; n = n.Parent() {
		st, ok := n.(syntax.Stmt)
		if !ok {
			continue
		}
		list, i, ok := astquery.StmtList(ctx.File, st)
		if !ok {
			continue
		}
		for _, next := range list[i+1:] {
			switch x := next.(type) {
			case *syntax.Return:
				return true
			case *syntax.Break:
				if pdoFetchIntoBreaksLoop(x, loop) {
					return true
				}
			case *syntax.ExprStmt:
				if syntax.Terminates(x) {
					return true
				}
			}
		}
	}
	return false
}

func pdoFetchIntoBreaksLoop(st *syntax.Break, loop *syntax.While) bool {
	for p := st.Parent(); p != loop; p = p.Parent() {
		switch p.(type) {
		case *syntax.While, *syntax.DoWhile, *syntax.For, *syntax.Foreach, *syntax.Switch:
			return false
		}
	}
	return true
}
