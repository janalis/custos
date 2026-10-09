package syntax

import (
	"fmt"
	"sync"
	"unsafe"

	"custos/internal/phpver"
)

// File is a parsed PHP source file.
type File struct {
	Path    string
	Src     []byte
	Version phpver.Version
	Tokens  []Token // complete token stream (trivia included)
	Stmts   []Stmt
	Errors  []Error

	memoMu sync.Mutex
	memo   map[any]any
}

// Memo returns the value cached on the file under key, computing it with fn
// on first use. Analysis helpers use it for derived per-file data (indexes
// over the tree) that would otherwise be recomputed per node — quadratic
// on large files. fn runs without the lock held (it may use Memo itself);
// concurrent first uses may both compute, the first stored value wins.
func (f *File) Memo(key any, fn func() any) any {
	f.memoMu.Lock()
	v, ok := f.memo[key]
	f.memoMu.Unlock()
	if ok {
		return v
	}
	v = fn()
	f.memoMu.Lock()
	defer f.memoMu.Unlock()
	if old, ok := f.memo[key]; ok {
		return old
	}
	if f.memo == nil {
		f.memo = map[any]any{}
	}
	f.memo[key] = v
	return v
}

// Error is a syntax error (the tree is still complete; see BadExpr/BadStmt).
type Error struct {
	Span Span
	Msg  string
}

func (e Error) Error() string { return fmt.Sprintf("%d: %s", e.Span.Start, e.Msg) }

// Options configure parsing.
type Options struct {
	Version      phpver.Version // default phpver.Default
	ShortOpenTag bool
	// Permissive additionally accepts syntax removed in Version (e.g. legacy
	// `$s{0}` offsets in PHP 8), the way IDEs parse code at any level.
	Permissive bool
}

// Parse parses src (which is retained by the returned File and must not be
// modified afterwards). It never fails: syntax errors are reported in
// File.Errors and the tree contains Bad* nodes where recovery happened.
func Parse(path string, src []byte, opt Options) *File {
	if opt.Version == 0 {
		opt.Version = phpver.Default
	}
	if len(src) > MaxFileSize {
		// Hostile or generated input: analysis memory is linear but with a
		// large constant (peak RSS up to ~250x the file size on garbage
		// input), and offsets are uint32. Report once, analyse nothing.
		f := &File{Path: path, Src: src, Version: opt.Version}
		f.Errors = []Error{{Span: Span{0, 0}, Msg: fmt.Sprintf("file larger than %d MB; not analysed", MaxFileSize>>20)}}
		return f
	}
	toks, lexErrs := Lex(src, LexOptions{Version: opt.Version, ShortOpenTag: opt.ShortOpenTag})
	p := &parser{src: src, toks: toks, ver: opt.Version, permissive: opt.Permissive}
	p.sig = make([]int32, 0, len(toks)/2+1)
	for i, t := range toks {
		if !t.Kind.IsTrivia() {
			p.sig = append(p.sig, int32(i))
		}
	}
	p.eof = Token{Kind: TEOF, Start: uint32(len(src)), End: uint32(len(src))}
	for _, e := range lexErrs {
		p.errs = append(p.errs, Error{Span: Span{e.Pos, e.Pos}, Msg: e.Msg})
	}
	f := &File{Path: path, Src: src, Version: opt.Version, Tokens: toks}
	f.Stmts = p.parseTopStmts()
	if len(p.errs) > MaxErrors {
		// Garbage input yields an error per token: keep the report bounded.
		p.errs = append(p.errs[:MaxErrors:MaxErrors], Error{
			Span: p.errs[MaxErrors].Span,
			Msg:  fmt.Sprintf("more than %d syntax errors; the rest are not reported", MaxErrors),
		})
	}
	if p.tooDeep || TreeDepth(f.Stmts) > MaxDepth {
		// Pathologically nested input (generated or hostile): recursive
		// consumers would exhaust the stack or go quadratic. Keep the
		// tokens, drop the tree, report once.
		f.Stmts = nil
		p.errs = append(p.errs, Error{Span: Span{0, 0}, Msg: fmt.Sprintf("nesting deeper than %d levels; file not analysed", MaxDepth)})
	}
	f.Errors = p.errs
	SetParents(f)
	return f
}

// MaxFileSize is the largest source analysed (bytes); larger files get one
// error and an empty tree.
const MaxFileSize = 10 << 20

