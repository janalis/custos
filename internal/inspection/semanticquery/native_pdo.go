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
	facts, known := nativePrepared(ctx, c.Var)
	if !known || len(facts.markers) == 0 {
		return
	}

	array := NativeArray(ctx, CallArgument(c.Args, 0, "params"))
	report = array != nil && len(array.Items) == 0 && pdoPriorBindingsRequired(ctx, c, facts)
	if report && len(c.Args.Args) == 1 && nativeLengthFixSafe(&syntax.FuncCall{Args: c.Args}) && nativeCleanEdit(ctx, c.Args.Span()) {
		fixes = append(fixes, nativeFix(c.Args.Span(), "()", "Retain previous bindings"))
	}

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// pdoPriorBindingsRequired proves that the statement has earlier
// bindValue/bindParam calls and that each one names a marker of the prepared
// SQL. A binding for an absent marker makes a bare execute() fail, so the empty
// array that discards it is not a mistake.
func pdoPriorBindingsRequired(ctx *analysis.Context, c *syntax.MethodCall, facts sqlFacts) bool {
	prior := NativePriorCalls(ctx, c, c.Var, "bindparam", "bindvalue")
	for _, node := range prior {
		// Anything but a method call with a proven marker leaves the key empty,
		// which no marker matches.
		key := ""
		if binding, ok := node.(*syntax.MethodCall); ok {
			value := CallArgument(binding.Args, 0, "param")
			if name, known := NativeString(ctx, value); known {
				key = strings.TrimPrefix(name, ":")
			} else if position, known := NativeInt(ctx, value); known && position > 0 {
				key = strconv.FormatInt(position-1, 10)
			}
		}
		if !facts.markers[key] {
			return false
		}
	}
	return len(prior) > 0
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
	if _, known := scanNativeSQL(sql); !known {
		return
	}
	report = pdoIdentifierMarker(sql)

	if report {
		ctx.ReportNode(c, message, fixes...)
	}
}

// pdoValueFromCalls are the SQL functions whose argument list uses FROM as a
// separator in front of an ordinary value.
var pdoValueFromCalls = map[string]bool{"EXTRACT": true, "SUBSTRING": true, "SUBSTR": true, "MID": true, "TRIM": true, "OVERLAY": true}

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
	if !known || CallArgument(c.Args, 0, "params") == nil && pdoStatementCaptured(ctx, c) {
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

// pdoStatementCaptured reports whether the statement variable of an execute
// call is captured by a closure or arrow function that may already have run.
// Such a callable can bind markers outside the scope's own call list, so the
// binding set collected from that list is not complete.
func pdoStatementCaptured(ctx *analysis.Context, c *syntax.MethodCall) bool {
	receiver := ctx.SpanText(c.Var.Span())
	// A call inside a loop never reaches this point: its bindings are unknown.
	scope := syntax.EnclosingVariableScope(c)
	captured := false
	visit := func(n syntax.Node) bool {
		if !syntax.IsVariableScope(n) {
			return !captured
		}
		if n.Span().Start > c.Span().Start {
			return false
		}
		switch fn := n.(type) {
		case *syntax.Closure:
			for _, use := range fn.Uses {
				captured = captured || ctx.SpanText(use.Var.Span()) == receiver
			}
		case *syntax.ArrowFunction:
			syntax.Inspect(fn.Expr, func(m syntax.Node) bool {
				if v, ok := m.(*syntax.Variable); ok && ctx.SpanText(v.Span()) == receiver {
					captured = true
				}
				return !captured
			})
		}
		return false
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(syntax.VariableScopeBody(scope), visit)
	}
	return captured
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

	if NativeMethod(ctx, c, "PDOStatement", "rowCount") {
		facts, known := nativePrepared(ctx, c.Var)
		report = known && facts.selectQuery
	}

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
		if !known || !state.Successful || nativeFinally(ctx, ret, begin.Var, "commit", "rollback") {
			continue
		}
		if transactionClosedByExpr(ctx, ret.Expr, begin.Var) || transactionGuardedFinally(ctx, ret, begin.Var) {
			continue
		}
		ctx.ReportNode(ret, message)
		return
	}
}

// transactionCloseCall reports whether e is a PDO commit or rollBack call on
// the connection that began the transaction.
func transactionCloseCall(ctx *analysis.Context, e, receiver syntax.Expr) bool {
	c, ok := e.(*syntax.MethodCall)
	return ok && (NativeMethod(ctx, c, "PDO", "commit") || NativeMethod(ctx, c, "PDO", "rollBack")) && NativeSameValue(ctx, receiver, c.Var)
}

