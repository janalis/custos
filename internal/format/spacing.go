package format

import (
	"slices"

	"github.com/janalis/custos/internal/php/syntax"
)

// wordish reports whether tokens of kind k need a space to stay apart from
// another word.
func wordish(k syntax.TokenKind) bool {
	switch k {
	case syntax.TVariable, syntax.TString, syntax.TNameQualified, syntax.TNameFullyQualified,
		syntax.TNameRelative, syntax.TLNumber, syntax.TDNumber, syntax.TConstantEncapsedString,
		syntax.TDoubleQuote, syntax.TStartHeredoc, syntax.TBacktick:
		return true
	}
	return k.IsKeyword() || (k >= syntax.TLine && k <= syntax.TPropertyC)
}

// controlKeyword reports whether k introduces a parenthesized condition.
func controlKeyword(k syntax.TokenKind) bool {
	switch k {
	case syntax.TIf, syntax.TElseif, syntax.TWhile, syntax.TFor, syntax.TForeach,
		syntax.TSwitch, syntax.TCatch, syntax.TMatch:
		return true
	}
	return false
}

// binaryRole reports whether r is spaced like a binary operator.
func binaryRole(r role) bool {
	switch r {
	case roleBinary, roleConcat, roleBoolean, roleAssign, roleTernaryQ, roleTernaryColon, roleArrow:
		return true
	}
	return false
}

// spacing returns the spaces wanted between units a and b on one line
// (-1 keeps them) and the setting that decides it.
//
//nolint:gocyclo // one decision per token pair family
func (e *engine) spacing(a, b int) (int, string) {
	s := e.s
	if !s.Spacing {
		return -1, ""
	}
	ta, tb := e.sig[a], e.sig[b]
	ka, kb := e.kind(a), e.kind(b)
	ra, rb := e.role[ta], e.role[tb]
	switch ka {
	case syntax.TOpenTag, syntax.TOpenTagWithEcho, syntax.TCloseTag, syntax.TInlineHTML, syntax.THaltCompilerData:
		return -1, ""
	}
	switch kb {
	case syntax.TOpenTag, syntax.TOpenTagWithEcho, syntax.TCloseTag, syntax.TInlineHTML, syntax.THaltCompilerData, syntax.TEOF:
		return -1, ""
	}
	switch {
	case ra == roleCast:
		return choice(s.CastSpace, "cast_space")
	case ra == roleNot:
		return choice(s.NotOperatorSpace, "not_operator_space")
	case rb == roleDeclareEq || ra == roleDeclareEq:
		return 0, "spacing"
	case ka == syntax.TComma && kb == syntax.TComma:
		return -1, ""
	case kb == syntax.TComma || kb == syntax.TSemicolon:
		return 0, "spacing"
	case ka == syntax.TComma:
		if closer(kb) {
			if o := e.match[b]; o >= 0 && !e.oneLine(o, b) {
				return -1, "" // a list closing on its last item's line
			}
			return 0, "spacing"
		}
		return 1, "spacing"
	case ka == syntax.TSemicolon && ra == roleForSemicolon:
		if kb == syntax.TRParen {
			return 0, "spacing"
		}
		return 1, "spacing"
	case ka == syntax.TAttribute || rb == roleAttrClose:
		return 0, "spacing"
	case ka == syntax.TLBrace && kb == syntax.TRBrace:
		return 0, "spacing"
	case ka == syntax.TLBrace || kb == syntax.TRBrace:
		br := e.br[ta]
		if kb == syntax.TRBrace && e.match[b] >= 0 {
			br = e.br[e.sig[e.match[b]]]
		}
		switch br {
		case brNone, brExprBrace, brGroupUse:
			return 0, "spacing"
		}
		return 1, "spacing"
	case ka == syntax.TLParen || ka == syntax.TLBracket || kb == syntax.TRParen || kb == syntax.TRBracket:
		return 0, "spacing"
	case ka == syntax.TObjectOperator || ka == syntax.TNullsafeObjectOperator || ka == syntax.TPaamayimNekudotayim,
		kb == syntax.TObjectOperator || kb == syntax.TNullsafeObjectOperator || kb == syntax.TPaamayimNekudotayim:
		return 0, "spacing"
	case ra == roleTernaryQ && rb == roleTernaryColon:
		return 0, "spacing"
	case ra == roleConcat || rb == roleConcat:
		return choice(s.ConcatSpace, "concat_space")
	case binaryRole(ra) || binaryRole(rb):
		if s.BinaryOperatorSpace == "at_least_one" {
			return atLeastOne, "binary_operator_space"
		}
		return 1, "binary_operator_space"
	case ra == roleRef || ra == roleSpread || ra == roleUnary || ra == roleNullable:
		return 0, "spacing"
	case ra == roleIncDec || rb == roleIncDec:
		return 0, "spacing"
	case ra == roleTypeOp || rb == roleTypeOp:
		return 0, "spacing"
	case ra == roleCatchBar || rb == roleCatchBar:
		return choice(e.s.CatchTypeSpace, "catch_type_space")
	case rb == roleReturnColon || rb == roleEnumColon || rb == roleNamedColon || rb == roleCaseColon ||
		rb == roleLabelColon || rb == roleAltColon:
		return 0, "spacing"
	case ra == roleReturnColon || ra == roleEnumColon || ra == roleNamedColon:
		return 1, "spacing"
	case kb == syntax.TLParen:
		return e.parenSpacing(a, b)
	case kb == syntax.TLBracket && e.br[tb] == brOffset:
		return 0, "spacing"
	case wordish(ka) && wordish(kb):
		return 1, "spacing"
	case ka == syntax.TRParen && kb == syntax.TLBrace:
		return 1, "spacing"
	case ka == syntax.TRBrace && wordish(kb):
		return 1, "spacing"
	case wordish(ka) && kb == syntax.TLBrace:
		return 1, "spacing"
	case ka == syntax.TRParen && (kb == syntax.TUse || wordish(kb)):
		return 1, "spacing"
	}
	return -1, ""
}