// MaxErrors caps the syntax errors reported for one file.
const MaxErrors = 1000

type parser struct {
	src  []byte
	toks []Token
	sig  []int32 // indices of significant tokens
	pos  int     // index into sig
	ver  phpver.Version
	eof  Token
	// permissive accepts removed legacy syntax regardless of ver.
	permissive bool
	// inDefault is set while parsing a property or parameter default: a
	// `{` after it opens the property hooks, never a legacy `$a{0}` offset.
	inDefault bool
	// depth is the current recursion depth of statement/expression parsing;
	// tooDeep is set once it exceeds MaxDepth (parsing then stops).
	depth   int
	tooDeep bool
	slabs   slabs
	errs    []Error
	lastEnd uint32
}

// ---- token access --------------------------------------------------------------

func (p *parser) tok() Token {
	if p.pos < len(p.sig) {
		return p.toks[p.sig[p.pos]]
	}
	return p.eof
}

func (p *parser) kind() TokenKind { return p.tok().Kind }

func (p *parser) peek(n int) Token {
	if i := p.pos + n; i < len(p.sig) {
		return p.toks[p.sig[i]]
	}
	return p.eof
}

func (p *parser) peekKind(n int) TokenKind { return p.peek(n).Kind }

func (p *parser) at(k TokenKind) bool { return p.kind() == k }

func (p *parser) advance() Token {
	t := p.tok()
	if p.pos < len(p.sig) {
		p.pos++
		p.lastEnd = t.End
	}
	return t
}

func (p *parser) accept(k TokenKind) (Token, bool) {
	if p.kind() == k {
		return p.advance(), true
	}
	return Token{}, false
}

func (p *parser) expect(k TokenKind) Token {
	if p.kind() == k {
		return p.advance()
	}
	p.errorAt(p.tok(), fmt.Sprintf("expected %s, found %s", k, p.describe(p.tok())))
	return Token{Kind: k, Start: p.tok().Start, End: p.tok().Start}
}

func (p *parser) describe(t Token) string {
	if len(p.errs) > MaxErrors {
		return "" // the message is dropped anyway (errorAt)
	}
	if t.Kind == TEOF {
		return "end of file"
	}
	s := string(p.src[t.Start:t.End])
	if len(s) > 20 {
		s = s[:20] + "…"
	}
	return fmt.Sprintf("%q", s)
}

func (p *parser) errorAt(t Token, msg string) {
	// One error per position keeps cascades readable.
	if n := len(p.errs); n > MaxErrors || n > 0 && p.errs[n-1].Span.Start == t.Start {
		return // beyond MaxErrors+1 errors nothing more is kept (see Parse)
	}
	p.errs = append(p.errs, Error{Span: Span{t.Start, t.End}, Msg: msg})
}

// text returns the source text of t without copying: node strings alias the
// source buffer, which Parse retains (callers must not mutate src afterwards).
func (p *parser) text(t Token) string {
	if t.End <= t.Start {
		return ""
	}
	return unsafe.String(&p.src[t.Start], int(t.End-t.Start))
}

func (p *parser) ref(t Token) TokenRef { return TokenRef{Kind: t.Kind, Span: Span{t.Start, t.End}} }

// missing is the zero-width span used for nodes that were expected but absent.
func (p *parser) missing() Span { return Span{p.lastEnd, p.lastEnd} }

// start returns the start offset of the current token (node start).
func (p *parser) start() uint32 { return p.tok().Start }

// fin sets a node span from start to the end of the last consumed token.
func fin[T Node](p *parser, n T, start uint32) T {
	b := n.base()
	if p.lastEnd < start {
		// Nothing consumed: zero-width at the end of the previous token.
		b.span = Span{p.lastEnd, p.lastEnd}
	} else {
		b.span = Span{start, p.lastEnd}
	}
	return n
}

// spanOf builds a node span from an explicit range.
func spanOf[T Node](n T, s Span) T {
	n.base().span = s
	return n
}

// ---- statements ------------------------------------------------------------------

func (p *parser) parseTopStmts() []Stmt {
	var out []Stmt
	for !p.at(TEOF) {
		before := p.pos
		if s := p.parseStmt(true); s != nil {
			out = append(out, s)
		}
		if p.pos == before {
			p.errorAt(p.tok(), "unexpected "+p.describe(p.tok()))
			p.advance()
		}
	}
	return out
}

