package format

import (
	"github.com/janalis/custos/internal/php/syntax"
)

// role is what a token is in its construct, where the token kind alone is
// ambiguous.
type role uint8

const (
	roleNone role = iota
	roleBinary
	roleConcat
	roleBoolean // && || and or xor
	roleAssign
	roleUnary
	roleNot
	roleCast
	roleIncDec
	roleTernaryQ
	roleTernaryColon
	roleNullable
	roleTypeOp    // | & in a type
	roleTypeParen // ( ) of a DNF type
	roleRef       // & taking a reference
	roleSpread    // ...
	roleArrow     // =>
	roleReturnColon
	roleNamedColon
	roleEnumColon
	roleCaseColon
	roleLabelColon
	roleAltColon
	roleDeclareEq
	roleForSemicolon
	roleClosureKeyword // function of a closure
	roleArrowKeyword   // fn
	roleUseKeyword     // use of a closure
	roleNewKeyword
	roleAttrClose // ] closing an attribute group
	roleCatchBar  // | between the types of a catch
	roleCaseLabel // case or default opening a switch case
)

// bracket is what an opening bracket opens.
type bracket uint8

const (
	brNone bracket = iota
	// Parentheses.
	brCall    // call, new, attribute, isset... argument list
	brParams  // parameter list
	brUse     // closure use list
	brControl // condition of a control structure
	brGroup   // grouping parentheses
	brArray   // array( list(
	brDeclare
	// Square brackets.
	brShortArray
	brOffset
	brAttr
	// Braces.
	brClass
	brFunction
	brClosure
	brAnonClass
	brControlBlock
	brSwitch
	brMatch
	brNamespace
	brHooks
	brHookBody
	brGroupUse
	brExprBrace // ${ } and ->{ }
)

// list reports whether items of b are separated by commas, one per line
// when split.
func (b bracket) list() bool {
	switch b {
	case brCall, brParams, brUse, brArray, brShortArray, brAttr, brMatch, brGroupUse:
		return true
	}
	return false
}

// block reports whether b holds statements or members.
func (b bracket) block() bool {
	switch b {
	case brClass, brFunction, brClosure, brAnonClass, brControlBlock, brSwitch, brNamespace, brHookBody:
		return true
	}
	return false
}

// region is a byte range whose lines are indented one more level.
type region struct{ start, end uint32 }

// annotate fills the per-token context of e from the syntax tree.
func (e *engine) annotate() {
	n := len(e.toks)
	e.role = make([]role, n)
	e.br = make([]bracket, n)
	e.start = make([]bool, n)
	e.ident = make([]bool, n)
	e.typeName = make([]bool, n)
	e.constName = make([]bool, n)
	e.destruct = make([]bool, n)
	for _, st := range e.f.Stmts {
		e.markStart(st)
		e.walk(st)
	}
}

// at returns the index of the token starting at off, or -1.
func (e *engine) at(off uint32) int {
	i := e.index(off)
	if i < len(e.toks) && e.toks[i].Start == off {
		return i
	}
	return -1
}

// index returns the index of the first token starting at or after off.
func (e *engine) index(off uint32) int {
	lo, hi := 0, len(e.toks)
	for lo < hi {
		m := int(uint(lo+hi) >> 1)
		if e.toks[m].Start < off {
			lo = m + 1
		} else {
			hi = m
		}
	}
	return lo
}

// find returns the index of the first token of kind k in [from, to), or -1.
func (e *engine) find(from, to uint32, k syntax.TokenKind) int {
	for i := e.index(from); i < len(e.toks) && e.toks[i].Start < to; i++ {
		if e.toks[i].Kind == k {
			return i
		}
	}
	return -1
}

// prevSig returns the index of the last significant token ending at or
// before off, or -1.
func (e *engine) prevSig(off uint32) int {
	for i := e.index(off) - 1; i >= 0; i-- {
		if !e.toks[i].Kind.IsTrivia() {
			return i
		}
	}
	return -1
}

func (e *engine) setRole(i int, r role) {
	if i >= 0 {
		e.role[i] = r
	}
}

func (e *engine) setBr(i int, b bracket) {
	if i >= 0 {
		e.br[i] = b
	}
}

