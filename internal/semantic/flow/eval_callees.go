package flow

import (
	"strings"

	"custos/internal/php/syntax"
)

// callOutcome is what finishCall establishes about a call while the handler
// of its callee kind runs: the value the call yields, the arguments bound to
// the parameters, and the resource states an invocation left.
type callOutcome struct {
	value                           Value
	bound                           []Value
	states                          map[uint32]State
	boundOK, invoked, immutablePure bool
}

// summarizedCall applies the summary of a callee that has one, and reports
// whether it had one.
func (e *Env) summarizedCall(n syntax.Node, values []Value, closedFollowed uint8, out *callOutcome, f *frame, r *scopeResult) bool {
	s, ok := e.contract(e.summaryName(n))
	if !ok {
		return false
	}
	r.argumentSensitive = r.argumentSensitive || s.argumentSensitive
	if out.boundOK && len(s.defaults) > 0 {
		out.bound, out.boundOK = e.defaultArguments(n, values, s.defaults)
	}
	out.value = Value{Complete: s.Return.Complete, Expr: n.(syntax.Expr), Safe: s.Return.Safe}
	out.value = substitute(s.Return, out.bound, out.boundOK)
	out.value.Expr = n.(syntax.Expr)
	out.value.Safe |= s.Return.Safe
	forwarded := false
	precise := e.preciseSummaryTarget(n, closedFollowed)
	if out.boundOK && s.Fixed && precise {
		result, states, ok := e.pureReturn(n, s.pure, out.bound, nil, out.value, f, r)
		out.immutablePure = ok && result.plain && result.deep && pureReadOnlyInputs([]Value{result}, nil)
		if !ok {
			result, states, ok = e.forwardedReturn(n, s.forward, out.bound, f, r)
		}
		if ok {
			out.value = result
			out.value.Expr = n.(syntax.Expr)
			out.states = states
			out.invoked = len(states) > 0
			forwarded = true
		}
	}
	if out.value.elements != nil && (!s.Fixed || !precise) {
		out.value.elements = nil
	}
	// An overridable or bodiless target may not be the body that runs:
	// its return keeps the sources it may yield but proves nothing more.
	if !precise {
		out.value.Complete = false
	}
	if out.boundOK {
		out.value.returned = e.bindReturned(n, s.returned, out.bound)
	}
	out.value.scalars = s.Return.scalars && len(out.value.Sources) > 0
	if !forwarded && out.boundOK && out.value.Complete && s.Quiet && precise {
		out.value, out.states = e.instantiateRecord(n, out.value, out.bound, f, r)
		out.invoked = len(out.states) > 0
	}
	if out.boundOK && out.value.Complete {
		if identity, states, known := e.instantiateResource(n, s.resource, out.bound, f, r); known {
			out.value.Identity = identity
			out.invoked, out.states = true, states
		}
	}
	if out.boundOK && len(s.effects) > 0 && precise {
		effects := e.resourceEffects(n, s.effects, out.bound, f, r)
		if len(effects) > 0 {
			if out.states == nil {
				out.states = effects
			} else {
				for id, state := range effects {
					out.states[id] = state
				}
			}
			out.invoked = true
		}
	}
	if s.Generator {
		out.value = Value{Complete: true, Expr: n.(syntax.Expr), Identity: n.Span().Start + 1}
	}
	return true
}

// unsummarizedCall resolves a project callee without a summary: an
// invocable target, a closed set of returns, or an unresolved dispatch.
func (e *Env) unsummarizedCall(n syntax.Node, values []Value, closedFollowed uint8, out *callOutcome, f *frame, r *scopeResult) {
	if returned, states, known := e.invokeCallable(n, values, f, r); known {
		out.invoked, out.states = true, states
		out.value = returned
		out.value.Expr = n.(syntax.Expr)
	} else if returned, known := e.closedReturn(n, values, closedFollowed); known {
		out.value = returned
		out.value.Expr = n.(syntax.Expr)
	} else {
		r.gap("unresolved-dispatch", n.Span())
		// An unresolved callee may validate or escape its arguments.
		out.value = Value{Expr: out.value.Expr}
	}
}

