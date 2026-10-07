// Package syntax implements a lossless, error-tolerant, version-aware PHP
// lexer and parser (PHP 5.3 – 8.5).
//
// The token stream covers the source contiguously (trivia included), so any
// byte range can be edited precisely and every node maps back to exact bytes.
package syntax

// TokenKind identifies a lexical token.
type TokenKind uint16

const (
	TEOF TokenKind = iota
	TBadCharacter

	// Trivia.
	TWhitespace
	TComment    // // # /* */
	TDocComment // /** */

	// Mode switches.
	TInlineHTML
	TOpenTag         // <?php or <?  (includes one trailing newline/space)
	TOpenTagWithEcho // <?=
	TCloseTag        // ?> (includes one trailing newline)

	// Literals and names.
	TVariable               // $name
	TString                 // identifier / label
	TNameQualified          // Foo\Bar
	TNameFullyQualified     // \Foo\Bar
	TNameRelative           // namespace\Foo
	TLNumber                // integer literal
	TDNumber                // float literal
	TConstantEncapsedString // '...' or "..." without interpolation
	TEncapsedAndWhitespace  // literal part inside an interpolated string
	TStringVarname          // name in "${name}"
	TNumString              // offset in "$a[0]"
	TStartHeredoc           // <<<ID / <<<'ID' (incl. newline)
	TEndHeredoc             // closing identifier
	TDollarOpenCurlyBraces  // ${ inside strings
	TCurlyOpen              // {$ inside strings
	TDoubleQuote            // " opening/closing an interpolated string
	TBacktick               // `
	THaltCompilerData       // everything after __halt_compiler();

	// Casts.
	TIntCast
	TDoubleCast // (float) (double) (real)
	TStringCast // (string) (binary)
	TArrayCast
	TObjectCast
	TBoolCast
	TUnsetCast
	TVoidCast // (void), PHP 8.5

	// Operators and punctuation (multi-char).
	TObjectOperator         // ->
	TNullsafeObjectOperator // ?->
	TDoubleArrow            // =>
	TPaamayimNekudotayim    // ::
	TNsSeparator            // \ (only alone, e.g. in group use)
	TEllipsis               // ...
	TCoalesce               // ??
	TCoalesceEqual          // ??=
	TPow                    // **
	TPowEqual               // **=
	TInc                    // ++
	TDec                    // --
	TIsEqual                // ==
	TIsNotEqual             // != <>
	TIsIdentical            // ===
	TIsNotIdentical         // !==
	TIsSmallerOrEqual       // <=
	TIsGreaterOrEqual       // >=
	TSpaceship              // <=>
	TPlusEqual              // +=
	TMinusEqual             // -=
	TMulEqual               // *=
	TDivEqual               // /=
	TConcatEqual            // .=
	TModEqual               // %=
	TAndEqual               // &=
	TOrEqual                // |=
	TXorEqual               // ^=
	TSlEqual                // <<=
	TSrEqual                // >>=
	TSl                     // <<
	TSr                     // >>
	TBooleanOr              // ||
	TBooleanAnd             // &&
	TPipe                   // |>
	TAttribute              // #[

	// Single-character punctuation.
	TSemicolon // ;
	TComma     // ,
	TDot       // .
	TLBrace    // {
	TRBrace    // }
	TLParen    // (
	TRParen    // )
	TLBracket  // [
	TRBracket  // ]
	TPlus      // +
	TMinus     // -
	TMul       // *
	TDiv       // /
	TMod       // %
	TEqual     // =
	TLess      // <
	TGreater   // >
	TExclaim   // !
	TQuestion  // ?
	TColon     // :
	TAmpersand // &
	TBar       // |
	TCaret     // ^
	TTilde     // ~
	TAt        // @
	TDollar    // $
	TBackslash // never produced; names absorb backslashes (kept for completeness)

	// Keywords (kwFirst..kwLast). Order matters for keyword tables.
	kwFirst
	TAbstract
	TAnd // and
	TArray
	TAs
	TBreak
	TCallable
	TCase
	TCatch
	TClass
	TClone
	TConst
	TContinue
	TDeclare
	TDefault
	TDo
	TEcho
	TElse
	TElseif
	TEmpty
	TEnddeclare
	TEndfor
	TEndforeach
	TEndif
	TEndswitch
	TEndwhile
	TEnum
	TEval
	TExit // exit / die
	TExtends
	TFinal
	TFinally
	TFn
	TFor
	TForeach
	TFunction
	TGlobal
	TGoto
	TIf
	TImplements
	TInclude
	TIncludeOnce
	TInstanceof
	TInsteadof
	TInterface
	TIsset
	TList
	TMatch
	TNamespace
	TNew
	TOr // or
	TPrint
	TPrivate
	TProtected
	TPublic
	TPrivateSet   // private(set)
	TProtectedSet // protected(set)
	TPublicSet    // public(set)
	TReadonly
	TRequire
	TRequireOnce
	TReturn
	TStatic
	TSwitch
	TThrow
	TTrait
	TTry
	TUnset
	TUse
	TVar
	TWhile
	TXor // xor
	TYield
	TYieldFrom
	THaltCompiler

	// Magic constants.
	TLine
	TFile
	TDir
	TClassC
	TTraitC
	TMethodC
	TFuncC
	TNsC
	TPropertyC
	kwLast
)

// Token is one lexical token: kind and byte span [Start, End).
type Token struct {
	Kind  TokenKind
	Start uint32
	End   uint32
}

