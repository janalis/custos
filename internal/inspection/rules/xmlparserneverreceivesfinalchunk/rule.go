// Package xmlparserneverreceivesfinalchunk implements the native XmlParserNeverReceivesFinalChunk inspection.
package xmlparserneverreceivesfinalchunk

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Finish parsing before freeing the XML parser."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XmlParserNeverReceivesFinalChunk" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "xml_parser_free") {
		return
	}
	e := semanticquery.CallArgument(c.Args, 0, "parser")
	if !semanticquery.ExpansionDUnaliased(ctx, e, n) {
		return
	}
	creation, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok || (!semanticquery.NativeBuiltin(ctx, creation, "xml_parser_create") && !semanticquery.NativeBuiltin(ctx, creation, "xml_parser_create_ns")) {
		return
	}
	calls := ctx.Flow().Calls(syntax.EnclosingVariableScope(n))
	if len(calls) > 256 || handedOver(ctx, e, n) {
		return
	}
	parsed := false
	for _, record := range calls {
		p, ok := record.Node.(*syntax.FuncCall)
		if !ok || p.Span().End >= n.Span().Start || !semanticquery.NativeSameValue(ctx, semanticquery.CallArgument(p.Args, 0, "parser"), e) {
			continue
		}
		if !semanticquery.NativeDominates(p, n) {
			return
		}
		switch semanticquery.NativeBuiltinName(ctx, p) {
		case "xml_parse":
			final := semanticquery.CallArgument(p.Args, 2, "is_final")
			if final != nil {
				v, k := semanticquery.NativeTruth(ctx, final)
				if !k || v {
					return
				}
			}
			parsed = true
		case "xml_parse_into_struct", "xml_parser_free":
			return
		}
	}
	if parsed {
		ctx.ReportNode(n, message)
	}
}

// handedOver reports whether the parser is passed, before the free, to code
// that may finish the parse: a project function, method, static or constructor
// call. A temporary array preserves its local binding but can still hand over
// the parser it contains. Arguments of
// nested builtin calls only read the parser and are not a hand-over.
func handedOver(ctx *analysis.Context, parser syntax.Expr, free syntax.Node) bool {
	scope := syntax.EnclosingVariableScope(free)
	found := false
	visit := func(node syntax.Node) bool {
		if found || node.Span().Start >= free.Span().Start {
			return false
		}
		var args *syntax.ArgList
		switch call := node.(type) {
		case *syntax.FuncCall:
			if semanticquery.NativeBuiltinName(ctx, call) == "" {
				args = call.Args
			}
		case *syntax.MethodCall:
			args = call.Args
		case *syntax.StaticCall:
			args = call.Args
		case *syntax.New:
			args = call.Args
		default:
			// Another function, method or closure has its own variables; an arrow
			// function shares the enclosing ones.
			if _, arrow := node.(*syntax.ArrowFunction); !arrow && node != scope && syntax.IsVariableScope(node) {
				return false
			}
		}
		if args != nil {
			syntax.Inspect(args, func(inner syntax.Node) bool {
				if nested, ok := inner.(*syntax.FuncCall); ok && semanticquery.NativeBuiltinName(ctx, nested) != "" {
					return false
				}
				found = found || astquery.Equivalent(ctx.File, inner, parser)
				return !found
			})
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}
