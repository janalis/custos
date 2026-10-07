package syntax

import "custos/internal/phpver"

// Operator precedence levels (higher binds tighter), following
// zend_language_parser.y.
const (
	precLowest     = 1 // or
	precXor        = 2
	precAnd        = 3
	precAssign     = 4
	precTernary    = 5
	precCoalesce   = 6
	precBoolOr     = 7
	precBoolAnd    = 8
	precBitOr      = 9
	precBitXor     = 10
	precBitAnd     = 11
	precEquality   = 12
	precCompare    = 13
	precPipe       = 14
	precConcat     = 15 // PHP >= 8.0; below that '.' is additive
	precShift      = 16
	precAdditive   = 17
	precMul        = 18
	precInstanceof = 19
	precPow        = 20
)

// binPrec returns the precedence of a binary operator token.
func (p *parser) binPrec(k TokenKind) (prec int, right bool, ok bool) {
	switch k {
	case TOr:
		return precLowest, false, true
	case TXor:
		return precXor, false, true
	case TAnd:
		return precAnd, false, true
	case TQuestion:
		return precTernary, false, true
	case TCoalesce:
		return precCoalesce, true, true
	case TBooleanOr:
		return precBoolOr, false, true
	case TBooleanAnd:
		return precBoolAnd, false, true
	case TBar:
		return precBitOr, false, true
	case TCaret:
		return precBitXor, false, true
	case TAmpersand:
		return precBitAnd, false, true
	case TIsEqual, TIsNotEqual, TIsIdentical, TIsNotIdentical, TSpaceship:
		return precEquality, false, true
	case TLess, TIsSmallerOrEqual, TGreater, TIsGreaterOrEqual:
		return precCompare, false, true
	case TPipe:
		return precPipe, false, true
	case TDot:
		if p.ver.Below(phpver.PHP80) {
			return precAdditive, false, true
		}
		return precConcat, false, true
	case TSl, TSr:
		return precShift, false, true
	case TPlus, TMinus:
		return precAdditive, false, true
	case TMul, TDiv, TMod:
		return precMul, false, true
	case TInstanceof:
		return precInstanceof, false, true
	case TPow:
		return precPow, true, true
	}
	return 0, false, false
}

func (p *parser) parseExpr(min int) Expr {
	start := p.start()
	return p.parseBinaryRHS(p.parseUnary(), min, start)
}

func (p *parser) parseBinaryRHS(left Expr, min int, start uint32) Expr {
	for {
		op := p.tok()
		prec, right, ok := p.binPrec(op.Kind)
		if !ok || prec < min {
			return left
		}
		p.advance()
		switch op.Kind {
		case TQuestion:
			n := &Ternary{Cond: left, Question: Span{op.Start, op.End}}
			if c, ok := p.accept(TColon); ok {
				n.Colon = Span{c.Start, c.End}
			} else {
				n.Then = p.parseExpr(precLowest)
				c := p.expect(TColon)
				n.Colon = Span{c.Start, c.End}
			}
			n.Else = p.parseExpr(precTernary + 1)
			left = fin(p, n, start)
		case TInstanceof:
			n := &Instanceof{Expr: left, Class: p.parseClassRef(false)}
			left = fin(p, n, start)
		default:
			next := prec + 1
			if right {
				next = prec
			}
			n := put(&p.slabs.sBinary, Binary{Left: left, Op: p.ref(op), Right: p.parseExpr(next)})
			left = fin(p, n, start)
		}
	}
}

