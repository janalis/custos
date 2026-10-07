package syntax

import (
	"bytes"
	"strconv"
	"strings"

	"custos/internal/phpver"
)

type keyword struct {
	kind TokenKind
	min  phpver.Version // lexed as TString below this version
}

// keywords maps lower-case reserved words to their token.
var keywords = map[string]keyword{
	"abstract": {TAbstract, 0}, "and": {TAnd, 0}, "array": {TArray, 0}, "as": {TAs, 0},
	"break": {TBreak, 0}, "callable": {TCallable, phpver.PHP54}, "case": {TCase, 0},
	"catch": {TCatch, 0}, "class": {TClass, 0}, "clone": {TClone, 0}, "const": {TConst, 0},
	"continue": {TContinue, 0}, "declare": {TDeclare, 0}, "default": {TDefault, 0}, "do": {TDo, 0},
	"echo": {TEcho, 0}, "else": {TElse, 0}, "elseif": {TElseif, 0}, "empty": {TEmpty, 0},
	"enddeclare": {TEnddeclare, 0}, "endfor": {TEndfor, 0}, "endforeach": {TEndforeach, 0},
	"endif": {TEndif, 0}, "endswitch": {TEndswitch, 0}, "endwhile": {TEndwhile, 0},
	"enum": {TEnum, phpver.PHP81}, "eval": {TEval, 0}, "exit": {TExit, 0}, "die": {TExit, 0},
	"extends": {TExtends, 0}, "final": {TFinal, 0}, "finally": {TFinally, phpver.PHP55},
	"fn": {TFn, phpver.PHP74}, "for": {TFor, 0}, "foreach": {TForeach, 0}, "function": {TFunction, 0},
	"global": {TGlobal, 0}, "goto": {TGoto, phpver.PHP53}, "if": {TIf, 0},
	"implements": {TImplements, 0}, "include": {TInclude, 0}, "include_once": {TIncludeOnce, 0},
	"instanceof": {TInstanceof, 0}, "insteadof": {TInsteadof, phpver.PHP54},
	"interface": {TInterface, 0}, "isset": {TIsset, 0}, "list": {TList, 0},
	"match": {TMatch, phpver.PHP80}, "namespace": {TNamespace, phpver.PHP53}, "new": {TNew, 0},
	"or": {TOr, 0}, "print": {TPrint, 0}, "private": {TPrivate, 0}, "protected": {TProtected, 0},
	"public": {TPublic, 0}, "readonly": {TReadonly, phpver.PHP81}, "require": {TRequire, 0},
	"require_once": {TRequireOnce, 0}, "return": {TReturn, 0}, "static": {TStatic, 0},
	"switch": {TSwitch, 0}, "throw": {TThrow, 0}, "trait": {TTrait, phpver.PHP54}, "try": {TTry, 0},
	"unset": {TUnset, 0}, "use": {TUse, 0}, "var": {TVar, 0}, "while": {TWhile, 0}, "xor": {TXor, 0},
	"yield": {TYield, phpver.PHP55}, "__halt_compiler": {THaltCompiler, 0},
	"__line__": {TLine, 0}, "__file__": {TFile, 0}, "__dir__": {TDir, phpver.PHP53},
	"__class__": {TClassC, 0}, "__trait__": {TTraitC, phpver.PHP54}, "__method__": {TMethodC, 0},
	"__function__": {TFuncC, 0}, "__namespace__": {TNsC, phpver.PHP53},
	"__property__": {TPropertyC, phpver.PHP84},
}

var castTypes = map[string]TokenKind{
	"int": TIntCast, "integer": TIntCast, "bool": TBoolCast, "boolean": TBoolCast,
	"float": TDoubleCast, "double": TDoubleCast, "real": TDoubleCast, "string": TStringCast,
	"binary": TStringCast, "array": TArrayCast, "object": TObjectCast, "unset": TUnsetCast,
}

var (
	openTagPrefix = []byte("<?")
	commentEnd    = []byte("*/")
)

// maxKeywordLen is the length of the longest keyword (__halt_compiler).
const maxKeywordLen = 15

type lexState uint8

