package syntax

// ---- attributes ---------------------------------------------------------------------

func (p *parser) parseAttributes() []*AttributeGroup {
	var groups []*AttributeGroup
	for p.at(TAttribute) {
		gs := p.start()
		p.advance()
		g := &AttributeGroup{}
		for !p.at(TRBracket) && !p.at(TEOF) {
			before := p.pos
			as := p.start()
			a := &Attribute{Name: p.parseName()}
			if p.at(TLParen) {
				a.Args = p.parseArgs()
			}
			g.Attrs = append(g.Attrs, fin(p, a, as))
			if _, ok := p.accept(TComma); !ok || p.pos == before {
				break
			}
		}
		p.expect(TRBracket)
		groups = append(groups, fin(p, g, gs))
	}
	return groups
}

// ---- types -------------------------------------------------------------------------

// canStartType reports whether the current token can begin a type.
func (p *parser) canStartType() bool {
	switch p.kind() {
	case TQuestion, TLParen, TString, TNameQualified, TNameFullyQualified, TNameRelative,
		TArray, TCallable, TStatic:
		return true
	}
	return false
}

// parseType parses a type declaration: ?T, A|B, A&B, (A&B)|C.
func (p *parser) parseType() Expr {
	start := p.start()
	if _, ok := p.accept(TQuestion); ok {
		t := p.tok()
		typ := p.parseTypeAtom()
		if _, grouped := typ.(*IntersectionType); grouped {
			p.errorAt(t, "nullable types require a single type name")
		}
		return fin(p, &NullableType{Type: typ}, start)
	}
	firstToken := p.tok()
	first := p.parseTypeAtom()
	if _, grouped := first.(*IntersectionType); grouped && !p.at(TBar) {
		p.errorAt(firstToken, "parenthesized intersection types require a union")
	}
	switch {
	case p.at(TBar):
		n := &UnionType{Types: []Expr{first}}
		for p.at(TBar) {
			p.advance()
			n.Types = append(n.Types, p.parseTypeAtom())
		}
		return fin(p, n, start)
	case p.at(TAmpersand) && p.isIntersectionAmp():
		n := &IntersectionType{Types: []Expr{first}}
		for p.at(TAmpersand) && p.isIntersectionAmp() {
			p.advance()
			t := p.tok()
			typ := p.parseTypeAtom()
			if _, grouped := typ.(*IntersectionType); grouped {
				p.errorAt(t, "intersection members require type names")
			}
			n.Types = append(n.Types, typ)
		}
		return fin(p, n, start)
	}
	return first
}

// isIntersectionAmp distinguishes `A&B $x` from by-ref `A &$x` / `A &...$x`.
func (p *parser) isIntersectionAmp() bool {
	k := p.peekKind(1)
	return k != TVariable && k != TEllipsis && k != TAmpersand
}

func (p *parser) parseTypeAtom() Expr {
	defer p.leave()
	if !p.enter() {
		return spanOf(&BadExpr{}, p.missing())
	}
	start := p.start()
	if p.at(TLParen) {
		// DNF group (A&B)
		opening := p.tok()
		p.advance()
		n := &IntersectionType{Types: []Expr{p.parseTypeAtom()}}
		for p.at(TAmpersand) {
			p.advance()
			n.Types = append(n.Types, p.parseTypeAtom())
		}
		p.expect(TRParen)
		if len(n.Types) < 2 {
			p.errorAt(opening, "parenthesized types require an intersection")
		}
		for _, typ := range n.Types {
			if _, grouped := typ.(*IntersectionType); grouped {
				p.errorAt(opening, "intersection members require type names")
			}
		}
		return fin(p, n, start)
	}
	return p.parseName()
}

// ---- parameters ------------------------------------------------------------------------

func (p *parser) parseParams() []*Param {
	p.expect(TLParen)
	var out []*Param
	for !p.at(TRParen) && !p.at(TEOF) {
		before := p.pos
		out = append(out, p.parseParam())
		if _, ok := p.accept(TComma); !ok || p.pos == before {
			break
		}
	}
	p.expect(TRParen)
	return out
}

