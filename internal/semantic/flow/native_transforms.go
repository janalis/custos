package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// everyContext holds the contexts a number or a string of digits is safe in.
const everyContext = HTML | HTMLSingleQuoted | HTMLDoubleQuoted | HTMLScript | HTMLRaw | HTMLURL | HTMLUnquoted | SQL | Shell | Header

// transform is what a PHP function does to the text it is given, as far as
// escaping for output is concerned. A function missing from transforms puts
// quote bytes of its own into its result and keeps no escaping proof that
// depends on them.
type transform struct {
	// safe lists the contexts its result is safe in, whatever it was given.
	safe Context
	// decodes marks a function that can turn escaped markup back into
	// markup: its result loses the escaping its input had.
	decodes bool
	// keepsQuotes marks a function that adds no quote byte of its own, so
	// an input escaped for a quoted attribute stays escaped for it.
	keepsQuotes bool
	// reshapes marks a function whose result is made of the bytes of its
	// first argument only: case changes, trimming, slicing, repetition.
	reshapes bool
}

var transforms = map[string]transform{
	// Their quote safety follows from their flags (see htmlQuoteSafety).
	"htmlspecialchars": {safe: HTML, keepsQuotes: true},
	"htmlentities":     {safe: HTML, keepsQuotes: true},

	"trim":       {keepsQuotes: true, reshapes: true},
	"ltrim":      {keepsQuotes: true, reshapes: true},
	"rtrim":      {keepsQuotes: true, reshapes: true},
	"strtolower": {keepsQuotes: true, reshapes: true},
	"strtoupper": {keepsQuotes: true, reshapes: true},
	"ucfirst":    {keepsQuotes: true, reshapes: true},
	"lcfirst":    {keepsQuotes: true, reshapes: true},
	"ucwords":    {keepsQuotes: true, reshapes: true},
	"str_repeat": {keepsQuotes: true, reshapes: true},
	"strrev":     {keepsQuotes: true, reshapes: true},
	"substr":     {keepsQuotes: true, reshapes: true},
	// What they add between or around the input is the program's own text
	// or another argument, which carries its own proof.
	"nl2br":   {keepsQuotes: true},
	"implode": {keepsQuotes: true},
	"join":    {keepsQuotes: true},
	"str_pad": {keepsQuotes: true},

	// Their output alphabets hold no markup or quote delimiter, nor a colon
	// that could start a URL scheme, nor a line break that could end a
	// header; only some hold no character that ends an unquoted attribute
	// value.
	"urlencode":     {safe: HTML | HTMLSingleQuoted | HTMLDoubleQuoted | HTMLURL | HTMLUnquoted | Header},
	"rawurlencode":  {safe: HTML | HTMLSingleQuoted | HTMLDoubleQuoted | HTMLURL | HTMLUnquoted | Header},
	"base64_encode": {safe: HTML | HTMLSingleQuoted | HTMLDoubleQuoted | HTMLURL | Header},
	"bin2hex":       {safe: HTML | HTMLSingleQuoted | HTMLDoubleQuoted | HTMLURL | HTMLUnquoted | Header},
	"intval":        {safe: everyContext},
	"floatval":      {safe: everyContext},

	"escapeshellarg": {safe: Shell},

	"htmlspecialchars_decode":     {decodes: true},
	"html_entity_decode":          {decodes: true},
	"urldecode":                   {decodes: true},
	"rawurldecode":                {decodes: true},
	"base64_decode":               {decodes: true},
	"hex2bin":                     {decodes: true},
	"str_replace":                 {decodes: true},
	"str_ireplace":                {decodes: true},
	"strtr":                       {decodes: true},
	"preg_replace":                {decodes: true},
	"preg_replace_callback":       {decodes: true},
	"preg_replace_callback_array": {decodes: true},
}

// Reshapes reports whether the result of the PHP function name is made of
// the bytes of its first argument only, so it exposes no markup the
// argument did not hold.
func Reshapes(name string) bool {
	return transforms[name].reshapes
}

// NativeSafety returns the contexts the result of the PHP function name is
// safe in, whatever it is given, before any flag it takes is read.
func NativeSafety(name string) Context {
	return transforms[name].safe
}

// filterSafety is what a filter_var filter leaves in its result: a number, a
// boolean or false for the validating filters, text with markup and quotes
// encoded for the escaping ones.
var filterSafety = map[string]Context{
	"FILTER_VALIDATE_INT":                everyContext,
	"FILTER_VALIDATE_FLOAT":              everyContext,
	"FILTER_VALIDATE_BOOLEAN":            everyContext,
	"FILTER_VALIDATE_BOOL":               everyContext,
	"FILTER_SANITIZE_NUMBER_INT":         everyContext,
	"FILTER_SANITIZE_SPECIAL_CHARS":      HTML | HTMLSingleQuoted | HTMLDoubleQuoted,
	"FILTER_SANITIZE_FULL_SPECIAL_CHARS": HTML | HTMLSingleQuoted | HTMLDoubleQuoted,
}

// filtered returns the contexts the result of a filter_var call is safe in,
// from the PHP constant naming its filter.
func (e *Env) filtered(filter Value) Context {
	constant, ok := syntax.UnwrapParens(filter.Expr).(*syntax.ConstFetch)
	if !filter.Complete || !ok || constant.Name == nil {
		return 0
	}
	fqn, fallback := e.types.Names.Const(constant.Name.Value, constant.Name.Span().Start)
	k := e.types.Index.Constant(fqn, e.types.PHP)
	if k == nil && fallback != "" {
		k = e.types.Index.Constant(fallback, e.types.PHP)
	}
	if k == nil || !k.Builtin {
		return 0
	}
	return filterSafety[strings.TrimPrefix(k.FQN, `\`)]
}

// plainQueryPart reports whether the http_build_query() argument at
// position, named name, is absent or a string literal holding no line
// break, so it cannot bring one into the query.
func plainQueryPart(list *syntax.ArgList, values []Value, name string, position int) bool {
	if absent(list, name, position) {
		return true
	}
	given, ok := passed(list, values, name, position)
	if !ok || !given.Complete {
		return false
	}
	lit, isLit := syntax.UnwrapParens(given.Expr).(*syntax.Literal)
	if !isLit || lit.LitKind != syntax.LitString {
		return false
	}
	text, known := literalText(lit)
	return known && !strings.ContainsAny(text, "\r\n")
}

// absent reports whether a call provably passes nothing for the parameter
// at position, named name: an unpacked list may pass anything.
func absent(list *syntax.ArgList, name string, position int) bool {
	at := 0
	for _, node := range list.Args {
		a, ok := node.(*syntax.Arg)
		if !ok || a.Unpack {
			return false
		}
		if a.Name != nil {
			if a.Name.Value == name {
				return false
			}
			continue
		}
		if at == position {
			return false
		}
		at++
	}
	return true
}