const (
	stInitial lexState = iota
	stScripting
	stDoubleQuotes
	stBackquote
	stHeredoc
	stVarOffset
	stVarname  // after ${ inside a string
	stProperty // after $var-> inside a string
)

type heredocInfo struct {
	label  string
	nowdoc bool
}

// LexError is a lexical problem (the token stream is still complete).
type LexError struct {
	Pos uint32
	Msg string
}

// LexOptions configure the lexer.
type LexOptions struct {
	Version phpver.Version
	// ShortOpenTag enables `<?` as an open tag (php.ini short_open_tag).
	ShortOpenTag bool
}

type lexer struct {
	src      []byte
	pos      int
	ver      phpver.Version
	short    bool
	states   []lexState
	heredocs []heredocInfo
	toks     []Token
	errs     []LexError
	// expectProperty: the next label is a member name (after -> / ?->).
	expectProperty bool
	// halt: number of significant tokens left before __halt_compiler data.
	halt int
}

// Lex tokenizes src. The returned tokens cover src contiguously.
func Lex(src []byte, opt LexOptions) ([]Token, []LexError) {
	if opt.Version == 0 {
		opt.Version = phpver.Default
	}
	l := &lexer{src: src, ver: opt.Version, short: opt.ShortOpenTag, states: []lexState{stInitial}, halt: -1}
	l.toks = make([]Token, 0, len(src)/4+16)
	return l.run()
}

func (l *lexer) run() ([]Token, []LexError) {
	for l.pos < len(l.src) {
		start := l.pos
		k := l.next()
		if l.pos == start { // safety net: never loop forever
			l.pos++
			k = TBadCharacter
		}
		l.emit(k, start)
	}
	return l.toks, l.errs
}

func (l *lexer) emit(k TokenKind, start int) {
	l.toks = append(l.toks, Token{Kind: k, Start: uint32(start), End: uint32(l.pos)})
	if l.halt > 0 && !k.IsTrivia() && k != THaltCompiler {
		// After `__halt_compiler ( ) ;` (or `?>`), the rest is raw data.
		l.halt--
		if l.halt == 0 || k == TSemicolon || k == TCloseTag {
			l.halt = -1
			if l.pos < len(l.src) {
				l.toks = append(l.toks, Token{Kind: THaltCompilerData, Start: uint32(l.pos), End: uint32(len(l.src))})
				l.pos = len(l.src)
			}
		}
	}
}

func (l *lexer) state() lexState { return l.states[len(l.states)-1] }
func (l *lexer) push(s lexState) { l.states = append(l.states, s) }
func (l *lexer) pop() {
	if len(l.states) > 1 {
		l.states = l.states[:len(l.states)-1]
	}
}
func (l *lexer) set(s lexState) { l.states[len(l.states)-1] = s }

func (l *lexer) peek(off int) byte {
	if p := l.pos + off; p < len(l.src) {
		return l.src[p]
	}
	return 0
}

func (l *lexer) hasPrefix(s string) bool {
	return len(l.src)-l.pos >= len(s) && string(l.src[l.pos:l.pos+len(s)]) == s
}

func (l *lexer) hasPrefixFold(s string) bool {
	return len(l.src)-l.pos >= len(s) && asciiEqualFold(l.src[l.pos:l.pos+len(s)], s)
}

// asciiEqualFold compares b to the lower-case ASCII string s, ignoring case.
func asciiEqualFold(b []byte, s string) bool {
	if len(b) != len(s) {
		return false
	}
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != s[i] {
			return false
		}
	}
	return true
}

func (l *lexer) errorf(msg string) { l.errs = append(l.errs, LexError{Pos: uint32(l.pos), Msg: msg}) }

func isLabelStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c >= 0x80
}

func isLabelChar(c byte) bool { return isLabelStart(c) || (c >= '0' && c <= '9') }

func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func (l *lexer) next() TokenKind {
	switch l.state() {
	case stInitial:
		return l.lexInitial()
	case stScripting:
		return l.lexScripting()
	case stDoubleQuotes:
		return l.lexEncapsed('"')
	case stBackquote:
		return l.lexEncapsed('`')
	case stHeredoc:
		return l.lexHeredoc()
	case stVarOffset:
		return l.lexVarOffset()
	case stVarname:
		return l.lexVarname()
	case stProperty:
		return l.lexStringProperty()
	}
	return TBadCharacter // unreachable state; run's guard consumes a byte
}