// parseStmtList parses statements until one of the terminators.
func (p *parser) parseStmtList(terms ...TokenKind) []Stmt {
	var out []Stmt
	for !p.at(TEOF) {
		k := p.kind()
		for _, t := range terms {
			if k == t {
				return out
			}
		}
		before := p.pos
		if s := p.parseStmt(false); s != nil {
			out = append(out, s)
		}
		if p.pos == before {
			p.errorAt(p.tok(), "unexpected "+p.describe(p.tok()))
			p.advance()
		}
	}
	return out
}

// endStmt consumes a statement terminator: ';' or '?>'.
func (p *parser) endStmt() {
	switch p.kind() {
	case TSemicolon, TCloseTag:
		p.advance()
	case TEOF:
		// Tolerated at end of file without close tag? PHP requires ';'.
		p.errorAt(p.tok(), "expected ';', found end of file")
	default:
		p.errorAt(p.tok(), "expected ';', found "+p.describe(p.tok()))
	}
}

// MaxDepth bounds syntax nesting (parser recursion and AST depth). Real code
// stays far below it; deeper input is reported and not analysed, so that no
// recursive consumer can overflow the stack.
const MaxDepth = 4000

// enter increments the recursion depth; when it exceeds MaxDepth parsing is
// abandoned (the rest of the input is skipped). Callers must defer p.leave().
func (p *parser) enter() bool {
	p.depth++
	if p.depth > MaxDepth && !p.tooDeep {
		p.tooDeep = true
		p.pos = len(p.sig) // skip to EOF
	}
	return !p.tooDeep
}

func (p *parser) leave() { p.depth-- }

func (p *parser) parseStmt(top bool) Stmt {
	defer p.leave()
	if !p.enter() {
		return nil
	}
	t := p.tok()
	start := t.Start
	switch t.Kind {
	case TOpenTag:
		p.advance()
		return nil
	case TCloseTag:
		p.advance()
		return nil
	case TInlineHTML:
		p.advance()
		return fin(p, &InlineHTML{Raw: p.text(t)}, start)
	case TOpenTagWithEcho:
		p.advance()
		n := &Echo{Short: true}
		if !p.atStmtEnd() {
			n.Exprs = p.parseExprList()
		}
		p.endStmt()
		return fin(p, n, start)
	case TLBrace:
		return p.parseBlock()
	case TSemicolon:
		p.advance()
		return fin(p, &Nop{}, start)
	case TIf:
		return p.parseIf()
	case TWhile:
		return p.parseWhile()
	case TDo:
		return p.parseDoWhile()
	case TFor:
		return p.parseFor()
	case TForeach:
		return p.parseForeach()
	case TSwitch:
		return p.parseSwitch()
	case TBreak, TContinue:
		p.advance()
		var num Expr
		if !p.atStmtEnd() {
			num = p.parseExpr(precLowest)
		}
		p.endStmt()
		if t.Kind == TBreak {
			return fin(p, &Break{Num: num}, start)
		}
		return fin(p, &Continue{Num: num}, start)
	case TReturn:
		p.advance()
		n := put(&p.slabs.sReturn, Return{})
		if !p.atStmtEnd() {
			n.Expr = p.parseExpr(precLowest)
		}
		p.endStmt()
		return fin(p, n, start)
	case TGlobal:
		p.advance()
		n := &Global{}
		for {
			n.Vars = append(n.Vars, p.parseSimpleVariable())
			if _, ok := p.accept(TComma); !ok {
				break
			}
		}
		p.endStmt()
		return fin(p, n, start)
	case TStatic:
		if p.peekKind(1) == TVariable {
			return p.parseStaticStmt()
		}
	case TEcho:
		p.advance()
		n := &Echo{}
		n.Exprs = p.parseExprList()
		p.endStmt()
		return fin(p, n, start)
	case TUnset:
		p.advance()
		n := &Unset{}
		p.expect(TLParen)
		for !p.at(TRParen) && !p.at(TEOF) {
			before := p.pos
			n.Vars = append(n.Vars, p.parseExpr(precLowest))
			if _, ok := p.accept(TComma); !ok || p.pos == before {
				break
			}
		}
		p.expect(TRParen)
		p.endStmt()
		return fin(p, n, start)
	case TTry:
		return p.parseTry()
	case TGoto:
		p.advance()
		n := &Goto{Label: p.parseIdentifier()}
		p.endStmt()
		return fin(p, n, start)
	case TString:
		if p.peekKind(1) == TColon {
			p.advance()
			id := spanOf(put(&p.slabs.sIdentifier, Identifier{Value: p.text(t)}), Span{t.Start, t.End})
			p.advance()
			return fin(p, &Label{Name: id}, start)
		}
	case TNamespace:
		if k := p.peekKind(1); k == TString || k == TNameQualified || k == TLBrace || k == TSemicolon {
			return p.parseNamespace()
		}
	case TUse:
		return p.parseUse()
	case TConst:
		return p.parseConstStmt(nil, start)
	case TFunction:
		if p.isFunctionDecl() {
			return p.parseFunction(nil, start)
		}
	case TAbstract, TFinal, TClass, TInterface, TTrait, TEnum:
		if t.Kind != TClass || p.peekKind(1) == TString || isSemiReserved(p.peekKind(1)) {
			return p.parseClassLike(nil, start)
		}
	case TReadonly:
		if k := p.peekKind(1); k == TClass || k == TFinal || k == TAbstract {
			return p.parseClassLike(nil, start)
		}
	case TAttribute:
		return p.parseAttributedStmt()
	case TDeclare:
		return p.parseDeclare()
	case THaltCompiler:
		p.advance()
		p.expect(TLParen)
		p.expect(TRParen)
		p.endStmt()
		n := &HaltCompiler{}
		if d, ok := p.accept(THaltCompilerData); ok {
			n.Data = p.text(d)
		}
		return fin(p, n, start)
	case TRBrace:
		// Stray closing brace: let the caller report/skip it.
		return nil
	}
	n := put(&p.slabs.sExprStmt, ExprStmt{Expr: p.parseExpr(precLowest)})
	p.endStmt()
	return fin(p, n, start)
}