func (p *parser) parseParam() *Param {
	start := p.start()
	n := put(&p.slabs.sParam, Param{})
	if p.at(TAttribute) {
		n.Attrs = p.parseAttributes()
	}
	for p.kind().IsModifier() {
		n.Modifiers = append(n.Modifiers, p.ref(p.advance()))
	}
	if !p.at(TAmpersand) && !p.at(TEllipsis) && !p.at(TVariable) {
		n.Type = p.parseType()
	}
	if _, ok := p.accept(TAmpersand); ok {
		n.ByRef = true
	}
	if _, ok := p.accept(TEllipsis); ok {
		n.Variadic = true
	}
	n.Var = p.parseVariableToken()
	if _, ok := p.accept(TEqual); ok {
		n.Default = p.parseDefault()
	}
	if p.at(TLBrace) {
		n.Hooks = p.parseHooks()
	}
	return fin(p, n, start)
}

func (p *parser) parseReturnType() Expr {
	if _, ok := p.accept(TColon); ok {
		return p.parseType()
	}
	return nil
}

// ---- functions / closures ------------------------------------------------------------

func (p *parser) parseFunction(attrs []*AttributeGroup, start uint32) Stmt {
	p.advance() // function
	n := &Function{Attrs: attrs}
	if _, ok := p.accept(TAmpersand); ok {
		n.ByRef = true
	}
	n.Name = p.parseIdentifier()
	n.Params = p.parseParams()
	n.ReturnType = p.parseReturnType()
	n.Body = p.parseBlock()
	return fin(p, n, start)
}

// parseClosureLike parses `[static] function ...` or `[static] fn ...`.
func (p *parser) parseClosureLike(attrs []*AttributeGroup, start uint32) Expr {
	static := false
	if _, ok := p.accept(TStatic); ok {
		static = true
	}
	switch p.kind() {
	case TFn:
		p.advance()
		n := &ArrowFunction{Attrs: attrs, Static: static}
		if _, ok := p.accept(TAmpersand); ok {
			n.ByRef = true
		}
		n.Params = p.parseParams()
		n.ReturnType = p.parseReturnType()
		p.expect(TDoubleArrow)
		n.Expr = p.parseExpr(precLowest)
		return fin(p, n, start)
	case TFunction:
		p.advance()
		n := &Closure{Attrs: attrs, Static: static}
		if _, ok := p.accept(TAmpersand); ok {
			n.ByRef = true
		}
		n.Params = p.parseParams()
		if _, ok := p.accept(TUse); ok {
			p.expect(TLParen)
			for !p.at(TRParen) && !p.at(TEOF) {
				before := p.pos
				us := p.start()
				u := &ClosureUse{}
				if _, ok := p.accept(TAmpersand); ok {
					u.ByRef = true
				}
				u.Var = p.parseVariableToken()
				n.Uses = append(n.Uses, fin(p, u, us))
				if _, ok := p.accept(TComma); !ok || p.pos == before {
					break
				}
			}
			p.expect(TRParen)
		}
		n.ReturnType = p.parseReturnType()
		n.Body = p.parseBlock()
		return fin(p, n, start)
	}
	// Only reached after attributes (or `static` following them): the
	// error node covers what was consumed.
	t := p.tok()
	p.errorAt(t, "expected function or fn, found "+p.describe(t))
	return fin(p, &BadExpr{}, start)
}

// ---- classes ---------------------------------------------------------------------------

func (p *parser) parseNameList() []*Name {
	var out []*Name
	for {
		out = append(out, p.parseName())
		if _, ok := p.accept(TComma); !ok {
			return out
		}
	}
}

func (p *parser) parseClassLike(attrs []*AttributeGroup, start uint32) Stmt {
	n := &ClassLike{Attrs: attrs}
	for p.at(TAbstract) || p.at(TFinal) || p.at(TReadonly) {
		n.Modifiers = append(n.Modifiers, p.ref(p.advance()))
	}
	switch p.kind() {
	case TClass:
		n.ClassKind = KindClass
	case TInterface:
		n.ClassKind = KindInterface
	case TTrait:
		n.ClassKind = KindTrait
	case TEnum:
		n.ClassKind = KindEnum
	default:
		p.errorAt(p.tok(), "expected class declaration, found "+p.describe(p.tok()))
		p.skipTo(TSemicolon, TRBrace, TLBrace)
		return fin(p, &BadStmt{}, start)
	}
	p.advance()
	n.Name = p.parseIdentifier()
	if n.ClassKind == KindEnum && p.at(TColon) {
		p.advance()
		n.EnumType = p.parseType()
	}
	p.parseClassClauses(n)
	p.parseClassBody(n)
	return fin(p, n, start)
}

