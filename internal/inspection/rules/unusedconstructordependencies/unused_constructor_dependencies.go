package unusedconstructordependencies

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// unusedConstructorDependencies reports private properties that are only
// touched by the constructor.
type unusedConstructorDependencies struct{}

func (unusedConstructorDependencies) ID() string { return "UnusedConstructorDependencies" }
func (unusedConstructorDependencies) Semantic()  {}
func (unusedConstructorDependencies) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KClassLike}
}

const unusedCtorDepMsg = "Private property is only used in the constructor; likely dead code."

func (unusedConstructorDependencies) Check(ctx *analysis.Context, n syntax.Node) {
	cl := n.(*syntax.ClassLike)
	if cl.ClassKind != syntax.KindClass || cl.Span().Len() == 0 { // D1
		return
	}
	var ctor *syntax.Method
	hasProps := false
	var traits []*syntax.Name
	for _, m := range cl.Members {
		switch x := m.(type) {
		case *syntax.Method:
			if x.Name != nil && strings.EqualFold(x.Name.Value, "__construct") {
				ctor = x
			}
		case *syntax.Property:
			hasProps = true
		case *syntax.TraitUse:
			traits = append(traits, x.Traits...)
		}
	}
	if ctor == nil || ctor.Body == nil || !hasProps || ucdReentered(ctor.Body) {
		return
	}
	candidates := map[string]bool{} // D2
	for _, m := range cl.Members {
		p, ok := m.(*syntax.Property)
		// custos: attributes (`#[ORM\Column]`) mark mapped properties like
		// annotations; assigning a hooked property runs its `set` hook.
		if !ok || !p.Modifiers.Has(syntax.TPrivate) || p.Modifiers.Has(syntax.TStatic) || semanticquery.DocHasAnnotation(ctx.File, p) || len(p.Attrs) > 0 || len(p.Hooks) > 0 {
			continue
		}
		for _, it := range p.Props {
			if it.Var.Name != "" {
				candidates[it.Var.Name] = true
			}
		}
	}
	if len(candidates) == 0 {
		return
	}
	fqn := ctx.Types().ClassFQN(cl)
	ctorRefs := map[string][]syntax.Expr{} // D4
	used := map[string]bool{}              // D5
	collectPropRefs(ctx, ctor.Body, fqn, candidates, func(name string, ref syntax.Expr) {
		if syntax.EnclosingFuncLike(ref) != syntax.Node(ctor) {
			// D4a: inside a closure/arrow function defined in the
			// constructor (it may run later): a read is a use, a plain
			// write is ignored.
			if !ucdPlainTarget(ref) {
				used[name] = true
			}
			return
		}
		ctorRefs[name] = append(ctorRefs[name], ref)
	})
	if len(ctorRefs) == 0 {
		return
	}
	if ucdClosureDynamicAccess(ctx, ctor.Body) || !ucdScanMembers(ctx, cl, ctor, fqn, candidates, used) {
		return
	}
	for _, t := range traits {
		tfqn := ctx.Names().Class(t.Value, t.Span().Start)
		decl := findClassLike(ctx, tfqn)
		if decl == nil {
			if ctx.Index().Class(tfqn, ctx.PHP) != nil {
				return // trait declared in another file: its bodies cannot be scanned
			}
			continue
		}
		if !ucdScanMembers(ctx, decl, nil, fqn, candidates, used) {
			return
		}
	}
	for _, m := range cl.Members { // keep source order of findings stable
		p, ok := m.(*syntax.Property)
		if !ok {
			continue
		}
		for _, it := range p.Props {
			name := it.Var.Name
			refs := ctorRefs[name]
			if !candidates[name] || used[name] || len(refs) == 0 {
				continue
			}
			for _, r := range refs { // D6
				if ucdPlainTarget(r) && !ucdBindsReferences(r) {
					ctx.ReportNode(r, unusedCtorDepMsg)
				}
			}
		}
	}
}

// ucdScanMembers marks the candidates used by the code of decl that runs
// outside the constructor ctor: the other method bodies and the property
// hooks (those of promoted constructor parameters included). It returns
// false when one of these scopes reads properties dynamically.
func ucdScanMembers(ctx *analysis.Context, decl *syntax.ClassLike, ctor *syntax.Method, fqn string, candidates, used map[string]bool) bool {
	mark := func(name string, _ syntax.Expr) { used[name] = true }
	scan := func(root syntax.Node) bool {
		if ucdDynamicAccess(ctx, root) {
			return false
		}
		collectPropRefs(ctx, root, fqn, candidates, mark)
		return true
	}
	for _, m := range decl.Members {
		var hooks []*syntax.PropertyHook
		switch x := m.(type) {
		case *syntax.Method:
			for _, p := range x.Params {
				hooks = append(hooks, p.Hooks...)
			}
			if x != ctor && x.Body != nil {
				if !scan(x.Body) {
					return false
				}
				if x.Name != nil && strings.EqualFold(x.Name.Value, "__sleep") {
					ucdSleepNames(x.Body, candidates, used)
				}
			}
		case *syntax.Property:
			hooks = x.Hooks
		}
		for _, h := range hooks {
			if !scan(h) {
				return false
			}
		}
	}
	return true
}

// ucdSleepNames marks the candidates named by a string literal in the body
// of `__sleep()`: the properties it lists are serialised (custos). A private
// property may be spelled `"\0Class\0name"`.
func ucdSleepNames(body syntax.Node, candidates, used map[string]bool) {
	syntax.Inspect(body, func(x syntax.Node) bool {
		if e, ok := x.(syntax.Expr); ok {
			if name, ok := astquery.QuotedStringValue(e); ok {
				if name = name[strings.LastIndexByte(name, 0)+1:]; candidates[name] {
					used[name] = true
				}
			}
		}
		return true
	})
}

