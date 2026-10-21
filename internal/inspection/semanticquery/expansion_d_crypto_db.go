package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// SameStableLocal proves a stable local variable even when its value comes from
// a parameter. Unknown calls, references and writes invalidate the relation.
func SameStableLocal(ctx *analysis.Context, a, b syntax.Expr) bool {
	if a == nil || b == nil {
		return false
	}
	facts := nativeContractScope(ctx, syntax.EnclosingVariableScope(a))
	events := facts.nodes
	if events == nil {
		return false
	}
	at := syntax.Node(b)
	if a.Span().Start > b.Span().Start {
		at = a
	}
	if !ExpansionDUnaliased(ctx, a, at) || !ExpansionDUnaliased(ctx, b, at) {
		return false
	}
	if NativeSameValue(ctx, a, b) {
		return true
	}
	if x, xk := NativeString(ctx, a); xk {
		if y, yk := NativeString(ctx, b); yk && x == y {
			return true
		}
	}
	if x, xk := NativeContractInt(ctx, a); xk {
		if y, yk := NativeContractInt(ctx, b); yk && x == y {
			return true
		}
	}
	av, aok := syntax.UnwrapParens(a).(*syntax.Variable)
	bv, bok := syntax.UnwrapParens(b).(*syntax.Variable)
	if !aok || !bok || av.Name != bv.Name || syntax.EnclosingVariableScope(a) != syntax.EnclosingVariableScope(b) {
		return false
	}
	start, end := a.Span().End, b.Span().Start
	if start > end {
		start, end = b.Span().End, a.Span().Start
	}
	// Only a write of the variable itself separates the two reads; using it
	// as an index or a property name of another target does not.
	if facts.localWrites(ctx, syntax.EnclosingVariableScope(a)).rebound(av.Name, start, end) {
		return false
	}
	for _, event := range events {
		if event.Span().Start <= start || event.Span().End >= end {
			continue
		}
		switch e := event.(type) {
		case *syntax.FuncCall:
			if NativeBuiltinName(ctx, e) == "" && astquery.MentionsVariable(e.Args, av.Name) {
				return false
			}
		case *syntax.MethodCall:
			if astquery.MentionsVariable(e.Args, av.Name) {
				return false
			}
		}
	}
	return true
}

// ScopeCallEvents returns the assignments, unsets and function and method
// calls of the scope of at, nested scopes excluded, cached per scope; none
// when there are more than 512.
func ScopeCallEvents(ctx *analysis.Context, at syntax.Node) []syntax.Node {
	return nativeContractEvents(ctx, syntax.EnclosingVariableScope(at))
}

// KnownByteLength returns the length in bytes of e when it is known: a
// string, or a local holding random_bytes() or str_repeat() of known size.
func KnownByteLength(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	if text, ok := NativeString(ctx, e); ok {
		return int64(len(text)), true
	}
	c, ok := NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok {
		return 0, false
	}
	switch NativeBuiltinName(ctx, c) {
	case "random_bytes":
		size, known := NativeContractInt(ctx, CallArgument(c.Args, 0, "length"))
		return size, known && size > 0
	case "str_repeat":
		text, known := NativeString(ctx, CallArgument(c.Args, 0, "string"))
		count, ok := NativeContractInt(ctx, CallArgument(c.Args, 1, "times"))
		if known && ok && count >= 0 && count <= 1<<30 {
			return int64(len(text)) * count, true
		}
	}
	return 0, false
}

// OriginCall returns the call to one of the builtins names that the local e
// holds, when e has no alias that could have changed it.
func OriginCall(ctx *analysis.Context, e syntax.Expr, names ...string) *syntax.FuncCall {
	if e == nil || !ExpansionDUnaliased(ctx, e, e) {
		return nil
	}
	c, ok := NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok {
		return nil
	}
	for _, name := range names {
		if NativeBuiltin(ctx, c, name) {
			return c
		}
	}
	return nil
}

// IsBuiltinConst reports whether e reads the builtin constant name.
func IsBuiltinConst(ctx *analysis.Context, e syntax.Expr, name string) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.ConstFetch)
	if !ok {
		return false
	}
	resolved := GlobalConstName(ctx, c)
	symbol := ctx.Index().Constant(resolved, ctx.PHP)
	return resolved == name && symbol != nil && symbol.Builtin
}