func (p *parser) parseUnary() Expr {
	t := p.tok()
	start := t.Start
	switch t.Kind {
	case TExclaim:
		p.advance()
		return fin(p, &Unary{Op: p.ref(t), Expr: p.parseExpr(precInstanceof)}, start)
	case TMinus, TPlus, TTilde, TAt, TIntCast, TDoubleCast, TStringCast, TArrayCast,
		TObjectCast, TBoolCast, TUnsetCast, TVoidCast:
		p.advance()
		return fin(p, &Unary{Op: p.ref(t), Expr: p.parseExpr(precPow)}, start)
	case TInc, TDec:
		p.advance()
		return fin(p, &IncDec{Op: p.ref(t), Prefix: true, Var: p.parseUnary()}, start)
	case TNew:
		return p.parsePostfix(p.parseNew(), start)
	case TClone:
		p.advance()
		if p.at(TLParen) && p.ver.AtLeast(phpver.PHP85) {
			// PHP 8.5: clone is function-like, `clone($obj, $withProperties)`.
			args := p.parseArgs()
			n := &Clone{}
			for i, a := range args.Args {
				if arg, ok := a.(*Arg); ok {
					switch i {
					case 0:
						n.Expr = arg.Value
					case 1:
						n.With = arg.Value
					}
				}
			}
			if n.Expr == nil {
				n.Expr = spanOf(&BadExpr{}, p.missing())
			}
			return p.parsePostfix(fin(p, n, start), start)
		}
		return fin(p, &Clone{Expr: p.parseUnary()}, start)
	case TPrint:
		p.advance()
		return fin(p, &Print{Expr: p.parseExpr(precAssign)}, start)
	case TYield:
		p.advance()
		n := &Yield{}
		if !p.atExprEnd() {
			v := p.parseExpr(precAssign)
			if _, ok := p.accept(TDoubleArrow); ok {
				n.Key = v
				n.Value = p.parseExpr(precAssign)
			} else {
				n.Value = v
			}
		}
		return fin(p, n, start)
	case TYieldFrom:
		p.advance()
		return fin(p, &YieldFrom{Expr: p.parseExpr(precAssign)}, start)
	case TInclude, TIncludeOnce, TRequire, TRequireOnce:
		p.advance()
		return fin(p, &Include{Keyword: p.ref(t), Expr: p.parseExpr(precLowest)}, start)
	case TThrow:
		p.advance()
		return fin(p, &Throw{Expr: p.parseExpr(precLowest)}, start)
	case TFunction, TFn:
		return p.parsePostfix(p.parseClosureLike(nil, start), start)
	case TStatic:
		if k := p.peekKind(1); k == TFunction || k == TFn {
			return p.parsePostfix(p.parseClosureLike(nil, start), start)
		}
	case TAttribute:
		attrs := p.parseAttributes()
		return p.parsePostfix(p.parseClosureLike(attrs, start), start)
	case TAmpersand:
		// By-reference marker in an unexpected place (e.g. legacy `=& new`):
		// parse the operand and keep going.
		p.errorAt(t, "unexpected '&'")
		p.advance()
		return p.parseUnary()
	}
	return p.parsePostfix(p.parsePrimary(), start)
}

// atExprEnd reports whether the current token cannot start an expression
// operand (used for bare yield).
func (p *parser) atExprEnd() bool {
	switch p.kind() {
	case TSemicolon, TCloseTag, TEOF, TRParen, TRBracket, TComma, TRBrace, TDoubleArrow, TColon, TAs:
		return true
	}
	return false
}

func isAssignable(e Expr) bool {
	switch e.(type) {
	case *Variable, *ArrayDimFetch, *PropertyFetch, *StaticPropertyFetch, *List, *Array:
		return true
	}
	return false
}

func isDereferenceable(e Expr) bool {
	switch e.(type) {
	case *BadExpr, *IncDec:
		return false
	}
	return true
}