func (e *engine) markStart(n syntax.Node) {
	if isNil(n) {
		return
	}
	if i := e.at(n.Span().Start); i >= 0 {
		e.start[i] = true
	}
}

func isNil(n syntax.Node) bool {
	if n == nil {
		return true
	}
	switch v := n.(type) {
	case *syntax.Block:
		return v == nil
	case *syntax.ArgList:
		return v == nil
	case *syntax.Name:
		return v == nil
	case *syntax.Identifier:
		return v == nil
	case *syntax.Variable:
		return v == nil
	case *syntax.Else:
		return v == nil
	case *syntax.Finally:
		return v == nil
	case *syntax.ClassLike:
		return v == nil
	}
	return false
}

// region adds a one-level indented region for the statement body when it
// is not a block.
func (e *engine) bodyRegion(body syntax.Stmt) {
	if isNil(body) {
		return
	}
	if _, ok := body.(*syntax.Block); ok {
		return
	}
	if _, ok := body.(*syntax.If); ok {
		return // else if
	}
	e.markStart(body)
	sp := body.Span()
	e.bodies = append(e.bodies, region{sp.Start, sp.End})
}

// keepIndent marks a byte range whose lines keep their indentation.
func (e *engine) keepIndent(sp syntax.Span) {
	e.kept = append(e.kept, region{sp.Start, sp.End})
}

// blockBraces marks the braces of a block body.
func (e *engine) blockBraces(b syntax.Stmt, kind bracket) {
	blk, ok := b.(*syntax.Block)
	if !ok || blk == nil || blk.Alt {
		return
	}
	sp := blk.Span()
	if i := e.at(sp.Start); i >= 0 && e.toks[i].Kind == syntax.TLBrace {
		e.br[i] = kind
	}
}

// parenAfter marks the first "(" in [from, to) as kind and returns its index.
func (e *engine) parenAfter(from, to uint32, kind bracket) int {
	i := e.find(from, to, syntax.TLParen)
	e.setBr(i, kind)
	return i
}

func (e *engine) argList(a *syntax.ArgList) {
	if a == nil {
		return
	}
	if i := e.at(a.Span().Start); i >= 0 && e.toks[i].Kind == syntax.TLParen {
		e.br[i] = brCall
	}
	for _, arg := range a.Args {
		e.markStart(arg)
	}
}

// signature marks the parameter list, return type colon and attributes of a
// function-like declaration starting at from.
func (e *engine) signature(from uint32, params []*syntax.Param, ret syntax.Expr) {
	end := uint32(len(e.src))
	if len(params) > 0 {
		end = params[0].Span().Start
	} else if !isNil(ret) {
		end = ret.Span().Start
	}
	e.parenAfter(from, end+1, brParams)
	for _, p := range params {
		e.markStart(p)
	}
	if !isNil(ret) {
		if i := e.prevSig(ret.Span().Start); i >= 0 && e.toks[i].Kind == syntax.TColon {
			e.role[i] = roleReturnColon
		}
	}
}

// afterAttrs returns the offset after the attribute groups of a declaration.
func afterAttrs(n syntax.Node, attrs []*syntax.AttributeGroup) uint32 {
	if len(attrs) > 0 {
		return attrs[len(attrs)-1].Span().End
	}
	return n.Span().Start
}

func (e *engine) attrs(groups []*syntax.AttributeGroup) {
	for _, g := range groups {
		sp := g.Span()
		e.setBr(e.at(sp.Start), brAttr)
		if i := e.index(sp.End) - 1; i >= 0 && e.toks[i].Kind == syntax.TRBracket {
			e.role[i] = roleAttrClose
		}
		if i := e.nextSigIndex(sp.End); i >= 0 {
			e.start[i] = true
		}
		for _, a := range g.Attrs {
			e.markStart(a)
		}
	}
}

func (e *engine) nextSigIndex(off uint32) int {
	for i := e.index(off); i < len(e.toks); i++ {
		if !e.toks[i].Kind.IsTrivia() {
			return i
		}
	}
	return -1
}

func (e *engine) stmts(list []syntax.Stmt) {
	for _, s := range list {
		e.markStart(s)
	}
}