func (p *parser) parseAnonymousClass(attrs []*AttributeGroup) *ClassLike {
	start := p.start()
	if len(attrs) > 0 {
		start = attrs[0].Span().Start
	}
	n := &ClassLike{Attrs: attrs, ClassKind: KindClass}
	for p.at(TReadonly) || p.at(TFinal) || p.at(TAbstract) {
		n.Modifiers = append(n.Modifiers, p.ref(p.advance()))
	}
	p.expect(TClass)
	if p.at(TLParen) {
		n.Args = p.parseArgs()
	}
	p.parseClassClauses(n)
	p.parseClassBody(n)
	return fin(p, n, start)
}

// parseClassClauses retains invalid clauses for recovery, including their
// names, so the declaration body and following statements remain available.
func (p *parser) parseClassClauses(n *ClassLike) {
	if clause, ok := p.accept(TExtends); ok {
		if n.ClassKind != KindClass && n.ClassKind != KindInterface {
			p.errorAt(clause, "extends is not allowed in this declaration")
		}
		n.Extends = []*Name{p.parseName()}
		if p.at(TComma) {
			if n.ClassKind == KindClass {
				p.errorAt(p.tok(), "classes may extend only one parent")
			}
			p.advance()
			n.Extends = append(n.Extends, p.parseNameList()...)
		}
	}
	if clause, ok := p.accept(TImplements); ok {
		if n.ClassKind != KindClass && n.ClassKind != KindEnum {
			p.errorAt(clause, "implements is not allowed in this declaration")
		}
		n.Implements = p.parseNameList()
	}
}

func (p *parser) parseClassBody(n *ClassLike) {
	lb := p.expect(TLBrace)
	n.LBrace = Span{lb.Start, lb.End}
	for !p.at(TRBrace) && !p.at(TEOF) {
		// parseMember always consumes a token: `;`, use, attributes, case,
		// modifiers, const/function, a type, or skipTo past a bad token.
		n.Members = append(n.Members, p.parseMember())
	}
	p.expect(TRBrace)
}

func (p *parser) parseMember() Stmt {
	start := p.start()
	switch p.kind() {
	case TUse:
		return p.parseTraitUse()
	case TSemicolon:
		p.advance()
		return fin(p, &Nop{}, start)
	}
	var attrs []*AttributeGroup
	if p.at(TAttribute) {
		attrs = p.parseAttributes()
	}
	if p.at(TCase) {
		p.advance()
		n := &EnumCase{Attrs: attrs, Name: p.parseIdentifier()}
		if _, ok := p.accept(TEqual); ok {
			n.Value = p.parseExpr(precLowest)
		}
		p.endStmt()
		return fin(p, n, start)
	}
	var mods Modifiers
	for p.kind().IsModifier() {
		mods = append(mods, p.ref(p.advance()))
	}
	switch p.kind() {
	case TConst:
		p.advance()
		n := &ClassConst{Attrs: attrs, Modifiers: mods}
		// Typed constant (8.3): `const TYPE NAME =`.
		if p.peekKind(1) != TEqual {
			n.Type = p.parseType()
		}
		for {
			cs := p.start()
			c := &ConstItem{Name: p.parseIdentifier()}
			p.expect(TEqual)
			c.Value = p.parseExpr(precLowest)
			n.Consts = append(n.Consts, fin(p, c, cs))
			if _, ok := p.accept(TComma); !ok {
				break
			}
		}
		p.endStmt()
		return fin(p, n, start)
	case TFunction:
		p.advance()
		n := &Method{Attrs: attrs, Modifiers: mods}
		if _, ok := p.accept(TAmpersand); ok {
			n.ByRef = true
		}
		n.Name = p.parseIdentifier()
		n.Params = p.parseParams()
		n.ReturnType = p.parseReturnType()
		if p.at(TLBrace) {
			n.Body = p.parseBlock()
		} else {
			p.endStmt()
		}
		return fin(p, n, start)
	}
	// Property: [type] $a [= x], $b ... ; or with hooks.
	n := &Property{Attrs: attrs, Modifiers: mods}
	if !p.at(TVariable) {
		if !p.canStartType() {
			p.errorAt(p.tok(), "unexpected "+p.describe(p.tok())+" in class body")
			p.skipTo(TSemicolon, TRBrace)
			p.accept(TSemicolon)
			return fin(p, &BadStmt{}, start)
		}
		n.Type = p.parseType()
	}
	for {
		ps := p.start()
		it := &PropertyItem{Var: p.parseVariableToken()}
		if _, ok := p.accept(TEqual); ok {
			it.Default = p.parseDefault()
		}
		n.Props = append(n.Props, fin(p, it, ps))
		if _, ok := p.accept(TComma); !ok {
			break
		}
	}
	if p.at(TLBrace) {
		n.Hooks = p.parseHooks()
	} else {
		p.endStmt()
	}
	return fin(p, n, start)
}

