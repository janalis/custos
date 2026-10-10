// Package zipextractionexceedsconfiguredbudget checks explicit extraction budgets.
package zipextractionexceedsconfiguredbudget

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Validate ZIP extraction against configured limits."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipExtractionExceedsConfiguredBudget" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ZipArchive", "extractTo") {
		return
	}
	opened := false
	for _, fact := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		m, ok := fact.Node.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeDominates(m, c) || !semanticquery.ExpansionCSame(ctx, m.Var, c.Var) {
			continue
		}
		if semanticquery.NativeMethod(ctx, m, "ZipArchive", "open") {
			opened = true
		}
		if semanticquery.NativeMethod(ctx, m, "ZipArchive", "close") {
			opened = false
		}
	}
	if !opened {
		return
	}
	maxBytes, maxEntries := ctx.Int("maxExpandedBytes"), ctx.Int("maxEntries")
	if maxBytes <= 0 || maxEntries <= 0 {
		return
	}
	countOK, bytesOK := false, false
	var countEnd, bytesEnd uint32
	prior := flowquery.NativePriorStatements(ctx.File, c)
	if len(prior) > 256 {
		return
	}
	for _, st := range prior {
		guard, ok := st.(*syntax.If)
		if ok && guard.Else == nil && syntax.Terminates(guard.Body) {
			b, ok := syntax.UnwrapParens(guard.Cond).(*syntax.Binary)
			if ok && b.Op.Kind == syntax.TGreater {
				limit, known := semanticquery.NativeInt(ctx, b.Right)
				p, ok := b.Left.(*syntax.PropertyFetch)
				if ok && known && limit > 0 && limit <= int64(maxEntries) && semanticquery.ExpansionCSame(ctx, p.Var, c.Var) {
					if name, ok := p.Name.(*syntax.Identifier); ok && name.Value == "numFiles" {
						countOK = true
						countEnd = guard.Span().End
					}
				}
			}
		}
		loop, ok := st.(*syntax.For)
		if ok {
			if boundedSizes(ctx, loop, c.Var, maxBytes) {
				bytesOK = true
				bytesEnd = loop.Span().End
			}
		}
	}
	for _, fact := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		m, ok := fact.Node.(*syntax.MethodCall)
		if !ok || m.Span().Start >= c.Span().Start || !semanticquery.ExpansionCSame(ctx, m.Var, c.Var) {
			continue
		}
		mutation := false
		for _, name := range []string{"addFile", "addFromString", "addEmptyDir", "replaceFile", "renameName", "renameIndex", "open", "close"} {
			mutation = mutation || semanticquery.NativeMethod(ctx, m, "ZipArchive", name)
		}
		if mutation {
			if m.Span().Start > countEnd {
				countOK = false
			}
			if m.Span().Start > bytesEnd {
				bytesOK = false
			}
		}
	}

	if countOK && bytesOK {
		return
	}
	// Unresolved external validators can perform the entire check. Exclude them;
	// direct metadata reads alone do not establish either configured bound.
	for _, fact := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		call, ok := fact.Node.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeDominates(call, c) || semanticquery.NativeBuiltinName(ctx, call) != "" || call.Args == nil {
			continue
		}
		for _, node := range call.Args.Args {
			if arg, ok := node.(*syntax.Arg); ok && semanticquery.ExpansionCSame(ctx, arg.Value, c.Var) {
				return
			}
		}
	}
	ctx.ReportNode(c, message)
}

