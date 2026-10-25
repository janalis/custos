package loopwhichdoesnotloop

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// loopWhichDoesNotLoop reports braced loops whose body always leaves on the
// first pass (or is empty).
type loopWhichDoesNotLoop struct{}

const loopNoLoopMsg = "Loop body exits on the first iteration; the loop never repeats."

func (loopWhichDoesNotLoop) ID() string { return "LoopWhichDoesNotLoop" }

// Semantic: iterated types and generator methods may live in other files.
func (loopWhichDoesNotLoop) Semantic() {}

func (loopWhichDoesNotLoop) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KForeach, syntax.KFor, syntax.KWhile, syntax.KDoWhile}
}

func (loopWhichDoesNotLoop) Check(ctx *analysis.Context, n syntax.Node) {
	if strings.HasSuffix(ctx.File.Path, ".blade.php") {
		return
	}
	var body syntax.Stmt
	switch l := n.(type) {
	case *syntax.Foreach:
		body = l.Body
	case *syntax.For:
		body = l.Body
	case *syntax.While:
		body = l.Body
	case *syntax.DoWhile:
		body = l.Body
	}
	blk, ok := body.(*syntax.Block) // D1
	if !ok || blk.Span().Len() == 0 {
		return
	}
	var last syntax.Stmt // D2
	for i := len(blk.Stmts) - 1; i >= 0; i-- {
		if blk.Stmts[i].Span().Len() > 0 {
			last = blk.Stmts[i]
			break
		}
	}
	if last == nil {
		// An empty while/do/for body still loops: the condition (or step)
		// does the work (`while (@ob_end_flush()) {}`).
		if _, ok := n.(*syntax.Foreach); !ok {
			return
		}
	}
	isThrow := false
	if last != nil {
		switch x := last.(type) {
		case *syntax.Break, *syntax.Return:
		case *syntax.ExprStmt:
			if _, ok := x.Expr.(*syntax.Throw); !ok {
				return
			}
			isThrow = true
		default:
			return
		}
	}
	if continuesLoop(blk, n) { // D3
		return
	}
	if fe, ok := n.(*syntax.Foreach); ok && !isThrow { // D4
		t := ctx.TypeOf(fe.Expr)
		// custos: an empty body iterates for the iterator's side effects
		// (initialising a lazy collection) unless the subject is known not
		// to be an object.
		if last == nil && (t.IsUnknown() || t.Has("object") || t.Has("iterable") || t.Has("mixed") || len(t.Classes()) > 0) {
			return
		}
		// custos: an empty body that stores every key or value in turn
		// keeps the last ones (`foreach ($a as $s->key => $s->value) {}`).
		if last == nil && (keepsLast(fe, fe.Key) || keepsLast(fe, fe.Value)) {
			return
		}
		// custos: `foreach ($this as …)` in a trait iterates the using
		// class (CakePHP's CollectionTrait::isEmpty()).
		if v, ok := syntax.UnwrapParens(fe.Expr).(*syntax.Variable); ok && v.Name == "this" {
			if cl := syntax.EnclosingClass(fe); cl != nil && cl.ClassKind == syntax.KindTrait {
				return
			}
		}
		// custos: a property of unknown type is often a collection that its
		// class fills elsewhere (Doctrine relations): taking its first
		// element may be the point.
		if _, prop := syntax.UnwrapParens(fe.Expr).(*syntax.PropertyFetch); prop && t.IsUnknown() && last != nil {
			return
		}
		if traversable(ctx, t) || (last != nil && callsGeneratorMethod(ctx, fe.Expr)) {
			return
		}
	}
	ctx.Report(astquery.KeywordSpan(ctx.File, n), loopNoLoopMsg)
}

// traversable reports whether one of t's known types is iterable or a
// \Traversable subtype (array-of types do not count).
func traversable(ctx *analysis.Context, t types.Type) bool {
	for _, a := range t.Atoms() {
		switch strings.ToLower(a) {
		case `\generator`, `\traversable`, `\iterator`, "iterable":
			return true
		}
		if strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "]") && ctx.Index().IsSubtype(a, `\Traversable`, ctx.PHP) {
			return true
		}
	}
	return false
}

// callsGeneratorMethod reports whether e calls a method that one of the
// receiver's classes declares as a generator. Inference leaves an overridable
// method without return type unknown, yet its own body yields: the loop takes
// the first element of a \Generator.
func callsGeneratorMethod(ctx *analysis.Context, e syntax.Expr) bool {
	c, ok := syntax.UnwrapParens(e).(*syntax.MethodCall)
	if !ok {
		return false
	}
	id, ok := c.Name.(*syntax.Identifier)
	if !ok {
		return false
	}
	for _, cls := range ctx.TypeOf(c.Var).Classes() {
		m := ctx.Index().FindMethod(strings.TrimPrefix(cls, `\`), id.Value, ctx.PHP)
		if m == nil {
			continue
		}
		if md := semanticquery.MethodDecl(ctx.File, ctx.Index(), m, ctx.PHP); md != nil {
			if semanticquery.NativeGeneratorBody(md.Body) {
				return true
			}
			continue
		}
		// Declared in another file: the index kept the body-derived type.
		if traversable(ctx, types.FromDoc(m.Inferred, nil)) {
			return true
		}
	}
	return false
}

func loopNoLoopIsLoop(n syntax.Node) bool {
	switch n.(type) {
	case *syntax.Foreach, *syntax.For, *syntax.While, *syntax.DoWhile:
		return true
	}
	return false
}

// continuesLoop reports whether a `continue` inside body targets loop.
func continuesLoop(body, loop syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		if found {
			return false
		}
		if syntax.IsFuncLike(x) {
			return false
		}
		c, ok := x.(*syntax.Continue)
		if !ok {
			return true
		}
		level := 1
		if c.Num != nil {
			if lit, ok := c.Num.(*syntax.Literal); ok {
				if v, err := strconv.Atoi(lit.Raw); err == nil {
					level = v
				}
			}
		}
		k := 0
		// The walk never crosses a function boundary: Inspect does not enter
		// function-likes, so the loop is reached first.
		for p := c.Parent(); p != nil; p = p.Parent() {
			_, isSwitch := p.(*syntax.Switch)
			if isSwitch || loopNoLoopIsLoop(p) {
				// PHP counts a switch as a loop structure for continue.
				k++
				if p == loop {
					if level == k {
						found = true
					}
					break
				}
			}
		}
		return true
	})
	return found
}

// keepsLast reports a foreach target that outlives the loop: anything but
// a plain variable (a property, an element, a list), or a variable read
// after the loop.
func keepsLast(fe *syntax.Foreach, target syntax.Expr) bool {
	if target == nil {
		return false
	}
	v, ok := syntax.UnwrapParens(target).(*syntax.Variable)
	if !ok || v.Name == "" {
		return true
	}
	read := false
	syntax.Inspect(syntax.EnclosingVariableScope(fe), func(n syntax.Node) bool {
		if read || n != syntax.EnclosingVariableScope(fe) && syntax.IsVariableScope(n) {
			return false
		}
		if u, ok := n.(*syntax.Variable); ok && u.Name == v.Name && u.Span().Start >= fe.Span().End {
			read = true
		}
		return true
	})
	return read
}
