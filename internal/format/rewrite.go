package format

import (
	"strings"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
)

// builtinTypes are the type names written in lowercase.
var builtinTypes = map[string]bool{
	"int": true, "float": true, "string": true, "bool": true, "array": true, "callable": true,
	"iterable": true, "object": true, "mixed": true, "void": true, "never": true, "null": true,
	"false": true, "true": true, "self": true, "parent": true, "static": true,
}

// shortCasts maps a lowercase cast name to its short form.
var shortCasts = map[string]string{
	"integer": "int", "boolean": "bool", "double": "float", "real": "float", "binary": "string",
}

// rewrite sets the text of unit k.
func (e *engine) rewrite(k int, text, rule string) {
	if e.frozenUnit(k) || e.deleted[k] || text == e.text[k] {
		return
	}
	e.text[k], e.textRule[k] = text, rule
}

// remove deletes unit k and the whitespace after it.
func (e *engine) remove(k int, rule string) {
	if e.frozenUnit(k) {
		return
	}
	e.text[k], e.textRule[k], e.deleted[k] = "", rule, true
	if k+1 < len(e.gaps) && !e.gaps[k+1].comment {
		e.gaps[k+1].drop = true
	}
}

// frozenUnit reports whether unit k is in an ignored region.
func (e *engine) frozenUnit(k int) bool {
	return e.inRegions(e.ignored, e.toks[e.sig[k]].Start)
}

// tokens applies the token rewrites that do not depend on the layout.
func (e *engine) tokens() {
	s := e.s
	for k := range e.sig {
		t := e.sig[k]
		kd := e.kind(k)
		txt := e.text[k]
		switch {
		case kd == syntax.TOpenTag:
			if s.KeywordCase == "lower" && strings.HasPrefix(strings.ToLower(txt), "<?php") {
				e.rewrite(k, "<?php"+txt[5:], "keyword_case")
			}
			if s.LineEnding != "keep" && endsLine(txt) {
				// The tag holds the line break that follows it.
				e.rewrite(k, strings.TrimRight(e.text[k], "\r\n")+e.eol, "line_ending")
			}
		case kd.IsCast():
			e.cast(k)
		case kd >= syntax.TLine && kd <= syntax.TPropertyC:
			if s.MagicConstantCase == "upper" {
				e.rewrite(k, strings.ToUpper(txt), "magic_constant_case")
			}
		case kd.IsKeyword() && (!e.ident[t] || e.typeName[t]) && s.KeywordCase == "lower":
			e.rewrite(k, strings.ToLower(txt), "keyword_case")
		case kd == syntax.TString && e.constName[t] && s.ConstantCase != "keep":
			e.constCase(k)
		case kd == syntax.TString && e.typeName[t] && s.TypeCase == "lower" && builtinTypes[strings.ToLower(txt)]:
			e.rewrite(k, strings.ToLower(txt), "type_case")
		case kd == syntax.TIsNotEqual && txt == "<>" && s.StandardizeNotEquals:
			e.rewrite(k, "!=", "standardize_not_equals")
		}
	}
}

func (e *engine) constCase(k int) {
	low := strings.ToLower(e.text[k])
	if low != "true" && low != "false" && low != "null" {
		return
	}
	if e.s.ConstantCase == "upper" {
		e.rewrite(k, strings.ToUpper(low), "constant_case")
		return
	}
	e.rewrite(k, low, "constant_case")
}

// cast normalizes a cast token: no inner spaces, lowercase, short names.
func (e *engine) cast(k int) {
	txt := e.text[k]
	name := strings.TrimSpace(txt[1 : len(txt)-1])
	rule := "keyword_case"
	if e.s.KeywordCase == "lower" {
		name = strings.ToLower(name)
	}
	if short, ok := shortCasts[strings.ToLower(name)]; ok && e.s.ShortCast {
		name, rule = short, "short_cast"
	}
	if e.s.KeywordCase != "lower" && !e.s.ShortCast {
		return
	}
	e.rewrite(k, "("+name+")", rule)
}

