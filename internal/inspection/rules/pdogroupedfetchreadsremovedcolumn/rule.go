// Package pdogroupedfetchreadsremovedcolumn implements PdoGroupedFetchReadsRemovedColumn.
package pdogroupedfetchreadsremovedcolumn

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Read the group column from the group key."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoGroupedFetchReadsRemovedColumn" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	fetch := n.(*syntax.ArrayDimFetch)
	if !astquery.ArrayFetchRequiresRead(fetch) {
		return
	}
	column, known := semanticquery.NativeString(ctx, fetch.Dim)
	if !known {
		return
	}
	row, ok := fetch.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	group, ok := row.Var.(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	if !semanticquery.ExpansionDPristine(ctx, group.Var, fetch) || pdoGroupedRowsRewritten(ctx, group.Var, fetch) {
		return
	}
	all, ok := semanticquery.NativeLocalValue(ctx, group.Var).(*syntax.MethodCall)
	if !ok {
		return
	}
	statement, ok := semanticquery.NativeLocalValue(ctx, all.Var).(*syntax.MethodCall)
	if !ok {
		return
	}
	legacy := pdoGroupedLegacyQuery(ctx, statement)
	if !legacy && !semanticquery.NativeMethod(ctx, statement, "PDO", "query") && !semanticquery.NativeMethod(ctx, statement, "PDO", "prepare") {
		return
	}
	if !semanticquery.NativeMethod(ctx, all, "PDOStatement", "fetchAll") && !(legacy && pdoGroupedLegacyFetchAll(ctx, all)) {
		return
	}
	modeExpr := semanticquery.CallArgument(all.Args, 0, "mode")
	mode, known := semanticquery.NativeContractInt(ctx, modeExpr)
	if b, ok := syntax.UnwrapParens(modeExpr).(*syntax.Binary); ok && b.Op.Kind == syntax.TPlus {
		left, lk := semanticquery.NativeContractInt(ctx, b.Left)
		right, rk := semanticquery.NativeContractInt(ctx, b.Right)
		mode, known = left|right, lk && rk && left&right == 0
	}
	flagsByName := false
	if b, ok := syntax.UnwrapParens(modeExpr).(*syntax.Binary); ok && (b.Op.Kind == syntax.TPlus || b.Op.Kind == syntax.TBar) {
		flagsByName = (semanticquery.ExpansionCConstant(ctx, b.Left, "PDO", "FETCH_GROUP") && semanticquery.ExpansionCConstant(ctx, b.Right, "PDO", "FETCH_ASSOC")) || (semanticquery.ExpansionCConstant(ctx, b.Right, "PDO", "FETCH_GROUP") && semanticquery.ExpansionCConstant(ctx, b.Left, "PDO", "FETCH_ASSOC"))
	}
	if !flagsByName && (!known || mode != 65538) {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(statement.Args, 0, "query"))
	if !known {
		return
	}
	columns, known := expansionDProjection(sql)
	if known && len(columns) > 0 && columns[0] == strings.ToUpper(column) {
		ctx.ReportNode(fetch, message)
	}
}

// pdoGroupedRowsRewritten reports whether the grouped rows may be changed
// through a reference before at: a by-reference foreach over them, a
// reference taken to them, or a call that receives them in a by-reference
// parameter or whose signature is unknown. Such code can restore the column.
func pdoGroupedRowsRewritten(ctx *analysis.Context, rows syntax.Expr, at syntax.Node) bool {
	name := pdoGroupedRoot(rows)
	if name == "" {
		return false
	}
	scope := syntax.EnclosingVariableScope(at)
	found := false
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= at.Span().Start || (n != scope && syntax.IsVariableScope(n)) {
			return false
		}
		switch x := n.(type) {
		case *syntax.Foreach:
			found = x.ByRef && pdoGroupedRoot(x.Expr) == name
		case *syntax.Assign:
			found = x.ByRef && pdoGroupedRoot(x.Value) == name
		case *syntax.Arg:
			if pdoGroupedRoot(x.Value) == name {
				found = semanticquery.ArgumentMayWrite(ctx.Types(), x)
			}
		}
		return !found
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return found
}

// pdoGroupedRoot names the plain variable an array element path starts from,
// or "" when the path has another base.
func pdoGroupedRoot(e syntax.Expr) string {
	for {
		switch x := syntax.UnwrapParens(e).(type) {
		case *syntax.ArrayDimFetch:
			e = x.Var
		case *syntax.Variable:
			return x.Name
		default:
			return ""
		}
	}
}

// pdoGroupedLegacyFetchAll accepts fetchAll() on the untyped result of a
// proven legacy PDO::query() call, which is a PDOStatement when it is used.
func pdoGroupedLegacyFetchAll(ctx *analysis.Context, c *syntax.MethodCall) bool {
	id, ok := c.Name.(*syntax.Identifier)
	return ok && strings.EqualFold(id.Value, "fetchAll") && len(ctx.Types().Native().TypeOf(c.Var).Classes()) == 0
}

// pdoGroupedLegacyQuery proves a PDO::query() call for targets below PHP 8.0.
// The stub index keeps only the 8.0 declaration of that method, so there the
// builtin is looked up without a version; a project override is not builtin.
func pdoGroupedLegacyQuery(ctx *analysis.Context, c *syntax.MethodCall) bool {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "query") || !ctx.PHP.Below(phpversion.PHP80) {
		return false
	}
	classes := ctx.Types().Native().TypeOf(c.Var).Classes()
	if len(classes) != 1 || !ctx.Index().IsSubtype(classes[0], "PDO", ctx.PHP) {
		return false
	}
	m := ctx.Index().FindMethod(classes[0], "query", 0)
	return m != nil && m.Builtin
}

func expansionDProjection(sql string) ([]string, bool) {
	tokens, known := semanticquery.SQLTokens(sql)
	if !known || len(tokens) < 2 || tokens[0] != "SELECT" {
		return nil, false
	}
	var columns []string
	start := 1
	for i := 1; ; i++ {
		if i < len(tokens) && tokens[i] != "FROM" && tokens[i] != "," {
			if tokens[i] == "*" || tokens[i] == "(" || tokens[i] == ")" || tokens[i] == ";" {
				return nil, false
			}
			continue
		}
		part := tokens[start:i]
		if len(part) == 0 {
			return nil, false
		}
		name := ""
		if len(part) == 1 && !strings.HasPrefix(part[0], "'") {
			name = part[0]
		} else if len(part) >= 3 && part[len(part)-2] == "AS" {
			name = part[len(part)-1]
		}
		if name == "" || strings.HasPrefix(name, "'") {
			return nil, false
		}
		name = strings.ToUpper(strings.TrimPrefix(name, "`"))
		for _, prior := range columns {
			if prior == name {
				return nil, false
			}
		}
		columns = append(columns, name)
		if i == len(tokens) || tokens[i] == "FROM" {
			return columns, true
		}
		start = i + 1
	}
}
