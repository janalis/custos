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
	if !sqlMethod(ctx, c) {
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
	// One predicate flag per open parenthesis: a subquery's clauses only
	// change their own level.
	levels := []bool{false}
	for i, token := range tokens {
		top := len(levels) - 1
		switch token {
		case "(":
			levels = append(levels, levels[top])
		case ")":
			if top > 0 {
				levels = levels[:top]
			}
		case "WHERE", "HAVING", "ON":
			levels[top] = true
		case "FROM":
			// IS [NOT] DISTINCT FROM is an operator, not a clause.
			if i == 0 || tokens[i-1] != "DISTINCT" {
				levels[top] = false
			}
		case "SELECT", "SET", "VALUES", "ORDER", "GROUP", "LIMIT", "UPDATE", "INSERT", "DELETE", "CREATE", "ALTER", "DROP":
			levels[top] = false
		}
		if token != "NULL" || !levels[top] {
			continue
		}
		if i > 0 && (tokens[i-1] == "=" || tokens[i-1] == "!=" || tokens[i-1] == "<>") || i+1 < len(tokens) && tokens[i+1] == "=" && !closesIsNull(tokens, i) {
			ctx.ReportNode(argument, message)
			return
		}
	}
}

// sqlMethod proves that c calls a builtin SQL-taking method of PDO or mysqli.
// The method is looked up without a target version, because the PDO::query()
// stub is not available below PHP 8.0; an override in a subclass is found
// first and is not builtin.
func sqlMethod(ctx *analysis.Context, c *syntax.MethodCall) bool {
	id, ok := c.Name.(*syntax.Identifier)
	if !ok {
		return false
	}
	name := strings.ToLower(id.Value)
	if name != "query" && name != "prepare" && name != "exec" {
		return false
	}
	classes := ctx.Types().Native().TypeOf(c.Var).Classes()
	if len(classes) != 1 {
		return false
	}
	if !ctx.Index().IsSubtype(classes[0], "PDO", ctx.PHP) && (name == "exec" || !ctx.Index().IsSubtype(classes[0], "mysqli", ctx.PHP)) {
		return false
	}
	m := ctx.Index().FindMethod(classes[0], name, 0)
	return m != nil && m.Builtin
}

// closesIsNull reports whether the NULL at i ends an IS NULL or IS NOT NULL
// operator, whose boolean result is what a following operator compares.
func closesIsNull(tokens []string, i int) bool {
	if i > 0 && tokens[i-1] == "NOT" {
		i--
	}
	return i > 0 && tokens[i-1] == "IS"
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
