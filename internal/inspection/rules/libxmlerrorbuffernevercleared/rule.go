// Package libxmlerrorbuffernevercleared implements the native LibxmlErrorBufferNeverCleared inspection.
package libxmlerrorbuffernevercleared

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Clear collected XML errors during repeated parsing."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "LibxmlErrorBufferNeverCleared" }
func (rule) Semantic()   {}
func (rule) Flow()       {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach, syntax.KFor, syntax.KWhile}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	prior := flowquery.NativePriorStatements(ctx.File, n)
	if len(prior) == 0 {
		return
	}
	enabled := false
	for i := len(prior) - 1; i >= 0 && len(prior)-i <= 64; i-- {
		statement, ok := prior[i].(*syntax.ExprStmt)
		if !ok {
			return
		}
		if assignment, ok := statement.Expr.(*syntax.Assign); ok {
			creation, ok := assignment.Value.(*syntax.New)
			if !ok || !semanticquery.NamesGlobalClass(ctx, creation.Class, "DOMDocument") || nestedCallChildren(creation.Args) {
				return
			}
			continue
		}
		enable, ok := statement.Expr.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, enable, "libxml_use_internal_errors") {
			return
		}
		truth, known := semanticquery.NativeTruth(ctx, semanticquery.CallArgument(enable.Args, 0, "use_errors"))
		if !known || !truth {
			return
		}
		enabled = true
		break
	}
	if !enabled {
		return
	}
	var body syntax.Stmt
	var controls []syntax.Expr
	switch loop := n.(type) {
	case *syntax.Foreach:
		body = loop.Body
		if entries, known := semanticquery.NativeArrayEntries(ctx, loop.Expr); known && len(entries) == 0 {
			return
		}
		controls = append(controls, loop.Expr)
	case *syntax.For:
		body = loop.Body
		controls = append(controls, loop.Init...)
		controls = append(controls, loop.Cond...)
		controls = append(controls, loop.Loop...)
		if len(loop.Cond) > 0 {
			if value, known := semanticquery.NativeTruth(ctx, loop.Cond[len(loop.Cond)-1]); known && !value {
				return
			}
		}
	case *syntax.While:
		body = loop.Body
		controls = append(controls, loop.Cond)
		if value, known := semanticquery.NativeTruth(ctx, loop.Cond); known && !value {
			return
		}
	}
	for _, control := range controls {
		unknown := false
		syntax.Inspect(control, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
				unknown = true
			}
			return !unknown
		})
		if unknown {
			return
		}
	}
	block, ok := body.(*syntax.Block)
	if !ok {
		return
	}
	parse := false
	for _, s := range block.Stmts {
		st, ok := s.(*syntax.ExprStmt)
		if !ok {
			return
		}
		if nestedCallChildren(st.Expr) {
			return
		}
		switch c := st.Expr.(type) {
		case *syntax.FuncCall:
			if !semanticquery.NativeBuiltin(ctx, c, "simplexml_load_string") {
				return
			}
			parse = true
		case *syntax.MethodCall:
			if !semanticquery.NativeMethod(ctx, c, "DOMDocument", "loadXML") {
				return
			}
			parse = true
		default:
			return
		}
	}
	// A guaranteed immediate clear outside the loop also bounds the buffer's lifetime.
	loop := n.(syntax.Stmt)
	next, ok := astquery.NextStmt(ctx.File, loop)
	if ok {
		if s, yes := next.(*syntax.ExprStmt); yes {
			if c, yes := s.Expr.(*syntax.FuncCall); yes && semanticquery.NativeBuiltin(ctx, c, "libxml_clear_errors") {
				return
			}
		}
	}
	if parse {
		ctx.ReportNode(n, message)
	}
}

func nestedCallChildren(root syntax.Node) bool {
	unknown := false
	syntax.Inspect(root, func(n syntax.Node) bool {
		if n == root {
			return true
		}
		switch n.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.New:
			unknown = true
		}
		return !unknown
	})
	return unknown
}
