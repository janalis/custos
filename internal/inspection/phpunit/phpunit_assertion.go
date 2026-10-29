package phpunit

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// assertionOwners lists the classes whose `assert…` and `expect…` methods are
// PHPUnit's own checks, under their current and their pre-namespace names.
var assertionOwners = map[string]bool{
	`phpunit\framework\assert`:   true,
	`phpunit\framework\testcase`: true,
	`phpunit_framework_assert`:   true,
	`phpunit_framework_testcase`: true,
}

// PuAssertion reports that call runs an assertion or an expectation PHPUnit
// itself declares: a method named `assert…` or `expect…` that the receiver
// (`$this`, a typed object, `self`, `static`, `parent` or a class name)
// finds in PHPUnit's Assert or TestCase, or one of the `assert…` functions
// of the PHPUnit\Framework namespace. A method of that name a class of the
// project declares is the project's own. Only the name of the declaring
// class is read, so the answer is the same however many copies of PHPUnit
// the index holds; without any the call resolves to nothing.
func PuAssertion(ctx *analysis.Context, call syntax.Node) bool {
	if f, ok := call.(*syntax.FuncCall); ok {
		fn := ctx.Types().ResolveFunction(f)
		return fn != nil && strings.HasPrefix(strings.ToLower(strings.TrimPrefix(fn.FQN, `\`)), `phpunit\framework\assert`)
	}
	c, ok := AsPuCall(call)
	if !ok || !checks(c.Ident.Value) {
		return false
	}
	var classes []string
	if n, named := c.Recv.(*syntax.Name); named {
		classes = []string{PuResolveClassName(ctx, n)}
	} else {
		classes = ctx.TypeOf(c.Recv).Classes()
	}
	var found *index.Method
	for _, class := range classes {
		if found == nil {
			found = ctx.Index().FindMethod(strings.TrimPrefix(class, `\`), c.Ident.Value, ctx.PHP)
		}
	}
	return found != nil && assertionOwners[strings.ToLower(strings.TrimPrefix(found.Class, `\`))]
}

// checks reports a method name PHPUnit gives to assertions and expectations.
func checks(name string) bool {
	name = strings.ToLower(name)
	return strings.HasPrefix(name, "assert") || strings.HasPrefix(name, "expect")
}

// UnindexedAssertion reports an `assert…` or `expect…` call on `$this`,
// `self`, `static` or `parent` in a class whose ancestry leaves the index at
// PHPUnit's TestCase, none of the indexed classes on the way declaring a
// method of that name: PHPUnit is then installed outside the indexed tree
// (symfony/phpunit-bridge), and the call reaches its own assertion.
func UnindexedAssertion(ctx *analysis.Context, call syntax.Node) bool {
	c, ok := AsPuCall(call)
	if !ok || !checks(c.Ident.Value) {
		return false
	}
	switch r := c.Recv.(type) {
	case *syntax.Variable:
		if r.Name != "this" {
			return false
		}
	case *syntax.Name:
		switch strings.ToLower(r.Value) {
		case "self", "static", "parent":
		default:
			return false
		}
	default:
		return false
	}
	ix := ctx.Index()
	fqn := ctx.Types().ClassFQN(syntax.EnclosingClass(call))
	if fqn == "" || ix.FindMethod(fqn, c.Ident.Value, ctx.PHP) != nil {
		return false
	}
	for depth := 0; fqn != "" && depth < 64; depth++ {
		cls := ix.Class(fqn, ctx.PHP)
		if cls == nil {
			return strings.EqualFold(strings.TrimPrefix(fqn, `\`), `PHPUnit\Framework\TestCase`)
		}
		fqn = cls.Parent
	}
	return false
}
