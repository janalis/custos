// Package sqlnullcomparedwithequality implements the native SqlNullComparedWithEquality inspection.
package sqlnullcomparedwithequality

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use IS NULL or IS NOT NULL for SQL null checks."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SqlNullComparedWithEquality" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDO", "query") && !semanticquery.NativeMethod(ctx, c, "PDO", "exec") && !semanticquery.NativeMethod(ctx, c, "PDO", "prepare") && !semanticquery.NativeMethod(ctx, c, "mysqli", "query") && !semanticquery.NativeMethod(ctx, c, "mysqli", "prepare") {
		return
	}
	argument := semanticquery.CallArgument(c.Args, 0, "query")
	sql, known := semanticquery.NativeString(ctx, argument)
	if !known {
		return
	}
	tokens, known := sqlTokens(sql)
	if !known {
		return
	}
	predicate := false
	for i, token := range tokens {
		switch token {
		case "WHERE", "HAVING", "ON":
			predicate = true
		case "SELECT", "SET", "VALUES", "FROM", "ORDER", "GROUP", "LIMIT", "UPDATE", "INSERT", "DELETE", "CREATE", "ALTER", "DROP":
			predicate = false
		}
		if !predicate || token != "NULL" {
			continue
		}
		if i > 0 && (tokens[i-1] == "=" || tokens[i-1] == "!=" || tokens[i-1] == "<>") || i+1 < len(tokens) && tokens[i+1] == "=" {
			ctx.ReportNode(argument, message)
			return
		}
	}
}

func sqlTokens(sql string) ([]string, bool) {
	var tokens []string
	for i := 0; i < len(sql); {
		c := sql[i]
		if c == '$' || c == '@' || c == '[' || c == ']' {
			return nil, false
		}
		if c == '\'' || c == '"' || c == '`' {
			quote := c
			i++
			closed := false
			for i < len(sql) {
				if sql[i] == '\\' {
					i += 2
					continue
				}
				if sql[i] == quote {
					if i+1 < len(sql) && sql[i+1] == quote {
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				i++
			}
			if !closed {
				return nil, false
			}
			tokens = append(tokens, "QUOTED")
			continue
		}
		if c == '#' || c == '-' && i+1 < len(sql) && sql[i+1] == '-' {
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			continue
		}
		if c == '/' && i+1 < len(sql) && sql[i+1] == '*' {
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 {
				return nil, false
			}
			i += end + 4
			continue
		}
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' {
			start := i
			i++
			for i < len(sql) && (sql[i] >= 'a' && sql[i] <= 'z' || sql[i] >= 'A' && sql[i] <= 'Z' || sql[i] >= '0' && sql[i] <= '9' || sql[i] == '_') {
				i++
			}
			tokens = append(tokens, strings.ToUpper(sql[start:i]))
			continue
		}
		if c == '=' || c == '!' || c == '<' || c == '>' {
			start := i
			i++
			if i < len(sql) && (sql[i] == '=' || c == '<' && sql[i] == '>') {
				i++
			}
			tokens = append(tokens, sql[start:i])
			continue
		}
		if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			tokens = append(tokens, string(c))
		}
		i++
	}
	return tokens, true
}