// atLeastOne marks "one space unless there already are more".
const atLeastOne = -2

func choice(v, rule string) (int, string) {
	switch v {
	case "one":
		return 1, rule
	case "none":
		return 0, rule
	}
	return -1, ""
}

// parenSpacing decides the space before an opening parenthesis.
func (e *engine) parenSpacing(a, b int) (int, string) {
	ta, tb := e.sig[a], e.sig[b]
	ka := e.kind(a)
	switch {
	case e.role[ta] == roleClosureKeyword:
		return choice(e.s.ClosureSpace, "closure_space")
	case e.role[ta] == roleArrowKeyword:
		return choice(e.s.ArrowFunctionSpace, "arrow_function_space")
	case e.role[ta] == roleUseKeyword:
		return 1, "spacing"
	case ka == syntax.TClass && !e.ident[ta]:
		return choice(e.s.AnonymousClassSpace, "anonymous_class_space") // only new class(...) puts ( after class
	case e.br[tb] == brControl && (controlKeyword(ka) || e.text[a] == "elseif"):
		return 1, "spacing"
	case e.br[tb] == brCall || e.br[tb] == brParams || e.br[tb] == brArray || e.br[tb] == brDeclare:
		return 0, "spacing"
	case ka == syntax.TClone || ka == syntax.TExit:
		return 0, "spacing" // written like a call
	case e.br[tb] == brGroup && ka.IsKeyword():
		return 1, "spacing"
	}
	return -1, ""
}

// spaceOK reports whether a and b, written with no space between them,
// still read as the same two tokens.
func (e *engine) spaceOK(a, b int) bool {
	if e.joinChecked == nil {
		e.joinChecked = map[[2]string]bool{}
	}
	key := [2]string{e.text[a], e.text[b]}
	if ok, seen := e.joinChecked[key]; seen {
		return ok
	}
	kinds := func(text string) []syntax.TokenKind {
		toks, _ := syntax.Lex([]byte("<?php "+text), syntax.LexOptions{Version: e.lex})
		var out []syntax.TokenKind
		for _, t := range toks {
			if !t.Kind.IsTrivia() && t.Kind != syntax.TEOF && t.Kind != syntax.TOpenTag {
				out = append(out, t.Kind)
			}
		}
		return out
	}
	ok := slices.Equal(kinds(key[0]+key[1]), append(kinds(key[0]), kinds(key[1])...))
	e.joinChecked[key] = ok
	return ok
}