func (p *parser) atStmtEnd() bool {
	k := p.kind()
	return k == TSemicolon || k == TCloseTag || k == TEOF
}

func (p *parser) parseExprList() []Expr {
	var out []Expr
	for {
		out = append(out, p.parseExpr(precLowest))
		if _, ok := p.accept(TComma); !ok {
			return out
		}
	}
}

func (p *parser) parseBlock() *Block {
	start := p.start()
	p.expect(TLBrace)
	n := put(&p.slabs.sBlock, Block{Stmts: p.parseStmtList(TRBrace)})
	p.expect(TRBrace)
	return fin(p, n, start)
}

// parseBody parses a control-structure body: a statement, or for alternative
// syntax (`:`) the statement list up to one of the end keywords (not consumed).
func (p *parser) parseBody(alt *bool, ends ...TokenKind) Stmt {
	if p.at(TColon) {
		*alt = true
		colon := p.advance()
		n := put(&p.slabs.sBlock, Block{Alt: true})
		n.Stmts = p.parseStmtList(ends...)
		end := p.lastEnd
		if len(n.Stmts) == 0 {
			end = colon.End
		}
		return spanOf(n, Span{colon.Start, end})
	}
	if p.at(TCloseTag) {
		// `if ($a) ?>html` — the close tag acts as ';': the body is empty and
		// the inline HTML is the next statement.
		start := p.advance().Start
		return fin(p, &Nop{}, start)
	}
	s := p.parseStmt(false)
	if s == nil {
		s = fin(p, &Nop{}, p.start())
		p.errorAt(p.tok(), "expected statement")
	}
	return s
}

func (p *parser) parseParenExpr() Expr {
	p.expect(TLParen)
	e := p.parseExpr(precLowest)
	p.expect(TRParen)
	return e
}