// parsePostfix parses member access, calls, offsets, postfix ++/-- and
// assignment after a primary expression.
func (p *parser) parsePostfix(e Expr, start uint32) Expr {
	for {
		t := p.tok()
		switch t.Kind {
		case TLBracket:
			if !isDereferenceable(e) {
				return e
			}
			p.advance()
			n := put(&p.slabs.sArrayDimFetch, ArrayDimFetch{Var: e})
			if !p.at(TRBracket) {
				n.Dim = p.parseExpr(precLowest)
			}
			p.expect(TRBracket)
			e = fin(p, n, start)
		case TLBrace:
			if !p.legacyCurlyOffsets() || !isOffsetTarget(e) {
				return e
			}
			p.advance()
			n := put(&p.slabs.sArrayDimFetch, ArrayDimFetch{Var: e, Curly: true, Dim: p.parseExpr(precLowest)})
			p.expect(TRBrace)
			e = fin(p, n, start)
		case TObjectOperator, TNullsafeObjectOperator:
			p.advance()
			name := p.parseMemberName()
			nullsafe := t.Kind == TNullsafeObjectOperator
			if p.at(TLParen) {
				e = fin(p, put(&p.slabs.sMethodCall, MethodCall{Var: e, Name: name, NullSafe: nullsafe, Args: p.parseArgs()}), start)
			} else {
				e = fin(p, put(&p.slabs.sPropertyFetch, PropertyFetch{Var: e, Name: name, NullSafe: nullsafe}), start)
			}
		case TPaamayimNekudotayim:
			p.advance()
			e = p.parseStaticMember(e, start)
		case TLParen:
			e = fin(p, put(&p.slabs.sFuncCall, FuncCall{Name: e, Args: p.parseArgs()}), start)
		case TInc, TDec:
			if !isAssignable(e) {
				return e
			}
			p.advance()
			e = fin(p, &IncDec{Var: e, Op: p.ref(t)}, start)
		case TEqual:
			if !isAssignable(e) {
				return e
			}
			p.advance()
			n := put(&p.slabs.sAssign, Assign{Var: e, Op: p.ref(t)})
			if _, ok := p.accept(TAmpersand); ok {
				n.ByRef = true
			}
			n.Value = p.parseExpr(precAssign)
			return fin(p, n, start)
		default:
			if t.Kind.IsAssignOp() && isAssignable(e) {
				p.advance()
				n := put(&p.slabs.sAssign, Assign{Var: e, Op: p.ref(t), Value: p.parseExpr(precAssign)})
				return fin(p, n, start)
			}
			return e
		}
	}
}

// legacyCurlyOffsets reports whether `$a{0}` offsets are accepted.
func (p *parser) legacyCurlyOffsets() bool { return p.permissive || p.ver.Below(phpver.PHP80) }

// isOffsetTarget limits legacy `$a{0}` offsets to variable-like bases.
func isOffsetTarget(e Expr) bool {
	switch e.(type) {
	case *Variable, *ArrayDimFetch, *PropertyFetch, *StaticPropertyFetch, *ConstFetch, *ClassConstFetch:
		return true
	}
	return false
}

// parseMemberName parses the name after -> / ?->.
func (p *parser) parseMemberName() Expr {
	t := p.tok()
	switch {
	case t.Kind == TString || t.Kind.IsKeyword():
		p.advance()
		return spanOf(put(&p.slabs.sIdentifier, Identifier{Value: p.text(t)}), Span{t.Start, t.End})
	case t.Kind == TVariable:
		return p.parseVariableToken()
	case t.Kind == TDollar:
		return p.parseSimpleVariable()
	case t.Kind == TLBrace:
		p.advance()
		e := p.parseExpr(precLowest)
		p.expect(TRBrace)
		return e
	}
	p.errorAt(t, "expected member name, found "+p.describe(t))
	return spanOf(&BadExpr{}, p.missing())
}

// parseStaticMember parses what follows `::`.
func (p *parser) parseStaticMember(class Expr, start uint32) Expr {
	t := p.tok()
	switch {
	case t.Kind == TVariable || t.Kind == TDollar:
		v := p.parseSimpleVariable()
		if p.at(TLParen) {
			return fin(p, put(&p.slabs.sStaticCall, StaticCall{Class: class, Name: v, Args: p.parseArgs()}), start)
		}
		return fin(p, &StaticPropertyFetch{Class: class, Name: v}, start)
	case t.Kind == TString || t.Kind.IsKeyword():
		p.advance()
		id := spanOf(put(&p.slabs.sIdentifier, Identifier{Value: p.text(t)}), Span{t.Start, t.End})
		if p.at(TLParen) {
			return fin(p, put(&p.slabs.sStaticCall, StaticCall{Class: class, Name: id, Args: p.parseArgs()}), start)
		}
		return fin(p, put(&p.slabs.sClassConstFetch, ClassConstFetch{Class: class, Name: id}), start)
	case t.Kind == TLBrace:
		p.advance()
		e := p.parseExpr(precLowest)
		p.expect(TRBrace)
		if p.at(TLParen) {
			return fin(p, put(&p.slabs.sStaticCall, StaticCall{Class: class, Name: e, Args: p.parseArgs()}), start)
		}
		return fin(p, put(&p.slabs.sClassConstFetch, ClassConstFetch{Class: class, Name: e}), start)
	}
	p.errorAt(t, "expected member name after '::', found "+p.describe(t))
	return fin(p, put(&p.slabs.sClassConstFetch, ClassConstFetch{Class: class, Name: spanOf(&BadExpr{}, p.missing())}), start)
}