// IsTrivia reports whether k is whitespace or a comment.
func (k TokenKind) IsTrivia() bool {
	return k == TWhitespace || k == TComment || k == TDocComment
}

// IsKeyword reports whether k is a reserved word (usable as identifier in
// member-name positions).
func (k TokenKind) IsKeyword() bool { return k > kwFirst && k < kwLast }

// IsMagicConst reports whether k is __LINE__ & co.
func (k TokenKind) IsMagicConst() bool { return k >= TLine && k <= TPropertyC }

// IsCast reports whether k is a cast token.
func (k TokenKind) IsCast() bool { return k >= TIntCast && k <= TVoidCast }

// IsAssignOp reports whether k is a compound assignment operator.
func (k TokenKind) IsAssignOp() bool {
	switch k {
	case TPlusEqual, TMinusEqual, TMulEqual, TDivEqual, TConcatEqual, TModEqual,
		TAndEqual, TOrEqual, TXorEqual, TSlEqual, TSrEqual, TPowEqual, TCoalesceEqual:
		return true
	}
	return false
}

// IsModifier reports whether k is a member/class modifier keyword.
func (k TokenKind) IsModifier() bool {
	switch k {
	case TPublic, TProtected, TPrivate, TStatic, TAbstract, TFinal, TReadonly, TVar,
		TPublicSet, TProtectedSet, TPrivateSet:
		return true
	}
	return false
}

var tokenNames = map[TokenKind]string{
	TEOF: "EOF", TBadCharacter: "BAD_CHARACTER", TWhitespace: "WHITESPACE", TComment: "COMMENT",
	TDocComment: "DOC_COMMENT", TInlineHTML: "INLINE_HTML", TOpenTag: "OPEN_TAG",
	TOpenTagWithEcho: "OPEN_TAG_WITH_ECHO", TCloseTag: "CLOSE_TAG", TVariable: "VARIABLE",
	TString: "STRING", TNameQualified: "NAME_QUALIFIED", TNameFullyQualified: "NAME_FULLY_QUALIFIED",
	TNameRelative: "NAME_RELATIVE", TLNumber: "LNUMBER", TDNumber: "DNUMBER",
	TConstantEncapsedString: "CONSTANT_ENCAPSED_STRING", TEncapsedAndWhitespace: "ENCAPSED_AND_WHITESPACE",
	TStringVarname: "STRING_VARNAME", TNumString: "NUM_STRING", TStartHeredoc: "START_HEREDOC",
	TEndHeredoc: "END_HEREDOC", TDollarOpenCurlyBraces: "DOLLAR_OPEN_CURLY_BRACES", TCurlyOpen: "CURLY_OPEN",
	TDoubleQuote: `"`, TBacktick: "`", THaltCompilerData: "HALT_COMPILER_DATA",
	TIntCast: "INT_CAST", TDoubleCast: "DOUBLE_CAST", TStringCast: "STRING_CAST", TArrayCast: "ARRAY_CAST",
	TObjectCast: "OBJECT_CAST", TBoolCast: "BOOL_CAST", TUnsetCast: "UNSET_CAST", TVoidCast: "VOID_CAST",
	TObjectOperator: "->", TNullsafeObjectOperator: "?->", TDoubleArrow: "=>", TPaamayimNekudotayim: "::",
	TNsSeparator: `\`, TEllipsis: "...", TCoalesce: "??", TCoalesceEqual: "??=", TPow: "**", TPowEqual: "**=",
	TInc: "++", TDec: "--", TIsEqual: "==", TIsNotEqual: "!=", TIsIdentical: "===", TIsNotIdentical: "!==",
	TIsSmallerOrEqual: "<=", TIsGreaterOrEqual: ">=", TSpaceship: "<=>", TPlusEqual: "+=", TMinusEqual: "-=",
	TMulEqual: "*=", TDivEqual: "/=", TConcatEqual: ".=", TModEqual: "%=", TAndEqual: "&=", TOrEqual: "|=",
	TXorEqual: "^=", TSlEqual: "<<=", TSrEqual: ">>=", TSl: "<<", TSr: ">>", TBooleanOr: "||",
	TBooleanAnd: "&&", TPipe: "|>", TAttribute: "#[", TSemicolon: ";", TComma: ",", TDot: ".",
	TLBrace: "{", TRBrace: "}", TLParen: "(", TRParen: ")", TLBracket: "[", TRBracket: "]", TPlus: "+",
	TMinus: "-", TMul: "*", TDiv: "/", TMod: "%", TEqual: "=", TLess: "<", TGreater: ">", TExclaim: "!",
	TQuestion: "?", TColon: ":", TAmpersand: "&", TBar: "|", TCaret: "^", TTilde: "~", TAt: "@",
	TDollar: "$", TBackslash: `\`, TPrivateSet: "private(set)", TProtectedSet: "protected(set)",
	TPublicSet: "public(set)", TYieldFrom: "yield from", THaltCompiler: "__halt_compiler",
	TLine: "__LINE__", TFile: "__FILE__", TDir: "__DIR__", TClassC: "__CLASS__", TTraitC: "__TRAIT__",
	TMethodC: "__METHOD__", TFuncC: "__FUNCTION__", TNsC: "__NAMESPACE__", TPropertyC: "__PROPERTY__",
}

func (k TokenKind) String() string {
	if s, ok := tokenNames[k]; ok {
		return s
	}
	for word, kw := range keywords {
		if kw.kind == k {
			return word
		}
	}
	return "TOKEN(?)"
}