func (p *parser) parseIf() Stmt {
	start := p.start()
	p.advance()
	n := &If{}
	n.Cond = p.parseParenExpr()
	n.Body = p.parseBody(&n.Alt, TElseif, TElse, TEndif)
	if n.Alt {
		for p.at(TElseif) {
			s := p.start()
			p.advance()
			ei := &ElseIf{Cond: p.parseParenExpr()}
			var alt bool
			ei.Body = p.parseBody(&alt, TElseif, TElse, TEndif)
			n.ElseIfs = append(n.ElseIfs, fin(p, ei, s))
		}
		if p.at(TElse) {
			s := p.start()
			p.advance()
			var alt bool
			el := &Else{Body: p.parseBody(&alt, TEndif)}
			n.Else = fin(p, el, s)
		}
		p.expect(TEndif)
		p.endStmt()
		return fin(p, n, start)
	}
	for p.at(TElseif) {
		s := p.start()
		p.advance()
		ei := &ElseIf{Cond: p.parseParenExpr()}
		var alt bool
		ei.Body = p.parseBody(&alt)
		n.ElseIfs = append(n.ElseIfs, fin(p, ei, s))
	}
	if p.at(TElse) {
		s := p.start()
		p.advance()
		var alt bool
		el := &Else{Body: p.parseBody(&alt)}
		n.Else = fin(p, el, s)
	}
	return fin(p, n, start)
}

func (p *parser) parseWhile() Stmt {
	start := p.start()
	p.advance()
	n := &While{Cond: p.parseParenExpr()}
	n.Body = p.parseBody(&n.Alt, TEndwhile)
	if n.Alt {
		p.expect(TEndwhile)
		p.endStmt()
	}
	return fin(p, n, start)
}

func (p *parser) parseDoWhile() Stmt {
	start := p.start()
	p.advance()
	n := &DoWhile{}
	var alt bool
	n.Body = p.parseBody(&alt)
	p.expect(TWhile)
	n.Cond = p.parseParenExpr()
	p.endStmt()
	return fin(p, n, start)
}

func (p *parser) parseForExprs(term TokenKind) []Expr {
	var out []Expr
	for !p.at(term) && !p.at(TEOF) {
		before := p.pos
		out = append(out, p.parseExpr(precLowest))
		if _, ok := p.accept(TComma); !ok || p.pos == before {
			break
		}
	}
	return out
}

func (p *parser) parseFor() Stmt {
	start := p.start()
	p.advance()
	n := &For{}
	p.expect(TLParen)
	n.Init = p.parseForExprs(TSemicolon)
	p.expect(TSemicolon)
	n.Cond = p.parseForExprs(TSemicolon)
	p.expect(TSemicolon)
	n.Loop = p.parseForExprs(TRParen)
	p.expect(TRParen)
	n.Body = p.parseBody(&n.Alt, TEndfor)
	if n.Alt {
		p.expect(TEndfor)
		p.endStmt()
	}
	return fin(p, n, start)
}

func (p *parser) parseForeach() Stmt {
	start := p.start()
	p.advance()
	n := &Foreach{}
	p.expect(TLParen)
	n.Expr = p.parseExpr(precLowest)
	p.expect(TAs)
	first, firstRef := p.parseForeachTarget()
	if _, ok := p.accept(TDoubleArrow); ok {
		n.Key = first
		n.Value, n.ByRef = p.parseForeachTarget()
	} else {
		n.Value, n.ByRef = first, firstRef
	}
	p.expect(TRParen)
	n.Body = p.parseBody(&n.Alt, TEndforeach)
	if n.Alt {
		p.expect(TEndforeach)
		p.endStmt()
	}
	return fin(p, n, start)
}

func (p *parser) parseForeachTarget() (Expr, bool) {
	_, byRef := p.accept(TAmpersand)
	t := p.tok()
	target := p.parseUnary()
	valid := isIncrementVariable(target)
	if !byRef {
		switch n := target.(type) {
		case *List:
			valid = true
		case *Array:
			valid = n.Short
		}
	}
	if !valid {
		p.errorAt(t, "foreach targets require a variable or destructuring pattern")
	}
	return target, byRef
}

func (p *parser) parseSwitch() Stmt {
	start := p.start()
	p.advance()
	n := &Switch{Cond: p.parseParenExpr()}
	end := TRBrace
	if _, ok := p.accept(TColon); ok {
		n.Alt = true
		end = TEndswitch
	} else {
		p.expect(TLBrace)
	}
	p.accept(TSemicolon)
	for !p.at(end) && !p.at(TEOF) {
		cs := p.start()
		c := &Case{}
		switch p.kind() {
		case TCase:
			p.advance()
			c.Cond = p.parseExpr(precLowest)
		case TDefault:
			p.advance()
		default:
			p.errorAt(p.tok(), "expected case or default, found "+p.describe(p.tok()))
			p.skipTo(end, TCase, TDefault)
			continue
		}
		if _, ok := p.accept(TColon); !ok {
			if _, ok := p.accept(TSemicolon); !ok {
				p.expect(TColon)
			}
		}
		c.Stmts = p.parseStmtList(TCase, TDefault, end)
		n.Cases = append(n.Cases, fin(p, c, cs))
	}
	p.expect(end)
	if n.Alt {
		p.endStmt()
	}
	return fin(p, n, start)
}

