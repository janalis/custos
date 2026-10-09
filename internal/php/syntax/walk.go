package syntax

// Children calls fn for each direct child of n, in source order. Nil
// children are skipped.
func Children(n Node, fn func(Node)) {
	e := func(x Expr) {
		if x != nil && !isNilNode(x) {
			fn(x)
		}
	}
	s := func(x Stmt) {
		if x != nil && !isNilNode(x) {
			fn(x)
		}
	}
	attrs := func(gs []*AttributeGroup) {
		for _, g := range gs {
			fn(g)
		}
	}
	params := func(ps []*Param) {
		for _, p := range ps {
			fn(p)
		}
	}
	stmts := func(ss []Stmt) {
		for _, x := range ss {
			s(x)
		}
	}
	exprs := func(xs []Expr) {
		for _, x := range xs {
			e(x)
		}
	}
	names := func(ns []*Name) {
		for _, x := range ns {
			fn(x)
		}
	}
	args := func(a *ArgList) {
		if a != nil {
			fn(a)
		}
	}
	switch n := n.(type) {
	case *Name, *Identifier, *Literal, *StringPart, *MagicConst, *BadExpr, *BadStmt,
		*VariadicPlaceholder, *InlineHTML, *Nop, *HaltCompiler:
	case *Variable:
		e(n.NameExpr)
	case *ArrayDimFetch:
		e(n.Var)
		e(n.Dim)
	case *PropertyFetch:
		e(n.Var)
		e(n.Name)
	case *StaticPropertyFetch:
		e(n.Class)
		e(n.Name)
	case *ClassConstFetch:
		e(n.Class)
		e(n.Name)
	case *ConstFetch:
		fn(n.Name)
	case *Arg:
		if n.Name != nil {
			fn(n.Name)
		}
		e(n.Value)
	case *ArgList:
		exprs(n.Args)
	case *FuncCall:
		e(n.Name)
		args(n.Args)
	case *MethodCall:
		e(n.Var)
		e(n.Name)
		args(n.Args)
	case *StaticCall:
		e(n.Class)
		e(n.Name)
		args(n.Args)
	case *New:
		if n.Class != nil {
			if cl, ok := n.Class.(*ClassLike); ok {
				// Anonymous class: its args come first in source.
				fn(cl)
			} else {
				e(n.Class)
				args(n.Args)
			}
		}
	case *Assign:
		e(n.Var)
		e(n.Value)
	case *Binary:
		e(n.Left)
		e(n.Right)
	case *Unary:
		e(n.Expr)
	case *IncDec:
		e(n.Var)
	case *Ternary:
		e(n.Cond)
		e(n.Then)
		e(n.Else)
	case *Instanceof:
		e(n.Expr)
		e(n.Class)
	case *Isset:
		exprs(n.Vars)
	case *Empty:
		e(n.Expr)
	case *Exit:
		args(n.Args)
	case *Print:
		e(n.Expr)
	case *Include:
		e(n.Expr)
	case *Eval:
		e(n.Expr)
	case *Clone:
		if n.Args != nil {
			args(n.Args)
		} else {
			e(n.Expr)
			e(n.With)
		}
	case *Throw:
		e(n.Expr)
	case *Yield:
		e(n.Key)
		e(n.Value)
	case *YieldFrom:
		e(n.Expr)
	case *ArrayItem:
		e(n.Key)
		e(n.Value)
	case *Array:
		for _, it := range n.Items {
			if it != nil {
				fn(it)
			}
		}
	case *List:
		for _, it := range n.Items {
			if it != nil {
				fn(it)
			}
		}
	case *ClosureUse:
		fn(n.Var)
	case *Closure:
		attrs(n.Attrs)
		params(n.Params)
		for _, u := range n.Uses {
			fn(u)
		}
		e(n.ReturnType)
		if n.Body != nil {
			fn(n.Body)
		}
	case *ArrowFunction:
		attrs(n.Attrs)
		params(n.Params)
		e(n.ReturnType)
		e(n.Expr)
	case *MatchArm:
		exprs(n.Conds)
		e(n.Body)
	case *Match:
		e(n.Cond)
		for _, a := range n.Arms {
			fn(a)
		}
	case *InterpolatedString:
		exprs(n.Parts)
	case *Paren:
		e(n.Expr)
	case *NullableType:
		e(n.Type)
	case *UnionType:
		exprs(n.Types)
	case *IntersectionType:
		exprs(n.Types)
	case *Attribute:
		fn(n.Name)
		args(n.Args)
	case *AttributeGroup:
		for _, a := range n.Attrs {
			fn(a)
		}
	case *PropertyHook:
		attrs(n.Attrs)
		fn(n.Name)
		params(n.Params)
		if n.Body != nil && !isNilNode(n.Body) {
			fn(n.Body)
		}
	case *Param:
		attrs(n.Attrs)
		e(n.Type)
		fn(n.Var)
		e(n.Default)
		for _, h := range n.Hooks {
			fn(h)
		}
	case *Block:
		stmts(n.Stmts)
	case *ExprStmt:
		e(n.Expr)
	case *Echo:
		exprs(n.Exprs)
	case *If:
		e(n.Cond)
		s(n.Body)
		for _, ei := range n.ElseIfs {
			fn(ei)
		}
		if n.Else != nil {
			fn(n.Else)
		}
	case *ElseIf:
		e(n.Cond)
		s(n.Body)
	case *Else:
		s(n.Body)
	case *While:
		e(n.Cond)
		s(n.Body)
	case *DoWhile:
		s(n.Body)
		e(n.Cond)
	case *For:
		exprs(n.Init)
		exprs(n.Cond)
		exprs(n.Loop)
		s(n.Body)
	case *Foreach:
		e(n.Expr)
		e(n.Key)
		e(n.Value)
		s(n.Body)
	case *Case:
		e(n.Cond)
		stmts(n.Stmts)
	case *Switch:
		e(n.Cond)
		for _, c := range n.Cases {
			fn(c)
		}
	case *Break:
		e(n.Num)
	case *Continue:
		e(n.Num)
	case *Return:
		e(n.Expr)
	case *Global:
		exprs(n.Vars)
	case *StaticVar:
		fn(n.Var)
		e(n.Default)
	case *StaticStmt:
		for _, v := range n.Vars {
			fn(v)
		}
	case *Unset:
		exprs(n.Vars)
	case *Goto:
		fn(n.Label)
	case *Label:
		fn(n.Name)
	case *Catch:
		names(n.Types)
		if n.Var != nil {
			fn(n.Var)
		}
		fn(n.Body)
	case *Finally:
		fn(n.Body)
	case *Try:
		fn(n.Body)
		for _, c := range n.Catches {
			fn(c)
		}
		if n.Finally != nil {
			fn(n.Finally)
		}
	case *DeclareItem:
		fn(n.Key)
		e(n.Value)
	case *Declare:
		for _, d := range n.Items {
			fn(d)
		}
		s(n.Body)
	case *UseItem:
		fn(n.Name)
		if n.Alias != nil {
			fn(n.Alias)
		}
	case *Use:
		if n.Prefix != nil {
			fn(n.Prefix)
		}
		for _, it := range n.Items {
			fn(it)
		}
	case *Namespace:
		if n.Name != nil {
			fn(n.Name)
		}
		stmts(n.Stmts)
	case *ConstItem:
		fn(n.Name)
		e(n.Value)
	case *ConstStmt:
		attrs(n.Attrs)
		for _, c := range n.Consts {
			fn(c)
		}
	case *Function:
		attrs(n.Attrs)
		fn(n.Name)
		params(n.Params)
		e(n.ReturnType)
		fn(n.Body)
	case *ClassLike:
		attrs(n.Attrs)
		if n.Name != nil {
			fn(n.Name)
		}
		args(n.Args)
		e(n.EnumType)
		names(n.Extends)
		names(n.Implements)
		stmts(n.Members)
	case *Method:
		attrs(n.Attrs)
		fn(n.Name)
		params(n.Params)
		e(n.ReturnType)
		if n.Body != nil {
			fn(n.Body)
		}
	case *PropertyItem:
		fn(n.Var)
		e(n.Default)
	case *Property:
		attrs(n.Attrs)
		e(n.Type)
		for _, it := range n.Props {
			fn(it)
		}
		for _, h := range n.Hooks {
			fn(h)
		}
	case *ClassConst:
		attrs(n.Attrs)
		e(n.Type)
		for _, c := range n.Consts {
			fn(c)
		}
	case *EnumCase:
		attrs(n.Attrs)
		fn(n.Name)
		e(n.Value)
	case *TraitAdaptation:
		if n.Trait != nil {
			fn(n.Trait)
		}
		fn(n.Method)
		names(n.Insteadof)
		if n.Alias != nil {
			fn(n.Alias)
		}
	case *TraitUse:
		names(n.Traits)
		for _, a := range n.Adaptations {
			fn(a)
		}
	}
}

