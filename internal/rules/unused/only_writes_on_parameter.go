package unused

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
	"custos/internal/types"
)

// onlyWritesOnParameter reports parameters, closure imports and locals that
// are written but never read, and unused closure imports / inline
// assignment targets.
type onlyWritesOnParameter struct{}

func init() { register(onlyWritesOnParameter{}) }

func (onlyWritesOnParameter) ID() string { return "OnlyWritesOnParameter" }

func (onlyWritesOnParameter) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFunction, syntax.KMethod, syntax.KClosure}
}

const (
	owpWriteMsg  = "Value is only written here and never read; the write is lost."
	owpUnusedMsg = "Variable is never used."
)

var owpSuperGlobals = map[string]bool{
	"_GET": true, "_POST": true, "_SESSION": true, "_REQUEST": true, "_FILES": true, "_COOKIE": true,
	"_ENV": true, "_SERVER": true, "GLOBALS": true, "HTTP_RAW_POST_DATA": true,
}

type owpScope struct {
	dynamic  map[string]bool // names read through compact(); "*" for get_defined_vars()
	ctx      *analysis.Context
	scope    syntax.Node
	body     *syntax.Block
	reported map[syntax.Span]bool
	includes int8 // -1 unknown, 0 no, 1 yes
}

func (s *owpScope) report(span syntax.Span, msg string) {
	if s.reported[span] {
		return
	}
	s.reported[span] = true
	s.ctx.Report(span, msg)
}

func (onlyWritesOnParameter) Check(ctx *analysis.Context, n syntax.Node) {
	body := util.FuncLikeBody(n)
	if body == nil || n.Span().Len() == 0 { // E2
		return
	}
	s := &owpScope{ctx: ctx, scope: n, body: body, reported: map[syntax.Span]bool{}, includes: -1}
	excluded := map[string]bool{}
	// Entry 1: parameters.
	for _, p := range util.FuncLikeParams(n) {
		if p.Var == nil || p.Var.NameExpr != nil || p.Var.Name == "" {
			continue
		}
		excluded[p.Var.Name] = true
		if p.ByRef || (!p.Variadic && owpHasObjectType(p.Type)) { // D1, E1
			continue
		}
		s.analyse(p.Var.Name, nil)
	}
	// Entry 2: closure imports.
	if c, ok := n.(*syntax.Closure); ok {
		includes := c.Body != nil && owpClosureIncludes(c.Body) // D4b
		for _, u := range c.Uses {
			if u.Var == nil || u.Var.NameExpr != nil || u.Var.Name == "" {
				continue
			}
			excluded[u.Var.Name] = true
			if u.ByRef { // D3
				if !includes && len(s.accesses(u.Var.Name)) == 0 && !s.dynamicReads()[u.Var.Name] && !s.dynamicReads()["*"] {
					s.report(u.Var.Span(), owpUnusedMsg)
				}
				continue
			}
			// D4; W findings dropped for object imports (D4a).
			v := u.Var
			objectImport := func() bool { return owpObjectType(ctx.TypeOf(v)) }
			if s.analyse(v.Name, objectImport) == 0 && !includes {
				s.report(v.Span(), owpUnusedMsg)
			}
		}
	}
	// Entry 3: local assignments in eligible contexts.
	done := map[string]bool{}
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch a := x.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Assign:
			if a.Op.Kind != syntax.TEqual { // D5
				return true
			}
			v, ok := a.Var.(*syntax.Variable)
			if !ok || v.NameExpr != nil || v.Name == "" || owpSuperGlobals[v.Name] || excluded[v.Name] || done[v.Name] {
				return true
			}
			if !owpEntryContext(a) { // D6
				return true
			}
			done[v.Name] = true
			s.analyse(v.Name, nil) // D8
		}
		return true
	})
}

// owpEntryContext implements D6 on the assignment's direct parent.
func owpEntryContext(a *syntax.Assign) bool {
	switch p := a.Parent().(type) {
	case *syntax.Paren, *syntax.ExprStmt:
		return true
	case *syntax.ArrayDimFetch:
		return p.Dim == syntax.Expr(a)
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical:
			return true
		}
	case *syntax.Assign:
		return p.Op.Kind == syntax.TEqual && p.Value == syntax.Expr(a)
	}
	return false
}

