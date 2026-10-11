package hostnamesubstitution

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// hostnameSubstitution reports client-controlled host names
// ($_SERVER['HTTP_HOST'] / ['SERVER_NAME']) used to build e-mail addresses
// or stored in host/domain/email-like variables and properties.
type hostnameSubstitution struct{}

const hostnameStoredMsg = "Client-controlled host name stored here; validate it against a whitelist."

func (hostnameSubstitution) ID() string { return "HostnameSubstitution" }
func (hostnameSubstitution) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KArrayDimFetch}
}

func (r hostnameSubstitution) Check(ctx *analysis.Context, n syntax.Node) {
	s := n.(*syntax.ArrayDimFetch)
	v, ok := s.Var.(*syntax.Variable)
	if !ok || v.NameExpr != nil || v.Name != "_SERVER" || s.Dim == nil {
		return
	}
	attr, ok := astquery.QuotedStringContent(s.Dim)
	if !ok || attr != "SERVER_NAME" && attr != "HTTP_HOST" { // D1
		return
	}
	// custos: `!empty($_SERVER['SERVER_NAME']) ? $_SERVER['SERVER_NAME'] : …`
	// tests the value; only the read flows anywhere (one report, not two).
	switch s.Parent().(type) {
	case *syntax.Isset, *syntax.Empty:
		return
	}
	scalar := false                                 // D2a: the walk crossed a bool/int/float-valued node
	for p := s.Parent(); p != nil; p = p.Parent() { // D2
		if syntax.IsFuncLike(p) {
			return
		}
		switch x := p.(type) {
		case *syntax.ClassLike:
			return
		case *syntax.Isset, *syntax.Empty:
			scalar = true
		case *syntax.Unary:
			switch x.Op.Kind {
			case syntax.TExclaim, syntax.TIntCast, syntax.TDoubleCast, syntax.TBoolCast:
				scalar = true
			}
		case *syntax.FuncCall:
			if nonStringBuiltins[ctx.GlobalFunctionName(x)] {
				scalar = true
			}
		case *syntax.Binary:
			switch x.Op.Kind {
			case syntax.TDot:
				r.checkConcat(ctx, x, s, s, attr)
				return
			case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
				syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual,
				syntax.TSpaceship, syntax.TBooleanAnd, syntax.TBooleanOr:
				scalar = true
			}
		case *syntax.Assign:
			if x.Op.Kind == syntax.TEqual {
				if !scalar {
					r.checkAssign(ctx, x, s, attr)
				}
				return
			}
		}
	}
}

// nonStringBuiltins lists global functions whose result is a bool, an int
// or a float whatever their arguments: a host name passed to one of them
// does not reach the assigned value (D2a).
var nonStringBuiltins = map[string]bool{
	"strlen": true, "mb_strlen": true, "iconv_strlen": true, "count": true, "sizeof": true,
	"in_array": true, "array_key_exists": true, "key_exists": true,
	"strcmp": true, "strcasecmp": true, "strncmp": true, "strncasecmp": true,
	"strnatcmp": true, "strnatcasecmp": true, "strcoll": true, "substr_compare": true,
	"strpos": true, "stripos": true, "strrpos": true, "strripos": true,
	"mb_strpos": true, "mb_stripos": true, "mb_strrpos": true, "mb_strripos": true,
	"str_contains": true, "str_starts_with": true, "str_ends_with": true,
	"substr_count": true, "mb_substr_count": true, "hash_equals": true,
	"preg_match": true, "preg_match_all": true, "fnmatch": true,
	"levenshtein": true, "similar_text": true, "crc32": true, "ord": true, "ip2long": true,
	"intval": true, "boolval": true, "floatval": true, "doubleval": true,
	"is_string": true, "is_numeric": true, "is_null": true, "is_scalar": true,
	"is_int": true, "is_bool": true, "is_array": true,
	"ctype_alnum": true, "ctype_alpha": true, "ctype_digit": true, "ctype_lower": true,
	"ctype_punct": true, "ctype_space": true, "ctype_upper": true, "ctype_xdigit": true,
	"checkdnsrr": true, "dns_check_record": true,
}

// checkConcat implements D3: x (the source or a variable occurrence) lies in
// the concatenation c. The whole concatenation chain around c (looking
// through parentheses) is flattened into its operands; the operand that
// contains x is reported when the operand right before it is a literal
// ending with `@`.
func (hostnameSubstitution) checkConcat(ctx *analysis.Context, c *syntax.Binary, x syntax.Node, s *syntax.ArrayDimFetch, attr string) {
	var root syntax.Node = c
	for p := c.Parent(); p != nil; p = p.Parent() {
		if b, ok := p.(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
			root = b
			continue
		}
		if _, ok := p.(*syntax.Paren); ok {
			continue
		}
		break
	}
	var leaves []syntax.Expr
	var flatten func(e syntax.Expr)
	flatten = func(e syntax.Expr) {
		if b, ok := syntax.UnwrapParens(e).(*syntax.Binary); ok && b.Op.Kind == syntax.TDot {
			flatten(b.Left)
			flatten(b.Right)
			return
		}
		if e != nil {
			leaves = append(leaves, e)
		}
	}
	flatten(root.(syntax.Expr))
	xs := x.Span()
	for i, l := range leaves {
		if l.Span().Contains(xs) {
			if i > 0 && endsWithAt(ctx, syntax.UnwrapParens(leaves[i-1])) && !whitelisted(ctx, s) {
				ctx.ReportNode(syntax.UnwrapParens(l), "E-mail address built from client-controlled $_SERVER['"+attr+"']; validate it against a whitelist.")
			}
			return
		}
	}
}