// transactionClosedByExpr proves that evaluating a returned expression always
// ends the transaction: the closing call is evaluated unconditionally, or by
// every arm of a ternary or match. Short-circuit right operands prove nothing.
func transactionClosedByExpr(ctx *analysis.Context, e, receiver syntax.Expr) bool {
	switch x := e.(type) {
	case *syntax.MethodCall:
		return transactionCloseCall(ctx, x, receiver)
	case *syntax.Paren:
		return transactionClosedByExpr(ctx, x.Expr, receiver)
	case *syntax.Unary:
		return transactionClosedByExpr(ctx, x.Expr, receiver)
	case *syntax.Assign:
		// `??=` only evaluates its value when the target is unset or null.
		return x.Op.Kind != syntax.TCoalesceEqual && transactionClosedByExpr(ctx, x.Value, receiver)
	case *syntax.Binary:
		if transactionClosedByExpr(ctx, x.Left, receiver) {
			return true
		}
		switch x.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TCoalesce:
			return false
		}
		return transactionClosedByExpr(ctx, x.Right, receiver)
	case *syntax.Ternary:
		if transactionClosedByExpr(ctx, x.Cond, receiver) {
			return true
		}
		return x.Then != nil && transactionClosedByExpr(ctx, x.Then, receiver) && transactionClosedByExpr(ctx, x.Else, receiver)
	case *syntax.Match:
		if transactionClosedByExpr(ctx, x.Cond, receiver) {
			return true
		}
		// An unmatched subject throws, which is not a normal return.
		for _, arm := range x.Arms {
			if !transactionClosedByExpr(ctx, arm.Body, receiver) {
				return false
			}
		}
		return len(x.Arms) > 0
	}
	return false
}

// transactionGuardedFinally accepts the cleanup idiom of a surrounding
// finally block that closes the transaction only while one is still active:
// `if ($pdo->inTransaction()) { $pdo->rollBack(); }`. Other conditions do not
// prove cleanup, and a return inside the finally block itself is not covered.
func transactionGuardedFinally(ctx *analysis.Context, ret *syntax.Return, receiver syntax.Expr) bool {
	for p := ret.Parent(); p != nil; p = p.Parent() {
		tr, ok := p.(*syntax.Try)
		if !ok || tr.Finally == nil || ret.Span().Start >= tr.Finally.Span().Start {
			continue
		}
		for _, s := range tr.Finally.Body.Stmts {
			guard, ok := s.(*syntax.If)
			if !ok {
				continue
			}
			cond, ok := guard.Cond.(*syntax.MethodCall)
			if !ok || !NativeMethod(ctx, cond, "PDO", "inTransaction") || !NativeSameValue(ctx, receiver, cond.Var) {
				continue
			}
			stmts := []syntax.Stmt{guard.Body}
			if block, ok := guard.Body.(*syntax.Block); ok {
				stmts = block.Stmts
			}
			for _, inner := range stmts {
				if st, ok := inner.(*syntax.ExprStmt); ok && transactionCloseCall(ctx, st.Expr, receiver) {
					return true
				}
			}
		}
	}
	return false
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
		if known && state.Successful && !nativeFinally(ctx, ret, begin.Var, "commit", "rollback") && !rollbackCatchEnds(ctx, ret, begin.Var) {
			ctx.ReportNode(ret, message)
			return
		}
	}
}

// rollbackCatchEnds reports whether a catch-path return ends the transaction
// itself: its expression always runs commit or rollBack on the receiver, or an
// enclosing finally block does so under an inTransaction guard on it.
func rollbackCatchEnds(ctx *analysis.Context, ret *syntax.Return, receiver syntax.Expr) bool {
	identity := ctx.Flow().Value(receiver).Identity
	on := func(e syntax.Expr, names ...string) bool {
		c, ok := e.(*syntax.MethodCall)
		if !ok || identity == 0 || ctx.Flow().Value(c.Var).Identity != identity {
			return false
		}
		for _, name := range names {
			if NativeMethod(ctx, c, "PDO", name) {
				return true
			}
		}
		return false
	}
	ends := func(e syntax.Expr) bool { return on(e, "commit", "rollBack") }
	if rollbackCatchEvaluates(ret.Expr, ends) {
		return true
	}
	for p := ret.Parent(); p != nil; p = p.Parent() {
		tr, ok := p.(*syntax.Try)
		if !ok || tr.Finally == nil {
			continue
		}
		for _, s := range tr.Finally.Body.Stmts {
			guard, ok := s.(*syntax.If)
			if !ok || !on(guard.Cond, "inTransaction") {
				continue
			}
			stmts := []syntax.Stmt{guard.Body}
			if block, ok := guard.Body.(*syntax.Block); ok {
				stmts = block.Stmts
			}
			for _, inner := range stmts {
				if st, ok := inner.(*syntax.ExprStmt); ok && ends(st.Expr) {
					return true
				}
			}
		}
	}
	return false
}

// rollbackCatchEvaluates reports whether every evaluation of e runs an
// expression accepted by ends: directly, through parentheses or a unary
// operator, in a ternary or match subject, or in each of their arms.
func rollbackCatchEvaluates(e syntax.Expr, ends func(syntax.Expr) bool) bool {
	switch x := e.(type) {
	case *syntax.Paren:
		return rollbackCatchEvaluates(x.Expr, ends)
	case *syntax.Unary:
		return rollbackCatchEvaluates(x.Expr, ends)
	case *syntax.Ternary:
		return rollbackCatchEvaluates(x.Cond, ends) ||
			x.Then != nil && rollbackCatchEvaluates(x.Then, ends) && rollbackCatchEvaluates(x.Else, ends)
	case *syntax.Match:
		if rollbackCatchEvaluates(x.Cond, ends) {
			return true
		}
		// No arm matching throws, so only the arms decide an ordinary return.
		for _, arm := range x.Arms {
			if !rollbackCatchEvaluates(arm.Body, ends) {
				return false
			}
		}
		return true
	}
	return ends(e)
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
