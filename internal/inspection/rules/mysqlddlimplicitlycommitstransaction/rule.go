// Package mysqlddlimplicitlycommitstransaction implements the native MysqlDdlImplicitlyCommitsTransaction inspection.
package mysqlddlimplicitlycommitstransaction

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Keep implicitly committing DDL outside rollback-dependent transactions."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqlDdlImplicitlyCommitsTransaction" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDO", "rollBack") {
		return
	}
	origin, ok := semanticquery.NativeValue(ctx, c.Var).(*syntax.New)
	if !ok {
		return
	}
	dsn, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(origin.Args, 0, "dsn"))
	if !known || !strings.HasPrefix(dsn, "mysql:") {
		return
	}
	prior := semanticquery.NativePriorCalls(ctx, c, c.Var, "beginTransaction", "exec", "commit", "rollBack")
	var ddl *syntax.MethodCall
	active := false
	for _, node := range prior {
		if call, ok := node.(*syntax.MethodCall); ok {
			if semanticquery.NativeMethod(ctx, call, "PDO", "beginTransaction") {
				active = true
				ddl = nil
				continue
			}
			if semanticquery.NativeMethod(ctx, call, "PDO", "commit") || semanticquery.NativeMethod(ctx, call, "PDO", "rollBack") {
				active = false
				ddl = nil
				continue
			}
			if active && semanticquery.NativeMethod(ctx, call, "PDO", "exec") {
				sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "statement"))
				if !known {
					return
				}
				parts := strings.Fields(strings.ToUpper(sql))
				if len(parts) >= 2 && (parts[0] == "CREATE" || parts[0] == "ALTER" || parts[0] == "DROP") && parts[1] == "TABLE" {
					ddl = call
				}
			}
		}
	}
	if ddl != nil {
		ctx.ReportNode(ddl, message)
	}
}