// parseSimpleVariable parses $name, $$var, ${expr}.
func (p *parser) parseSimpleVariable() Expr {
	t := p.tok()
	switch t.Kind {
	case TVariable:
		return p.parseVariableToken()
	case TDollar:
		p.advance()
		if p.at(TLBrace) {
			p.advance()
			e := p.parseExpr(precLowest)
			p.expect(TRBrace)
			return fin(p, put(&p.slabs.sVariable, Variable{NameExpr: e}), t.Start)
		}
		inner := p.parseSimpleVariable()
		return fin(p, put(&p.slabs.sVariable, Variable{NameExpr: inner}), t.Start)
	}
	p.errorAt(t, "expected variable, found "+p.describe(t))
	return spanOf(&BadExpr{}, p.missing())
}

func (p *parser) parseArgs() *ArgList {
	start := p.start()
	p.expect(TLParen)
	n := put(&p.slabs.sArgList, ArgList{})
	for !p.at(TRParen) && !p.at(TEOF) {
		before := p.pos
		as := p.start()
		switch {
		case p.at(TEllipsis) && p.peekKind(1) == TRParen:
			p.advance()
			n.Args = append(n.Args, fin(p, &VariadicPlaceholder{}, as))
		case p.at(TEllipsis):
			p.advance()
			n.Args = append(n.Args, fin(p, put(&p.slabs.sArg, Arg{Unpack: true, Value: p.parseExpr(precLowest)}), as))
		case (p.at(TString) || p.kind().IsKeyword()) && p.peekKind(1) == TColon:
			id := p.parseIdentifier()
			p.advance()
			n.Args = append(n.Args, fin(p, put(&p.slabs.sArg, Arg{Name: id, Value: p.parseExpr(precLowest)}), as))
		case p.at(TAmpersand):
			p.advance()
			n.Args = append(n.Args, fin(p, put(&p.slabs.sArg, Arg{ByRef: true, Value: p.parseExpr(precLowest)}), as))
		default:
			n.Args = append(n.Args, fin(p, put(&p.slabs.sArg, Arg{Value: p.parseExpr(precLowest)}), as))
		}
		if _, ok := p.accept(TComma); !ok || p.pos == before {
			break
		}
	}
	p.expect(TRParen)
	return fin(p, n, start)
}

// parseClassRef parses a class reference after `new` or `instanceof`:
// a name, static/self/parent, a variable with property/offset/static-property
// chains (no calls), or a parenthesised expression.
func (p *parser) parseClassRef(forNew bool) Expr {
	t := p.tok()
	start := t.Start
	switch t.Kind {
	case TString, TNameQualified, TNameFullyQualified, TNameRelative, TStatic:
		return p.parseName()
	case TLParen:
		p.advance()
		e := p.parseExpr(precLowest)
		p.expect(TRParen)
		return fin(p, put(&p.slabs.sParen, Paren{Expr: e}), start)
	case TVariable, TDollar:
		var e Expr = p.parseSimpleVariable()
		for {
			switch p.kind() {
			case TLBracket:
				p.advance()
				n := put(&p.slabs.sArrayDimFetch, ArrayDimFetch{Var: e})
				if !p.at(TRBracket) {
					n.Dim = p.parseExpr(precLowest)
				}
				p.expect(TRBracket)
				e = fin(p, n, start)
			case TLBrace:
				if !p.legacyCurlyOffsets() {
					return e
				}
				p.advance()
				n := put(&p.slabs.sArrayDimFetch, ArrayDimFetch{Var: e, Curly: true, Dim: p.parseExpr(precLowest)})
				p.expect(TRBrace)
				e = fin(p, n, start)
			case TObjectOperator, TNullsafeObjectOperator:
				ns := p.advance().Kind == TNullsafeObjectOperator
				e = fin(p, put(&p.slabs.sPropertyFetch, PropertyFetch{Var: e, Name: p.parseMemberName(), NullSafe: ns}), start)
			case TPaamayimNekudotayim:
				if k := p.peekKind(1); k != TVariable && k != TDollar {
					return e
				}
				p.advance()
				e = fin(p, &StaticPropertyFetch{Class: e, Name: p.parseSimpleVariable()}, start)
			default:
				return e
			}
		}
	}
	if !forNew {
		// instanceof accepts arbitrary expressions in recovery mode.
		return p.parseUnary()
	}
	p.errorAt(t, "expected class name, found "+p.describe(t))
	return spanOf(&BadExpr{}, p.missing())
}

