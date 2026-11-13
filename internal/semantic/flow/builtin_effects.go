package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// importBindings applies extract() and parse_str(): each key of the array
// may replace a local binding of that name, or create one, with an element
// of the array.
func (e *Env) importBindings(n syntax.Node, name string, args *syntax.ArgList, values []Value, f *frame, r *scopeResult) {
	if name == "extract" && e.assigns(n, args, f, r) {
		return
	}
	f.definite = false
	if array, given := passed(args, values, "array", 0); given && name == "extract" || len(values) == 1 {
		if !given {
			array = values[0]
		}
		if array.scalars {
			// The variables hold the numbers; the keys name them.
			array.Safe |= everyContext
		}
		// With a skipping or prefixing flag existing variables stay.
		if flags, chosen := passed(args, values, "flags", 1); !chosen || name != "extract" || !keepsExisting(flags.Expr) {
			f.unsettle(r, array)
		}
		f.importing(array)
	}
}

// mappedScalars applies array_map() with a callback returning scalars: each
// element is the callback's result; one array keeps its keys. A list has
// the keys PHP numbered it with.
func (e *Env) mappedScalars(values []Value, out *callOutcome) {
	if len(values) > 0 && e.scalarCallback(values[0].Expr) && len(out.value.Sources) > 0 {
		if len(values) == 2 && !e.list(values[1].Expr) {
			out.value.scalars = true
		} else {
			out.value.Safe |= everyContext
		}
	}
}

