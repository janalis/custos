package astquery

import "custos/internal/php/syntax"

type stableArrayReadKey struct{ statement syntax.Stmt }

// StableArrayRead excludes writable accesses and statements that can replace
// the array before a read. The memo contains syntax facts only.
func StableArrayRead(f *syntax.File, fetch *syntax.ArrayDimFetch) bool {
	var base syntax.Expr = fetch
	for {
		if dim, ok := base.(*syntax.ArrayDimFetch); ok {
			base = dim.Var
		} else {
			break
		}
	}
	v, ok := base.(*syntax.Variable)
	if !ok {
		return false
	}
	for child, parent := syntax.Node(fetch), fetch.Parent(); parent != nil; child, parent = parent, parent.Parent() {
		switch p := parent.(type) {
		case *syntax.Arg, *syntax.Unset, *syntax.IncDec:
			return false
		case *syntax.Assign:
			if p.Var == child {
				return false
			}
		}
		if statement, ok := parent.(syntax.Stmt); ok {
			writes := f.Memo(stableArrayReadKey{statement}, func() any {
				names := map[string]bool{}
				syntax.Inspect(statement, func(n syntax.Node) bool {
					if syntax.IsVariableScope(n) {
						return false
					}
					var target syntax.Node
					switch x := n.(type) {
					case *syntax.Assign:
						target = x.Var
					case *syntax.IncDec:
						target = x.Var
					case *syntax.FuncCall:
						target = x.Args
					case *syntax.MethodCall:
						target = x.Args
					case *syntax.StaticCall:
						target = x.Args
					case *syntax.New:
						target = x.Args
					case *syntax.Foreach:
						if x.Value == child || x.Key == child {
							names[v.Name] = true
						}
					}
					if target != nil {
						syntax.Inspect(target, func(n syntax.Node) bool {
							if variable, ok := n.(*syntax.Variable); ok {
								names[variable.Name] = true
							}
							return true
						})
					}
					return true
				})
				return names
			}).(map[string]bool)
			return !writes[v.Name]
		}
	}
	return false
}
