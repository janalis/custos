package phpunit

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

// unnecessaryAssertion reports assertions already guaranteed by a declared
// return type, and `expects($this->any())`.
type unnecessaryAssertion struct{}

func init() { register(unnecessaryAssertion{}) }

func (unnecessaryAssertion) ID() string { return "UnnecessaryAssertion" }

func (unnecessaryAssertion) Semantic() {}

func (unnecessaryAssertion) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall}
}

const (
	uaTypedMsg = "The declared return type already guarantees this; the assertion can go."
	uaAnyMsg   = "expects(any()) verifies nothing; drop the expects() call."
)

func (unnecessaryAssertion) Check(ctx *analysis.Context, n syntax.Node) {
	c, ok := asPuCall(n)
	if !ok || n.Span().Len() == 0 {
		return
	}
	if strings.HasPrefix(strings.ToLower(c.Name), "assert") {
		uaCheckTyped(ctx, c)
		return
	}
	if c.Name != "expects" { // D7
		return
	}
	args, ok := c.args()
	if !ok || len(args) != 1 {
		return
	}
	if _, ok := puMethodCallNamed(args[0], "any"); !ok { // D8
		return
	}
	call, recv := c.Node, c.Recv
	f := ctx.File
	ctx.ReportNode(args[0], uaAnyMsg, analysis.Fix{
		Title: "Drop expects(any())",
		Edits: func() []analysis.TextEdit {
			rs := recv.Span()
			return []analysis.TextEdit{{Span: call.Span(), NewText: string(f.Src[rs.Start:rs.End])}}
		},
	})
}

func uaCheckTyped(ctx *analysis.Context, c puCall) {
	if ctx.PHP < phpver.PHP70 { // E1
		return
	}
	pos := 0 // D1
	switch c.Name {
	case "assertNull", "assertEmpty":
	case "assertInstanceOf", "assertInternalType":
		pos = 1
	default:
		return
	}
	args, ok := c.args()
	if !ok || len(args) < pos+1 {
		return
	}
	vals := util.PossibleValues(ctx.File, args[pos]) // D2
	if len(vals) != 1 {
		return
	}
	t, ok := uaDeclaredReturn(ctx, vals[0]) // D3–D5
	if !ok || len(t.Atoms()) != 1 {
		return
	}
	atom := t.Atoms()[0]
	switch c.Name { // D6
	case "assertNull", "assertEmpty":
		if !strings.EqualFold(strings.TrimPrefix(atom, `\`), "void") {
			return
		}
	case "assertInstanceOf":
		cn, ok := puClassConstClass(args[0])
		if !ok {
			return
		}
		fqn := puResolveClassName(ctx, cn)
		if fqn == "" || ctx.Index().Class(fqn, ctx.PHP) == nil {
			return
		}
		if !strings.HasPrefix(atom, `\`) || !strings.EqualFold(atom[1:], fqn) {
			return
		}
	case "assertInternalType":
		lit, ok := args[0].(*syntax.Literal)
		if !ok {
			return
		}
		name, ok := util.StringLiteralValue(lit.Raw)
		if !ok || !uaInternalTypeMatches(name, atom) {
			return
		}
	}
	ctx.ReportNode(c.Node, uaTypedMsg)
}

// uaDeclaredReturn resolves a function/method call to its declaration and
// returns its declared return type merged with the @return doc type. ok is
// false when the call does not resolve or has no native return type.
func uaDeclaredReturn(ctx *analysis.Context, e syntax.Expr) (types.Type, bool) {
	var ret, doc, cls string
	switch x := e.(type) {
	case *syntax.FuncCall:
		f := ctx.Types().ResolveFunction(x)
		if f == nil {
			return types.Unknown, false
		}
		ret, doc = f.Return, f.DocReturn
	case *syntax.MethodCall, *syntax.StaticCall:
		c, ok := asPuCall(x)
		if !ok {
			return types.Unknown, false
		}
		if mc, ok := x.(*syntax.MethodCall); ok {
			cs := ctx.TypeOf(mc.Var).Classes()
			if len(cs) != 1 {
				return types.Unknown, false
			}
			cls = strings.TrimPrefix(cs[0], `\`)
		} else if nm, ok := c.Recv.(*syntax.Name); ok {
			cls = puResolveClassName(ctx, nm)
		} else {
			cs := ctx.TypeOf(c.Recv).Classes()
			if len(cs) != 1 {
				return types.Unknown, false
			}
			cls = strings.TrimPrefix(cs[0], `\`)
		}
		if cls == "" {
			return types.Unknown, false
		}
		m := ctx.Index().FindMethod(cls, c.Name, ctx.PHP)
		if m == nil {
			return types.Unknown, false
		}
		ret, doc = m.Return, m.DocReturn
	default:
		return types.Unknown, false
	}
	if ret == "" { // D4
		return types.Unknown, false
	}
	t := types.FromDoc(ret, nil)
	if doc != "" {
		t = types.Union(t, types.FromDoc(doc, nil))
	}
	if t.IsUnknown() {
		return t, false
	}
	if cls != "" {
		atoms := make([]string, 0, len(t.Atoms()))
		for _, a := range t.Atoms() {
			if a == "static" || a == "self" {
				a = `\` + cls
			}
			atoms = append(atoms, a)
		}
		t = types.Of(atoms...)
	}
	return t, true
}

// uaInternalTypeMatches reports whether a value of the single declared type
// atom always satisfies assertInternalType(name, ...).
func uaInternalTypeMatches(name, atom string) bool {
	atom = strings.ToLower(atom)
	isClass := strings.HasPrefix(atom, `\`)
	isArray := atom == "array" || strings.HasSuffix(atom, "[]")
	switch strings.ToLower(name) {
	case "array":
		return isArray
	case "bool", "boolean":
		return atom == "bool" || atom == "true" || atom == "false"
	case "float", "double", "real":
		return atom == "float"
	case "int", "integer":
		return atom == "int"
	case "numeric":
		return atom == "int" || atom == "float"
	case "string":
		return atom == "string"
	case "scalar":
		switch atom {
		case "int", "float", "string", "bool", "true", "false":
			return true
		}
	case "null":
		return atom == "null" || atom == "void"
	case "object":
		return (isClass && !isArray) || atom == "object"
	case "resource":
		return atom == "resource"
	case "callable":
		return atom == "callable" || atom == `\closure`
	case "iterable":
		return atom == "iterable" || isArray
	}
	return false
}