// resultSafety adds the contexts the result of a PHP function is safe in
// once its flags and other arguments are read, and gives a fresh resource
// its identity.
func (e *Env) resultSafety(n syntax.Node, name string, args *syntax.ArgList, values []Value, out *callOutcome, r *scopeResult) {
	switch name {
	case "htmlspecialchars", "htmlentities":
		// An unpacked list may pass flags that leave quotes alone.
		if out.boundOK && len(out.bound) > 1 {
			if flags, given := passed(args, values, "flags", 1); given || absent(args, "flags", 1) {
				out.value.Safe |= e.htmlQuoteSafety(flags, !given, r)
			}
		}
	case "json_encode":
		out.value.Safe |= e.jsonSafety(args, values, r)
	case "http_build_query":
		// Every key and value is URL-encoded and joined by `=` and the
		// separator, so no line break unless the numeric prefix or the
		// separator brings one; an `&` may still start an entity.
		if plainQueryPart(args, values, "numeric_prefix", 1) && plainQueryPart(args, values, "arg_separator", 2) {
			out.value.Safe |= Header
		}
	case "number_format":
		// Digits, a sign and the separators, which PHP chooses unless
		// they are given.
		if absent(args, "decimal_separator", 2) && absent(args, "thousands_separator", 3) {
			out.value.Safe |= everyContext
		}
	case "filter_var":
		if filter, given := passed(args, values, "filter", 1); given {
			out.value.Safe |= e.filtered(filter)
		}
	case "quote":
		if m := e.method(n); m != nil && strings.EqualFold(strings.TrimPrefix(m.Class, `\`), "PDO") {
			out.value.Safe |= SQL
		}
	case "md5", "sha1", "hash", "hash_hmac":
		out.value.Safe |= digestSafety(name, out)
	case "fopen", "tmpfile", "popen", "curl_init":
		// A fresh handle: no code ran on it, whatever its arguments
		// went through.
		out.value.Identity, out.value.Invalidated = n.Span().Start+1, 0
		if r.allocations == nil {
			r.allocations = map[uint32]bool{}
		}
		r.allocations[out.value.Identity] = true
	case "prepare", "query":
		out.value = Value{Complete: true, Expr: n.(syntax.Expr), Identity: n.Span().Start + 1}
	}
}

// jsonSafety returns the contexts a json_encode() result is safe in from
// the flags it is given.
func (e *Env) jsonSafety(args *syntax.ArgList, values []Value, r *scopeResult) Context {
	var safe Context
	// A JSON text holds no unescaped slash, so no `</script>`, unless asked
	// for one without escaping `<` and `>` instead.
	flags, given := passed(args, values, "flags", 1)
	bits, known := int64(0), absent(args, "flags", 1)
	if given && flags.Complete {
		bits, known = e.htmlFlagBits(flags.Expr, r, 16)
	}
	if known && (bits&64 == 0 || bits&1 != 0) {
		safe |= HTMLScript
	}
	// Strings escape CR and LF as `\r` and `\n`; only JSON_PRETTY_PRINT
	// (128) puts line breaks between members.
	if known && bits&128 == 0 {
		safe |= Header
	}
	return safe
}

// digestSafety returns every context for a hash function asked for its
// hexadecimal digest: the binary flag is absent or provably false.
func digestSafety(name string, out *callOutcome) Context {
	position := 1
	switch name {
	case "hash":
		position = 2
	case "hash_hmac":
		position = 3
	}
	if !out.boundOK || position >= len(out.bound) {
		return 0
	}
	flag := out.bound[position]
	if flag.Complete && flag.Expr == nil {
		return everyContext
	}
	switch flag := syntax.UnwrapParens(flag.Expr).(type) {
	case *syntax.ConstFetch:
		if strings.EqualFold(flag.Name.Value, "false") {
			return everyContext
		}
	case *syntax.Literal:
		if flag.LitKind == syntax.LitInt && flag.Raw == "0" {
			return everyContext
		}
	}
	return 0
}

// frameEffects records what a PHP function does to the session, the
// response, the output buffers and the headers of the frame.
func frameEffects(n syntax.Node, name string, first Value, f *frame) {
	switch name {
	case "session_start", "session_write_close", "session_commit", "session_abort", "session_destroy":
		f.setState(sessionID, State{Operation: name, Span: n.Span()})
	case "http_response_code":
		f.setState(responseID, State{Operation: name, Span: n.Span()})
	case "flush", "ob_end_flush":
		if _, output := f.states[outputID]; output {
			if _, buffering := f.states[bufferID]; !buffering || name == "ob_end_flush" {
				f.setState(committedID, State{Operation: name, Span: n.Span()})
			}
		}
		if name == "ob_end_flush" {
			delete(f.states, bufferID)
		}
	case "ob_start":
		f.setState(bufferID, State{Operation: name, Span: n.Span()})
	case "ob_end_clean":
		delete(f.states, bufferID)
	case "header":
		if text, known := literalText(first.Expr); known && strings.HasPrefix(strings.ToLower(text), "set-cookie:") {
			f.setState(cookieID, State{Operation: name, Span: n.Span()})
		}
	}
}

// assigns applies `extract` given nothing but an array written with literal
// text keys that name ordinary variables: each is given the element, like an
// assignment, which releases what the variable held. It reports that it did.
func (e *Env) assigns(call syntax.Node, list *syntax.ArgList, f *frame, r *scopeResult) bool {
	if list == nil || len(list.Args) != 1 {
		return false
	}
	arg, ok := list.Args[0].(*syntax.Arg)
	if !ok || arg.Unpack || arg.Name != nil {
		return false
	}
	array, ok := syntax.UnwrapParens(arg.Value).(*syntax.Array)
	if !ok {
		return false
	}
	names := make([]string, len(array.Items))
	for i, item := range array.Items {
		if item.Key == nil || item.ByRef || item.Unpack {
			return false
		}
		// A request table, `$this` and the globals table are not variables
		// extract() assigns.
		key, known := literalText(item.Key)
		if !known || external(key) || key == "this" || key == "GLOBALS" {
			return false
		}
		names[i] = key
	}
	for i, item := range array.Items {
		e.release(call, names[i], f, r)
		f.write(r, names[i], r.values.at(item.Value))
	}
	return true
}

// keepsExisting reports an extract() flag under which a variable that already
// exists is not overwritten.
func keepsExisting(flag syntax.Expr) bool {
	if both, ok := syntax.UnwrapParens(flag).(*syntax.Binary); ok && both.Op.Kind == syntax.TBar {
		// A modifier such as EXTR_REFS is combined with the mode.
		return keepsExisting(both.Left) || keepsExisting(both.Right)
	}
	c, ok := syntax.UnwrapParens(flag).(*syntax.ConstFetch)
	if !ok || c.Name == nil {
		return false
	}
	switch strings.ToUpper(strings.TrimPrefix(c.Name.Value, `\`)) {
	case "EXTR_SKIP", "EXTR_PREFIX_SAME", "EXTR_PREFIX_ALL", "EXTR_PREFIX_IF_EXISTS":
		return true
	}
	return false
}

// scalarCallback reports a callback named by a string literal that is a PHP
// function returning only numbers, booleans or null.
func (e *Env) scalarCallback(callback syntax.Expr) bool {
	name, known := literalText(callback)
	if !known {
		return false
	}
	f := e.types.Index.Function(strings.TrimPrefix(name, `\`), e.types.PHP)
	return f != nil && f.Builtin && scalarResult(f.Return)
}

// importable reports a name an import can create: not a superglobal, the
// globals table or the object in scope.
func importable(name string) bool {
	switch name {
	case "this", "GLOBALS", "_SESSION", "_ENV":
		return false
	}
	return true
}