func (p *parser) parseNew() Expr {
	start := p.start()
	p.advance() // new
	n := &New{}
	if p.at(TClass) || p.at(TAttribute) || ((p.at(TReadonly) || p.at(TFinal) || p.at(TAbstract)) && p.peekKind(1) == TClass) {
		var attrs []*AttributeGroup
		if p.at(TAttribute) {
			attrs = p.parseAttributes()
		}
		n.Class = p.parseAnonymousClass(attrs)
		return fin(p, n, start)
	}
	n.Class = p.parseClassRef(true)
	if p.at(TLParen) {
		n.Args = p.parseArgs()
	}
	return fin(p, n, start)
}

func (p *parser) parsePrimary() Expr {
	t := p.tok()
	start := t.Start
	switch t.Kind {
	case TVariable, TDollar:
		return p.parseSimpleVariable()
	case TString, TNameQualified, TNameFullyQualified, TNameRelative:
		name := p.parseName()
		switch p.kind() {
		case TLParen:
			return fin(p, put(&p.slabs.sFuncCall, FuncCall{Name: name, Args: p.parseArgs()}), start)
		case TPaamayimNekudotayim:
			return name // class reference; handled by parsePostfix
		}
		return fin(p, put(&p.slabs.sConstFetch, ConstFetch{Name: name}), start)
	case TStatic:
		// static::  (static closures are handled in parseUnary)
		return p.parseName()
	case TArray:
		if p.peekKind(1) == TLParen {
			return p.parseArrayLiteral(false)
		}
		return p.parseName()
	case TLBracket:
		return p.parseArrayLiteral(true)
	case TList:
		return p.parseList()
	case TLNumber:
		p.advance()
		return spanOf(put(&p.slabs.sLiteral, Literal{LitKind: LitInt, Raw: p.text(t)}), Span{t.Start, t.End})
	case TDNumber:
		p.advance()
		return spanOf(put(&p.slabs.sLiteral, Literal{LitKind: LitFloat, Raw: p.text(t)}), Span{t.Start, t.End})
	case TConstantEncapsedString:
		p.advance()
		return spanOf(put(&p.slabs.sLiteral, Literal{LitKind: LitString, Raw: p.text(t)}), Span{t.Start, t.End})
	case TDoubleQuote:
		return p.parseInterpolated(TDoubleQuote, false, false)
	case TBacktick:
		return p.parseInterpolated(TBacktick, false, true)
	case TStartHeredoc:
		return p.parseInterpolated(TEndHeredoc, true, false)
	case TIsset:
		p.advance()
		n := &Isset{}
		p.expect(TLParen)
		for !p.at(TRParen) && !p.at(TEOF) {
			before := p.pos
			n.Vars = append(n.Vars, p.parseExpr(precLowest))
			if _, ok := p.accept(TComma); !ok || p.pos == before {
				break
			}
		}
		p.expect(TRParen)
		return fin(p, n, start)
	case TEmpty:
		p.advance()
		p.expect(TLParen)
		n := &Empty{Expr: p.parseExpr(precLowest)}
		p.expect(TRParen)
		return fin(p, n, start)
	case TEval:
		p.advance()
		p.expect(TLParen)
		n := &Eval{Expr: p.parseExpr(precLowest)}
		p.expect(TRParen)
		return fin(p, n, start)
	case TExit:
		p.advance()
		n := &Exit{Keyword: p.ref(t)}
		if p.at(TLParen) {
			n.Args = p.parseArgs()
		}
		return fin(p, n, start)
	case TLParen:
		p.advance()
		e := p.parseExpr(precLowest)
		p.expect(TRParen)
		return fin(p, put(&p.slabs.sParen, Paren{Expr: e}), start)
	case TMatch:
		return p.parseMatch()
	case TLine, TFile, TDir, TClassC, TTraitC, TMethodC, TFuncC, TNsC, TPropertyC:
		p.advance()
		return spanOf(&MagicConst{Token: p.ref(t)}, Span{t.Start, t.End})
	}
	if t.Kind.IsKeyword() && p.peekKind(1) == TPaamayimNekudotayim {
		// e.g. `parent::` is TString, but tolerate other keywords as class refs.
		return p.parseName()
	}
	p.errorAt(t, "unexpected "+p.describe(t))
	if !p.atRecoveryPoint() {
		p.advance()
		return fin(p, &BadExpr{}, start)
	}
	return spanOf(&BadExpr{}, p.missing())
}