// owpHasObjectType reports whether a native type hint has an object member.
func owpHasObjectType(t syntax.Expr) bool {
	if t == nil {
		return false
	}
	found := false
	syntax.Inspect(t, func(x syntax.Node) bool {
		var v string
		switch n := x.(type) {
		case *syntax.Name:
			v = n.Value
		case *syntax.Identifier:
			v = n.Value
		default:
			return true
		}
		switch strings.ToLower(strings.TrimPrefix(v, `\`)) {
		case "array", "iterable", "string", "int", "float", "bool", "false", "true", "null", "void",
			"mixed", "callable", "resource", "never", "closure":
		default:
			found = true
		}
		return false
	})
	return found
}

type owpAccess struct {
	v      *syntax.Variable
	arrow  bool // occurrence inside a nested arrow function (captured read)
	second bool // the read half of a consumed inline assignment
}

// accesses returns the reachable accesses to name in control-flow order
// (approximated by source order); a consumed plain assignment target yields
// a write followed by a read on the same node.
func (s *owpScope) accesses(name string) []owpAccess {
	var out []owpAccess
	for _, a := range util.VarAccesses(s.ctx.File, s.scope, name) {
		if !util.Reachable(a.Var, s.scope) {
			continue
		}
		acc := owpAccess{v: a.Var, arrow: owpInArrow(a.Var, s.scope)}
		out = append(out, acc)
		if asg, ok := a.Var.Parent().(*syntax.Assign); ok && !acc.arrow && asg.Var == syntax.Expr(a.Var) && asg.Op.Kind == syntax.TEqual {
			if _, stmt := asg.Parent().(*syntax.ExprStmt); !stmt {
				out = append(out, owpAccess{v: a.Var, second: true})
			}
		}
	}
	return out
}

func owpInArrow(v syntax.Node, scope syntax.Node) bool {
	for p := v.Parent(); p != nil && p != scope; p = p.Parent() {
		if _, ok := p.(*syntax.ArrowFunction); ok {
			return true
		}
	}
	return false
}

func owpIsVar(e syntax.Expr, name string) bool {
	v, ok := e.(*syntax.Variable)
	return ok && v.NameExpr == nil && v.Name == name
}

func owpDirectStmt(n syntax.Node) bool {
	_, ok := n.Parent().(*syntax.ExprStmt)
	return ok
}

// analyse runs the access analysis for name and returns the number of
// accesses found. When dropWrites is non-nil and returns true, the W
// findings of this analysis are not reported (D4a).
func (s *owpScope) analyse(name string, dropWrites func() bool) int {
	accs := s.accesses(name)
	if s.dynamicReads()[name] || s.dynamicReads()["*"] { // E8: compact('v'), get_defined_vars()
		return len(accs) + 1
	}
	if len(accs) == 0 {
		return 0
	}
	reads, writes := 0, 0
	ref := false
	var targets []syntax.Node
	for _, acc := range accs {
		x := acc.v
		if acc.arrow {
			reads++
			continue
		}
		switch p := x.Parent().(type) {
		case *syntax.ArrayDimFetch:
			if p.Var != syntax.Expr(x) {
				reads++ // used as an index
				break
			}
			// A1
			var t syntax.Node = p
			for {
				q, ok := t.Parent().(*syntax.ArrayDimFetch)
				if !ok || q.Var != t {
					break
				}
				t = q
			}
			switch q := t.Parent().(type) {
			case *syntax.Assign:
				if q.Var != t {
					reads++
					break
				}
				writes++
				if ref || q.Op.Kind != syntax.TEqual {
					reads++
				} else {
					targets = append(targets, t)
				}
			case *syntax.IncDec:
				targets = append(targets, t)
				writes++
			default:
				reads++
			}
		case *syntax.IncDec: // A2
			writes++
			if ref {
				reads++
			} else {
				targets = append(targets, p)
			}
			if !owpDirectStmt(p) {
				reads++
			}
		case *syntax.Unary: // A2
			if !owpDirectStmt(p) {
				reads++
			}
		case *syntax.Binary: // A3 (+ A7 read)
			reads += 2
		case *syntax.Assign:
			if p.Op.Kind != syntax.TEqual {
				if owpIsVar(p.Var, name) && p.Var == syntax.Expr(x) { // A4
					writes++
					if ref {
						reads++
					} else {
						targets = append(targets, x)
					}
					if !owpDirectStmt(p) {
						reads++
					}
				} else {
					reads++
				}
				break
			}
			// A5
			if owpIsVar(p.Var, name) {
				writes++
				if ref {
					reads++
				}
				if p.ByRef {
					ref = true
				}
				if len(accs) == 2 && accs[0].v == accs[1].v && accs[0].v == x {
					s.report(p.Var.Span(), owpUnusedMsg)
					return len(accs)
				}
			} else {
				reads++ // the value is $v (or nested in it): a read
			}
		case *syntax.Arg, *syntax.ArgList, *syntax.ClosureUse, *syntax.Unset, *syntax.Empty, *syntax.Isset, *syntax.Foreach: // A6
			reads++
		default: // A7
			if owpIsWriteNature(x) {
				targets = append(targets, x)
				writes++
			} else {
				reads++
			}
		}
	}
	if reads != 0 || writes == 0 || len(targets) == 0 {
		return len(accs)
	}
	if !s.ctx.Bool("IGNORE_INCLUDES") && s.hasIncludes() { // E5
		return len(accs)
	}
	if s.suppressed(targets) { // E4
		return len(accs)
	}
	if dropWrites != nil && dropWrites() { // D4a, E5b
		return len(accs)
	}
	for _, t := range targets {
		s.report(t.Span(), owpWriteMsg)
	}
	return len(accs)
}

// dynamicReads returns the variables the scope reads by name: string
// arguments of compact() (nested arrays included), or "*" when it calls
// get_defined_vars(). Nested functions, closures and classes are separate
// scopes; arrow functions capture the enclosing scope and are included.
func (s *owpScope) dynamicReads() map[string]bool {
	if s.dynamic != nil {
		return s.dynamic
	}
	s.dynamic = map[string]bool{}
	var names func(e syntax.Expr)
	names = func(e syntax.Expr) {
		if c, _, ok := util.QuotedStringRaw(e); ok {
			s.dynamic[c] = true
			return
		}
		if arr, ok := util.UnwrapParens(e).(*syntax.Array); ok {
			for _, it := range arr.Items {
				if it != nil && it.Value != nil {
					names(it.Value)
				}
			}
		}
	}
	syntax.Inspect(s.body, func(x syntax.Node) bool {
		switch c := x.(type) {
		case *syntax.Function, *syntax.Method, *syntax.Closure, *syntax.ClassLike:
			return false
		case *syntax.FuncCall:
			switch {
			case s.ctx.IsGlobalFunctionCall(c, "compact"):
				if args, ok := util.CallArgValues(c); ok {
					for _, a := range args {
						names(a)
					}
				}
			case s.ctx.IsGlobalFunctionCall(c, "get_defined_vars"):
				s.dynamic["*"] = true
			}
		}
		return true
	})
	return s.dynamic
}

// owpIsWriteNature reports writes not covered by A1–A6: destructuring
// targets, catch variables, global/static declarations.
func owpIsWriteNature(x *syntax.Variable) bool {
	switch p := x.Parent().(type) {
	case *syntax.Catch, *syntax.Global, *syntax.StaticVar:
		return true
	case *syntax.ArrayItem:
		if p.Value != syntax.Expr(x) {
			return false
		}
		// Only a destructuring pattern (the array/list is the assignment
		// target or the foreach value) writes; an array literal used as a
		// value merely reads its elements.
		prev := syntax.Node(p)
		for c := syntax.Node(p.Parent()); c != nil; prev, c = c, c.Parent() {
			switch q := c.(type) {
			case *syntax.List, *syntax.Array:
				continue
			case *syntax.ArrayItem:
				if q.Value != prev {
					return false
				}
				continue
			case *syntax.Assign:
				return q.Op.Kind == syntax.TEqual && q.Var == prev
			case *syntax.Foreach:
				return q.Value == prev
			}
			return false
		}
	}
	return false
}

// owpObjectType reports whether an inferred type has an object member by
// the D1 criterion: `object` or a class/interface other than Closure.
func owpObjectType(t types.Type) bool {
	if t.Has("object") {
		return true
	}
	for _, c := range t.Classes() {
		if !strings.EqualFold(c, `\Closure`) {
			return true
		}
	}
	return false
}

// owpClosureIncludes reports whether a closure body contains an
// include/require outside nested closures, functions and classes (arrow
// function bodies count) (D4b).
func owpClosureIncludes(body syntax.Node) bool {
	if body == nil {
		return false
	}
	found := false
	syntax.Inspect(body, func(x syntax.Node) bool {
		switch x.(type) {
		case *syntax.Closure, *syntax.Function, *syntax.Method, *syntax.ClassLike:
			return false
		case *syntax.Include:
			found = true
		}
		return !found
	})
	return found
}

func (s *owpScope) hasIncludes() bool {
	if s.includes < 0 {
		s.includes = 0
		syntax.Inspect(s.body, func(x syntax.Node) bool {
			if _, ok := x.(*syntax.Include); ok {
				s.includes = 1
			}
			return s.includes == 0
		})
	}
	return s.includes == 1
}

// suppressed implements E4.
func (s *owpScope) suppressed(targets []syntax.Node) bool {
	f := s.ctx.File
	for _, t := range targets {
		a, ok := t.Parent().(*syntax.Assign)
		if !ok {
			continue
		}
		st, ok := a.Parent().(*syntax.ExprStmt)
		if !ok {
			continue
		}
		i := util.TokenIndex(f, st.Span().Start) - 1
		for i >= 0 && (f.Tokens[i].Kind == syntax.TWhitespace || f.Tokens[i].Kind == syntax.TComment) {
			i--
		}
		if i < 0 || f.Tokens[i].Kind != syntax.TDocComment {
			continue
		}
		text := string(f.Src[f.Tokens[i].Start:f.Tokens[i].End])
		if (strings.Contains(text, "@noinspection") && strings.Contains(text, "OnlyWritesOnParameterInspection")) ||
			strings.Contains(text, "@custos-ignore OnlyWritesOnParameter") {
			return true
		}
	}
	return false
}