// CurlSetting returns the value c sets for the cURL option constant option,
// through curl_setopt() or a literal curl_setopt_array(); nil otherwise.
func CurlSetting(ctx *analysis.Context, c *syntax.FuncCall, option string) syntax.Expr {
	if NativeBuiltin(ctx, c, "curl_setopt") && IsBuiltinConst(ctx, CallArgument(c.Args, 1, "option"), option) {
		return CallArgument(c.Args, 2, "value")
	}
	if NativeBuiltin(ctx, c, "curl_setopt_array") {
		a := NativeArray(ctx, CallArgument(c.Args, 1, "options"))
		if a == nil {
			return nil
		}
		for _, item := range a.Items {
			if IsBuiltinConst(ctx, item.Key, option) {
				return item.Value
			}
		}
	}
	return nil
}

// IsFalse reports whether e is the constant false.
func IsFalse(e syntax.Expr) bool {
	v, k := astquery.BoolConst(syntax.UnwrapParens(e))
	return k && !v
}

// SQLTokens tokenizes a deliberately bounded SQL subset. Strings and
// comments never become keywords or placeholders; uncertain syntax is refused.
func SQLTokens(sql string) ([]string, bool) {
	if len(sql) > 65536 {
		return nil, false
	}
	var out []string
	for i := 0; i < len(sql); {
		ch := sql[i]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			i++
			continue
		}
		if ch == '$' {
			return nil, false
		}
		if ch == '#' || (ch == '-' && i+1 < len(sql) && sql[i+1] == '-') {
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			continue
		}
		if ch == '/' && i+1 < len(sql) && sql[i+1] == '*' {
			end := strings.Index(sql[i+2:], "*/")
			if end < 0 || strings.Contains(sql[i+2:i+2+end], "/*") {
				return nil, false
			}
			i += end + 4
			continue
		}
		if ch == '\'' || ch == '"' || ch == '`' {
			q := ch
			i++
			var text strings.Builder
			closed := false
			for i < len(sql) {
				if sql[i] == '\\' {
					if i+1 >= len(sql) {
						return nil, false
					}
					text.WriteByte(sql[i+1])
					i += 2
					continue
				}
				if sql[i] == q {
					if i+1 < len(sql) && sql[i+1] == q {
						text.WriteByte(q)
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				text.WriteByte(sql[i])
				i++
			}
			if !closed {
				return nil, false
			}
			prefix := "'"
			if q == '`' {
				prefix = "`"
			}
			out = append(out, prefix+text.String())
			continue
		}
		if ch == ':' && i+1 < len(sql) && sql[i+1] == ':' {
			out = append(out, "::")
			i += 2
			continue
		}
		if ch == ':' || ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' {
			start := i
			i++
			for i < len(sql) {
				c := sql[i]
				if c != '_' && !(c >= 'a' && c <= 'z') && !(c >= 'A' && c <= 'Z') && !(c >= '0' && c <= '9') {
					break
				}
				i++
			}
			word := sql[start:i]
			if ch != ':' {
				word = strings.ToUpper(word)
			}
			out = append(out, word)
			continue
		}
		out = append(out, string(ch))
		i++
	}
	return out, true
}

// EarlierMethodCalls returns the calls, before at in its scope, of the
// methods names of class on the same stable local as receiver that run
// before at: they dominate it or at runs only when they succeeded.
func EarlierMethodCalls(ctx *analysis.Context, at, receiver syntax.Expr, class string, names ...string) []*syntax.MethodCall {
	var out []*syntax.MethodCall
	for _, event := range ScopeCallEvents(ctx, at) {
		c, ok := event.(*syntax.MethodCall)
		if !ok || c.Span().End >= at.Span().Start || (!NativeDominates(c, at) && !NativeCallSucceeded(ctx, c, at)) || !SameStableLocal(ctx, c.Var, receiver) {
			continue
		}
		for _, name := range names {
			if NativeMethod(ctx, c, class, name) {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// PDODriver reports whether e holds a PDO created with a DSN for driver.
func PDODriver(ctx *analysis.Context, e syntax.Expr, driver string) bool {
	construct := NativeConstruction(ctx, e, "PDO")
	if construct == nil {
		return false
	}
	dsn, known := NativeString(ctx, CallArgument(construct.Args, 0, "dsn"))
	return known && strings.HasPrefix(strings.ToLower(dsn), driver+":")
}
