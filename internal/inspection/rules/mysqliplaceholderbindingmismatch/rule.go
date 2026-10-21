// Package mysqliplaceholderbindingmismatch implements the native MysqliPlaceholderBindingMismatch inspection.
package mysqliplaceholderbindingmismatch

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Bind one value for each SQL placeholder."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliPlaceholderBindingMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "mysqli_stmt", "bind_param") || c.Args == nil {
		return
	}
	producer, ok := semanticquery.NativeValue(ctx, c.Var).(*syntax.MethodCall)
	if !ok || !semanticquery.NativeMethod(ctx, producer, "mysqli", "prepare") {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(producer.Args, 0, "query"))
	if !known || ambiguousComment(sql) {
		return
	}
	count, known := sQLPositionalCount(sql)
	if !known {
		return
	}
	args := 0
	for _, node := range c.Args.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok || arg.Unpack || arg.Name != nil {
			return
		}
		args++
	}
	if count != args-1 {
		ctx.ReportNode(c, message)
	}
}

// ambiguousComment reports comment-like text the server does not skip as a
// plain comment: executable or hint comments, and a double dash that is not
// followed by whitespace. The check is textual, so the same bytes inside
// quoted data also leave the placeholder count unproven.
func ambiguousComment(sql string) bool {
	for _, opener := range []string{"/*!", "/*M!", "/*+"} {
		if strings.Contains(sql, opener) {
			return true
		}
	}
	for i := strings.Index(sql, "--"); i >= 0; {
		if i+2 < len(sql) && sql[i+2] > ' ' {
			return true
		}
		next := strings.Index(sql[i+2:], "--")
		if next < 0 {
			break
		}
		i += 2 + next
	}
	return false
}

// sQLPositionalCount counts positional placeholders outside quoted data
// and comments in the supported SQL subset. Vendor syntax stays unknown.
func sQLPositionalCount(sql string) (int, bool) {
	facts, known := semanticquery.ScanSQL(sql)
	return facts.Positional, known && len(facts.Markers) == facts.Positional
}