// walk annotates n and its descendants.
func (e *engine) walk(n syntax.Node) {
	if isNil(n) {
		return
	}
	e.node(n)
	syntax.Children(n, e.walk)
}

func (e *engine) typeOps(t syntax.Expr) {
	sp := t.Span()
	for i := e.index(sp.Start); i < len(e.toks) && e.toks[i].End <= sp.End; i++ {
		switch e.toks[i].Kind {
		case syntax.TBar, syntax.TAmpersand:
			e.role[i] = roleTypeOp
		case syntax.TLParen, syntax.TRParen:
			e.role[i] = roleTypeParen
		}
	}
}

//nolint:gocyclo // one case per node kind
func (e *engine) node(n syntax.Node) {
	switch v := n.(type) {
	case *syntax.Binary:
		i := e.at(v.Op.Span.Start)
		switch v.Op.Kind {
		case syntax.TDot:
			e.setRole(i, roleConcat)
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor:
			e.setRole(i, roleBoolean)
		default:
			e.setRole(i, roleBinary)
		}
	case *syntax.Instanceof:
		e.setRole(e.find(v.Expr.Span().End, v.Class.Span().Start, syntax.TInstanceof), roleBinary)
	case *syntax.Assign:
		if arr, ok := v.Var.(*syntax.Array); ok {
			e.destructure(arr)
		}
		i := e.at(v.Op.Span.Start)
		e.setRole(i, roleAssign)
		if v.ByRef {
			e.setRole(e.find(v.Op.Span.End, v.Value.Span().Start+1, syntax.TAmpersand), roleRef)
		}
	case *syntax.Unary:
		i := e.at(v.Op.Span.Start)
		switch {
		case v.Op.Kind == syntax.TExclaim:
			e.setRole(i, roleNot)
		case v.Op.Kind.IsCast():
			e.setRole(i, roleCast)
		default:
			e.setRole(i, roleUnary)
		}
	case *syntax.IncDec:
		e.setRole(e.at(v.Op.Span.Start), roleIncDec)
	case *syntax.Ternary:
		e.setRole(e.at(v.Question.Start), roleTernaryQ)
		e.setRole(e.at(v.Colon.Start), roleTernaryColon)
	case *syntax.NullableType:
		e.setRole(e.at(v.Span().Start), roleNullable)
		e.typeNames(v)
	case *syntax.UnionType:
		e.typeOps(v)
		e.typeNames(v)
	case *syntax.IntersectionType:
		e.typeOps(v)
		e.typeNames(v)
	case *syntax.Name:
		e.markIdent(v.Span())
		e.typeNameAt(v)
	case *syntax.Identifier:
		e.markIdent(v.Span())
	case *syntax.ConstFetch:
		if i := e.at(v.Span().Start); i >= 0 && v.Name != nil && v.Name.NameKind == syntax.NameUnqualified {
			e.constName[i] = true
		}
	case *syntax.Param:
		e.attrs(v.Attrs)
		if v.ByRef {
			e.setRole(e.find(afterAttrs(v, v.Attrs), v.Var.Span().Start, syntax.TAmpersand), roleRef)
		}
		if v.Variadic {
			e.setRole(e.find(afterAttrs(v, v.Attrs), v.Var.Span().Start, syntax.TEllipsis), roleSpread)
		}
		if !isNil(v.Default) {
			e.setRole(e.find(v.Var.Span().End, v.Default.Span().Start, syntax.TEqual), roleAssign)
		}
		if len(v.Hooks) > 0 {
			e.setBr(e.find(v.Var.Span().End, v.Hooks[0].Span().Start, syntax.TLBrace), brHooks)
			for _, h := range v.Hooks {
				e.markStart(h)
			}
		}
	case *syntax.Arg:
		if v.Unpack {
			e.setRole(e.at(v.Span().Start), roleSpread)
		}
		if v.Name != nil {
			e.setRole(e.find(v.Name.Span().End, v.Value.Span().Start, syntax.TColon), roleNamedColon)
		}
		if v.ByRef {
			e.setRole(e.at(v.Span().Start), roleRef)
		}
	case *syntax.VariadicPlaceholder:
		e.setRole(e.at(v.Span().Start), roleSpread)
	case *syntax.ArgPlaceholder:
		if v.Name != nil {
			e.setRole(e.find(v.Name.Span().End, v.Span().End, syntax.TColon), roleNamedColon)
		}
	case *syntax.ArgList:
		e.argList(v)
	case *syntax.ArrayItem:
		if v.Unpack {
			e.setRole(e.at(v.Span().Start), roleSpread)
		}
		if v.ByRef && !isNil(v.Value) {
			e.setRole(e.prevSig(v.Value.Span().Start), roleRef)
		}
		if !isNil(v.Key) && !isNil(v.Value) {
			e.setRole(e.find(v.Key.Span().End, v.Value.Span().Start, syntax.TDoubleArrow), roleArrow)
		}
	case *syntax.Array:
		sp := v.Span()
		if v.Short {
			e.setBr(e.at(sp.Start), brShortArray)
		} else {
			e.setBr(e.find(sp.Start, sp.End, syntax.TLParen), brArray)
		}
		for _, it := range v.Items {
			if it != nil {
				e.markStart(it)
			}
		}
	case *syntax.List:
		// list(...): short destructuring parses as an array.
		sp := v.Span()
		i := e.find(sp.Start, sp.End, syntax.TLParen)
		e.setBr(i, brArray)
		if i >= 0 {
			e.destruct[i] = true
		}
		for _, it := range v.Items {
			if it != nil {
				e.markStart(it)
			}
		}
	case *syntax.ArrayDimFetch:
		if !isNil(v.Var) && !v.Curly {
			e.setBr(e.find(v.Var.Span().End, v.Span().End, syntax.TLBracket), brOffset)
		}
	case *syntax.Foreach:
		if arr, ok := v.Value.(*syntax.Array); ok {
			e.destructure(arr)
		}
		e.parenAfter(v.Span().Start, v.Expr.Span().Start+1, brControl)
		if !isNil(v.Key) {
			e.setRole(e.find(v.Key.Span().End, v.Value.Span().Start, syntax.TDoubleArrow), roleArrow)
		}
		if v.ByRef {
			e.setRole(e.prevSig(v.Value.Span().Start), roleRef)
		}
		e.loop(v.Body, v.Alt, v.Span())
	case *syntax.Yield:
		if !isNil(v.Key) && !isNil(v.Value) {
			e.setRole(e.find(v.Key.Span().End, v.Value.Span().Start, syntax.TDoubleArrow), roleArrow)
		}
	case *syntax.MatchArm:
		from := v.Span().Start
		if len(v.Conds) > 0 {
			from = v.Conds[len(v.Conds)-1].Span().End
		}
		e.setRole(e.find(from, v.Body.Span().Start, syntax.TDoubleArrow), roleArrow)
	case *syntax.Match:
		e.parenAfter(v.Span().Start, v.Cond.Span().Start+1, brControl)
		e.setBr(e.find(v.Cond.Span().End, v.Span().End, syntax.TLBrace), brMatch)
		for _, a := range v.Arms {
			e.markStart(a)
		}
	case *syntax.ArrowFunction:
		e.attrs(v.Attrs)
		kw := e.find(afterAttrs(v, v.Attrs), v.Span().End, syntax.TFn)
		e.setRole(kw, roleArrowKeyword)
		if kw >= 0 {
			e.signature(e.toks[kw].End, v.Params, v.ReturnType)
		}
		from := v.Span().Start
		if !isNil(v.ReturnType) {
			from = v.ReturnType.Span().End
		} else if len(v.Params) > 0 {
			from = v.Params[len(v.Params)-1].Span().End
		}
		e.setRole(e.find(from, v.Expr.Span().Start, syntax.TDoubleArrow), roleArrow)
	case *syntax.Closure:
		e.attrs(v.Attrs)
		kw := e.find(afterAttrs(v, v.Attrs), v.Span().End, syntax.TFunction)
		e.setRole(kw, roleClosureKeyword)
		if kw >= 0 {
			e.signature(e.toks[kw].End, v.Params, v.ReturnType)
		}
		if len(v.Uses) > 0 {
			u := e.find(v.Uses[0].Span().Start-uint32(min(int(v.Uses[0].Span().Start), 64)), v.Uses[0].Span().Start, syntax.TUse)
			if u >= 0 {
				e.role[u] = roleUseKeyword
				e.parenAfter(e.toks[u].End, v.Uses[0].Span().Start, brUse)
			}
			for _, cu := range v.Uses {
				e.markStart(cu)
				if cu.ByRef {
					e.setRole(e.at(cu.Span().Start), roleRef)
				}
			}
		}
		e.blockBraces(v.Body, brClosure)
		if v.Body != nil {
			e.stmts(v.Body.Stmts)
		}
	case *syntax.Function:
		e.attrs(v.Attrs)
		e.signature(v.Name.Span().End, v.Params, v.ReturnType)
		e.blockBraces(v.Body, brFunction)
		if v.Body != nil {
			e.stmts(v.Body.Stmts)
		}
	case *syntax.Method:
		e.attrs(v.Attrs)
		e.signature(v.Name.Span().End, v.Params, v.ReturnType)
		e.blockBraces(v.Body, brFunction)
		if v.Body != nil {
			e.stmts(v.Body.Stmts)
		}
	case *syntax.PropertyHook:
		e.attrs(v.Attrs)
		if len(v.Params) > 0 {
			e.signature(v.Name.Span().End, v.Params, nil)
		}
		if b, ok := v.Body.(*syntax.Block); ok && b != nil {
			e.blockBraces(b, brHookBody)
			e.stmts(b.Stmts)
		}
	case *syntax.Property:
		e.attrs(v.Attrs)
		if len(v.Hooks) > 0 {
			from := v.Props[len(v.Props)-1].Span().End
			e.setBr(e.find(from, v.Hooks[0].Span().Start, syntax.TLBrace), brHooks)
			for _, h := range v.Hooks {
				e.markStart(h)
			}
		}
		for _, p := range v.Props {
			if !isNil(p.Default) {
				e.setRole(e.find(p.Var.Span().End, p.Default.Span().Start, syntax.TEqual), roleAssign)
			}
		}
	case *syntax.ConstItem:
		if !isNil(v.Value) {
			e.setRole(e.find(v.Name.Span().End, v.Value.Span().Start, syntax.TEqual), roleAssign)
		}
	case *syntax.StaticVar:
		if !isNil(v.Default) {
			e.setRole(e.find(v.Var.Span().End, v.Default.Span().Start, syntax.TEqual), roleAssign)
		}
	case *syntax.ClassConst:
		e.attrs(v.Attrs)
	case *syntax.ConstStmt:
		e.attrs(v.Attrs)
	case *syntax.EnumCase:
		e.attrs(v.Attrs)
		if !isNil(v.Value) {
			e.setRole(e.find(v.Name.Span().End, v.Value.Span().Start, syntax.TEqual), roleAssign)
		}
	case *syntax.ClassLike:
		e.attrs(v.Attrs)
		kind := brClass
		if v.Name == nil {
			kind = brAnonClass
			e.argList(v.Args)
		}
		e.setBr(e.at(v.LBrace.Start), kind)
		if !isNil(v.EnumType) {
			if i := e.prevSig(v.EnumType.Span().Start); i >= 0 && e.toks[i].Kind == syntax.TColon {
				e.role[i] = roleEnumColon
			}
		}
		e.stmts(v.Members)
	case *syntax.New:
		e.setRole(e.at(v.Span().Start), roleNewKeyword)
	case *syntax.Attribute:
		e.argList(v.Args)
	case *syntax.FuncCall:
		e.argList(v.Args)
	case *syntax.MethodCall:
		e.argList(v.Args)
	case *syntax.StaticCall:
		e.argList(v.Args)
	case *syntax.Exit:
		e.argList(v.Args)
	case *syntax.Clone:
		e.argList(v.Args)
	case *syntax.Isset:
		e.setBr(e.find(v.Span().Start, v.Span().End, syntax.TLParen), brCall)
		for _, x := range v.Vars {
			e.markStart(x)
		}
	case *syntax.Empty:
		e.setBr(e.find(v.Span().Start, v.Span().End, syntax.TLParen), brCall)
	case *syntax.Eval:
		e.setBr(e.find(v.Span().Start, v.Span().End, syntax.TLParen), brCall)
	case *syntax.Unset:
		e.setBr(e.find(v.Span().Start, v.Span().End, syntax.TLParen), brCall)
		for _, x := range v.Vars {
			e.markStart(x)
		}
	case *syntax.Paren:
		e.setBr(e.at(v.Span().Start), brGroup)
	case *syntax.If:
		e.setBr(e.find(v.Span().Start, v.Cond.Span().Start+1, syntax.TLParen), brControl)
		e.conditional(v.Body, v.Alt, v.Span())
		for _, ei := range v.ElseIfs {
			e.markStart(ei)
			e.setBr(e.find(ei.Span().Start, ei.Cond.Span().Start+1, syntax.TLParen), brControl)
			e.conditional(ei.Body, v.Alt, v.Span())
		}
		if v.Else != nil {
			e.markStart(v.Else)
			e.conditional(v.Else.Body, v.Alt, v.Span())
		}
	case *syntax.While:
		e.setBr(e.find(v.Span().Start, v.Cond.Span().Start+1, syntax.TLParen), brControl)
		e.loop(v.Body, v.Alt, v.Span())
	case *syntax.DoWhile:
		e.loop(v.Body, false, v.Span())
		e.setBr(e.find(v.Body.Span().End, v.Cond.Span().Start+1, syntax.TLParen), brControl)
		if i := e.nextSigIndex(v.Body.Span().End); i >= 0 {
			e.start[i] = true // while
		}
	case *syntax.For:
		first := e.find(v.Span().Start, v.Span().End, syntax.TLParen)
		e.setBr(first, brControl)
		if first >= 0 {
			depth := 0
			for i := first + 1; i < len(e.toks); i++ {
				switch e.toks[i].Kind {
				case syntax.TLParen, syntax.TLBracket, syntax.TLBrace:
					depth++
				case syntax.TRParen, syntax.TRBracket, syntax.TRBrace:
					depth--
				case syntax.TSemicolon:
					if depth == 0 {
						e.role[i] = roleForSemicolon
					}
				}
				if depth < 0 {
					break
				}
			}
		}
		e.loop(v.Body, v.Alt, v.Span())
	case *syntax.Switch:
		e.setBr(e.find(v.Span().Start, v.Cond.Span().Start+1, syntax.TLParen), brControl)
		if v.Alt {
			e.keepIndent(v.Span())
		} else {
			e.setBr(e.find(v.Cond.Span().End, v.Span().End, syntax.TLBrace), brSwitch)
		}
		for _, c := range v.Cases {
			e.markStart(c)
			from := c.Span().Start
			if !isNil(c.Cond) {
				from = c.Cond.Span().End
			}
			colon := -1
			for i := e.index(from); i < len(e.toks); i++ {
				if k := e.toks[i].Kind; k == syntax.TColon || k == syntax.TSemicolon {
					colon = i
					break
				}
			}
			e.setRole(colon, roleCaseColon)
			e.setRole(e.index(c.Span().Start), roleCaseLabel)
			e.stmts(c.Stmts)
			if len(c.Stmts) > 0 {
				e.regions = append(e.regions, region{c.Stmts[0].Span().Start, c.Span().End})
			}
		}
	case *syntax.Try:
		e.blockBraces(v.Body, brControlBlock)
		e.stmts(v.Body.Stmts)
		for _, c := range v.Catches {
			e.markStart(c)
			e.setBr(e.find(c.Span().Start, c.Body.Span().Start, syntax.TLParen), brControl)
			for i := e.index(c.Span().Start); i < len(e.toks) && e.toks[i].Start < c.Body.Span().Start; i++ {
				if e.toks[i].Kind == syntax.TBar {
					e.role[i] = roleCatchBar
				}
			}
			e.blockBraces(c.Body, brControlBlock)
			e.stmts(c.Body.Stmts)
		}
		if v.Finally != nil {
			e.markStart(v.Finally)
			e.blockBraces(v.Finally.Body, brControlBlock)
			e.stmts(v.Finally.Body.Stmts)
		}
	case *syntax.Block:
		if !v.Alt {
			if i := e.at(v.Span().Start); i >= 0 && e.br[i] == brNone && e.toks[i].Kind == syntax.TLBrace {
				e.br[i] = brControlBlock
			}
		}
		e.stmts(v.Stmts)
	case *syntax.Declare:
		e.setBr(e.find(v.Span().Start, v.Span().End, syntax.TLParen), brDeclare)
		for _, it := range v.Items {
			if !isNil(it.Value) {
				e.setRole(e.find(it.Key.Span().End, it.Value.Span().Start, syntax.TEqual), roleDeclareEq)
			}
		}
		if !isNil(v.Body) {
			if v.Alt {
				e.keepIndent(v.Span())
			}
			e.blockBraces(v.Body, brControlBlock)
			e.bodyRegion(v.Body)
		}
	case *syntax.Namespace:
		if v.Braced {
			from := v.Span().Start
			if v.Name != nil {
				from = v.Name.Span().End
			}
			e.setBr(e.find(from, v.Span().End, syntax.TLBrace), brNamespace)
		}
		e.stmts(v.Stmts)
	case *syntax.Use:
		if v.Prefix != nil {
			e.setBr(e.find(v.Prefix.Span().End, v.Span().End, syntax.TLBrace), brGroupUse)
			for _, it := range v.Items {
				e.markStart(it)
			}
		}
	case *syntax.Label:
		e.setRole(e.nextSigIndex(v.Name.Span().End), roleLabelColon)
	case *syntax.TraitUse:
		if len(v.Adaptations) > 0 {
			e.setBr(e.find(v.Traits[len(v.Traits)-1].Span().End, v.Span().End, syntax.TLBrace), brControlBlock)
			for _, a := range v.Adaptations {
				e.markStart(a)
			}
		}
	}
}

