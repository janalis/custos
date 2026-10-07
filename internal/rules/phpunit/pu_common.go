package phpunit

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// puCall is the shape shared by instance (`->`, `?->`) and static (`::`)
// method calls.
type puCall struct {
	Node  syntax.Expr
	Recv  syntax.Expr // receiver expression or class reference
	Ident *syntax.Identifier
	Name  string // method name, in its declared spelling when known (puName)
	Args  *syntax.ArgList
}

// puKnownNames lists the PHPUnit methods these rules match by name.
var puKnownNames = map[string]string{}

func init() {
	for _, n := range []string{
		"expects", "will", "willReturn", "method", "exactly", "once", "any",
		"returnValue", "returnValueMap", "returnCallback", "returnArgument",
		"assert", "assertTrue", "assertFalse", "assertNotTrue", "assertNotFalse",
		"assertEquals", "assertNotEquals", "assertSame", "assertNotSame",
		"assertNull", "assertEmpty", "assertInstanceOf", "assertInternalType",
		"assertFileEquals", "assertStringEqualsFile",
		"assertFileNotExists", "assertDirectoryNotExists",
	} {
		puKnownNames[strings.ToLower(n)] = n
	}
}

// puName returns the declared spelling of a known PHPUnit method name
// written in any case (PHP method names are case-insensitive), or the name
// as written when it is not one the rules match.
func puName(written string) string {
	if n, ok := puKnownNames[strings.ToLower(written)]; ok {
		return n
	}
	return written
}

// asPuCall views e as a method call with a literal name.
func asPuCall(e syntax.Node) (puCall, bool) {
	switch c := e.(type) {
	case *syntax.MethodCall:
		if id, ok := c.Name.(*syntax.Identifier); ok && c.Args != nil {
			return puCall{Node: c, Recv: c.Var, Ident: id, Name: puName(id.Value), Args: c.Args}, true
		}
	case *syntax.StaticCall:
		if id, ok := c.Name.(*syntax.Identifier); ok && c.Args != nil {
			return puCall{Node: c, Recv: c.Class, Ident: id, Name: puName(id.Value), Args: c.Args}, true
		}
	}
	return puCall{}, false
}

// args returns the plain positional argument values (nil, false when the
// list uses spreads, named arguments or placeholders).
func (c puCall) args() ([]syntax.Expr, bool) { return util.ArgValues(c.Args) }

// puMethodCallNamed reports whether e (no parentheses stripped) is a method
// call whose name is one of names (case-insensitive, as PHP compares
// method names).
func puMethodCallNamed(e syntax.Node, names ...string) (puCall, bool) {
	c, ok := asPuCall(e)
	if !ok {
		return puCall{}, false
	}
	for _, n := range names {
		if strings.EqualFold(c.Ident.Value, n) {
			return c, true
		}
	}
	return puCall{}, false
}

// puResolveClassName resolves a class reference name (self/static/parent
// included) to an FQN without leading backslash ("" when unresolvable).
func puResolveClassName(ctx *analysis.Context, n *syntax.Name) string {
	switch strings.ToLower(n.Value) {
	case "self", "static":
		return ctx.Types().ClassFQN(syntax.EnclosingClass(n))
	case "parent":
		return ctx.Names().ParentFQN(syntax.EnclosingClass(n))
	}
	return strings.TrimPrefix(ctx.Names().Class(n.Value, n.Span().Start), `\`)
}

// puClassConstClass returns the class name node of an `X::class` fetch.
func puClassConstClass(e syntax.Expr) (*syntax.Name, bool) {
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
