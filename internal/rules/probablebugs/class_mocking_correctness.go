package probablebugs

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/syntax"
)

// classMockingCorrectness reports test-double factories called with classes
// that cannot be mocked that way.
type classMockingCorrectness struct{}

func init() { register(classMockingCorrectness{}) }

func (classMockingCorrectness) ID() string { return "ClassMockingCorrectness" }

func (classMockingCorrectness) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethodCall, syntax.KStaticCall, syntax.KClassLike}
}

// Semantic marks the rule as needing the project index.
func (classMockingCorrectness) Semantic() {}

const (
	mockFinal        = "Final classes cannot be mocked."
	mockTrait        = "Traits cannot be mocked this way."
	mockNeedAbstract = "This factory expects an abstract class."
	mockNeedTrait    = "This factory expects a trait."
	mockUseAbstract  = "Abstract class: build the double with getMockForAbstractClass()."
	mockUseTrait     = "Trait: build the double with getMockForTrait()."
	mockCtor         = "Mocked constructor needs arguments; pass them or disable the constructor."
)

type mockFactory struct{ class, method string }

// mockFactories lists the test-double factory methods by declaring class
// (FQN without leading backslash) and method name.
var mockFactories = map[mockFactory]bool{
	{"PHPUnit_Framework_TestCase", "getMockBuilder"}:          true,
	{"PHPUnit_Framework_TestCase", "getMock"}:                 true,
	{"PHPUnit_Framework_TestCase", "getMockClass"}:            true,
	{"PHPUnit_Framework_MockObject_Generator", "getMock"}:     true,
	{"PHPUnit_Framework_MockObject_MockBuilder", "getMock"}:   true,
	{`PHPUnit\Framework\TestCase`, "getMockBuilder"}:          true,
	{`PHPUnit\Framework\TestCase`, "getMockForTrait"}:         true,
	{`PHPUnit\Framework\TestCase`, "getMockForAbstractClass"}: true,
	{`PHPUnit\Framework\TestCase`, "getMockClass"}:            true,
	{`PHPUnit\Framework\TestCase`, "createMock"}:              true,
	{`Prophecy\Prophet`, "prophesize"}:                        true,
	{`Prophecy\Prophecy\ObjectProphecy`, "willExtend"}:        true,
}

func (classMockingCorrectness) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.MethodCall:
		checkMockCall(ctx, x, x.Name, x.Args, func() []string { return semClassesOf(ctx, x.Var) })
	case *syntax.StaticCall:
		checkMockCall(ctx, x, x.Name, x.Args, func() []string { return semClassesOf(ctx, x.Class) })
	case *syntax.ClassLike:
		checkPhpSpec(ctx, x)
	}
}

// finalsBypassed reports whether the project ships dg/bypass-finals, which
// strips `final` at load time so final classes can be doubled.
func finalsBypassed(ctx *analysis.Context) bool {
	return ctx.Memo("ClassMockingCorrectness.bypassFinals", func() any {
		return ctx.Index().Class(`DG\BypassFinals`, ctx.PHP) != nil
	}).(bool)
}