// syntaxRewrites applies the rewrites read from the tree.
func (e *engine) syntaxRewrites(n syntax.Node) {
	if isNil(n) || e.inString(n.Span().Start) {
		return
	}
	switch v := n.(type) {
	case *syntax.Array:
		if e.isDestructuring(v) {
			// A destructuring target is a list: array() cannot be assigned to.
			e.arraySyntax(v.Span(), v.Short, e.s.ListSyntax, phpversion.PHP71, "list_syntax")
		} else {
			e.arraySyntax(v.Span(), v.Short, e.s.ArraySyntax, phpversion.PHP54, "array_syntax")
		}
	case *syntax.List:
		e.arraySyntax(v.Span(), false, e.s.ListSyntax, phpversion.PHP71, "list_syntax")
	case *syntax.Else:
		e.elseIf(v)
	case *syntax.New:
		e.newParens(v)
	}
	syntax.Children(n, e.syntaxRewrites)
}

// arraySyntax converts between array(...) and [...].
func (e *engine) arraySyntax(sp syntax.Span, short bool, want string, min phpversion.Version, rule string) {
	k := e.unitAt(sp.Start)
	switch {
	case want == "short" && !short && e.php.AtLeast(min):
		o := k + 1
		if e.kind(o) != syntax.TLParen || e.match[o] < 0 || e.gaps[o].comment {
			return
		}
		e.remove(k, rule)
		e.rewrite(o, "[", rule)
		e.rewrite(e.match[o], "]", rule)
	case want == "long" && short:
		c := e.match[k]
		word := "array("
		if rule == "list_syntax" {
			word = "list("
		}
		e.rewrite(k, word, rule)
		e.rewrite(c, ")", rule)
	}
}

// elseIf writes `else if` as `elseif`.
func (e *engine) elseIf(v *syntax.Else) {
	inner, ok := v.Body.(*syntax.If)
	if !ok || !e.s.ElseIf || inner.Alt {
		return
	}
	k := e.unitAt(v.Span().Start)
	i := e.unitAt(inner.Span().Start)
	if k < 0 || i != k+1 || e.kind(k) != syntax.TElse || e.gaps[i].comment {
		return
	}
	e.rewrite(k, "elseif", "elseif")
	e.remove(i, "elseif")
	e.gaps[i].drop = true
	e.gaps[i+1].drop = false
}

// newParens adds or removes the parentheses of `new` without arguments.
func (e *engine) newParens(v *syntax.New) {
	switch e.s.NewParentheses {
	case "add":
		if v.Args != nil {
			return
		}
		if _, anon := v.Class.(*syntax.ClassLike); anon {
			return
		}
		k := e.unitBefore(v.Span().End)
		e.rewrite(k, e.text[k]+"()", "new_parentheses")
	case "remove":
		if v.Args == nil || len(v.Args.Args) > 0 {
			return
		}
		o := e.unitAt(v.Args.Span().Start)
		if o < 0 || e.match[o] != o+1 || e.gaps[o+1].comment || e.gaps[o].comment {
			return
		}
		c := o + 1
		if c+1 < len(e.sig) {
			switch e.kind(c + 1) {
			case syntax.TObjectOperator, syntax.TNullsafeObjectOperator, syntax.TPaamayimNekudotayim,
				syntax.TLBracket, syntax.TLParen:
				return
			}
		}
		e.remove(o, "new_parentheses")
		e.remove(c, "new_parentheses")
		e.gaps[o].drop = true
		if c+1 < len(e.gaps) {
			e.gaps[c+1].drop = false
		}
	}
}

// listKind returns the trailing_comma kind of the list opened at unit o,
// and whether the PHP version allows a trailing comma there.
func (e *engine) listKind(o int) (string, bool) {
	t := e.sig[o]
	switch e.br[t] {
	case brShortArray, brArray:
		if e.destruct[t] {
			return "destructuring", true
		}
		return "arrays", true
	case brMatch:
		return "match", true
	case brParams:
		return "parameters", e.php.AtLeast(phpversion.PHP80)
	case brUse:
		return "closure_use", e.php.AtLeast(phpversion.PHP80)
	case brCall:
		if o > 0 {
			switch e.kind(o - 1) {
			case syntax.TIsset, syntax.TUnset, syntax.TEmpty, syntax.TEval, syntax.TExit, syntax.TList:
				return "", false
			}
		}
		if e.kind(o+1) == syntax.TEllipsis && e.match[o] == o+2 {
			return "", false
		}
		return "arguments", e.php.AtLeast(phpversion.PHP73)
	}
	return "", false
}