// conditional marks one branch body of an if statement.
func (e *engine) conditional(body syntax.Stmt, alt bool, whole syntax.Span) {
	if alt {
		e.keepIndent(whole)
		return
	}
	e.blockBraces(body, brControlBlock)
	e.bodyRegion(body)
}

// loop marks the body of a loop.
func (e *engine) loop(body syntax.Stmt, alt bool, whole syntax.Span) {
	if alt {
		e.keepIndent(whole)
		return
	}
	e.blockBraces(body, brControlBlock)
	e.bodyRegion(body)
}

func (e *engine) markIdent(sp syntax.Span) {
	for i := e.index(sp.Start); i < len(e.toks) && e.toks[i].End <= sp.End; i++ {
		e.ident[i] = true
	}
}

// typeNames marks the built-in type names inside the type t.
func (e *engine) typeNames(t syntax.Expr) {
	syntax.Children(t, func(c syntax.Node) {
		if nm, ok := c.(*syntax.Name); ok {
			e.typeNameAt(nm)
		}
	})
}

// typeNameAt marks nm as a type name when it is in a type position.
func (e *engine) typeNameAt(nm *syntax.Name) {
	if !isTypePosition(nm) {
		return
	}
	if i := e.at(nm.Span().Start); i >= 0 {
		e.typeName[i] = true
	}
}

