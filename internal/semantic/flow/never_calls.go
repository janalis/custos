package flow

import (
	"slices"
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// never reports a call that does not return: its callee is declared `never`,
// or is a function, or a method no subclass can replace, every path of which
// ends in a throw, an exit or such a call.
func (e *Env) never(call syntax.Node) bool {
	// A null-safe call may not happen, and `f(...)` only makes a closure.
	var list *syntax.ArgList
	switch n := call.(type) {
	case *syntax.FuncCall:
		list = n.Args
	case *syntax.MethodCall:
		if list = n.Args; conditional(n) {
			return false
		}
	case *syntax.StaticCall:
		if list = n.Args; conditional(n.Class) {
			return false
		}
	}
	if list != nil && len(list.Args) == 1 {
		if _, callable := list.Args[0].(*syntax.VariadicPlaceholder); callable {
			return false
		}
	}
	declared, name := "", ""
	if fn, ok := call.(*syntax.FuncCall); ok {
		if f := e.types.ResolveFunction(fn); f != nil {
			declared, name = f.Return, f.FQN
		}
	} else if m := e.method(call); m == nil {
		return e.unindexedTestFailure(call)
	} else {
		declared = m.Return
		if m.Final || m.Visibility == index.Private {
			name = m.Class + "::" + m.Name
		}
		if within := syntax.EnclosingClass(call); within != nil && within.ClassKind == syntax.KindTrait && !m.Abstract && strings.EqualFold(m.Class, e.types.ClassFQN(within)) {
			// A class using the trait can replace the method, whatever its
			// modifiers, with one that returns: only an abstract method
			// binds the class to the declared type.
			declared, name = "", ""
		}
	}
	if strings.EqualFold(strings.TrimPrefix(declared, `\`), "never") {
		return true
	}
	summary, ok := e.contract(strings.ToLower(name))
	return ok && summary.Never
}

// unindexedTestFailure reports `$this->fail()`, `self::markTestSkipped()`
// and the like in a class whose ancestry leaves the index at PHPUnit's
// TestCase: PHPUnit's failures throw, also where PHPUnit is installed
// outside the indexed tree (symfony/phpunit-bridge).
func (e *Env) unindexedTestFailure(call syntax.Node) bool {
	var name syntax.Expr
	switch n := call.(type) {
	case *syntax.MethodCall:
		if this, ok := n.Var.(*syntax.Variable); !ok || this.Name != "this" {
			return false
		}
		name = n.Name
	case *syntax.StaticCall:
		class, ok := n.Class.(*syntax.Name)
		if !ok || !slices.Contains([]string{"self", "static", "parent"}, strings.ToLower(class.Value)) {
			return false
		}
		name = n.Name
	}
	id, ok := name.(*syntax.Identifier)
	if !ok || !slices.Contains([]string{"fail", "marktestskipped", "marktestincomplete"}, strings.ToLower(id.Value)) {
		return false
	}
	within := syntax.EnclosingClass(call)
	if within == nil {
		return false
	}
	for fqn, depth := e.types.ClassFQN(within), 0; fqn != "" && depth < 64; depth++ {
		c := e.types.Index.Class(fqn, e.types.PHP)
		if c == nil {
			return strings.EqualFold(strings.TrimPrefix(fqn, `\`), `PHPUnit\Framework\TestCase`)
		}
		fqn = c.Parent
	}
	return false
}
