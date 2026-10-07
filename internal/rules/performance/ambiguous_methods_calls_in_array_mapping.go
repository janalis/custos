package performance

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// ambiguousMethodsCallsInArrayMapping reports `$map[f($x)] = f($x);` in loop
// bodies: the same call evaluated for the key and the value.
type ambiguousMethodsCallsInArrayMapping struct{}

func init() { register(ambiguousMethodsCallsInArrayMapping{}) }

func (ambiguousMethodsCallsInArrayMapping) ID() string { return "AmbiguousMethodsCallsInArrayMapping" }

func (ambiguousMethodsCallsInArrayMapping) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFor, syntax.KForeach}
}

func (ambiguousMethodsCallsInArrayMapping) Check(ctx *analysis.Context, n syntax.Node) {
	var body syntax.Stmt
	switch n := n.(type) {
	case *syntax.For:
		body = n.Body
	case *syntax.Foreach:
		body = n.Body
	}
	stmts, ok := util.BlockStmts(body) // D1
	if !ok {
		return
	}
	for _, s := range stmts { // D2
		es, ok := s.(*syntax.ExprStmt)
		if !ok {
			continue
		}
		a, ok := syntax.UnwrapParens(es.Expr).(*syntax.Assign)
		if !ok || a.Value == nil {
			continue
		}
		if _, ok := a.Var.(*syntax.ArrayDimFetch); !ok {
			continue
		}
		left := collectCalls(a.Var, false) // D3
		if len(left) == 0 {
			continue
		}
		right := collectCalls(a.Value, false) // D4
		if isCall(a.Value) {
			right = append(right, a.Value)
		}
	match:
		for _, r := range right { // D5
			if ambiguousHasSideEffects(ctx, r) {
				continue // E6: evaluating it once would change behaviour
			}
			for _, l := range left {
				if util.EquivalentFoldNames(ctx.File, l, r) { // names compared as PHP does (case-insensitive)
					ctx.ReportNode(r, "This call is repeated in the key; store its result in a local variable.")
					break match
				}
			}
		}
	}
}

func isCall(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		return true
	}
	return false
}

// collectCalls returns the function/method calls strictly below root (and
// root itself when self is set), in pre-order.
func collectCalls(root syntax.Node, self bool) []syntax.Node {
	var out []syntax.Node
	syntax.Inspect(root, func(n syntax.Node) bool {
		if (n != root || self) && isCall(n) {
			out = append(out, n)
		}
		return true
	})
	return out
}

// ambiguousImpureFunctions lists built-in functions whose result differs
// between two evaluations or that change state (internal pointers, random
// generators, clocks, streams), lower-case.
var ambiguousImpureFunctions = func() map[string]bool {
	m := map[string]bool{}
	for _, f := range strings.Fields(`next prev reset end each array_shift array_pop
		array_unshift array_push array_splice shuffle str_shuffle array_rand
		rand mt_rand random_int random_bytes lcg_value uniqid microtime hrtime time
		openssl_random_pseudo_bytes mcrypt_create_iv
		fgets fgetc fgetss fread fgetcsv fscanf fwrite fputs fputcsv readline
		stream_get_line stream_get_contents socket_read file_put_contents
		func_get_args`) {
		m[f] = true
	}
	return m
}()

// ambiguousHasSideEffects reports whether evaluating the call r (or any
// sub-expression, closures excluded) has a side effect or a result that is
// not stable between evaluations (E6): assignments, increments, call-time or
// declared by-reference arguments, and the built-ins listed above.
func ambiguousHasSideEffects(ctx *analysis.Context, r syntax.Node) bool {
	found := false
	syntax.Inspect(r, func(x syntax.Node) bool {
		if found {
			return false
		}
		switch c := x.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			return false
		case *syntax.Assign, *syntax.IncDec:
			found = true
		case *syntax.Arg:
			found = c.ByRef
		case *syntax.FuncCall:
			name, ok := c.Name.(*syntax.Name)
			if !ok {
				return true
			}
			f := ctx.Types().ResolveFunction(c)
			fqn := name.Value
			if f != nil {
				fqn = f.FQN
			}
			if ambiguousImpureFunctions[strings.ToLower(strings.TrimPrefix(fqn, `\`))] {
				found = true
			} else if f != nil && util.ArgBindsByRef(c.Args, f.Params, false) {
				found = true
			}
		}
		return !found
	})
	return found
}
