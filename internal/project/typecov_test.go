package project_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/project"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

// TestTypeCoverage measures how precise type inference is on real code: the
// share of variable reads, calls, property and element reads whose inferred
// type is unknown, by kind, and the most frequent causes (how an unknown
// variable got its value, what an unknown element read reads from). It is
// opt-in (`make typecov`): CUSTOS_TYPECOV lists root=srcdir pairs separated
// by ';', and CUSTOS_TYPECOV_OUT, when set, is the prefix of a .summary
// file and of a .unknown.tsv file listing every unknown expression.
// CUSTOS_TYPECOV_ORIGIN=1 also classifies unknown method calls and where
// element reads' bases come from.
func TestTypeCoverage(t *testing.T) {
	spec := os.Getenv("CUSTOS_TYPECOV") // root=srcdir;root=srcdir
	out := os.Getenv("CUSTOS_TYPECOV_OUT")
	if spec == "" {
		t.Skip("no CUSTOS_TYPECOV")
	}
	counts := map[string][2]int{}
	var unknown []string
	causes := map[string]int{}
	for _, pair := range strings.Split(spec, ";") {
		root, src, _ := strings.Cut(pair, "=")
		files, err := project.Discover([]string{filepath.Join(root, src)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := config.Load(root)
		if err != nil {
			t.Fatal(err)
		}
		opt := syntax.Options{Version: cfg.PHP}
		ix := project.BuildIndex(project.IndexSources(root, files, []string{filepath.Join(root, src)}, nil), opt, infer.Options{})
		for _, path := range files {
			b, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			f := syntax.Parse(path, b, opt)
			local := index.New(ix)
			local.Add(index.Extract(f))
			env := infer.NewEnv(f, names.New(f), local, cfg.PHP, infer.Options{})
			syntax.InspectFile(f, func(n syntax.Node) bool {
				var kind string
				switch x := n.(type) {
				case *syntax.Variable:
					if x.NameExpr != nil || x.Name == "this" {
						return true
					}
					if a, ok := x.Parent().(*syntax.Assign); ok && a.Var == syntax.Expr(x) {
						return true
					}
					switch x.Parent().(type) {
					case *syntax.StaticPropertyFetch, *syntax.Param, *syntax.PropertyItem, *syntax.ClosureUse, *syntax.Global, *syntax.StaticVar, *syntax.Catch, *syntax.Foreach, *syntax.ArrayItem, *syntax.Unset, *syntax.Isset:
						return true
					}
					kind = "var"
				case *syntax.MethodCall:
					kind = "mcall"
				case *syntax.PropertyFetch:
					kind = "prop"
				case *syntax.StaticCall:
					kind = "scall"
				case *syntax.FuncCall:
					kind = "fcall"
				case *syntax.ArrayDimFetch:
					if x.Dim == nil || writeTarget(x) {
						return true
					}
					kind = "dim"
				default:
					return true
				}
				e := n.(syntax.Expr)
				c := counts[kind]
				c[0]++
				if ty := env.TypeOf(e); ty.HasAny("self", "parent", "static") {
					line := 1 + strings.Count(string(b[:e.Span().Start]), "\n")
					causes["relative-leak:"+kind]++
					unknown = append(unknown, fmt.Sprintf("leak\t%s:%d\t%s\t%s", path, line, strings.ReplaceAll(string(b[e.Span().Start:min(e.Span().End, e.Span().Start+60)]), "\n", " "), ty.String()))
				}
				if env.TypeOf(e).IsUnknown() {
					c[1]++
					line := 1 + strings.Count(string(b[:e.Span().Start]), "\n")
					snip := string(b[e.Span().Start:min(e.Span().End, e.Span().Start+60)])
					cause := ""
					if v, ok := e.(*syntax.Variable); ok {
						cause = defCause(v, env)
						causes[cause]++
					}
					if mc, ok := e.(*syntax.MethodCall); ok && os.Getenv("CUSTOS_TYPECOV_ORIGIN") != "" {
						cause = "mcall/" + mcallCause(mc, env, local)
						causes[cause]++
					}
					if d, ok := e.(*syntax.ArrayDimFetch); ok {
						bt := env.TypeOf(d.Var)
						switch {
						case bt.IsUnknown():
							cause = "dim/base-unknown"
							if os.Getenv("CUSTOS_TYPECOV_ORIGIN") != "" {
								cause += "<-" + dimOrigin(d.Var, env)
							}
						case bt.Has("mixed"):
							cause = "dim/base-mixed"
						case bt.OnlyOf("array") || bt.Without("null", "false").OnlyOf("array"):
							cause = "dim/base-plain-array"
							if os.Getenv("CUSTOS_TYPECOV_ORIGIN") != "" {
								cause += "<-" + dimOrigin(d.Var, env)
							}
						default:
							cause = "dim/base=" + bt.Without("null").ShapeString()
							if len(cause) > 70 {
								cause = cause[:70]
							}
						}
						causes[cause]++
					}
					unknown = append(unknown, fmt.Sprintf("%s\t%s:%d\t%s\t%s", kind, path, line, strings.ReplaceAll(snip, "\n", " "), cause))
				}
				counts[kind] = c
				return true
			})
		}
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var report strings.Builder
	for _, k := range keys {
		c := counts[k]
		fmt.Fprintf(&report, "%-6s total=%7d unknown=%7d (%.1f%%)\n", k, c[0], c[1], 100*float64(c[1])/float64(max(c[0], 1)))
	}
	ck := make([]string, 0, len(causes))
	for k := range causes {
		ck = append(ck, k)
	}
	sort.Slice(ck, func(i, j int) bool { return causes[ck[i]] > causes[ck[j]] })
	for _, k := range ck {
		fmt.Fprintf(&report, "  %6d %s\n", causes[k], k)
	}
	t.Log("\n" + report.String())
	if out != "" {
		_ = os.WriteFile(out+".summary", []byte(report.String()), 0o644)
		_ = os.WriteFile(out+".unknown.tsv", []byte(strings.Join(unknown, "\n")), 0o644)
	}
}

// writeTarget reports an element that is (part of) the target of a write.
func writeTarget(n syntax.Expr) bool {
	cur := syntax.Node(n)
	for {
		switch p := cur.Parent().(type) {
		case *syntax.ArrayDimFetch:
			if p.Var == cur {
				cur = p
				continue
			}
		case *syntax.Assign:
			return p.Var == cur
		case *syntax.Unset, *syntax.Isset, *syntax.ArrayItem, *syntax.Foreach:
			return true
		}
		return false
	}
}

// defCause classifies how a variable whose read is unknown gets its value:
// the last plain definition before the read in its scope.
func defCause(v *syntax.Variable, env *infer.Env) string {
	scope := syntax.EnclosingVariableScope(v)
	var params []*syntax.Param
	switch f := scope.(type) {
	case *syntax.Function:
		params = f.Params
	case *syntax.Method:
		params = f.Params
	case *syntax.Closure:
		params = f.Params
		for _, u := range f.Uses {
			if u.Var.Name == v.Name {
				return "closure-use"
			}
		}
	case *syntax.ArrowFunction:
		params = f.Params
	}
	for _, p := range params {
		if p.Var.Name == v.Name {
			if p.Type == nil {
				if _, ok := scope.(*syntax.Closure); ok {
					return "param-untyped-closure"
				}
				if _, ok := scope.(*syntax.ArrowFunction); ok {
					return "param-untyped-closure"
				}
				return "param-untyped"
			}
			return "param-typed"
		}
	}
	best := "none-found"
	root := scope
	if root == nil {
		return "file-scope"
	}
	syntax.Inspect(root, func(n syntax.Node) bool {
		if n != root {
			switch n.(type) {
			case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
				return false
			}
		}
		if n.Span().Start >= v.Span().Start {
			return false
		}
		switch x := n.(type) {
		case *syntax.Assign:
			if w, ok := x.Var.(*syntax.Variable); ok && w.Name == v.Name {
				best = "assign:" + fmt.Sprintf("%T", syntax.UnwrapParens(x.Value))
				if x.Op.Kind != syntax.TEqual {
					best = "assign-op"
				}
			}
		case *syntax.Foreach:
			if w, ok := x.Value.(*syntax.Variable); ok && w.Name == v.Name {
				it := env.TypeOf(x.Expr)
				switch {
				case it.IsUnknown():
					best = "foreach-value/iter-unknown:" + fmt.Sprintf("%T", syntax.UnwrapParens(x.Expr))
				case it.Has("array") || it.Has("mixed"):
					best = "foreach-value/plain-array-or-mixed"
				default:
					best = "foreach-value/iter=" + it.ShapeString()
				}
			}
			if w, ok := x.Key.(*syntax.Variable); ok && w.Name == v.Name {
				best = "foreach-key"
			}
			if _, ok := x.Value.(*syntax.List); ok {
				syntax.Inspect(x.Value, func(m syntax.Node) bool {
					if w, ok := m.(*syntax.Variable); ok && w.Name == v.Name {
						best = "foreach-destructure"
					}
					return true
				})
			}
		case *syntax.List, *syntax.Array:
			if a, ok := x.Parent().(*syntax.Assign); ok && a.Var == x.(syntax.Expr) {
				syntax.Inspect(x, func(m syntax.Node) bool {
					if w, ok := m.(*syntax.Variable); ok && w.Name == v.Name {
						best = "destructure"
					}
					return true
				})
			}
		case *syntax.Catch:
			if x.Var != nil && x.Var.Name == v.Name {
				best = "catch"
			}
		}
		return true
	})
	return best
}

// dimOrigin names where the plain-array base of an element read comes from.
func dimOrigin(x syntax.Expr, env *infer.Env) string {
	switch b := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		c := defCause(b, env)
		if strings.HasPrefix(c, "assign:") || c == "none-found" {
			// find the last assignment value for detail
			scope := syntax.EnclosingVariableScope(b)
			var last syntax.Expr
			if scope != nil {
				syntax.Inspect(scope, func(n syntax.Node) bool {
					if n.Span().Start >= b.Span().Start {
						return false
					}
					if a, ok := n.(*syntax.Assign); ok {
						if w, ok := a.Var.(*syntax.Variable); ok && w.Name == b.Name {
							last = a.Value
						}
					}
					return true
				})
			}
			switch v := syntax.UnwrapParens(last).(type) {
			case *syntax.FuncCall:
				if nm, ok := v.Name.(*syntax.Name); ok {
					return c + ":" + strings.ToLower(nm.Value)
				}
			case *syntax.MethodCall:
				if id, ok := v.Name.(*syntax.Identifier); ok {
					return c + ":" + id.Value
				}
			case *syntax.StaticCall:
				if id, ok := v.Name.(*syntax.Identifier); ok {
					return c + ":" + id.Value
				}
			}
		}
		return c
	case *syntax.PropertyFetch:
		return "prop"
	case *syntax.StaticPropertyFetch:
		return "sprop"
	case *syntax.ArrayDimFetch:
		return "dim"
	case *syntax.MethodCall:
		if id, ok := b.Name.(*syntax.Identifier); ok {
			return "mcall:" + id.Value
		}
	case *syntax.FuncCall:
		if nm, ok := b.Name.(*syntax.Name); ok {
			return "fcall:" + strings.ToLower(nm.Value)
		}
	}
	return fmt.Sprintf("%T", x)
}

// mcallCause classifies why method call mc is unknown.
func mcallCause(mc *syntax.MethodCall, env *infer.Env, ix *index.Index) string {
	rt := env.TypeOf(mc.Var)
	id, named := mc.Name.(*syntax.Identifier)
	switch {
	case !named:
		return "dynamic-name"
	case rt.IsUnknown():
		return "recv-unknown:" + fmt.Sprintf("%T", syntax.UnwrapParens(mc.Var))
	case rt.Has("mixed") || rt.Has("object"):
		return "recv-mixed-object"
	}
	cs := rt.Classes()
	if len(cs) == 0 {
		return "recv-nonclass=" + rt.String()
	}
	c := strings.TrimPrefix(cs[0], `\\`)
	m := ix.FindMethod(c, id.Value, phpversion.PHP84)
	switch {
	case m == nil && ix.Class(c, phpversion.PHP84) == nil:
		return "class-missing"
	case m == nil:
		return "method-missing"
	case m.Return == "" && m.DocReturn == "":
		return "no-return-type"
	case m.Tpl != nil || m.GenReturn != "":
		return "template"
	}
	return "declared=" + m.Return + "|doc=" + m.DocReturn
}