// ---- inline HTML ---------------------------------------------------------------

func (l *lexer) lexInitial() TokenKind {
	if l.peek(0) == '<' && l.peek(1) == '?' {
		if l.hasPrefixFold("<?php") && (l.pos+5 == len(l.src) || isSpace(l.src[l.pos+5])) {
			l.pos += 5
			l.eatOneNewlineOrSpace()
			l.set(stScripting)
			return TOpenTag
		}
		if l.hasPrefix("<?=") {
			l.pos += 3
			l.set(stScripting)
			return TOpenTagWithEcho
		}
		if l.short {
			l.pos += 2
			l.set(stScripting)
			return TOpenTag
		}
	}
	for l.pos < len(l.src) {
		i := bytes.Index(l.src[l.pos:], openTagPrefix)
		if i < 0 {
			l.pos = len(l.src)
			break
		}
		l.pos += i
		if l.hasPrefixFold("<?php") && (l.pos+5 == len(l.src) || isSpace(l.src[l.pos+5])) ||
			l.hasPrefix("<?=") || l.short {
			break
		}
		l.pos += 2
	}
	return TInlineHTML
}

func (l *lexer) eatOneNewlineOrSpace() {
	switch l.peek(0) {
	case '\r':
		l.pos++
		if l.peek(0) == '\n' {
			l.pos++
		}
	case '\n', ' ', '\t':
		l.pos++
	}
}

// ---- scripting ----------------------------------------------------------------

func (l *lexer) lexScripting() TokenKind {
	c := l.src[l.pos]
	if isSpace(c) {
		for l.pos < len(l.src) && isSpace(l.src[l.pos]) {
			l.pos++
		}
		return TWhitespace
	}
	if l.expectProperty {
		// After -> / ?->, comments keep the member-name context (`#[` is a
		// comment there too, as in PHP).
		switch {
		case c == '#' || (c == '/' && l.peek(1) == '/'):
			return l.lineComment()
		case c == '/' && l.peek(1) == '*':
			return l.blockComment()
		}
		l.expectProperty = false
		if isLabelStart(c) {
			l.scanLabel()
			return TString
		}
	}
	switch c {
	case '#':
		if l.peek(1) == '[' && l.ver.AtLeast(phpver.PHP80) {
			l.pos += 2
			return TAttribute
		}
		return l.lineComment()
	case '/':
		switch l.peek(1) {
		case '/':
			return l.lineComment()
		case '*':
			return l.blockComment()
		case '=':
			l.pos += 2
			return TDivEqual
		}
		l.pos++
		return TDiv
	case '?':
		if l.peek(1) == '>' {
			l.pos += 2
			if l.peek(0) == '\n' {
				l.pos++
			} else if l.peek(0) == '\r' {
				l.pos++
				if l.peek(0) == '\n' {
					l.pos++
				}
			}
			l.set(stInitial)
			return TCloseTag
		}
		if l.peek(1) == '-' && l.peek(2) == '>' {
			l.pos += 3
			l.expectProperty = true
			return TNullsafeObjectOperator
		}
		if l.peek(1) == '?' {
			if l.peek(2) == '=' {
				l.pos += 3
				return TCoalesceEqual
			}
			l.pos += 2
			return TCoalesce
		}
		l.pos++
		return TQuestion
	case '$':
		if isLabelStart(l.peek(1)) {
			l.pos++
			l.scanLabel()
			return TVariable
		}
		l.pos++
		return TDollar
	case '\'':
		return l.singleQuoted()
	case '"':
		return l.doubleQuoteStart()
	case '`':
		l.pos++
		l.push(stBackquote)
		return TBacktick
	case '{':
		l.pos++
		l.push(stScripting)
		return TLBrace
	case '}':
		l.pos++
		if len(l.states) > 1 {
			l.pop()
		}
		return TRBrace
	case '(':
		if k, ok := l.tryCast(); ok {
			return k
		}
		l.pos++
		return TLParen
	case '\\':
		if isLabelStart(l.peek(1)) {
			l.pos++
			l.scanLabel()
			l.scanNameTail()
			return TNameFullyQualified
		}
		l.pos++
		return TNsSeparator
	case '<':
		if l.hasPrefix("<<<") {
			if k, ok := l.heredocStart(); ok {
				return k
			}
		}
	case 'b', 'B':
		switch l.peek(1) {
		case '\'':
			l.pos++
			return l.singleQuoted()
		case '"':
			l.pos++
			return l.doubleQuoteStart()
		case '<':
			if l.peek(2) == '<' && l.peek(3) == '<' {
				save := l.pos
				l.pos++
				if k, ok := l.heredocStart(); ok {
					return k
				}
				l.pos = save
			}
		}
	case '.':
		if isDigit(l.peek(1)) {
			return l.number()
		}
	}
	if isDigit(c) {
		return l.number()
	}
	if isLabelStart(c) {
		return l.labelOrKeyword()
	}
	return l.operator()
}

