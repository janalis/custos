package semanticquery

import (
	"strconv"
	"strings"
	"unicode"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// SQLFacts is what ScanSQL reads in an SQL string: its named markers, the
// names written as quoted strings, how many positional markers it has,
// whether a marker stands for an identifier, and whether it is a SELECT.
type SQLFacts struct {
	Markers     map[string]bool
	Quoted      map[string]bool
	Positional  int
	Identifiers bool
	SelectQuery bool
}

// ScanSQL reads the placeholders of sql, skipping comments and quoted
// text; false when part of it may not read as written: a dollar sign, an
// unclosed string or comment, a nested or executable vendor comment, `??`,
// or `--` with no space after it.
func ScanSQL(sql string) (SQLFacts, bool) {
	f := SQLFacts{Markers: map[string]bool{}, Quoted: map[string]bool{}, SelectQuery: strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "SELECT")}
	previous := ""
	for i := 0; i < len(sql); {
		c := sql[i]
		if c == '$' {
			return f, false
		}
		if c == '\'' || c == '"' || c == '`' {
			quote := c
			start := i + 1
			i++
			var value strings.Builder
			closed := false
			for i < len(sql) {
				if sql[i] == '\\' {
					i += 2
					continue
				}
				if sql[i] == quote {
					if i+1 < len(sql) && sql[i+1] == quote {
						value.WriteByte(quote)
						i += 2
						continue
					}
					closed = true
					i++
					break
				}
				value.WriteByte(sql[i])
				i++
			}
			if !closed {
				return f, false
			}
			if quote != '`' && i > start {
				text := value.String()
				if strings.HasPrefix(text, ":") && nativeSQLName(text[1:]) {
					f.Quoted[text[1:]] = true
				}
			}
			previous = ""
			continue
		}
		if c == '-' && i+1 < len(sql) && sql[i+1] == '-' && i+2 < len(sql) && sql[i+2] > ' ' {
			// SQL dialects disagree on whether this opens a comment.
			return f, false
		}
		if c == '#' || c == '-' && i+1 < len(sql) && sql[i+1] == '-' {
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			continue
		}
		if c == '/' && i+1 < len(sql) && sql[i+1] == '*' {
			if i+2 < len(sql) && (sql[i+2] == '!' || strings.HasPrefix(sql[i+2:], "M!")) {
				// Executable vendor comments depend on the server and version.
				return f, false
			}
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 || strings.Contains(sql[i+2:i+2+end], "/*") {
				return f, false
			}
			i += end + 4
			continue
		}
		if c == '?' {
			if i+1 < len(sql) && sql[i+1] == '?' {
				return f, false
			}
			f.Markers[strconv.Itoa(f.Positional)] = true
			f.Positional++
			i++
			continue
		}
		if c == ':' && i+1 < len(sql) && (i == 0 || sql[i-1] != ':') && (sql[i+1] == '_' || unicode.IsLetter(rune(sql[i+1]))) {
			start := i + 1
			i = start
			for i < len(sql) && (sql[i] == '_' || unicode.IsLetter(rune(sql[i])) || unicode.IsDigit(rune(sql[i]))) {
				i++
			}
			f.Markers[sql[start:i]] = true
			switch previous {
			case "FROM", "JOIN", "INTO", "UPDATE", "TABLE":
				f.Identifiers = true
			}
			previous = ""
			continue
		}
		if unicode.IsLetter(rune(c)) {
			start := i
			for i < len(sql) && (unicode.IsLetter(rune(sql[i])) || unicode.IsDigit(rune(sql[i])) || sql[i] == '_') {
				i++
			}
			previous = strings.ToUpper(sql[start:i])
			continue
		}
		if !unicode.IsSpace(rune(c)) {
			previous = ""
		}
		i++
	}
	return f, true
}

func nativeSQLName(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if c != '_' && !unicode.IsLetter(c) && !(i > 0 && unicode.IsDigit(c)) {
			return false
		}
	}
	return true
}

// PreparedSQL returns the facts of the SQL that the PDO statement receiver
// was prepared or queried with, when that is a known string.
func PreparedSQL(ctx *analysis.Context, receiver syntax.Expr) (SQLFacts, bool) {
	c, ok := NativeValue(ctx, receiver).(*syntax.MethodCall)
	if !ok || !(NativeMethod(ctx, c, "PDO", "prepare") || NativeMethod(ctx, c, "PDO", "query")) {
		return SQLFacts{}, false
	}
	sql, known := NativeString(ctx, CallArgument(c.Args, 0, "query"))
	if !known {
		return SQLFacts{}, false
	}
	return ScanSQL(sql)
}

// ExecuteBindings returns the names an execute() call c binds: the keys of
// the array it passes, or else those bound on its statement beforehand;
// false when they are not known.
func ExecuteBindings(ctx *analysis.Context, c *syntax.MethodCall) (map[string]bool, bool) {
	bindings := map[string]bool{}
	values, state := CallArgumentBinding(c.Args, 0, "params")
	if state == ArgumentUnknown {
		return nil, false
	}
	if state == ArgumentBound {
		entries, known := NativeArrayEntries(ctx, values)
		if !known {
			return nil, false
		}
		for key := range entries {
			bindings[strings.TrimPrefix(key[2:], ":")] = true
		}
		return bindings, true
	}
	if _, known := ctx.Flow().StateBefore(c, c.Var); !known {
		return nil, false
	}
	prior, historyKnown := NativeCallHistory(ctx, c, c.Var, "bindparam", "bindvalue").Certain()
	if !historyKnown {
		return nil, false
	}
	for _, node := range prior {
		binding := node.(*syntax.MethodCall)
		value := CallArgument(binding.Args, 0, "param")
		if name, known := NativeString(ctx, value); known {
			bindings[strings.TrimPrefix(name, ":")] = true
		} else if position, known := NativeInt(ctx, value); known && position > 0 {
			bindings[strconv.FormatInt(position-1, 10)] = true
		} else {
			return nil, false
		}
	}
	return bindings, true
}
