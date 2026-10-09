package phpunit

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// PuCall is the shape shared by instance (`->`, `?->`) and static (`::`)
// method calls.
type PuCall struct {
	Node  syntax.Expr
	Recv  syntax.Expr // receiver expression or class reference
	Ident *syntax.Identifier
	Name  string // method name, in its declared spelling when known (puName)
	Args  *syntax.ArgList
}

// puKnownNames lists the PHPUnit methods these rules match by name.
var knownNames = func() map[string]string {
	known := map[string]string{}
	for _, n := range []string{
		"expects", "will", "willReturn", "method", "exactly", "once", "any",
		"returnValue", "returnValueMap", "returnCallback", "returnArgument",
		"assert", "assertTrue", "assertFalse", "assertNotTrue", "assertNotFalse",
		"assertEquals", "assertNotEquals", "assertSame", "assertNotSame",
		"assertNull", "assertEmpty", "assertInstanceOf", "assertInternalType",
		"assertFileEquals", "assertStringEqualsFile",
		"assertFileNotExists", "assertDirectoryNotExists",
	} {
		known[strings.ToLower(n)] = n
	}
	return known
}()

// PuName returns the declared spelling of a known PHPUnit method name
// written in any case (PHP method names are case-insensitive), or the name
// as written when it is not one the rules match.
func PuName(written string) string {
	if n, ok := knownNames[strings.ToLower(written)]; ok {
		return n
	}
	return written
}

// AsPuCall views e as a method call with a literal name.
func AsPuCall(e syntax.Node) (PuCall, bool) {
	switch c := e.(type) {
	case *syntax.MethodCall:
		if id, ok := c.Name.(*syntax.Identifier); ok && c.Args != nil {
			return PuCall{Node: c, Recv: c.Var, Ident: id, Name: PuName(id.Value), Args: c.Args}, true
		}
	case *syntax.StaticCall:
		if id, ok := c.Name.(*syntax.Identifier); ok && c.Args != nil {
			return PuCall{Node: c, Recv: c.Class, Ident: id, Name: PuName(id.Value), Args: c.Args}, true
		}
	}
	return PuCall{}, false
}

// Values returns the plain positional argument values (nil, false when the
// list uses spreads, named arguments or placeholders).
func (c PuCall) Values() ([]syntax.Expr, bool) { return astquery.ArgValues(c.Args) }

// PuMethodCallNamed reports whether e (no parentheses stripped) is a method
// call whose name is one of names (case-insensitive, as PHP compares
// method names).
func PuMethodCallNamed(e syntax.Node, names ...string) (PuCall, bool) {
	c, ok := AsPuCall(e)
	if !ok {
		return PuCall{}, false
	}
	for _, n := range names {
		if strings.EqualFold(c.Ident.Value, n) {
			return c, true
		}
	}
	return PuCall{}, false
}

// PuResolveClassName resolves a class reference name (self/static/parent
// included) to an FQN without leading backslash ("" when unresolvable).
func PuResolveClassName(ctx *analysis.Context, n *syntax.Name) string {
	switch strings.ToLower(n.Value) {
	case "self", "static":
		return ctx.Types().ClassFQN(syntax.EnclosingClass(n))
	case "parent":
		return ctx.Names().ParentFQN(syntax.EnclosingClass(n))
	}
	return strings.TrimPrefix(ctx.Names().Class(n.Value, n.Span().Start), `\`)
}

// PuClassConstClass returns the class name node of an `X::class` fetch.
func PuClassConstClass(e syntax.Expr) (*syntax.Name, bool) {
	f, ok := e.(*syntax.ClassConstFetch)
	if !ok {
		return nil, false
	}
	id, ok := f.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "class") {
		return nil, false
	}
	n, ok := f.Class.(*syntax.Name)
	return n, ok
}
