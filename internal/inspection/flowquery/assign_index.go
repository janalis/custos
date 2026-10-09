package flowquery

import (
	"custos/internal/php/syntax"
)

// assignIndex summarises the writes under a root node (nested closures
// included), built in one walk and cached on the file: value discovery used
// to re-walk the whole function body for every variable use, which is
// quadratic on large functions.
type assignIndex struct {
	plain    []*syntax.Assign            // plain `=` (not by-ref) assignments, preorder
	byVar    map[string][]*syntax.Assign // the plain ones whose target is $name
	unstable map[string]bool             // UnstableVariableIn(f, root, name)
	byRefUse map[string]bool             // imported by reference into a closure under root
	// the plain ones by target kind (non-variable targets) and, for
	// property fetches, by target source text
	byKind     map[syntax.NodeKind][]*syntax.Assign
	propByText map[string][]*syntax.Assign
}
type assignIndexKey struct{ root syntax.Node }

func assignsUnder(f *syntax.File, root syntax.Node) *assignIndex {
	if root == nil {
		return &assignIndex{}
	}
	build := func() any { return buildAssignIndex(f, root) }
	if f == nil {
		return build().(*assignIndex)
	}
	return f.Memo(assignIndexKey{root}, build).(*assignIndex)
}

func buildAssignIndex(f *syntax.File, root syntax.Node) *assignIndex {
	ix := &assignIndex{
		byVar: map[string][]*syntax.Assign{}, unstable: map[string]bool{}, byRefUse: map[string]bool{},
		byKind: map[syntax.NodeKind][]*syntax.Assign{}, propByText: map[string][]*syntax.Assign{},
	}
	simple := func(e syntax.Expr) (string, bool) {
		v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
		if !ok || v.NameExpr != nil {
			return "", false
		}
		return v.Name, true
	}
	syntax.Inspect(root, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.ClosureUse:
			if x.ByRef && x.Var != nil {
				ix.byRefUse[x.Var.Name] = true
			}
		case *syntax.IncDec:
			if name, ok := simple(x.Var); ok {
				ix.unstable[name] = true
			}
		case *syntax.Assign:
			if x.Op.Kind != syntax.TEqual {
				if name, ok := simple(x.Var); ok {
					ix.unstable[name] = true
				}
				break
			}
			if x.ByRef {
				break
			}
			ix.plain = append(ix.plain, x)
			if t, ok := x.Var.(*syntax.Variable); ok {
				if t.NameExpr == nil {
					ix.byVar[t.Name] = append(ix.byVar[t.Name], x)
				}
			} else if x.Var != nil {
				k := x.Var.Kind()
				ix.byKind[k] = append(ix.byKind[k], x)
				if k == syntax.KPropertyFetch && f != nil {
					sp := x.Var.Span()
					text := string(f.Src[sp.Start:sp.End])
					ix.propByText[text] = append(ix.propByText[text], x)
				}
			}
		}
		return true
	})
	return ix
}

// UnstableVariableIn reports whether variable $name is, anywhere under root
// (nested closures included), the operand of `++`/`--` or the target of a
// compound assignment (`+=`, `.=`, `??=`, ...). Value discovery treats such
// a variable's value set as unknown. The answer for every name of root is
// computed in one walk and cached on f (nil: not cached).
func UnstableVariableIn(f *syntax.File, root syntax.Node, name string) bool {
	if root == nil || name == "" {
		return false
	}
	return assignsUnder(f, root).unstable[name]
}

// maxPossibleValues caps the candidate sets of value discovery: beyond it
// the result is reported unknown (consumers stay silent). Real code has a
// handful of candidates; hostile code with thousands of assignments to one
// variable and thousands of uses would otherwise be quadratic.
const (
	MaxPossibleValues = 512
	// maxAssignScan bounds the assignments value discovery compares one by one
	// against a property fetch (Equivalent); more makes the result unknown.
	MaxAssignScan = 4 * MaxPossibleValues
)

type varOccurrencesKey struct{ root syntax.Node }

// VarOccurrences returns every simple variable node (`$name`, no
// variable-variable) under root, nested function-likes included, grouped
// by name in source order; cached on f (nil: not cached). Use it instead of
// walking a whole body for each variable of interest.
func VarOccurrences(f *syntax.File, root syntax.Node) map[string][]*syntax.Variable {
	build := func() any {
		m := map[string][]*syntax.Variable{}
		if root == nil {
			return m
		}
		syntax.Inspect(root, func(x syntax.Node) bool {
			if v, ok := x.(*syntax.Variable); ok && v.NameExpr == nil {
				m[v.Name] = append(m[v.Name], v)
			}
			return true
		})
		return m
	}
	if f == nil {
		return build().(map[string][]*syntax.Variable)
	}
	return f.Memo(varOccurrencesKey{root}, build).(map[string][]*syntax.Variable)
}

// AssignmentsOfKind returns plain assignments targeting nodes of kind under root.
func AssignmentsOfKind(f *syntax.File, root syntax.Node, kind syntax.NodeKind) []*syntax.Assign {
	return assignsUnder(f, root).byKind[kind]
}
