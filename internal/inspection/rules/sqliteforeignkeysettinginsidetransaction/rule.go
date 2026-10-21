// Package sqliteforeignkeysettinginsidetransaction implements SqliteForeignKeySettingInsideTransaction.
package sqliteforeignkeysettinginsidetransaction

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Change foreign-key enforcement before the transaction."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqliteForeignKeySettingInsideTransaction" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	class := "SQLite3"
	if !semanticquery.NativeMethod(ctx, c, class, "exec") {
		class = "PDO"
		if !semanticquery.NativeMethod(ctx, c, class, "exec") || !semanticquery.PDODriver(ctx, c.Var, "sqlite") {
			return
		}
	}
	// active is true only while an open transaction is proven; every unknown
	// state is tracked as "not open" so it can never produce a report.
	active := false
	for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
		prior, ok := event.(*syntax.MethodCall)
		if !ok || prior.Span().End >= c.Span().Start || !semanticquery.SameStableLocal(ctx, prior.Var, c.Var) {
			continue
		}
		// A call on another branch may or may not have run before the PRAGMA.
		certain := semanticquery.NativeDominates(prior, c) || semanticquery.NativeCallSucceeded(ctx, prior, c)
		if !certain && sqliteCallLeaves(prior, c) {
			// The path through this call never reaches the PRAGMA.
			continue
		}
		name := ""
		if id, ok := prior.Name.(*syntax.Identifier); ok && semanticquery.NativeMethod(ctx, prior, class, id.Value) {
			name = strings.ToLower(id.Value)
		}
		switch name {
		case "":
			// User-defined or dynamic method: it may run any SQL.
			active = false
		case "begintransaction":
			active = certain
		case "commit", "rollback", "close", "open":
			active = false
		case "exec", "query", "querysingle", "prepare":
			tokens, known := sqliteStatementTokens(ctx, prior, class, name)
			if name == "prepare" {
				// A prepared statement runs later, at a point not tracked here.
				opened, _ := expansionDTransaction(tokens, false)
				kept, _ := expansionDTransaction(tokens, true)
				if !known || opened || !kept {
					return
				}
				continue
			}
			next, _ := expansionDTransaction(tokens, active)
			active = known && next && (certain || active)
		}
	}
	tokens, known := sqliteStatementTokens(ctx, c, class, "exec")
	if !known {
		return
	}
	_, bad := expansionDTransaction(tokens, active)
	if bad {
		ctx.ReportNode(c, message)
	}
}

// sqliteCallLeaves reports whether the statement holding prior is followed, in
// a block that does not contain at, by plain expression statements and then a
// return, throw or exit: control then leaves the function without reaching at.
// A try statement on the way out could resume execution, so it proves nothing.
func sqliteCallLeaves(prior, at syntax.Node) bool {
	left := false
	for node := prior; node != nil && !node.Span().Contains(at.Span()); node = node.Parent() {
		if _, ok := node.(*syntax.Try); ok || syntax.IsFuncLike(node) {
			return false
		}
		block, ok := node.Parent().(*syntax.Block)
		if _, plain := node.(*syntax.ExprStmt); left || !ok || !plain || block.Span().Contains(at.Span()) {
			continue
		}
		for _, st := range block.Stmts {
			if st.Span().Start < node.Span().End {
				continue
			}
			if _, ok := st.(*syntax.Return); ok {
				left = true
				break
			}
			next, ok := st.(*syntax.ExprStmt)
			if !ok {
				break
			}
			if _, throws := syntax.UnwrapParens(next.Expr).(*syntax.Throw); throws || syntax.ExitInvocation(next.Expr) {
				left = true
				break
			}
		}
	}
	return left
}

// sqliteStatementTokens tokenizes the literal SQL passed to a SQLite3 or PDO
// statement method. Only exec() is known to run a whole script: how much of a
// multi-statement string the other methods run differs between the two
// extensions, so such a string is unknown for them.
func sqliteStatementTokens(ctx *analysis.Context, call *syntax.MethodCall, class, method string) ([]string, bool) {
	param := "query"
	if class == "PDO" && method == "exec" {
		param = "statement"
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, param))
	if !known {
		return nil, false
	}
	tokens, known := semanticquery.SQLTokens(sql)
	if !known {
		return nil, false
	}
	if method != "exec" {
		for i, token := range tokens {
			if token == ";" && i+1 < len(tokens) {
				return nil, false
			}
		}
	}
	return tokens, true
}

func expansionDTransaction(tokens []string, active bool) (bool, bool) {
	start := 0
	violation := false
	for i := 0; i <= len(tokens); i++ {
		if i < len(tokens) && tokens[i] != ";" {
			continue
		}
		part := tokens[start:i]
		start = i + 1
		if len(part) == 0 {
			continue
		}
		switch part[0] {
		case "BEGIN":
			active = true
		case "COMMIT", "END", "ROLLBACK":
			active = false
		case "PRAGMA":
			if len(part) >= 4 && part[1] == "FOREIGN_KEYS" && (part[2] == "=" || part[2] == "(") {
				violation = violation || active
			}
		}
	}
	return active, violation
}