// boundedSizes recognizes a full ascending index traversal with a remaining
// budget check before every addition. This prevents both overflow and overshoot.
func boundedSizes(ctx *analysis.Context, loop *syntax.For, archive syntax.Expr, maximum int) bool {
	if len(loop.Init) != 1 || len(loop.Cond) != 1 || len(loop.Loop) != 1 {
		return false
	}
	init, ok := loop.Init[0].(*syntax.Assign)
	if !ok || init.Op.Kind != syntax.TEqual {
		return false
	}
	zero, known := semanticquery.NativeInt(ctx, init.Value)
	if !known || zero != 0 {
		return false
	}
	cond, ok := loop.Cond[0].(*syntax.Binary)
	if !ok || cond.Op.Kind != syntax.TLess || !semanticquery.ExpansionCSame(ctx, cond.Left, init.Var) {
		return false
	}
	end, ok := cond.Right.(*syntax.PropertyFetch)
	if !ok || !semanticquery.ExpansionCSame(ctx, end.Var, archive) {
		return false
	}
	name, ok := end.Name.(*syntax.Identifier)
	if !ok || name.Value != "numFiles" {
		return false
	}
	step, ok := loop.Loop[0].(*syntax.IncDec)
	if !ok || step.Op.Kind != syntax.TInc || !semanticquery.ExpansionCSame(ctx, step.Var, init.Var) {
		return false
	}
	body, ok := loop.Body.(*syntax.Block)
	if !ok || len(body.Stmts) != 5 {
		return false
	}
	assignment, ok := body.Stmts[0].(*syntax.ExprStmt)
	if !ok {
		return false
	}
	entry, ok := assignment.Expr.(*syntax.Assign)
	if !ok || entry.Op.Kind != syntax.TEqual {
		return false
	}
	stat, ok := entry.Value.(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, stat, "ZipArchive", "statIndex") || !semanticquery.ExpansionCSame(ctx, stat.Var, archive) || !semanticquery.ExpansionCSame(ctx, semanticquery.CallArgument(stat.Args, 0, "index"), init.Var) {
		return false
	}
	fail, ok := body.Stmts[1].(*syntax.If)
	if !ok || fail.Else != nil || !syntax.Terminates(fail.Body) {
		return false
	}
	failCond, ok := fail.Cond.(*syntax.Binary)
	if !ok || failCond.Op.Kind != syntax.TIsIdentical || !semanticquery.ExpansionCSame(ctx, failCond.Left, entry.Var) {
		return false
	}
	isFalse, known := astquery.BoolConst(failCond.Right)
	if !known || isFalse {
		return false
	}
	negative, ok := body.Stmts[2].(*syntax.If)
	if !ok || negative.Else != nil || !syntax.Terminates(negative.Body) {
		return false
	}
	negativeCond, ok := negative.Cond.(*syntax.Binary)
	if !ok || negativeCond.Op.Kind != syntax.TLess {
		return false
	}
	size, ok := negativeCond.Left.(*syntax.ArrayDimFetch)
	if !ok || !semanticquery.ExpansionCSame(ctx, size.Var, entry.Var) {
		return false
	}
	key, known := semanticquery.NativeString(ctx, size.Dim)
	if !known || key != "size" {
		return false
	}
	zero, known = semanticquery.NativeInt(ctx, negativeCond.Right)
	if !known || zero != 0 {
		return false
	}
	budget, ok := body.Stmts[3].(*syntax.If)
	if !ok || budget.Else != nil || !syntax.Terminates(budget.Body) {
		return false
	}
	bound, ok := budget.Cond.(*syntax.Binary)
	if !ok || bound.Op.Kind != syntax.TGreater || !semanticquery.ExpansionCSame(ctx, bound.Left, size) {
		return false
	}
	remaining, ok := bound.Right.(*syntax.Binary)
	if !ok || remaining.Op.Kind != syntax.TMinus {
		return false
	}
	limit, known := semanticquery.NativeInt(ctx, remaining.Left)
	if !known || limit <= 0 || limit > int64(maximum) {
		return false
	}
	addition, ok := body.Stmts[4].(*syntax.ExprStmt)
	if !ok {
		return false
	}
	add, ok := addition.Expr.(*syntax.Assign)
	if !ok || add.Op.Kind != syntax.TPlusEqual || !semanticquery.ExpansionCSame(ctx, add.Var, remaining.Right) || !semanticquery.ExpansionCSame(ctx, add.Value, size) {
		return false
	}
	// The accumulator must begin at zero immediately before this loop.
	previous, exists := astquery.PrevStmt(ctx.File, loop)
	if !exists {
		return false
	}
	previousExpr, ok := previous.(*syntax.ExprStmt)
	if !ok {
		return false
	}
	initial, ok := previousExpr.Expr.(*syntax.Assign)
	if !ok || initial.Op.Kind != syntax.TEqual || !semanticquery.ExpansionCSame(ctx, initial.Var, add.Var) {
		return false
	}
	zero, known = semanticquery.NativeInt(ctx, initial.Value)
	return known && zero == 0
}
