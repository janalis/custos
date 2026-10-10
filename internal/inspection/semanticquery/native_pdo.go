package semanticquery

import (
	"strconv"
	"strings"
	"unicode"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

type sqlFacts struct {
	markers     map[string]bool
	quoted      map[string]bool
	positional  int
	identifiers bool
	selectQuery bool
}

func scanNativeSQL(sql string) (sqlFacts, bool) {
	f := sqlFacts{markers: map[string]bool{}, quoted: map[string]bool{}, selectQuery: strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sql)), "SELECT")}
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
					f.quoted[text[1:]] = true
				}
			}
			previous = ""
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
			f.markers[strconv.Itoa(f.positional)] = true
			f.positional++
			i++
			continue
		}
		if c == ':' && i+1 < len(sql) && (i == 0 || sql[i-1] != ':') && (sql[i+1] == '_' || unicode.IsLetter(rune(sql[i+1]))) {
			start := i + 1
			i = start
			for i < len(sql) && (sql[i] == '_' || unicode.IsLetter(rune(sql[i])) || unicode.IsDigit(rune(sql[i]))) {
				i++
			}
			f.markers[sql[start:i]] = true
			switch previous {
			case "FROM", "JOIN", "INTO", "UPDATE", "TABLE":
				f.identifiers = true
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

func nativePrepared(ctx *analysis.Context, receiver syntax.Expr) (sqlFacts, bool) {
	c, ok := NativeValue(ctx, receiver).(*syntax.MethodCall)
	if !ok || !(NativeMethod(ctx, c, "PDO", "prepare") || NativeMethod(ctx, c, "PDO", "query")) {
		return sqlFacts{}, false
	}
	sql, known := NativeString(ctx, CallArgument(c.Args, 0, "query"))
	if !known {
		return sqlFacts{}, false
	}
	return scanNativeSQL(sql)
}

func nativeBindings(ctx *analysis.Context, c *syntax.MethodCall) (map[string]bool, bool) {
	bindings := map[string]bool{}
	if values := CallArgument(c.Args, 0, "params"); values != nil {
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
	for _, node := range NativePriorCalls(ctx, c, c.Var, "bindparam", "bindvalue") {
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

// CheckPdoExecuteArrayReplacesBindings checks the independently specified contract.
func CheckPdoExecuteArrayReplacesBindings(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return
	}
	var fixes []diagnostic.Fix
	report := false

	if !NativeMethod(ctx, c, "PDOStatement", "execute") {
		return
	}
	_, known := nativePrepared(ctx, c.Var)
	if !known {
		return
	}

	array := NativeArray(ctx, CallArgument(c.Args, 0, "params"))
	report = array != nil && len(array.Items) == 0 && len(NativePriorCalls(ctx, c, c.Var, "bindparam", "bindvalue")) > 0
	if report && len(c.Args.Args) == 1 && nativeLengthFixSafe(&syntax.FuncCall{Args: c.Args}) && nativeCleanEdit(ctx, c.Args.Span()) {
		fixes = append(fixes, nativeFix(c.Args.Span(), "()", "Retain previous bindings"))
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckPdoFetchColumnFalsyValueLoss checks the independently specified contract.
func CheckPdoFetchColumnFalsyValueLoss(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return
	}

	if !NativeMethod(ctx, c, "PDOStatement", "fetchColumn") {
		return
	}
	assignment, ok := c.Parent().(*syntax.Assign)
	if !ok {
		return
	}
	p := assignment.Parent()
	for {
		par, ok := p.(*syntax.Paren)
		if !ok {
			break
		}
		p = par.Parent()
	}
	if branch, ok := p.(*syntax.If); ok && branch.Cond != nil {
		ctx.ReportNode(assignment, message)
	}
}

// CheckPdoIdentifierPlaceholder checks the independently specified contract.
func CheckPdoIdentifierPlaceholder(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return
	}
	var fixes []diagnostic.Fix
	report := false

	if !NativeMethod(ctx, c, "PDO", "prepare") {
		return
	}
	sql, known := NativeString(ctx, CallArgument(c.Args, 0, "query"))
	if !known {
		return
	}
	facts, known := scanNativeSQL(sql)
	if !known {
		return
	}
	report = facts.identifiers

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckPdoMixedPlaceholderStyles checks the independently specified contract.
func CheckPdoMixedPlaceholderStyles(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return
	}
	var fixes []diagnostic.Fix
	report := false

	if !NativeMethod(ctx, c, "PDO", "prepare") {
		return
	}
	sql, known := NativeString(ctx, CallArgument(c.Args, 0, "query"))
	if !known {
		return
	}
	facts, known := scanNativeSQL(sql)
	if !known {
		return
	}
	report = facts.positional > 0 && len(facts.markers) > facts.positional

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckPdoPlaceholderBindingMismatch checks the independently specified contract.
func CheckPdoPlaceholderBindingMismatch(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return
	}
	var fixes []diagnostic.Fix
	report := false

	if !NativeMethod(ctx, c, "PDOStatement", "execute") {
		return
	}
	facts, known := nativePrepared(ctx, c.Var)
	if !known {
		return
	}

	bindings, known := nativeBindings(ctx, c)
	if !known {
		return
	}
	if len(bindings) != len(facts.markers) {
		report = true
	}
	for key := range facts.markers {
		if !bindings[key] {
			report = true
		}
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckPdoQuotedPlaceholder checks the independently specified contract.
func CheckPdoQuotedPlaceholder(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return
	}
	var fixes []diagnostic.Fix
	report := false

	if !NativeMethod(ctx, c, "PDOStatement", "execute") {
		return
	}
	facts, known := nativePrepared(ctx, c.Var)
	if !known {
		return
	}

	bindings, known := nativeBindings(ctx, c)
	if !known {
		return
	}
	for key := range bindings {
		if facts.quoted[key] && !facts.markers[key] {
			report = true
		}
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckPdoReferenceBindingVariableReuse checks the independently specified contract.
func CheckPdoReferenceBindingVariableReuse(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return
	}
	var fixes []diagnostic.Fix
	report := false

	if !NativeMethod(ctx, c, "PDOStatement", "execute") {
		return
	}
	_, known := nativePrepared(ctx, c.Var)
	if !known {
		return
	}

	if c.Args != nil && len(c.Args.Args) > 0 {
		return
	}
	type referenceBinding struct{ param, value string }
	variables := map[string]referenceBinding{}
	for _, node := range NativePriorCalls(ctx, c, c.Var, "bindparam") {
		binding := node.(*syntax.MethodCall)
		param, known := NativeString(ctx, CallArgument(binding.Args, 0, "param"))
		variable, ok := CallArgument(binding.Args, 1, "var").(*syntax.Variable)
		if !known || !ok {
			return
		}
		value, known := nativeScalarBinding(ctx, variable)
		if !known {
			return
		}
		if old, exists := variables[variable.Name]; exists && old.param != param && old.value != value {
			report = true
		}
		variables[variable.Name] = referenceBinding{param: param, value: value}
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckPdoSelectRowCountAssumption checks the independently specified contract.
func CheckPdoSelectRowCountAssumption(ctx *analysis.Context, n syntax.Node, message string) {
	c, ok := n.(*syntax.MethodCall)
	if !ok {
		return
	}
	var fixes []diagnostic.Fix
	report := false

	if !NativeMethod(ctx, c, "PDOStatement", "rowCount") {
		return
	}
	facts, known := nativePrepared(ctx, c.Var)
	report = known && facts.selectQuery

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// CheckTransactionEarlyReturn checks the independently specified contract.
func CheckTransactionEarlyReturn(ctx *analysis.Context, n syntax.Node, message string) {
	ret, ok := n.(*syntax.Return)
	if !ok {
		return
	}

	caught := false
	for p := ret.Parent(); p != nil; p = p.Parent() {
		if _, ok := p.(*syntax.Catch); ok {
			caught = true
			break
		}
	}
	if caught {
		return
	}
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(ret)) {
		begin, ok := call.Node.(*syntax.MethodCall)
		if !ok || !NativeMethod(ctx, begin, "PDO", "beginTransaction") || begin.Span().End >= ret.Span().Start {
			continue
		}
		state, known := ctx.Flow().StateBefore(ret, begin.Var, "begintransaction")
		if known && state.Successful && !nativeFinally(ctx, ret, begin.Var, "commit", "rollback") {
			ctx.ReportNode(ret, message)
			return
		}
	}
}

// CheckTransactionExceptionWithoutRollback checks the independently specified contract.
func CheckTransactionExceptionWithoutRollback(ctx *analysis.Context, n syntax.Node, message string) {
	ret, ok := n.(*syntax.Return)
	if !ok {
		return
	}

	caught := false
	for p := ret.Parent(); p != nil; p = p.Parent() {
		if _, ok := p.(*syntax.Catch); ok {
			caught = true
			break
		}
	}
	if !caught {
		return
	}
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(ret)) {
		begin, ok := call.Node.(*syntax.MethodCall)
		if !ok || !NativeMethod(ctx, begin, "PDO", "beginTransaction") || begin.Span().End >= ret.Span().Start {
			continue
		}
		state, known := ctx.Flow().StateBefore(ret, begin.Var, "begintransaction")
		if known && state.Successful && !nativeFinally(ctx, ret, begin.Var, "commit", "rollback") {
			ctx.ReportNode(ret, message)
			return
		}
	}
}

// nativeScalarBinding proves a scalar value at one binding point.
func nativeScalarBinding(ctx *analysis.Context, expr syntax.Expr) (string, bool) {
	if value, known := NativeInt(ctx, expr); known {
		return "integer:" + strconv.FormatInt(value, 10), true
	}
	if value, known := NativeString(ctx, expr); known {
		return "string:" + value, true
	}
	return "", false
}