// atRecoveryPoint reports tokens that should not be swallowed by expression
// error recovery.
func (p *parser) atRecoveryPoint() bool {
	switch p.kind() {
	case TSemicolon, TCloseTag, TEOF, TRBrace, TRParen, TRBracket, TComma, TLBrace:
		return true
	}
	return false
}

func (p *parser) parseArrayItems(end TokenKind) []*ArrayItem {
	var items []*ArrayItem
	for !p.at(end) && !p.at(TEOF) {
		before := p.pos
		if p.at(TComma) {
			// Skipped slot (destructuring).
			items = append(items, nil)
			p.advance()
			continue
		}
		is := p.start()
		it := put(&p.slabs.sArrayItem, ArrayItem{})
		switch {
		case p.at(TEllipsis):
			p.advance()
			it.Unpack = true
			it.Value = p.parseExpr(precLowest)
		case p.at(TAmpersand):
			p.advance()
			it.ByRef = true
			it.Value = p.parseExpr(precLowest)
		default:
			v := p.parseExpr(precLowest)
			if _, ok := p.accept(TDoubleArrow); ok {
				it.Key = v
				if _, ok := p.accept(TAmpersand); ok {
					it.ByRef = true
				}
				it.Value = p.parseExpr(precLowest)
			} else {
				it.Value = v
			}
		}
		items = append(items, fin(p, it, is))
		if _, ok := p.accept(TComma); !ok || p.pos == before {
			break
		}
	}
	return items
}

func (p *parser) parseArrayLiteral(short bool) Expr {
	start := p.start()
	n := &Array{Short: short}
	if short {
		p.advance()
		n.Items = p.parseArrayItems(TRBracket)
		p.expect(TRBracket)
	} else {
		p.advance()
		p.expect(TLParen)
		n.Items = p.parseArrayItems(TRParen)
		p.expect(TRParen)
	}
	return fin(p, n, start)
}

func (p *parser) parseList() Expr {
	start := p.start()
	p.advance()
	p.expect(TLParen)
	n := &List{Items: p.parseArrayItems(TRParen)}
	p.expect(TRParen)
	return fin(p, n, start)
}

func (p *parser) parseMatch() Expr {
	start := p.start()
	p.advance()
	n := &Match{Cond: p.parseParenExpr()}
	p.expect(TLBrace)
	for !p.at(TRBrace) && !p.at(TEOF) {
		before := p.pos
		as := p.start()
		arm := &MatchArm{}
		if p.at(TDefault) {
			p.advance()
			p.accept(TComma)
		} else {
			for !p.at(TDoubleArrow) && !p.at(TEOF) {
				b := p.pos
				arm.Conds = append(arm.Conds, p.parseExpr(precLowest))
				if _, ok := p.accept(TComma); !ok || p.pos == b {
					break
				}
			}
		}
		p.expect(TDoubleArrow)
		arm.Body = p.parseExpr(precLowest)
		n.Arms = append(n.Arms, fin(p, arm, as))
		if _, ok := p.accept(TComma); !ok || p.pos == before {
			break
		}
	}
	p.expect(TRBrace)
	return fin(p, n, start)
}