func (l *lexer) lineComment() TokenKind {
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		if c == '\n' || c == '\r' {
			break
		}
		if c == '?' && l.peek(1) == '>' {
			break
		}
		l.pos++
	}
	return TComment
}

func (l *lexer) blockComment() TokenKind {
	doc := l.peek(2) == '*' && isSpace(l.peek(3))
	end := bytes.Index(l.src[l.pos+2:], commentEnd)
	if end < 0 {
		l.errorf("unterminated comment")
		l.pos = len(l.src)
	} else {
		l.pos += 2 + end + 2
	}
	if doc {
		return TDocComment
	}
	return TComment
}

func (l *lexer) scanLabel() {
	for l.pos < len(l.src) && isLabelChar(l.src[l.pos]) {
		l.pos++
	}
}

// scanNameTail consumes (\label)* and reports whether anything was consumed.
func (l *lexer) scanNameTail() bool {
	any := false
	for l.peek(0) == '\\' && isLabelStart(l.peek(1)) {
		l.pos++
		l.scanLabel()
		any = true
	}
	return any
}

func (l *lexer) labelOrKeyword() TokenKind {
	start := l.pos
	l.scanLabel()
	word := l.src[start:l.pos]
	if l.scanNameTail() {
		if asciiEqualFold(word, "namespace") {
			return TNameRelative
		}
		return TNameQualified
	}
	if len(word) > maxKeywordLen {
		return TString
	}
	var buf [maxKeywordLen]byte
	for i, c := range word {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		buf[i] = c
	}
	kw, ok := keywords[string(buf[:len(word)])] // no allocation: map lookup conversion
	if !ok || l.ver.Below(kw.min) {
		return TString
	}
	switch kw.kind {
	case TYield:
		// yield from (7.0+): one token spanning the whitespace.
		if l.ver.AtLeast(phpver.PHP70) {
			p := l.pos
			for p < len(l.src) && isSpace(l.src[p]) {
				p++
			}
			if p > l.pos && len(l.src)-p >= 4 && asciiEqualFold(l.src[p:p+4], "from") &&
				(p+4 == len(l.src) || !isLabelChar(l.src[p+4])) {
				l.pos = p + 4
				return TYieldFrom
			}
		}
	case TEnum:
		// enum is contextual: keyword only when followed by a name that is
		// not extends/implements.
		p := l.skipTriviaFrom(l.pos)
		if p == l.pos || p >= len(l.src) || !isLabelStart(l.src[p]) {
			return TString
		}
		q := p
		for q < len(l.src) && isLabelChar(l.src[q]) {
			q++
		}
		if w := l.src[p:q]; asciiEqualFold(w, "extends") || asciiEqualFold(w, "implements") {
			return TString
		}
	case TReadonly:
		// PHP 8.1 lexes `readonly (` (whitespace only) as a name so that
		// readonly() calls keep working; 8.2+ always emits the keyword (the
		// grammar accepts it as a function name, and DNF types may follow).
		if l.ver.Below(phpver.PHP82) {
			p := l.pos
			for p < len(l.src) && isSpace(l.src[p]) {
				p++
			}
			if p < len(l.src) && l.src[p] == '(' {
				return TString
			}
		}
	case TPublic, TProtected, TPrivate:
		if l.ver.AtLeast(phpver.PHP84) {
			if end, ok := l.matchSetSuffix(l.pos); ok {
				l.pos = end
				return map[TokenKind]TokenKind{TPublic: TPublicSet, TProtected: TProtectedSet, TPrivate: TPrivateSet}[kw.kind]
			}
		}
	case THaltCompiler:
		l.halt = 3
	}
	return kw.kind
}