// trailingCommas adds the trailing comma of lists split over several
// lines and removes it from lists on one line.
func (e *engine) trailingCommas() {
	for o := range e.sig {
		if !opener(e.kind(o)) || e.match[o] <= o+1 {
			continue
		}
		kind, ok := e.listKind(o)
		if kind == "" {
			continue
		}
		c := e.match[o]
		last := c - 1
		for last > o && e.deleted[last] {
			last--
		}
		hasComma := e.kind(last) == syntax.TComma
		multi := e.startsLine(c)
		single := !multi && e.oneLine(o, c)
		switch {
		case multi && !hasComma && ok && contains(e.s.TrailingComma, kind) && last > o:
			if e.kind(last) == syntax.TComma || e.frozenUnit(last) || strings.HasSuffix(e.text[last], ",") {
				continue
			}
			e.rewrite(last, e.text[last]+",", "trailing_comma")
		case single && hasComma && !e.s.SingleLineTrailingComma && e.kind(last-1) != syntax.TComma && last-1 > o:
			e.remove(last, "single_line_trailing_comma")
			e.gaps[last+1].drop = false // the space before the closing bracket follows the spacing rules
		}
	}
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// fileEnds handles the tags, the byte order mark and the final newline.
func (e *engine) fileEnds() {
	s := e.s
	if e.kind(0) == syntax.TInlineHTML && e.text[0] == "\xEF\xBB\xBF" && s.RemoveBOM {
		e.remove(0, "remove_bom")
	}
	first := 0
	if e.kind(0) == syntax.TInlineHTML && len(e.sig) > 1 {
		first = 1 // after a byte order mark or a shebang line
	}
	if e.template {
		return
	}
	eof := len(e.sig) - 1
	last := eof - 1
	halt := false
	for k := range e.sig {
		// Everything after __halt_compiler() is data: the end stays as written.
		halt = halt || e.kind(k) == syntax.THaltCompiler
	}
	if s.ClosingTag == "remove" && !halt {
		if last > first && e.kind(last) == syntax.TInlineHTML && strings.TrimSpace(e.text[last]) == "" && e.kind(last-1) == syntax.TCloseTag {
			last--
		}
		if last > first && e.kind(last) == syntax.TCloseTag {
			if p := e.kind(last - 1); p == syntax.TSemicolon || p == syntax.TRBrace && e.match[last-1] >= 0 && e.br[e.sig[e.match[last-1]]].block() {
				for j := last; j < eof; j++ {
					e.remove(j, "closing_tag")
				}
				last--
			} else {
				last = eof - 1
			}
		} else {
			last = eof - 1
		}
	}
	for last > first && e.deleted[last] && !e.gaps[last+1].comment {
		last-- // removed by a rule, and no comment follows it
	}
	if !halt && (s.FinalNewline && last >= first && !isMarkup(e.kind(last)) || s.FinalNewline && last == first && e.kind(last) == syntax.TOpenTag) {
		e.final = last + 1
	}
	if e.kind(first) == syntax.TOpenTag && s.OpenTagOwnLine && first+1 < len(e.sig) && e.kind(first+1) != syntax.TEOF {
		txt := e.text[first]
		if !strings.HasSuffix(txt, "\n") && !strings.HasSuffix(txt, "\r") {
			e.rewrite(first, strings.TrimRight(txt, " \t")+e.eol, "open_tag_own_line")
		}
		if s.BlankLineAfterOpenTag && !e.gaps[first+1].comment && e.leadBreaks(&e.gaps[first+1]) == 0 {
			e.setNL(first+1, 1, "blank_line_after_open_tag")
		}
	}
	if s.BlankLineAfterOpenTag && e.kind(first) == syntax.TOpenTag && first+1 < eof &&
		endsLine(e.text[first]) && e.leadBreaks(&e.gaps[first+1]) == 0 && !e.gaps[first+1].comment {
		e.setNL(first+1, 1, "blank_line_after_open_tag")
	}
}

// oneLine reports whether units a to b are on one line in the output.
func (e *engine) oneLine(a, b int) bool {
	for k := a + 1; k <= b; k++ {
		if e.startsLine(k) || e.gaps[k].comment && strings.ContainsAny(string(e.src[e.gaps[k].start:e.gaps[k].end]), "\n\r") {
			return false
		}
	}
	return true
}

// isDestructuring reports whether a is the target of a destructuring.
func (e *engine) isDestructuring(a *syntax.Array) bool {
	i := e.at(a.Span().Start)
	return a.Short && i >= 0 && e.destruct[i]
}