func endsWithAt(ctx *analysis.Context, e syntax.Expr) bool {
	if content, ok := astquery.QuotedStringContent(e); ok {
		return strings.HasSuffix(content, "@")
	}
	if is, ok := e.(*syntax.InterpolatedString); ok && !is.Heredoc && !is.Backtick {
		sp := is.Span()
		return sp.Len() >= 3 && ctx.Src[sp.End-2] == '@'
	}
	return false
}

// checkAssign implements D4–D6.
func (r hostnameSubstitution) checkAssign(ctx *analysis.Context, a *syntax.Assign, s *syntax.ArrayDimFetch, attr string) {
	switch t := a.Var.(type) {
	case *syntax.PropertyFetch:
		if id, ok := t.Name.(*syntax.Identifier); ok && hostLikeName(id.Value) && !whitelisted(ctx, s) {
			ctx.ReportNode(s, hostnameStoredMsg)
		}
	case *syntax.StaticPropertyFetch:
		if v, ok := t.Name.(*syntax.Variable); ok && v.NameExpr == nil && hostLikeName(v.Name) && !whitelisted(ctx, s) {
			ctx.ReportNode(s, hostnameStoredMsg)
		}
	case *syntax.Variable:
		if t.NameExpr != nil || t.Name == "" {
			return
		}
		fn := syntax.EnclosingFuncLike(s)
		if fn == nil { // D6
			if hostLikeName(t.Name) {
				ctx.ReportNode(s, hostnameStoredMsg)
			}
			return
		}
		if validated(ctx, fn, a, t.Name) { // D7
			return
		}
		passed, killed := false, false // D5
		syntax.Inspect(fn, func(n syntax.Node) bool {
			if killed {
				return false
			}
			v, ok := n.(*syntax.Variable)
			if !ok {
				return true
			}
			if v == t {
				passed = true
				return true
			}
			if !passed || v.NameExpr != nil || v.Name != t.Name {
				return true
			}
			if hostOverwrites(v, a) { // D5a
				killed = true
				return false
			}
			par, _ := astquery.ParentSkipParens(v)
			if c, ok := par.(*syntax.Binary); ok && c.Op.Kind == syntax.TDot {
				r.checkConcat(ctx, c, v, s, attr)
			}
			return true
		})
	}
}

// hostOverwrites implements D5a: v is the target of a plain assignment
// statement that runs on every path after the source assignment src (its
// statement list encloses src) and whose value does not mention v's name.
func hostOverwrites(v *syntax.Variable, src *syntax.Assign) bool {
	a, ok := v.Parent().(*syntax.Assign)
	if !ok || a.Var != v || a.Op.Kind != syntax.TEqual || a.Value == nil {
		return false
	}
	st, ok := a.Parent().(*syntax.ExprStmt)
	if !ok {
		return false
	}
	list := st.Parent()
	enclosing := false
	for p := src.Parent(); p != nil; p = p.Parent() {
		if p == list {
			enclosing = true
			break
		}
		if syntax.IsFuncLike(p) {
			break
		}
	}
	if !enclosing {
		return false
	}
	mentions := false
	syntax.Inspect(a.Value, func(n syntax.Node) bool {
		if w, ok := n.(*syntax.Variable); ok && (w.NameExpr != nil || w.Name == v.Name) {
			mentions = true
		}
		return !mentions
	})
	return !mentions
}

func hostLikeName(name string) bool {
	l := strings.ToLower(name)
	return strings.Contains(l, "domain") || strings.Contains(l, "email") || strings.Contains(l, "host")
}

// whitelisted implements D7 for the source access itself.
func whitelisted(ctx *analysis.Context, s *syntax.ArrayDimFetch) bool {
	fn := syntax.EnclosingFuncLike(s)
	if fn == nil {
		return false
	}
	want := ctx.Text(s)
	return hasInArray(ctx, fn, func(_ *syntax.FuncCall, first syntax.Expr) bool {
		d, ok := first.(*syntax.ArrayDimFetch)
		return ok && ctx.Text(d) == want
	})
}

// validated implements the variable form of D7: fn itself (not a nested
// closure, which has its own variables) tests the variable name with
// in_array() after the source assignment src.
func validated(ctx *analysis.Context, fn syntax.Node, src *syntax.Assign, name string) bool {
	return hasInArray(ctx, fn, func(call *syntax.FuncCall, first syntax.Expr) bool {
		v, ok := first.(*syntax.Variable)
		return ok && v.NameExpr == nil && v.Name == name &&
			call.Span().Start >= src.Span().End && syntax.EnclosingFuncLike(call) == fn
	})
}

// hasInArray reports whether fn contains a call to the global in_array()
// whose first argument satisfies match.
func hasInArray(ctx *analysis.Context, fn syntax.Node, match func(call *syntax.FuncCall, first syntax.Expr) bool) bool {
	found := false
	syntax.Inspect(fn, func(n syntax.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*syntax.FuncCall)
		if !ok || call.Args == nil || len(call.Args.Args) == 0 {
			return true
		}
		if !ctx.IsGlobalFunctionCall(call, "in_array") {
			return true
		}
		found = match(call, syntax.UnwrapParens(astquery.ArgValue(call.Args.Args[0])))
		return !found
	})
	return found
}
