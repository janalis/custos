package flow

import "github.com/janalis/custos/internal/php/syntax"

// IntEntry returns an exact integer at a read on a concrete, acyclic entry
// prefix. Unknown branches and side effects end this proof. Its environment
// and expression values are separate from the fixed-point analysis.
func (e *Env) IntEntry(x syntax.Expr) (int64, bool) {
	if x == nil {
		return 0, false
	}
	if plainVariable(x) == "" {
		return 0, false
	}
	scope := syntax.EnclosingVariableScope(x)
	for _, p := range syntax.VariableScopeParams(scope) {
		if p.ByRef {
			return 0, false
		}
	}
	normal := e.scope(scope)
	if value := normal.values.at(x); !normal.complete || value.Expr == nil && value.interval() == nil && value.origins() == nil {
		return 0, false
	}
	twin := *e
	twin.scopes = make(map[syntax.Node]*scopeResult, len(e.scopes))
	for key, value := range e.scopes {
		twin.scopes[key] = value
	}
	// The replay visits an acyclic prefix of the already complete scope.
	// Its cloned memo retains normal callee results, and it cannot require
	// more transfer work than that completed analysis.
	r := twin.followMode(syntax.EnclosingVariableScope(x), nil, true, x)
	value := r.values.at(x)
	lo, hi, known := integerBounds(value)
	return lo, value.Complete && known && lo == hi
}

func (g *graph) entryPrefix(at int, f frame, r *scopeResult, target syntax.Expr) {
	seen := make([]bool, len(g.blocks))
	for steps := 0; at > 0 && steps < MaxDepth*32; steps++ {
		if seen[at] {
			return
		}
		seen[at] = true
		b := g.blocks[at]
		nodes := b.nodes
		if b.node != nil {
			nodes = []syntax.Node{b.node}
		}
		for j := len(nodes) - 1; j >= 0; j-- {
			node := nodes[j]
			contains := false
			syntax.Inspect(node, func(n syntax.Node) bool {
				if n == target {
					contains = true
				}
				return !contains
			})
			if !entrySafe(node, target, contains, f) {
				return
			}
			g.env.transfer(node, &f, r)
			if contains {
				return
			}
		}
		next := g.successors(at)
		if len(next) == 1 {
			at = int(next[0])
			continue
		}
		condition, ok := b.node.(syntax.Expr)
		truth, known := entryTruth(condition, r)
		if !ok || !known || len(next) != 2 {
			return
		}
		branch := 1
		if truth {
			branch = 0
		}
		g.env.refine(condition, &f, truth)
		at = int(next[branch])
	}
}

// Only scalar local operations are replayed before the read. A call containing
// the requested argument is terminal, and no other call may occur in it.
func entrySafe(node syntax.Node, target syntax.Expr, contains bool, f frame) bool {
	safe := true
	syntax.Inspect(node, func(n syntax.Node) bool {
		if !safe {
			return false
		}
		switch n := n.(type) {
		case *syntax.ExprStmt, *syntax.Paren, *syntax.Literal, *syntax.Variable, *syntax.Arg, *syntax.ArgList, *syntax.Name:
			if v, ok := n.(*syntax.Variable); ok && (v.NameExpr != nil || external(v.Name)) {
				safe = false
			}
		case *syntax.Return:
			safe = contains
		case *syntax.Echo:
			// Earlier echo operands can invoke an output handler before this read.
			first := false
			syntax.Inspect(n.Exprs[0], func(child syntax.Node) bool {
				if child == target {
					first = true
				}
				return !first
			})
			safe = first
		case *syntax.Assign:
			name := plainVariable(n.Var)
			old, held := f.vars.lookup(name)
			safe = name != "" && !n.ByRef && f.aliased[name] == 0 && (held && old.plain && old.deep || !held && f.definite)
		case *syntax.IncDec:
			safe = plainVariable(n.Var) != "" && f.aliased[plainVariable(n.Var)] == 0
		case *syntax.Binary:
			switch n.Op.Kind {
			case syntax.TPlus, syntax.TMinus, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual:
			default:
				safe = false
			}
		case *syntax.Unary:
			safe = n.Op.Kind == syntax.TPlus || n.Op.Kind == syntax.TMinus || n.Op.Kind == syntax.TExclaim
		case *syntax.ConstFetch:
			safe = constantName(n) != ""
		case *syntax.FuncCall:
			// The target must be a direct argument expression of this call.
			found := false
			if contains && n.Args != nil {
				for _, arg := range n.Args.Args {
					value := arg
					if named, ok := arg.(*syntax.Arg); ok {
						value = named.Value
					}
					if value == target {
						found = true
					}
				}
			}
			safe = found
		default:
			safe = false
		}
		return safe
	})
	return safe
}

func entryTruth(x syntax.Expr, r *scopeResult) (bool, bool) {
	if n, ok := syntax.UnwrapParens(x).(*syntax.ConstFetch); ok {
		switch constantName(n) {
		case "true":
			return true, true
		case "false", "null":
			return false, true
		}
	}
	if n, ok := syntax.UnwrapParens(x).(*syntax.Binary); ok {
		lo, hi, a := integerBounds(r.values.at(n.Left))
		bl, bh, b := integerBounds(r.values.at(n.Right))
		if a && b && lo == hi && bl == bh {
			switch n.Op.Kind {
			case syntax.TLess:
				return lo < bl, true
			case syntax.TGreater:
				return lo > bl, true
			case syntax.TIsSmallerOrEqual:
				return lo <= bl, true
			case syntax.TIsGreaterOrEqual:
				return lo >= bl, true
			}
		}
	}
	lo, hi, known := integerBounds(r.values.at(x))
	return lo != 0, known && lo == hi
}