// ucdClosureDynamicAccess reports whether a closure or arrow function
// defined in the constructor reads properties dynamically: it may run after
// construction (D4a), unlike the constructor's own scope.
func ucdClosureDynamicAccess(ctx *analysis.Context, body syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch x.(type) {
		case *syntax.Closure, *syntax.ArrowFunction:
			found = found || ucdDynamicAccess(ctx, x)
			return false
		}
		return !found
	})
	return found
}

// ucdDynamicAccess reports whether body reads properties of `$this` by a
// computed name (`$this->$name`, `$this->{$k}`) or all at once
// (`get_object_vars($this)`, `foreach ($this as …)`, `(array) $this`): any
// private property may be used there (custos divergence from D3, found on
// Moodle's `cm_info::__get()`).
func ucdDynamicAccess(ctx *analysis.Context, body syntax.Node) bool {
	isThis := func(e syntax.Expr) bool {
		v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
		return ok && v.Name == "this"
	}
	found := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch f := x.(type) {
		case *syntax.PropertyFetch:
			if _, ok := f.Name.(*syntax.Identifier); !ok && isThis(f.Var) {
				found = true
			}
		case *syntax.Foreach:
			found = found || isThis(f.Expr)
		case *syntax.Unary:
			found = found || (f.Op.Kind == syntax.TArrayCast && isThis(f.Expr))
		case *syntax.FuncCall:
			if ctx.IsGlobalFunctionCall(f, "get_object_vars") {
				if args, ok := astquery.CallArgValues(f); ok && len(args) == 1 && isThis(args[0]) {
					found = true
				}
			}
		}
		return !found
	})
	return found
}

// ucdReentered reports whether the constructor calls itself on the same
// object (`$this->__construct(…)`, SimplePie's redirect handling): the
// re-entered constructor sees what the first run stored (custos).
func ucdReentered(body syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if c, ok := x.(*syntax.MethodCall); ok {
			v, isVar := syntax.UnwrapParens(c.Var).(*syntax.Variable)
			id, isID := c.Name.(*syntax.Identifier)
			found = found || isVar && v.Name == "this" && isID && strings.EqualFold(id.Value, "__construct")
		}
		return !found
	})
	return found
}

// ucdPlainTarget reports whether ref is the target of a plain `=` assignment.
func ucdPlainTarget(ref syntax.Expr) bool {
	parent, child := astquery.ParentSkipParens(ref)
	a, ok := parent.(*syntax.Assign)
	return ok && a.Op.Kind == syntax.TEqual && syntax.Node(a.Var) == child && child == syntax.Node(ref)
}

// ucdBindsReferences reports whether the plain assignment to ref stores an
// array holding references (`[&$this->data]`): building it turns what they
// point to into references shared by clones of the object (E6).
func ucdBindsReferences(ref syntax.Expr) bool {
	parent, _ := astquery.ParentSkipParens(ref)
	arr, ok := syntax.UnwrapParens(parent.(*syntax.Assign).Value).(*syntax.Array)
	if !ok {
		return false
	}
	for _, it := range arr.Items {
		if it != nil && it.ByRef {
			return true
		}
	}
	return false
}

// collectPropRefs calls fn for every property access under root named after
// a candidate that resolves to the class fqn or cannot be resolved.
func collectPropRefs(ctx *analysis.Context, root syntax.Node, fqn string, candidates map[string]bool, fn func(string, syntax.Expr)) {
	syntax.Inspect(root, func(x syntax.Node) bool {
		switch f := x.(type) {
		case *syntax.PropertyFetch:
			if id, ok := f.Name.(*syntax.Identifier); ok && candidates[id.Value] && refersTo(ctx, f.Var, nil, id.Value, fqn) {
				fn(id.Value, f)
			}
		case *syntax.StaticPropertyFetch:
			if v, ok := f.Name.(*syntax.Variable); ok && v.NameExpr == nil && candidates[v.Name] && refersTo(ctx, nil, f.Class, v.Name, fqn) {
				fn(v.Name, f)
			}
		}
		return true
	})
}

func refersTo(ctx *analysis.Context, recv, class syntax.Expr, prop, fqn string) bool {
	var classes []string
	if recv != nil {
		classes = ctx.TypeOf(recv).Classes()
	} else if nm, ok := class.(*syntax.Name); ok {
		switch strings.ToLower(nm.Value) {
		case "self", "static":
			return true
		case "parent":
			return false
		}
		classes = []string{ctx.Names().Class(nm.Value, nm.Span().Start)}
	} else {
		classes = ctx.TypeOf(class).Classes()
	}
	if len(classes) == 0 {
		return true
	}
	for _, c := range classes {
		p := ctx.Index().FindProperty(strings.TrimPrefix(c, `\`), prop, ctx.PHP)
		if p == nil || strings.EqualFold(strings.TrimPrefix(p.Class, `\`), fqn) {
			return true
		}
	}
	return false
}

// findClassLike returns the class-like declared in this file with the FQN.
func findClassLike(ctx *analysis.Context, fqn string) *syntax.ClassLike {
	var found *syntax.ClassLike
	syntax.InspectFile(ctx.File, func(x syntax.Node) bool {
		if found != nil {
			return false
		}
		if c, ok := x.(*syntax.ClassLike); ok && c.Name != nil && strings.EqualFold(ctx.Types().ClassFQN(c), strings.TrimPrefix(fqn, `\`)) {
			found = c
			return false
		}
		return true
	})
	return found
}