func checkMockCall(ctx *analysis.Context, call syntax.Node, name syntax.Expr, list *syntax.ArgList, classes func() []string) {
	id, ok := name.(*syntax.Identifier)
	if !ok {
		return
	}
	lname := strings.ToLower(id.Value) // method names are case-insensitive
	switch lname {                     // D1
	case "getmockbuilder", "getmock", "getmockclass", "getmockfortrait", "getmockforabstractclass", "createmock", "prophesize", "willextend":
	default:
		return
	}
	args := semArgs(list)
	if len(args) == 0 {
		return
	}
	arg, ok := args[0].(*syntax.Arg)
	if !ok || arg.Unpack {
		return
	}
	ix := ctx.Index()
	resolved := false // D2
	for _, c := range classes() {
		if m := ix.FindMethod(c, id.Value, ctx.PHP); m != nil {
			resolved = mockFactories[mockFactory{strings.TrimPrefix(m.Class, `\`), m.Name}]
			break
		}
	}
	if !resolved {
		return
	}
	a := arg.Value
	var k *index.Class // D3
	switch v := a.(type) {
	case *syntax.ClassConstFetch:
		cid, ok := v.Name.(*syntax.Identifier)
		if !ok || !strings.EqualFold(cid.Value, "class") {
			return
		}
		if _, ok := v.Class.(*syntax.Name); !ok {
			return
		}
		for _, c := range semClassesOf(ctx, v.Class) {
			if k = ix.Class(c, ctx.PHP); k != nil {
				break
			}
		}
	default:
		content, _, ok := util.QuotedStringRaw(a)
		if !ok || len(content) <= 3 {
			return
		}
		fqn := strings.ReplaceAll(content, `\\`, `\`)
		if c := ix.Class(strings.TrimPrefix(fqn, `\`), ctx.PHP); c != nil && c.Final {
			k = c
		}
	}
	if k == nil {
		return
	}
	isTrait := k.Kind == syntax.KindTrait
	isIface := k.Kind == syntax.KindInterface
	switch lname { // D4
	case "createmock":
		if isTrait {
			ctx.ReportNode(a, mockTrait)
		} else if k.Final && !finalsBypassed(ctx) {
			ctx.ReportNode(a, mockFinal)
		}
	case "getmockbuilder":
		chained := ""
		if p, ok := call.Parent().(*syntax.MethodCall); ok && p.Var == call {
			if pid, ok := p.Name.(*syntax.Identifier); ok {
				chained = pid.Value
			} else {
				chained = "?"
			}
		}
		switch {
		case k.Abstract && !isIface:
			if chained == "" {
				ctx.ReportNode(a, mockUseAbstract)
			}
		case isTrait:
			if chained == "" {
				ctx.ReportNode(a, mockUseTrait)
			}
		case k.Final:
			if !finalsBypassed(ctx) {
				ctx.ReportNode(a, mockFinal)
			}
		}
		if strings.EqualFold(chained, "getMock") {
			if ctor := ix.FindMethod(strings.TrimPrefix(k.FQN, `\`), "__construct", ctx.PHP); ctor != nil {
				for _, p := range ctor.Params {
					if !p.Optional && !p.Variadic {
						ctx.ReportNode(a, mockCtor)
						break
					}
				}
			}
		}
	case "getmockfortrait":
		if !isTrait {
			ctx.ReportNode(a, mockNeedTrait)
		}
	case "getmockforabstractclass":
		if !k.Abstract && !isIface {
			ctx.ReportNode(a, mockNeedAbstract)
		}
	default:
		if k.Final && !finalsBypassed(ctx) {
			ctx.ReportNode(a, mockFinal)
		}
	}
}

// checkPhpSpec implements Part B (PhpSpec collaborators).
func checkPhpSpec(ctx *analysis.Context, cl *syntax.ClassLike) {
	if cl.ClassKind != syntax.KindClass || len(cl.Extends) == 0 { // D5
		return
	}
	ix := ctx.Index()
	parent := ix.Class(ctx.Names().Class(cl.Extends[0].Value, cl.Extends[0].Span().Start), ctx.PHP)
	if parent == nil || strings.TrimPrefix(parent.FQN, `\`) != `PhpSpec\ObjectBehavior` {
		return
	}
	for _, mem := range cl.Members { // D6
		m, ok := mem.(*syntax.Method)
		if !ok {
			continue
		}
		for _, p := range m.Params {
			var hit *syntax.Name
			visit := func(root syntax.Node) {
				if root == nil || hit != nil {
					return
				}
				syntax.Inspect(root, func(x syntax.Node) bool {
					if hit != nil {
						return false
					}
					nm, ok := x.(*syntax.Name)
					if !ok {
						return true
					}
					switch nm.Parent().(type) {
					case *syntax.ConstFetch, *syntax.FuncCall:
						return false
					}
					if names.IsBuiltinType(nm.Value) {
						return false
					}
					if c := ix.Class(ctx.Names().Class(nm.Value, nm.Span().Start), ctx.PHP); c != nil && c.Final {
						hit = nm
					}
					return false
				})
			}
			if p.Type != nil { // only the type: defaults are not collaborators
				visit(p.Type)
			}
			if hit != nil && !finalsBypassed(ctx) {
				ctx.ReportNode(hit, mockFinal)
			}
		}
	}
}