// matchSetSuffix matches `(set)` (any case, no inner whitespace: PHP lexes
// `private( set )` as separate tokens, a syntax error) at p.
func (l *lexer) matchSetSuffix(p int) (int, bool) {
	if len(l.src)-p < 5 || !asciiEqualFold(l.src[p:p+5], "(set)") {
		return 0, false
	}
	return p + 5, true
}

// skipTriviaFrom skips whitespace and comments starting at p.
func (l *lexer) skipTriviaFrom(p int) int {
	for p < len(l.src) {
		c := l.src[p]
		switch {
		case isSpace(c):
			p++
		case c == '/' && p+1 < len(l.src) && l.src[p+1] == '*':
			e := bytes.Index(l.src[p+2:], commentEnd)
			if e < 0 {
				return len(l.src)
			}
			p += 2 + e + 2
		case (c == '/' && p+1 < len(l.src) && l.src[p+1] == '/') || (c == '#' && (p+1 >= len(l.src) || l.src[p+1] != '[')):
			for p < len(l.src) && l.src[p] != '\n' {
				p++
			}
		default:
			return p
		}
	}
	return p
}

func (l *lexer) tryCast() (TokenKind, bool) {
	p := l.pos + 1
	for p < len(l.src) && (l.src[p] == ' ' || l.src[p] == '\t') {
		p++
	}
	s := p
	for p < len(l.src) && ((l.src[p] >= 'a' && l.src[p] <= 'z') || (l.src[p] >= 'A' && l.src[p] <= 'Z')) {
		p++
	}
	if p == s {
		return 0, false
	}
	if p-s > 7 {
		return 0, false
	}
	var buf [7]byte
	for i, c := range l.src[s:p] {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		buf[i] = c
	}
	k, ok := castTypes[string(buf[:p-s])]
	if !ok && string(buf[:p-s]) == "void" && l.ver.AtLeast(phpver.PHP85) {
		k, ok = TVoidCast, true
	}
	if !ok {
		return 0, false
	}
	for p < len(l.src) && (l.src[p] == ' ' || l.src[p] == '\t') {
		p++
	}
	if p >= len(l.src) || l.src[p] != ')' {
		return 0, false
	}
	l.pos = p + 1
	return k, true
}

func (l *lexer) number() TokenKind {
	start := l.pos
	c := l.src[l.pos]
	if c == '0' && (l.peek(1) == 'x' || l.peek(1) == 'X') && isHex(l.peek(2)) {
		l.pos += 2
		for l.pos < len(l.src) && (isHex(l.src[l.pos]) || l.src[l.pos] == '_') {
			l.pos++
		}
		return l.intOrFloat(start, 16)
	}
	if c == '0' && (l.peek(1) == 'b' || l.peek(1) == 'B') && (l.peek(2) == '0' || l.peek(2) == '1') {
		l.pos += 2
		for l.pos < len(l.src) && (l.src[l.pos] == '0' || l.src[l.pos] == '1' || l.src[l.pos] == '_') {
			l.pos++
		}
		return l.intOrFloat(start, 2)
	}
	if c == '0' && (l.peek(1) == 'o' || l.peek(1) == 'O') && l.peek(2) >= '0' && l.peek(2) <= '7' {
		l.pos += 2
		for l.pos < len(l.src) && ((l.src[l.pos] >= '0' && l.src[l.pos] <= '7') || l.src[l.pos] == '_') {
			l.pos++
		}
		return l.intOrFloat(start, 8)
	}
	float := false
	l.digits()
	if l.peek(0) == '.' && isDigit(l.peek(1)) {
		float = true
		l.pos++
		l.digits()
	} else if l.peek(0) == '.' && l.pos > start {
		// "1." is a float (also in `1..2`: "1." then ".2", as PHP).
		float = true
		l.pos++
	}
	if (l.peek(0) == 'e' || l.peek(0) == 'E') &&
		(isDigit(l.peek(1)) || ((l.peek(1) == '+' || l.peek(1) == '-') && isDigit(l.peek(2)))) {
		float = true
		l.pos += 2
		l.digits()
	}
	if float {
		return TDNumber
	}
	base := 10
	if c == '0' && l.pos-start > 1 {
		base = 8
	}
	return l.intOrFloat(start, base)
}

