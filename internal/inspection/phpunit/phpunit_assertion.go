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