// skipTo advances until one of kinds (not consumed) or EOF.
func (p *parser) skipTo(kinds ...TokenKind) {
	for !p.at(TEOF) {
		for _, k := range kinds {
			if p.at(k) {
				return
			}
		}
		p.advance()
	}
}

func (p *parser) parseStaticStmt() Stmt {
	start := p.start()
	p.advance()
	n := &StaticStmt{}
	for {
		vs := p.start()
		v := &StaticVar{Var: p.parseVariableToken()}
		if _, ok := p.accept(TEqual); ok {
			v.Default = p.parseExpr(precLowest)
		}
		n.Vars = append(n.Vars, fin(p, v, vs))
		if _, ok := p.accept(TComma); !ok {
			break
		}
	}
	p.endStmt()
	return fin(p, n, start)
}

func (p *parser) parseVariableToken() *Variable {
	if !p.at(TVariable) {
		p.expect(TVariable)
		return spanOf(put(&p.slabs.sVariable, Variable{}), p.missing())
	}
	t := p.advance()
	return spanOf(put(&p.slabs.sVariable, Variable{Name: p.text(Token{Start: t.Start + 1, End: t.End})}), Span{t.Start, t.End})
}

func (p *parser) parseTry() Stmt {
	start := p.start()
	p.advance()
	n := &Try{Body: p.parseBlock()}
	for p.at(TCatch) {
		cs := p.start()
		p.advance()
		c := &Catch{}
		p.expect(TLParen)
		for {
			c.Types = append(c.Types, p.parseName())
			if _, ok := p.accept(TBar); !ok {
				break
			}
		}
		if p.at(TVariable) {
			c.Var = p.parseVariableToken()
		}
		p.expect(TRParen)
		c.Body = p.parseBlock()
		n.Catches = append(n.Catches, fin(p, c, cs))
	}
	if p.at(TFinally) {
		fs := p.start()
		p.advance()
		n.Finally = fin(p, &Finally{Body: p.parseBlock()}, fs)
	}
	return fin(p, n, start)
}

func (p *parser) parseNamespace() Stmt {
	start := p.start()
	p.advance()
	n := &Namespace{}
	if p.at(TString) || p.at(TNameQualified) {
		n.Name = p.parseName()
	}
	if p.at(TLBrace) {
		n.Braced = true
		p.advance()
		n.Stmts = p.parseStmtList(TRBrace)
		p.expect(TRBrace)
		return fin(p, n, start)
	}
	p.endStmt()
	// Unbraced: the namespace extends to the next namespace declaration.
	for !p.at(TEOF) {
		if p.at(TNamespace) {
			if k := p.peekKind(1); k == TString || k == TNameQualified || k == TLBrace || k == TSemicolon {
				break
			}
		}
		before := p.pos
		if s := p.parseStmt(true); s != nil {
			n.Stmts = append(n.Stmts, s)
		}
		if p.pos == before {
			p.errorAt(p.tok(), "unexpected "+p.describe(p.tok()))
			p.advance()
		}
	}
	return fin(p, n, start)
}

func (p *parser) useKind() UseKind {
	switch {
	case p.at(TFunction):
		p.advance()
		return UseFunction
	case p.at(TConst):
		p.advance()
		return UseConst
	}
	return UseNormal
}

func (p *parser) parseUse() Stmt {
	start := p.start()
	p.advance()
	n := &Use{Type: p.useKind()}
	for {
		is := p.start()
		name := p.parseName()
		if p.at(TNsSeparator) && p.peekKind(1) == TLBrace {
			// Group use: prefix\{ ... }
			p.advance()
			p.advance()
			n.Prefix = name
			for !p.at(TRBrace) && !p.at(TEOF) {
				before := p.pos
				gs := p.start()
				it := &UseItem{Type: p.useKind()}
				if it.Type == UseNormal {
					it.Type = n.Type
				}
				it.Name = p.parseName()
				if _, ok := p.accept(TAs); ok {
					it.Alias = p.parseIdentifier()
				}
				n.Items = append(n.Items, fin(p, it, gs))
				if _, ok := p.accept(TComma); !ok || p.pos == before {
					break
				}
			}
			p.expect(TRBrace)
			break
		}
		it := &UseItem{Type: n.Type, Name: name}
		if _, ok := p.accept(TAs); ok {
			it.Alias = p.parseIdentifier()
		}
		n.Items = append(n.Items, fin(p, it, is))
		if _, ok := p.accept(TComma); !ok {
			break
		}
	}
	p.endStmt()
	return fin(p, n, start)
}