func (l *lexer) digits() {
	for l.pos < len(l.src) && (isDigit(l.src[l.pos]) || (l.src[l.pos] == '_' && isDigit(l.peek(1)))) {
		l.pos++
	}
}

func isHex(c byte) bool { return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') }

// intOrFloat classifies an integer literal: values overflowing int64 are floats.
func (l *lexer) intOrFloat(start, base int) TokenKind {
	if l.pos-start <= 17 { // cannot overflow int64 in any base we accept
		return TLNumber
	}
	s := strings.ReplaceAll(string(l.src[start:l.pos]), "_", "")
	switch base {
	case 16, 2:
		s = s[2:]
	case 8:
		if len(s) > 1 && (s[1] == 'o' || s[1] == 'O') {
			s = s[2:]
		}
	}
	if _, err := strconv.ParseInt(s, base, 64); err != nil {
		if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
			return TDNumber
		}
	}
	return TLNumber
}

func (l *lexer) singleQuoted() TokenKind {
	l.pos++ // opening quote
	for l.pos < len(l.src) {
		switch l.src[l.pos] {
		case '\\':
			l.pos += 2
			continue
		case '\'':
			l.pos++
			return TConstantEncapsedString
		}
		l.pos++
	}
	l.pos = len(l.src)
	l.errorf("unterminated string")
	return TConstantEncapsedString
}

// doubleQuoteStart lexes a whole "..." string when it has no interpolation,
// otherwise emits the opening quote and enters the interpolation state.
func (l *lexer) doubleQuoteStart() TokenKind {
	p := l.pos + 1
	for p < len(l.src) {
		c := l.src[p]
		switch {
		case c == '\\':
			p += 2
			continue
		case c == '"':
			l.pos = p + 1
			return TConstantEncapsedString
		case c == '$' && p+1 < len(l.src) && (isLabelStart(l.src[p+1]) || l.src[p+1] == '{'):
			l.pos++
			l.push(stDoubleQuotes)
			return TDoubleQuote
		case c == '{' && p+1 < len(l.src) && l.src[p+1] == '$':
			l.pos++
			l.push(stDoubleQuotes)
			return TDoubleQuote
		}
		p++
	}
	l.pos = len(l.src)
	l.errorf("unterminated string")
	return TConstantEncapsedString
}

// lexEncapsed handles the inside of "..." and `...`.
func (l *lexer) lexEncapsed(term byte) TokenKind {
	if l.src[l.pos] == term {
		l.pos++
		l.pop()
		if term == '"' {
			return TDoubleQuote
		}
		return TBacktick
	}
	if k, ok := l.interpolationStart(); ok {
		return k
	}
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		if c == '\\' {
			l.pos += 2
			continue
		}
		if c == term || l.atInterpolation() {
			break
		}
		l.pos++
	}
	if l.pos > len(l.src) {
		l.pos = len(l.src)
	}
	return TEncapsedAndWhitespace
}

func (l *lexer) atInterpolation() bool {
	c := l.peek(0)
	return (c == '$' && (isLabelStart(l.peek(1)) || l.peek(1) == '{')) || (c == '{' && l.peek(1) == '$')
}

