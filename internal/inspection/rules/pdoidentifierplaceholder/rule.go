// Package pdoidentifierplaceholder implements the native PdoIdentifierPlaceholder inspection.
package pdoidentifierplaceholder

import (
	"strings"
	"unicode"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Select SQL identifiers from a trusted allowlist."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoIdentifierPlaceholder" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "PDO", "prepare") {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "query"))
	if !known {
		return
	}
	if _, known := semanticquery.ScanSQL(sql); !known {
		return
	}
	if pdoIdentifierMarker(sql) {
		ctx.ReportNode(c, message)
	}
}

// pdoIdentifierMarker reports whether SQL already accepted by scanNativeSQL has
// a named or positional marker where a table name is required: directly after
// JOIN, INTO, UPDATE or TABLE, or after a FROM that opens a table reference.
// FROM does not open one in IS [NOT] DISTINCT FROM, nor directly inside the
// parentheses of a function listed in pdoValueFromCalls.
func pdoIdentifierMarker(sql string) bool {
	var openers []string
	previous, before := "", ""
	for i := 0; i < len(sql); {
		c := sql[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			for i++; i < len(sql); i++ {
				if sql[i] == '\\' {
					i++
				} else if sql[i] == c {
					if i+1 == len(sql) || sql[i+1] != c {
						break
					}
					i++
				}
			}
			i++
			previous, before = "", ""
		case c == '#' || c == '-' && strings.HasPrefix(sql[i:], "--"):
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
		case strings.HasPrefix(sql[i:], "/*"):
			_, rest, _ := strings.Cut(sql[i+2:], "*/")
			i = len(sql) - len(rest)
		case c == '?' || c == ':' && i+1 < len(sql) && (i == 0 || sql[i-1] != ':') && (sql[i+1] == '_' || unicode.IsLetter(rune(sql[i+1]))):
			switch previous {
			case "JOIN", "INTO", "UPDATE", "TABLE":
				return true
			case "FROM":
				if before != "DISTINCT" && (len(openers) == 0 || !pdoValueFromCalls[openers[len(openers)-1]]) {
					return true
				}
			}
			i++
			for c == ':' && i < len(sql) && (sql[i] == '_' || unicode.IsLetter(rune(sql[i])) || unicode.IsDigit(rune(sql[i]))) {
				i++
			}
			previous, before = "", ""
		case unicode.IsLetter(rune(c)):
			start := i
			for i < len(sql) && (unicode.IsLetter(rune(sql[i])) || unicode.IsDigit(rune(sql[i])) || sql[i] == '_') {
				i++
			}
			before, previous = previous, strings.ToUpper(sql[start:i])
		case unicode.IsSpace(rune(c)):
			i++
		default:
			if c == '(' {
				openers = append(openers, previous)
			} else if c == ')' && len(openers) > 0 {
				openers = openers[:len(openers)-1]
			}
			previous, before = "", ""
			i++
		}
	}
	return false
}

// pdoValueFromCalls are the SQL functions whose argument list uses FROM as a
// separator in front of an ordinary value.
var pdoValueFromCalls = map[string]bool{"EXTRACT": true, "SUBSTRING": true, "SUBSTR": true, "MID": true, "TRIM": true, "OVERLAY": true}