func (p *parser) parseConstStmt(attrs []*AttributeGroup, start uint32) Stmt {
	p.advance()
	n := &ConstStmt{Attrs: attrs}
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
}

func (p *parser) parseDeclare() Stmt {
	start := p.start()
	p.advance()
	n := &Declare{}
	p.expect(TLParen)
	for !p.at(TRParen) && !p.at(TEOF) {
		before := p.pos
		ds := p.start()
		d := &DeclareItem{Key: p.parseIdentifier()}
		p.expect(TEqual)
		d.Value = p.parseExpr(precLowest)
		n.Items = append(n.Items, fin(p, d, ds))
		if _, ok := p.accept(TComma); !ok || p.pos == before {
			break
		}
	}
	p.expect(TRParen)
	switch {
	case p.at(TSemicolon) || p.at(TCloseTag):
		p.endStmt()
	case p.at(TColon):
		n.Body = p.parseBody(&n.Alt, TEnddeclare)
		p.expect(TEnddeclare)
		p.endStmt()
	default:
		var alt bool
		n.Body = p.parseBody(&alt)
	}
	return fin(p, n, start)
}

// parseAttributedStmt handles `#[...]` before a declaration or closure.
func (p *parser) parseAttributedStmt() Stmt {
	start := p.start()
	attrs := p.parseAttributes()
	switch p.kind() {
	case TFunction:
		if p.isFunctionDecl() {
			return p.parseFunction(attrs, start)
		}
	case TAbstract, TFinal, TClass, TInterface, TTrait, TEnum, TReadonly:
		return p.parseClassLike(attrs, start)
	case TConst:
		return p.parseConstStmt(attrs, start)
	}
	// Closure / arrow function expression statement.
	e := p.parseClosureLike(attrs, start)
	e = p.parsePostfix(e, start)
	e = p.parseBinaryRHS(e, precLowest, start)
	n := put(&p.slabs.sExprStmt, ExprStmt{Expr: e})
	p.endStmt()
	return fin(p, n, start)
}

// isFunctionDecl distinguishes `function name(` from a closure.
func (p *parser) isFunctionDecl() bool {
	k := p.peekKind(1)
	if k == TAmpersand {
		k = p.peekKind(2)
	}
	return k == TString || isSemiReserved(k)
}

// isSemiReserved reports whether k is a keyword usable as an identifier in
// member/declaration name positions.
func isSemiReserved(k TokenKind) bool { return k.IsKeyword() }

// ---- names / identifiers -------------------------------------------------------------

func (p *parser) parseName() *Name {
	t := p.tok()
	var nk NameKind
	switch t.Kind {
	case TString:
		nk = NameUnqualified
	case TNameQualified:
		nk = NameQualified
	case TNameFullyQualified:
		nk = NameFullyQualified
	case TNameRelative:
		nk = NameRelative
	default:
		if t.Kind.IsKeyword() {
			// static/self-like keywords (static, array, callable…) in name positions.
			nk = NameUnqualified
		} else {
			p.errorAt(t, "expected name, found "+p.describe(t))
			return spanOf(put(&p.slabs.sName, Name{}), p.missing())
		}
	}
	p.advance()
	return spanOf(put(&p.slabs.sName, Name{Value: p.text(t), NameKind: nk}), Span{t.Start, t.End})
}

// parseIdentifier parses an identifier; keywords are accepted.
func (p *parser) parseIdentifier() *Identifier {
	t := p.tok()
	if t.Kind == TString || t.Kind.IsKeyword() {
		p.advance()
		return spanOf(put(&p.slabs.sIdentifier, Identifier{Value: p.text(t)}), Span{t.Start, t.End})
	}
	p.errorAt(t, "expected identifier, found "+p.describe(t))
	return spanOf(put(&p.slabs.sIdentifier, Identifier{}), p.missing())
}