// interpolationStart lexes $var, ${ and {$ inside strings/heredocs.
func (l *lexer) interpolationStart() (TokenKind, bool) {
	c := l.peek(0)
	switch {
	case c == '$' && isLabelStart(l.peek(1)):
		l.pos++
		l.scanLabel()
		switch {
		case l.peek(0) == '[':
			l.push(stVarOffset)
		case l.peek(0) == '-' && l.peek(1) == '>' && isLabelStart(l.peek(2)):
			l.push(stProperty)
		case l.peek(0) == '?' && l.peek(1) == '-' && l.peek(2) == '>' && isLabelStart(l.peek(3)) && l.ver.AtLeast(phpver.PHP80):
			l.push(stProperty)
		}
		return TVariable, true
	case c == '$' && l.peek(1) == '{':
		l.pos += 2
		l.push(stVarname)
		return TDollarOpenCurlyBraces, true
	case c == '{' && l.peek(1) == '$':
		l.pos++
		l.push(stScripting)
		return TCurlyOpen, true
	}
	return 0, false
}

func (l *lexer) lexVarOffset() TokenKind {
	c := l.src[l.pos]
	switch {
	case c == '[':
		l.pos++
		return TLBracket
	case c == ']':
		l.pos++
		l.pop()
		return TRBracket
	case isDigit(c):
		for l.pos < len(l.src) && isLabelChar(l.src[l.pos]) {
			l.pos++
		}
		return TNumString
	case c == '$' && isLabelStart(l.peek(1)):
		l.pos++
		l.scanLabel()
		return TVariable
	case isLabelStart(c):
		l.scanLabel()
		return TString
	case c == '-':
		l.pos++
		return TMinus
	}
	// Anything else ends the offset (PHP reports an error).
	l.pop()
	l.pos++
	l.errorf("unexpected character in string offset")
	return TBadCharacter
}

func (l *lexer) lexVarname() TokenKind {
	if isLabelStart(l.peek(0)) {
		p := l.pos
		for p < len(l.src) && isLabelChar(l.src[p]) {
			p++
		}
		if p < len(l.src) && (l.src[p] == '[' || l.src[p] == '}') {
			l.pos = p
			l.set(stScripting)
			return TStringVarname
		}
	}
	l.set(stScripting)
	return l.lexScripting()
}

func (l *lexer) lexStringProperty() TokenKind {
	if l.hasPrefix("->") {
		l.pos += 2
		return TObjectOperator
	}
	if l.hasPrefix("?->") {
		l.pos += 3
		return TNullsafeObjectOperator
	}
	l.scanLabel()
	l.pop()
	return TString
}

// ---- heredoc / nowdoc ------------------------------------------------------------

func (l *lexer) heredocStart() (TokenKind, bool) {
	p := l.pos + 3
	for p < len(l.src) && (l.src[p] == ' ' || l.src[p] == '\t') {
		p++
	}
	quote := byte(0)
	if p < len(l.src) && (l.src[p] == '\'' || l.src[p] == '"') {
		quote = l.src[p]
		p++
	}
	if p >= len(l.src) || !isLabelStart(l.src[p]) {
		return 0, false
	}
	ls := p
	for p < len(l.src) && isLabelChar(l.src[p]) {
		p++
	}
	label := string(l.src[ls:p])
	if quote != 0 {
		if p >= len(l.src) || l.src[p] != quote {
			return 0, false
		}
		p++
	}
	switch {
	case p < len(l.src) && l.src[p] == '\n':
		p++
	case p < len(l.src) && l.src[p] == '\r':
		p++
		if p < len(l.src) && l.src[p] == '\n' {
			p++
		}
	default:
		return 0, false
	}
	l.pos = p
	l.heredocs = append(l.heredocs, heredocInfo{label: label, nowdoc: quote == '\''})
	l.push(stHeredoc)
	return TStartHeredoc, true
}

