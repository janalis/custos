// Package pgidentifierescapedasliteral implements PgIdentifierEscapedAsLiteral.
package pgidentifierescapedasliteral

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape PostgreSQL identifiers with the identifier API."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PgIdentifierEscapedAsLiteral" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pg_escape_literal") {
		return
	}
	found := false
	unsafe := false
	for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		query, ok := record.Node.(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, query, "pg_query") || query.Span().Start < c.Span().End && !query.Span().Contains(c.Span()) {
			continue
		}
		sql := semanticquery.CallArgument(query.Args, 1, "query")
		if sql == nil {
			sql = semanticquery.CallArgument(query.Args, 0, "query")
		}
		parts, values, known := semanticquery.ExpansionSQLParts(ctx, sql)
		if !known {
			continue
		}
		for i, value := range values {
			if semanticquery.NativeValue(ctx, value) != c {
				continue
			}
			if identifierPosition(parts[i], parts[i+1]) {
				found = true
			} else {
				unsafe = true
			}
		}
	}
	if !found {
		return
	}
	var fixes []diagnostic.Fix
	if !unsafe && safeUses(ctx, c) {
		name, ok := c.Name.(*syntax.Name)
		if ok {
			replacement := semanticquery.QualifiedBuiltinFor(ctx, "pg_escape_identifier", c)
			fixes = append(fixes, edit(name.Span(), replacement))
		}
	}
	ctx.ReportNode(c, message, fixes...)
}

func identifierPosition(before, after string) bool {
	if strings.ContainsAny(before, `'\";$`) || strings.Contains(before, "--") || strings.Contains(before, "/*") {
		return false
	}
	fields := strings.Fields(strings.ToUpper(before))
	if len(fields) == 0 {
		return false
	}
	switch fields[len(fields)-1] {
	case "FROM", "JOIN", "UPDATE", "INTO", "TABLE":
	default:
		return false
	}
	if !strings.HasSuffix(before, " ") && !strings.HasSuffix(before, "\n") && !strings.HasSuffix(before, "\t") {
		return false
	}
	return after == "" || strings.HasPrefix(after, " ") || strings.HasPrefix(after, "\n") || strings.HasPrefix(after, "\t") || strings.HasPrefix(after, ";")
}

func safeUses(ctx *analysis.Context, c *syntax.FuncCall) bool {
	parent, _ := astquery.ParentSkipParens(c)
	assignment, ok := parent.(*syntax.Assign)
	if !ok {
		return true
	}
	variable, ok := assignment.Var.(*syntax.Variable)
	if !ok {
		return false
	}
	safe, budget := true, 256
	visit := func(n syntax.Node) bool {
		budget--
		if budget < 0 {
			safe = false
			return false
		}

		if assignment, ok := n.(*syntax.Assign); ok && assignment.ByRef {
			safe = false
			return false
		}
		v, ok := n.(*syntax.Variable)
		if ok && v.Name == variable.Name && syntax.EnclosingVariableScope(v) != syntax.EnclosingVariableScope(c) {
			safe = false
			return false
		}
		if !ok || v.Name != variable.Name || v == variable || v.Span().Start < c.Span().End {
			return true
		}
		producer := semanticquery.NativeValue(ctx, v)
		if producer != c {
			return true
		}
		p := v.Parent()
		for p != nil {
			if call, ok := p.(*syntax.FuncCall); ok {
				if semanticquery.NativeBuiltin(ctx, call, "pg_query") {
					sql := semanticquery.CallArgument(call.Args, 1, "query")
					if sql == nil {
						sql = semanticquery.CallArgument(call.Args, 0, "query")
					}
					parts, values, known := semanticquery.ExpansionSQLParts(ctx, sql)
					if known {
						for i, val := range values {
							if val == v && identifierPosition(parts[i], parts[i+1]) {
								return true
							}
						}
					}
				}
				break
			}
			if _, ok := p.(syntax.Stmt); ok {
				break
			}
			p = p.Parent()
		}
		safe = false
		return false
	}
	if scope := syntax.EnclosingVariableScope(c); scope != nil {
		syntax.Inspect(scope, visit)
	} else {
		syntax.InspectFile(ctx.File, visit)
	}
	return safe
}

func edit(span syntax.Span, text string) diagnostic.Fix {
	return diagnostic.Fix{Title: message, Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: text}} }}
}