// parseDefault parses a property or parameter default value, which may be
// followed by a property hook list.
func (p *parser) parseDefault() Expr {
	saved := p.inDefault
	p.inDefault = true
	e := p.parseExpr(precLowest)
	p.inDefault = saved
	return e
}

// parseHooks parses a property hook list `{ get => ...; set { ... } }`.
func (p *parser) parseHooks() []*PropertyHook {
	p.expect(TLBrace)
	var hooks []*PropertyHook
	for !p.at(TRBrace) && !p.at(TEOF) {
		before := p.pos
		hs := p.start()
		h := &PropertyHook{}
		if p.at(TAttribute) {
			h.Attrs = p.parseAttributes()
		}
		for p.kind().IsModifier() {
			h.Modifiers = append(h.Modifiers, p.ref(p.advance()))
		}
		if _, ok := p.accept(TAmpersand); ok {
			h.ByRef = true
		}
		h.Name = p.parseIdentifier()
		if p.at(TLParen) {
			h.Params = p.parseParams()
		}
		switch p.kind() {
		case TDoubleArrow:
			p.advance()
			h.Body = p.parseExpr(precLowest)
			p.endStmt()
		case TLBrace:
			h.Body = p.parseBlock()
		default:
			p.endStmt()
		}
		hooks = append(hooks, fin(p, h, hs))
		if p.pos == before {
			p.advance()
		}
	}
	p.expect(TRBrace)
	return hooks
}

func (p *parser) parseTraitUse() Stmt {
	start := p.start()
	p.advance()
	n := &TraitUse{Traits: p.parseNameList()}
	if !p.at(TLBrace) {
		p.endStmt()
		return fin(p, n, start)
	}
	p.advance()
	for !p.at(TRBrace) && !p.at(TEOF) {
		before := p.pos
		as := p.start()
		a := &TraitAdaptation{}
		if p.peekKind(1) == TPaamayimNekudotayim {
			a.Trait = p.parseName()
			p.advance()
		}
		a.Method = p.parseIdentifier()
		switch p.kind() {
		case TInsteadof:
			p.advance()
			a.Insteadof = p.parseNameList()
		case TAs:
			p.advance()
			if p.kind().IsModifier() {
				r := p.ref(p.advance())
				a.Modifier = &r
			}
			if p.at(TString) || p.kind().IsKeyword() {
				a.Alias = p.parseIdentifier()
			} else if a.Modifier == nil {
				p.errorAt(p.tok(), "expected visibility or alias, found "+p.describe(p.tok()))
			}
		default:
			p.errorAt(p.tok(), "expected insteadof or as, found "+p.describe(p.tok()))
		}
		p.endStmt()
		n.Adaptations = append(n.Adaptations, fin(p, a, as))
		if p.pos == before {
			p.advance()
		}
	}
	p.expect(TRBrace)
	return fin(p, n, start)
}
