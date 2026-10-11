package missingissetimplementation

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

// missingIssetImplementation reports isset()/empty() on undeclared
// properties of classes without __isset().
type missingIssetImplementation struct{}

func (missingIssetImplementation) ID() string { return "MissingIssetImplementation" }
func (missingIssetImplementation) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KIsset, syntax.KEmpty}
}

// Semantic marks the rule as needing the project index.
func (missingIssetImplementation) Semantic() {}

func issetExempt(fqn string) bool {
	switch strings.ToLower(strings.TrimPrefix(fqn, `\`)) {
	case "simplexmlelement", "stdclass", "domdocument":
		return true
	case "arrayobject", "arrayiterator":
		// custos: with ARRAY_AS_PROPS their native handler answers
		// isset($o->key) from the stored array.
		return true
	}
	return false
}

func (missingIssetImplementation) Check(ctx *analysis.Context, n syntax.Node) {
	switch x := n.(type) {
	case *syntax.Isset:
		for _, a := range x.Vars {
			checkMissingIsset(ctx, a)
		}
	case *syntax.Empty:
		checkMissingIsset(ctx, x.Expr)
	}
}

func checkMissingIsset(ctx *analysis.Context, a syntax.Expr) {
	pf, ok := a.(*syntax.PropertyFetch) // D1
	if !ok || pf.NullSafe {
		return
	}
	id, ok := pf.Name.(*syntax.Identifier)
	if !ok {
		return
	}
	if ctx.Text(pf.Var) == "$this" { // D3
		return
	}
	ix := ctx.Index()
	typ := ctx.TypeOf(pf.Var)
	types := typ.Classes()
	for _, t := range types { // D2
		if ix.FindProperty(strings.TrimPrefix(t, `\`), id.Value, ctx.PHP) != nil {
			return
		}
	}
	// custos: the check is "always false" only when every possible value
	// is an object of a concrete class without __isset() and without
	// dynamic properties; a non-class member (object, mixed, array…), an
	// unresolvable class, an interface or abstract class (implementations
	// may declare __isset()) or an #[\AllowDynamicProperties] hierarchy
	// makes the outcome undecidable.
	for _, a := range typ.Atoms() {
		if a != "null" && !strings.HasPrefix(a, `\`) {
			return
		}
	}
	report := ""
	for _, t := range types { // D4
		if issetExempt(t) {
			return
		}
		cls := strings.TrimPrefix(t, `\`)
		c := ix.Class(cls, ctx.PHP)
		if c == nil || issetExempt(c.FQN) || c.Kind != syntax.KindClass || c.Abstract {
			return
		}
		if ix.FindMethod(cls, "__isset", ctx.PHP) != nil || misAllowsDynamic(ctx, cls) {
			return
		}
		if !c.Final && misSubclassMayHave(ctx, cls, id.Value) {
			return
		}
		// custos: a missing ancestor may declare the property or
		// __isset().
		if !semanticquery.HierarchyResolved(ix, cls, ctx.PHP) {
			return
		}
		// custos: without __set() a write creates a real (dynamic)
		// property, which isset() does see; skip it when the file writes
		// properties of that name or by a computed name (an importer's
		// `$entity->$field = $value`).
		// An indirect write (element access, reference) only creates the
		// property when no __get() intercepts the fetch.
		if ix.FindMethod(cls, "__set", ctx.PHP) == nil &&
			misDynamicWrite(ctx, id.Value, ix.FindMethod(cls, "__get", ctx.PHP) == nil) {
			return
		}
		if report == "" {
			report = t
		}
	}
	if report != "" {
		ctx.ReportNode(pf, report+" has no __isset(); this isset/empty check is always false.")
	}
}

// misAllowsDynamic reports whether a class of cls' hierarchy carries
// #[\AllowDynamicProperties] (recorded by the index for every file) or is
// one of the exempt builtins.
func misAllowsDynamic(ctx *analysis.Context, cls string) bool {
	for _, c := range ctx.Index().Ancestors(cls, ctx.PHP) {
		// custos: subclasses of stdClass (always dynamic) and of the
		// exempt builtins (SimpleXMLElement's native isset handler).
		if c.HasAttr("AllowDynamicProperties") || issetExempt(c.FQN) {
			return true
		}
	}
	return false
}

// misDynamicWrite reports whether the file writes a property named name on
// a receiver other than $this, or writes a property with a computed name.
// Writes through an element access or a reference are recorded under
// misIndirect+name and count only when indirect is set.
func misDynamicWrite(ctx *analysis.Context, name string, indirect bool) bool {
	w := ctx.Memo("writes", func() any {
		names := map[string]bool{}
		for _, st := range ctx.File.Stmts {
			syntax.Inspect(st, func(n syntax.Node) bool {
				switch x := n.(type) {
				case *syntax.Assign:
					misNoteWrite(ctx, names, x.Var, false)
					if x.ByRef { // $r = &$o->name
						misNoteWrite(ctx, names, x.Value, true)
					}
				case *syntax.IncDec:
					misNoteWrite(ctx, names, x.Var, false)
				case *syntax.Foreach:
					misNoteWrite(ctx, names, x.Key, false)
					misNoteWrite(ctx, names, x.Value, false)
				case *syntax.Arg:
					if _, ok := misElementBase(x.Value).(*syntax.PropertyFetch); ok && misArgByRef(ctx, x) {
						misNoteWrite(ctx, names, x.Value, true)
					}
				}
				return true
			})
		}
		return names
	}).(map[string]bool)
	return w[""] || w[name] || indirect && w[misIndirect+name]
}

// misIndirect prefixes the recorded name of a property only written through
// an element access or a reference; "$" cannot start a static property name.
const misIndirect = "$"

// misElementBase strips element accesses: writing `$o->name[] = 1` or
// `$o->name['k'] = 1` creates the property too.
func misElementBase(e syntax.Expr) syntax.Expr {
	for {
		d, ok := e.(*syntax.ArrayDimFetch)
		if !ok {
			return e
		}
		e = d.Var
	}
}

// misNoteWrite records the properties a write target creates: the fetch
// itself, the fetch under element accesses, or the slots of a destructuring
// pattern. "" stands for a computed property name. indirect marks a write
// through a reference; a write under an element access is indirect too.
func misNoteWrite(ctx *analysis.Context, names map[string]bool, target syntax.Expr, indirect bool) {
	var items []*syntax.ArrayItem
	base := misElementBase(target)
	prefix := ""
	if indirect || base != target {
		prefix = misIndirect
	}
	switch x := base.(type) {
	case *syntax.PropertyFetch:
		if id, ok := x.Name.(*syntax.Identifier); !ok {
			names[""] = true // computed name
		} else if ctx.Text(x.Var) != "$this" {
			names[prefix+id.Value] = true
		}
	case *syntax.Array:
		items = x.Items
	case *syntax.List:
		items = x.Items
	}
	for _, it := range items {
		if it != nil {
			misNoteWrite(ctx, names, it.Value, false)
		}
	}
}

// misArgByRef reports whether arg is passed by reference: call-time `&`, or
// bound to a by-reference parameter of the resolved callee. An unresolved
// callee is taken as by-value, the common reading of `f($o->name)`.
func misArgByRef(ctx *analysis.Context, arg *syntax.Arg) bool {
	if arg.ByRef {
		return true
	}
	list := arg.Parent().(*syntax.ArgList)
	var params []index.Param
	if n, ok := list.Parent().(*syntax.New); ok {
		ctor := ctx.Index().FindMethod(semanticquery.StaticCallClass(ctx, n.Class), "__construct", ctx.PHP)
		if ctor == nil {
			return false
		}
		params = ctor.Params
	} else if callee, ok := semanticquery.ResolveCallee(ctx, list.Parent()); ok {
		params = callee.Params
	}
	if arg.Name != nil {
		for _, p := range params {
			if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
				return p.ByRef
			}
		}
		return false
	}
	pos := 0
	for list.Args[pos] != syntax.Expr(arg) {
		pos++
	}
	// Arguments past the last parameter bind to a variadic tail.
	if n := len(params); pos >= n {
		if n == 0 || !params[n-1].Variadic {
			return false
		}
		pos = n - 1
	}
	return params[pos].ByRef
}

// misSubclassMayHave reports whether a descendant of cls declares the
// property or __isset() or allows dynamic properties (custos): the value may be such a subclass instance
// (`isset($e->errorcode)` on `Exception`, `isset($node->tagName)` on
// `DOMNode`).
func misSubclassMayHave(ctx *analysis.Context, cls, prop string) bool {
	ix := ctx.Index()
	queue := []string{cls}
	seen := map[string]bool{strings.ToLower(cls): true}
	for len(queue) > 0 {
		k := queue[0]
		queue = queue[1:]
		for _, ch := range ix.ChildrenAll(k) {
			lk := strings.ToLower(strings.TrimPrefix(ch, `\`))
			if c := ix.Class(ch, ctx.PHP); c != nil && !seen[lk] {
				if c.Props[prop] != nil || c.Methods["__isset"] != nil || c.HasAttr("AllowDynamicProperties") {
					return true
				}
				seen[lk] = true
				queue = append(queue, ch)
			}
		}
	}
	return false
}