// closingLabelAt reports whether a closing heredoc label starts at line
// start p, returning the end offset of the label (indentation included).
func (l *lexer) closingLabelAt(p int, label string) (int, bool) {
	q := p
	if l.ver.AtLeast(phpver.PHP73) {
		for q < len(l.src) && (l.src[q] == ' ' || l.src[q] == '\t') {
			q++
		}
	}
	if len(l.src)-q < len(label) || string(l.src[q:q+len(label)]) != label {
		return 0, false
	}
	e := q + len(label)
	if e < len(l.src) && isLabelChar(l.src[e]) {
		return 0, false
	}
	if l.ver.Below(phpver.PHP73) {
		r := e
		if r < len(l.src) && l.src[r] == ';' {
			r++
		}
		if r < len(l.src) && l.src[r] != '\n' && l.src[r] != '\r' {
			return 0, false
		}
	}
	return e, true
}

func (l *lexer) lexHeredoc() TokenKind {
	h := l.heredocs[len(l.heredocs)-1]
	atLineStart := l.pos == 0 || l.src[l.pos-1] == '\n' || l.src[l.pos-1] == '\r'
	if atLineStart {
		if e, ok := l.closingLabelAt(l.pos, h.label); ok {
			l.pos = e
			l.heredocs = l.heredocs[:len(l.heredocs)-1]
			l.pop()
			return TEndHeredoc
		}
	}
	if !h.nowdoc {
		if k, ok := l.interpolationStart(); ok {
			return k
		}
	}
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		if c == '\\' && !h.nowdoc {
			l.pos += 2
			continue
		}
		if c == '\n' || c == '\r' {
			l.pos++
			if c == '\r' && l.peek(0) == '\n' {
				l.pos++
			}
			if _, ok := l.closingLabelAt(l.pos, h.label); ok {
				break
			}
			continue
		}
		if !h.nowdoc && l.atInterpolation() {
			break
		}
		l.pos++
	}
	if l.pos >= len(l.src) {
		l.pos = len(l.src)
		if _, ok := l.closingLabelAt(l.pos, h.label); !ok {
			l.errorf("unterminated heredoc")
		}
	}
	return TEncapsedAndWhitespace
}

// ---- operators -----------------------------------------------------------------

type opEntry struct {
	s string
	k TokenKind
}

// operators sorted longest first within each leading byte.
var operators = func() map[byte][]opEntry {
	list := []opEntry{
		{"<<=", TSlEqual}, {">>=", TSrEqual}, {"**=", TPowEqual}, {"...", TEllipsis}, {"<=>", TSpaceship},
		{"===", TIsIdentical}, {"!==", TIsNotIdentical},
		{"->", TObjectOperator}, {"=>", TDoubleArrow}, {"::", TPaamayimNekudotayim}, {"++", TInc},
		{"--", TDec}, {"==", TIsEqual}, {"!=", TIsNotEqual}, {"<>", TIsNotEqual}, {"<=", TIsSmallerOrEqual},
		{">=", TIsGreaterOrEqual}, {"+=", TPlusEqual}, {"-=", TMinusEqual}, {"*=", TMulEqual},
		{".=", TConcatEqual}, {"%=", TModEqual}, {"&=", TAndEqual}, {"|=", TOrEqual}, {"^=", TXorEqual},
		{"<<", TSl}, {">>", TSr}, {"||", TBooleanOr}, {"&&", TBooleanAnd}, {"**", TPow}, {"|>", TPipe},
		{";", TSemicolon}, {",", TComma}, {".", TDot}, {"[", TLBracket}, {"]", TRBracket}, {")", TRParen},
		{"+", TPlus}, {"-", TMinus}, {"*", TMul}, {"%", TMod}, {"=", TEqual}, {"<", TLess}, {">", TGreater},
		{"!", TExclaim}, {":", TColon}, {"&", TAmpersand}, {"|", TBar}, {"^", TCaret}, {"~", TTilde},
		{"@", TAt},
	}
	m := map[byte][]opEntry{}
	for _, e := range list {
		m[e.s[0]] = append(m[e.s[0]], e)
	}
	return m
}()

func (l *lexer) operator() TokenKind {
	for _, e := range operators[l.src[l.pos]] {
		if l.hasPrefix(e.s) {
			if e.k == TPipe && l.ver.Below(phpver.PHP85) {
				continue
			}
			l.pos += len(e.s)
			if e.k == TObjectOperator {
				l.expectProperty = true
			}
			return e.k
		}
	}
	l.pos++
	return TBadCharacter
}
