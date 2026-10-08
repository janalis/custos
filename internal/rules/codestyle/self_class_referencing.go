package codestyle

import (
	"custos/internal/analysis/util"
	"strings"

	"custos/internal/analysis"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// selfClassReferencing enforces a consistent way of referring to the
// enclosing class inside its methods: `self`/`__CLASS__` by default, or the
// explicit class name with PREFER_CLASS_NAMES.
type selfClassReferencing struct{}

func init() { register(selfClassReferencing{}) }

func (selfClassReferencing) ID() string { return "SelfClassReferencing" }

func (selfClassReferencing) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethod}
}

func (r selfClassReferencing) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	cls, ok := m.Parent().(*syntax.ClassLike)
	if !ok || cls.Name == nil || cls.Name.Value == "" || cls.ClassKind == syntax.KindTrait { // S1
		return
	}
	if m.Body == nil || m.Modifiers.Has(syntax.TAbstract) { // S2
		return
	}
	c := selfRefCheck{ctx: ctx, name: cls.Name.Value, reverse: ctx.Bool("PREFER_CLASS_NAMES"), final: cls.Modifiers.Has(syntax.TFinal) || cls.ClassKind == syntax.KindEnum}
	if ns := ctx.Names().Namespace(cls.Span().Start); ns != "" {
		c.fqn = ns + `\` + c.name
	} else {
		c.fqn = c.name
	}
	for _, p := range m.Params {
		if p.Type != nil {
			c.walk(p.Type)
		}
		if p.Default != nil {
			c.walk(p.Default)
		}
	}
	if m.ReturnType != nil {
		c.walk(m.ReturnType)
	}
	c.walk(m.Body)
}

type selfRefCheck struct {
	ctx     *analysis.Context
	name    string // class short name
	fqn     string // class FQN without leading backslash
	reverse bool
	final   bool // final class or enum: no subclass can rebind static::
}

// walk visits n's subtree, skipping nested function-likes (S3), attributes
// and anonymous class bodies (only their constructor arguments are visited).
func (c *selfRefCheck) walk(n syntax.Node) {
	switch n := n.(type) {
	case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.AttributeGroup:
		return
	case *syntax.ClassLike:
		if n.Args != nil {
			c.walk(n.Args)
		}
		return
	case *syntax.Name:
		c.ref(n)
		return
	case *syntax.MagicConst:
		c.magic(n)
		return
	}
	syntax.Children(n, c.walk)
}

func (c *selfRefCheck) ref(n *syntax.Name) {
	if n.Span().Len() == 0 || !selfRefPosition(n) {
		return
	}
	ctx := c.ctx
	if c.reverse {
		if !strings.EqualFold(n.Value, "self") { // D4 (any case)
			return
		}
		span := n.Span()
		ctx.Report(span, "Spell the class reference '"+n.Value+"' as '"+c.name+"'.", selfRefFix("Use the class name", span, c.name))
		return
	}
	v := n.Value
	if !strings.EqualFold(util.LastNamePart(v), c.name) { // D1 (any case)
		return
	}
	resolved := ctx.Names().Class(v, n.Span().Start)
	if !strings.EqualFold(resolved, c.fqn) { // E4
		return
	}
	if f, ok := n.Parent().(*syntax.ClassConstFetch); ok && f.Class == syntax.Expr(n) {
		if id, ok := f.Name.(*syntax.Identifier); ok && strings.EqualFold(id.Value, "class") { // D2
			if resolved != c.fqn {
				// E6: Name::class yields the name as written; __CLASS__ or
				// self::class would change the string's letter case.
				return
			}
			span := f.Span()
			ctx.Report(span, "Refer to the class as '__CLASS__' instead of '"+c.name+"::class'.", selfRefFix("Use __CLASS__", span, "__CLASS__"))
			return
		}
	}
	span := n.Span() // D3
	msg := "Refer to the class as 'self' instead of '" + c.name + "'."
	if _, call := n.Parent().(*syntax.StaticCall); call && !c.final && len(ctx.Index().ChildrenAll(c.fqn)) > 0 {
		// custos: self::m() forwards the late static binding, Name::m()
		// resets it: inside m(), static:: and `new static` would name the
		// calling subclass.
		ctx.Report(span, msg)
		return
	}
	ctx.Report(span, msg, selfRefFix("Use self", span, "self"))
}

func (c *selfRefCheck) magic(n *syntax.MagicConst) {
	if !c.reverse || n.Token.Kind != syntax.TClassC || c.ctx.PHP < phpver.PHP55 {
		return
	}
	span := n.Span()
	repl := c.name + "::class"
	c.ctx.Report(span, "Spell the class reference '"+c.ctx.SpanText(span)+"' as '"+repl+"'.", selfRefFix("Use the class name", span, repl))
}

// selfRefPosition reports whether name n is used as a class reference:
// new, static access, instanceof, catch and type declarations.
func selfRefPosition(n *syntax.Name) bool {
	e := syntax.Expr(n)
	switch p := n.Parent().(type) {
	case *syntax.New:
		return p.Class == e
	case *syntax.ClassConstFetch:
		return p.Class == e
	case *syntax.StaticCall:
		return p.Class == e
	case *syntax.StaticPropertyFetch:
		return p.Class == e
	case *syntax.Instanceof:
		return p.Class == e
	case *syntax.Catch:
		return true
	case *syntax.Param:
		return p.Type == e
	case *syntax.Method:
		return p.ReturnType == e
	case *syntax.NullableType, *syntax.UnionType:
		return true
	}
	// An intersection type cannot contain self (compile error).
	return false
}

func selfRefFix(title string, span syntax.Span, text string) analysis.Fix {
	return analysis.Fix{
		Title: title,
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: text}} },
	}
}