// parseInterpolated parses "...", `...` and heredoc bodies.
func (p *parser) parseInterpolated(end TokenKind, heredoc, backtick bool) Expr {
	start := p.start()
	p.advance() // opening token
	n := &InterpolatedString{Heredoc: heredoc, Backtick: backtick}
	for !p.at(end) && !p.at(TEOF) {
		before := p.pos
		t := p.tok()
		switch t.Kind {
		case TEncapsedAndWhitespace:
			p.advance()
			n.Parts = append(n.Parts, spanOf(put(&p.slabs.sStringPart, StringPart{Raw: p.text(t)}), Span{t.Start, t.End}))
		case TVariable:
			n.Parts = append(n.Parts, p.parseEncapsVar())
		case TCurlyOpen:
			p.advance()
			n.Parts = append(n.Parts, p.parseExpr(precLowest))
			p.expect(TRBrace)
		case TDollarOpenCurlyBraces:
			ds := t.Start
			p.advance()
			if vn, ok := p.accept(TStringVarname); ok {
				var e Expr = spanOf(put(&p.slabs.sVariable, Variable{Name: p.text(vn)}), Span{vn.Start, vn.End})
				if _, ok := p.accept(TLBracket); ok {
					d := put(&p.slabs.sArrayDimFetch, ArrayDimFetch{Var: e, Dim: p.parseExpr(precLowest)})
					p.expect(TRBracket)
					e = fin(p, d, vn.Start)
				}
				p.expect(TRBrace)
				// Span covers ${...} for exact replacement ranges.
				e.base().span = Span{ds, p.lastEnd}
				n.Parts = append(n.Parts, e)
			} else {
				e := p.parseExpr(precLowest)
				p.expect(TRBrace)
				n.Parts = append(n.Parts, fin(p, put(&p.slabs.sVariable, Variable{NameExpr: e}), ds))
			}
		default:
			p.errorAt(t, "unexpected "+p.describe(t)+" in string")
			p.advance()
		}
		if p.pos == before {
			p.advance()
		}
	}
	p.expect(end)
	if !backtick {
		allLiteral := true
		for _, part := range n.Parts {
			if _, ok := part.(*StringPart); !ok {
				allLiteral = false
				break
			}
		}
		if allLiteral && heredoc {
			return fin(p, put(&p.slabs.sLiteral, Literal{LitKind: LitString, Raw: string(p.src[start:p.lastEnd])}), start)
		}
	}
	return fin(p, n, start)
}

// parseEncapsVar parses $var, $var[..], $var->prop inside strings.
func (p *parser) parseEncapsVar() Expr {
	start := p.start()
	var e Expr = p.parseVariableToken()
	switch p.kind() {
	case TLBracket:
		p.advance()
		d := put(&p.slabs.sArrayDimFetch, ArrayDimFetch{Var: e})
		t := p.tok()
		switch t.Kind {
		case TNumString, TString:
			p.advance()
			kind := LitString
			if t.Kind == TNumString {
				kind = LitInt
			}
			d.Dim = spanOf(put(&p.slabs.sLiteral, Literal{LitKind: kind, Raw: p.text(t)}), Span{t.Start, t.End})
		case TVariable:
			d.Dim = p.parseVariableToken()
		case TMinus:
			p.advance()
			num := p.expect(TNumString)
			d.Dim = spanOf(put(&p.slabs.sLiteral, Literal{LitKind: LitInt, Raw: string(p.src[t.Start:num.End])}), Span{t.Start, num.End})
		default:
			p.errorAt(t, "invalid string offset")
		}
		p.expect(TRBracket)
		e = fin(p, d, start)
	case TObjectOperator, TNullsafeObjectOperator:
		ns := p.advance().Kind == TNullsafeObjectOperator
		name := p.tok()
		p.expect(TString)
		id := spanOf(put(&p.slabs.sIdentifier, Identifier{Value: p.text(name)}), Span{name.Start, name.End})
		e = fin(p, put(&p.slabs.sPropertyFetch, PropertyFetch{Var: e, Name: id, NullSafe: ns}), start)
	}
	return e
}