// isNilNode detects typed-nil interface values (e.g. (*Block)(nil) as Stmt).
func isNilNode(n Node) bool {
	switch v := n.(type) {
	case *Block:
		return v == nil
	case *Name:
		return v == nil
	case *Identifier:
		return v == nil
	case *Variable:
		return v == nil
	case *ClassLike:
		return v == nil
	}
	return false
}

// Inspect walks the tree depth-first in source order. If fn returns false
// the node's children are skipped.
func Inspect(n Node, fn func(Node) bool) {
	if !fn(n) {
		return
	}
	Children(n, func(c Node) { Inspect(c, fn) })
}

// InspectFile walks every top-level statement of f.
func InspectFile(f *File, fn func(Node) bool) {
	for _, s := range f.Stmts {
		Inspect(s, fn)
	}
}

// SetParents links every node to its parent (top-level statements have a
// nil parent). It also moves parent span starts back to cover their
// children, which only matters for zero-width nodes created by error
// recovery. Ends need no widening: a parser node ends at the last consumed
// token, after all its children were parsed (lastEnd never decreases).
func SetParents(f *File) {
	var link func(parent Node)
	link = func(parent Node) {
		b := parent.base()
		Children(parent, func(c Node) {
			c.base().parent = parent
			link(c)
			cs := c.Span()
			if cs.Start < b.span.Start {
				b.span.Start = cs.Start
			}
		})
	}
	for _, s := range f.Stmts {
		s.base().parent = nil
		link(s)
	}
}