// isTypePosition reports whether nm is (part of) a declared type.
func isTypePosition(nm *syntax.Name) bool {
	var child syntax.Node = nm
	for p := nm.Parent(); ; child, p = p, p.Parent() {
		switch v := p.(type) {
		case *syntax.NullableType, *syntax.UnionType, *syntax.IntersectionType:
			continue
		case *syntax.Param:
			return v.Type == child
		case *syntax.Property:
			return v.Type == child
		case *syntax.ClassConst:
			return v.Type == child
		case *syntax.Function:
			return v.ReturnType == child
		case *syntax.Method:
			return v.ReturnType == child
		case *syntax.Closure:
			return v.ReturnType == child
		case *syntax.ArrowFunction:
			return v.ReturnType == child
		case *syntax.ClassLike:
			return v.EnumType == child
		default:
			return false
		}
	}
}

// destructure marks the array a and the arrays nested in it as the target
// of a destructuring assignment.
func (e *engine) destructure(a *syntax.Array) {
	if i := e.at(a.Span().Start); i >= 0 {
		e.destruct[i] = true // only [...] can be assigned to
	}
	for _, it := range a.Items {
		if it == nil {
			continue
		}
		if inner, ok := it.Value.(*syntax.Array); ok {
			e.destructure(inner)
		}
	}
}