// builtinCall applies what a PHP function does to its result and to the
// states of the frame.
func (e *Env) builtinCall(n syntax.Node, name string, args *syntax.ArgList, values []Value, first Value, result string, out *callOutcome, f *frame, r *scopeResult) {
	contract := transforms[name]
	if contract.decodes {
		out.value = decodedHTML(out.value)
	}
	if !contract.keepsQuotes {
		out.value.Safe &^= HTMLSingleQuoted | HTMLDoubleQuoted
	}
	if len(out.value.Sources) > 0 && scalarResult(result) {
		out.value.Safe |= everyContext
	}
	switch name {
	case "extract", "parse_str":
		if name == "extract" && e.assigns(n, args, f, r) {
			break
		}
		f.definite = false
		// Each key of the array may replace a local binding of that
		// name, or create one, with an element of the array.
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
	case "array_map":
		// Each element is the callback's result; one array keeps its keys.
		// A list has the keys PHP numbered it with.
		if len(values) > 0 && e.scalarCallback(values[0].Expr) && len(out.value.Sources) > 0 {
			if len(values) == 2 && !e.list(values[1].Expr) {
				out.value.scalars = true
			} else {
				out.value.Safe |= everyContext
			}
		}
	}
	if len(values) > 0 && values[0].scalars && keepsElements(name) {
		// The result is that array with fewer elements: the other
		// arguments only choose among them.
		out.value = Value{Complete: out.value.Complete, Expr: out.value.Expr, Sources: values[0].Sources, Safe: values[0].Safe, scalars: true}
	}
	// Invoking a closure object runs project code like calling it.
	if name == "__invoke" || name == "call" || e.callbackParameter(n, args, values) {
		invalidateCallbackRecords(values, f, r)
		reenter(n, f, r)
		if e.exposed(f, r) {
			e.invalidateShared(n, f, r)
		}
	} else if e.handed(n, name, args, values) {
		e.implicit(n, f, r)
	}
	out.value.Safe |= transforms[name].safe
	switch name {
	case "htmlspecialchars", "htmlentities":
		// An unpacked list may pass flags that leave quotes alone.
		if out.boundOK && len(out.bound) > 1 {
			if flags, given := passed(args, values, "flags", 1); given || absent(args, "flags", 1) {
				out.value.Safe |= e.htmlQuoteSafety(flags, !given, r)
			}
		}
	case "json_encode":
		// A JSON text holds no unescaped slash, so no `</script>`,
		// unless asked for one without escaping `<` and `>` instead.
		flags, given := passed(args, values, "flags", 1)
		bits, known := int64(0), absent(args, "flags", 1)
		if given && flags.Complete {
			bits, known = e.htmlFlagBits(flags.Expr, r, 16)
		}
		if known && (bits&64 == 0 || bits&1 != 0) {
			out.value.Safe |= HTMLScript
		}
		// Strings escape CR and LF as `\r` and `\n`; only
		// JSON_PRETTY_PRINT (128) puts line breaks between members.
		if known && bits&128 == 0 {
			out.value.Safe |= Header
		}
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
		position := 1
		switch name {
		case "hash":
			position = 2
		case "hash_hmac":
			position = 3
		}
		if out.boundOK && position < len(out.bound) {
			flag := out.bound[position]
			if flag.Complete && flag.Expr == nil {
				out.value.Safe |= everyContext
			} else {
				switch flag := syntax.UnwrapParens(flag.Expr).(type) {
				case *syntax.ConstFetch:
					if strings.EqualFold(flag.Name.Value, "false") {
						out.value.Safe |= everyContext
					}
				case *syntax.Literal:
					if flag.LitKind == syntax.LitInt && flag.Raw == "0" {
						out.value.Safe |= everyContext
					}
				}
			}
		}
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

// userCall invalidates what a call to project code may change.
func (e *Env) userCall(n syntax.Node, args *syntax.ArgList, values []Value, rv Value, f *frame, r *scopeResult) {
	e.invalidateEscaped(n, append(values, rv), f, r)
	if !e.quiet(n, args, values) {
		reenter(n, f, r)
	}
	e.invalidateShared(n, f, r)
	for id := range f.states {
		if id >= bufferID {
			delete(f.states, id)
		}
	}
}
