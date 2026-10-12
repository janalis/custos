package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func (e *Env) transfer(n syntax.Node, f *frame, r *scopeResult) {
	remember(r, n, *f)
	switch n := n.(type) {
	case syntax.Expr:
		e.eval(n, f, r, 0)
	case *syntax.Return:
		r.returns = append(r.returns, e.eval(n.Expr, f, r, 0))
	case *syntax.ExprStmt:
		e.eval(n.Expr, f, r, 0)
	case *syntax.Foreach:
		iterated := e.eval(n.Expr, f, r, 0)
		if n.ByRef || referenceTarget(n.Value, f) {
			// The loop variable stays bound to an element after the loop.
			f.alias(rootVariable(n.Value), aliasAny)
			f.alias(rootVariable(n.Expr), aliasAny)
		}
		kill(n.Key, f, r)
		kill(n.Value, f, r)
		// The keys and elements of request data are request data.
		if tainted(iterated) {
			spread(n.Key, Value{Complete: true, Sources: iterated.Sources, Safe: iterated.Safe}, f, r)
			item := iterated
			if item.scalars {
				item.Safe |= HTML | SQL | Shell | Header
			}
			spread(n.Value, item, f, r)
			if name := plainVariable(n.Value); name != "" && iterated.upload == uploadTable {
				// Each element of the table is one uploaded file.
				entry := f.vars[name]
				entry.upload = uploadEntry
				f.vars[name] = entry
			}
		}
	case *syntax.Catch:
		kill(n.Var, f, r)
	case *syntax.Echo:
		for _, x := range n.Exprs {
			e.eval(x, f, r, 0)
		}
		f.states[outputID] = State{Operation: "echo", Span: n.Span()}
	case *syntax.Global:
		for _, x := range n.Vars {
			// The name leaves what it was bound to for the global variable.
			rebind(x, f, r)
			if name := rootVariable(x); name != "" {
				f.alias(name, aliasHeld)
				if r.globalNames == nil {
					r.globalNames = map[string]bool{}
				}
				r.globalNames[name] = true
			}
		}
	case *syntax.Unset:
		for _, x := range n.Vars {
			rebind(x, f, r)
		}
	case *syntax.StaticStmt:
		for _, x := range n.Vars {
			rebind(x.Var, f, r)
		}
	default:
		syntax.Children(n, func(child syntax.Node) {
			if x, ok := child.(syntax.Expr); ok {
				e.eval(x, f, r, 0)
			}
		})
	}
}

func remember(r *scopeResult, n syntax.Node, f frame) {
	if old, ok := r.before[n]; ok {
		for id, s := range old {
			if common, ok := agree(s, f.states[id]); ok {
				old[id] = common
			} else {
				delete(old, id)
			}
		}
	} else {
		if !reserve(r, len(f.states)) {
			return
		}
		m := map[uint32]State{}
		for id, s := range f.states {
			m[id] = s
		}
		r.before[n] = m
	}
}

func (e *Env) eval(x syntax.Expr, f *frame, r *scopeResult, depth int) Value {
	if x == nil {
		return Value{Complete: true}
	}
	if !r.complete {
		return Value{}
	}
	r.operations++
	if r.operations > MaxTransfers || r.copies > MaxCopied || depth >= MaxDepth*8 {
		r.complete = false
		return Value{}
	}
	remember(r, x, *f)
	v := Value{Complete: true, Expr: x}
	switch n := x.(type) {
	case *syntax.Variable:
		if n.NameExpr != nil {
			v = Value{}
		} else if strings.HasPrefix(n.Name, "_") && external(n.Name) {
			v.Sources = []Source{{Kind: n.Name, Path: e.file.Path, Span: n.Span(), Parameter: -1}}
			if n.Name == "_FILES" {
				v.upload = uploadTable
			}
		} else {
			var tracked bool
			if v, tracked = f.vars[n.Name]; !tracked && importable(n.Name) {
				// An import of request data may have created the variable.
				if spread := f.vars[imported]; tainted(spread) {
					v = Value{Sources: spread.Sources, Safe: spread.Safe}
				}
			}
			v.elements = nil
		}
	case *syntax.Paren:
		v = e.eval(n.Expr, f, r, depth+1)
	case *syntax.Assign:
		if n.Op.Kind == syntax.TCoalesceEqual {
			// The right side runs only when the target is null.
			assigned := clone(*f)
			v = e.assign(n, &assigned, r, depth)
			*f = rejoin(*f, assigned, r)
			break
		}
		v = e.assign(n, f, r, depth)
	case *syntax.FuncCall:
		if _, named := n.Name.(*syntax.Name); !named {
			// A callee expression runs first: an invoked closure literal
			// binds its reference captures before the call.
			e.eval(n.Name, f, r, depth+1)
		}
		v = e.call(n, n.Args, nil, f, r, depth)
	case *syntax.MethodCall:
		if conditional(n) {
			// The arguments are not evaluated when the receiver is null.
			called := clone(*f)
			v = e.call(n, n.Args, n.Var, &called, r, depth)
			*f = rejoin(*f, called, r)
			break
		}
		v = e.call(n, n.Args, n.Var, f, r, depth)
	case *syntax.StaticCall:
		if conditional(n.Class) {
			// The call is not made when a null-safe link before it is null.
			called := clone(*f)
			v = e.call(n, n.Args, nil, &called, r, depth)
			*f = rejoin(*f, called, r)
			break
		}
		v = e.call(n, n.Args, nil, f, r, depth)
	case *syntax.New:
		values := []Value{}
		if n.Args != nil {
			for _, a := range n.Args.Args {
				values = append(values, e.eval(a, f, r, depth+1))
			}
		}
		e.invalidateEscaped(n, values, f, r)
		e.invalidateReferences(n, n.Args, f, r)
		if e.exposed(f, r) && !e.nativeClass(n) {
			e.invalidateShared(f, r)
		}
		for id := range f.states {
			if id >= bufferID {
				delete(f.states, id)
			}
		}
		v.Identity = n.Span().Start + 1
	case *syntax.Arg:
		v = e.eval(n.Value, f, r, depth+1)
		if n.ByRef || n.Unpack {
			// The values such an argument binds are not followed.
			v = Value{Expr: v.Expr}
		}
	case *syntax.Binary:
		left := e.eval(n.Left, f, r, depth+1)
		var right Value
		switch n.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
			// The right operand runs only on the left operand's deciding outcome.
			// The path that skips it saw the left operand decide the other way.
			deciding := n.Op.Kind == syntax.TBooleanAnd || n.Op.Kind == syntax.TAnd
			skipped := clone(*f)
			e.refine(n.Left, &skipped, !deciding)
			for name, refined := range skipped.vars {
				// Only what the outcome excludes is kept: a value pinned
				// or cleared by the test is another value to later reads.
				if held := f.vars[name]; refined.Expr != held.Expr || len(refined.Sources) != len(held.Sources) {
					skipped.vars[name] = held
				}
			}
			e.refine(n.Left, f, deciding)
			right = e.eval(n.Right, f, r, depth+1)
			if stored := assignedInPlace(n.Right); stored != nil {
				// A value the same assignment produced earlier may arrive on
				// the path that skips it: its test then says nothing of it.
				reached := skipped.vars[stored.Var.(*syntax.Variable).Name]
				produced := Value{Expr: syntax.UnwrapParens(stored.Value)}
				if reached.Expr == produced.Expr || reached.Expr == nil && reached.origins != nil && (reached.origins.full || reached.origins.holds(&Value{Expr: produced.Expr, NonFalse: true, NonNull: true})) {
					if r.stale == nil {
						r.stale = map[*syntax.Assign]bool{}
					}
					r.stale[stored] = true
				}
			}
			*f = rejoin(skipped, *f, r)
		case syntax.TCoalesce:
			before := clone(*f)
			right = e.eval(n.Right, f, r, depth+1)
			*f = rejoin(before, *f, r)
		default:
			right = e.eval(n.Right, f, r, depth+1)
		}
		v = merge(left, right)
		v.Expr = x
		v.Identity = 0
		switch n.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
			syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual, syntax.TLess, syntax.TGreater,
			syntax.TSpaceship, syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor,
			// These operators yield a number or throw; `+` may also unite arrays.
			syntax.TMinus, syntax.TMul, syntax.TDiv, syntax.TMod, syntax.TPow, syntax.TSl, syntax.TSr:
			v.Safe |= HTML | SQL | Shell | Header
		}
	case *syntax.Unary:
		v = e.eval(n.Expr, f, r, depth+1)
		v.Expr = x
		v.Identity = 0
		if n.Op.Kind != syntax.TArrayCast {
			v.upload, v.scalars = 0, false
		}
		switch n.Op.Kind {
		case syntax.TExclaim, syntax.TIntCast, syntax.TDoubleCast, syntax.TBoolCast, syntax.TMinus, syntax.TPlus:
			v.Safe |= HTML | SQL | Shell | Header
		}
	case *syntax.IncDec:
		v = e.eval(n.Var, f, r, depth+1)
		if target, ok := n.Var.(*syntax.Variable); ok && target.NameExpr == nil {
			// The stepped variable holds the step applied to a known value,
			// which a reader can evaluate, and nothing provable otherwise.
			next := Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe}
			if v.Expr != nil {
				next.Expr = x
			}
			f.write(r, target.Name, next)
			if n.Prefix {
				v = next
				v.Expr = x
			}
		} else {
			kill(n.Var, f, r)
			v = Value{}
		}
	case *syntax.Ternary:
		cond := e.eval(n.Cond, f, r, depth+1)
		other := clone(*f)
		a := cond // The short form yields its condition.
		if n.Then != nil {
			e.refine(n.Cond, f, true)
			a = e.eval(n.Then, f, r, depth+1)
		}
		e.refine(n.Cond, &other, false)
		b := e.eval(n.Else, &other, r, depth+1)
		*f = rejoin(*f, other, r)
		v = merge(a, b)
	case *syntax.Match:
		// Only an arm result becomes the value; the subject merely selects it.
		e.eval(n.Cond, f, r, depth+1)
		entry := clone(*f)
		for i, arm := range n.Arms {
			branch := clone(entry)
			for _, c := range arm.Conds {
				e.eval(c, &branch, r, depth+1)
			}
			body := e.eval(arm.Body, &branch, r, depth+1)
			if i == 0 {
				v, *f = body, branch
			} else {
				v = merge(v, body)
				*f = rejoin(*f, branch, r)
			}
		}
		v.Expr = x
		v.Identity = 0
	case *syntax.ArrayDimFetch:
		v = e.eval(n.Var, f, r, depth+1)
		if conditional(n.Var) {
			indexed := clone(*f)
			e.eval(n.Dim, &indexed, r, depth+1)
			*f = rejoin(*f, indexed, r)
		} else {
			e.eval(n.Dim, f, r, depth+1)
		}
		v.Expr = x
		v.Identity = 0
		root := n.Var
		levels := 1
		for {
			fetch, ok := root.(*syntax.ArrayDimFetch)
			if !ok {
				break
			}
			root = fetch.Var
			levels++
		}
		// What an element of the uploads table is depends on how deep it sits.
		whole := v.upload
		v.upload = 0
		if v.scalars {
			// An element of such an array is one of its values.
			v.scalars = false
			v.Safe |= HTML | SQL | Shell | Header
		}
		if variable, ok := root.(*syntax.Variable); ok {
			key, known := literalText(n.Dim)
			if variable.NameExpr == nil && !external(variable.Name) {
				own, tracked := Value{}, false
				if slot, constant := arrayKey(n.Dim); constant && levels == 1 {
					own, tracked = f.vars[variable.Name].elements.lookup(slot)
				}
				switch {
				case tracked:
					// The element's own last write decides what it holds.
					v = Value{Complete: own.Complete, Sources: own.Sources, Safe: own.Safe, Expr: x}
				case whole == uploadEntry && known && uploadField(key):
					// The server chooses these fields of an uploaded file.
					v = Value{Expr: x, Complete: true}
				case whole == uploadTable:
					v.upload = uploadEntry
				case variable.Name == "GLOBALS" && levels == 1 && known && (r.global || r.globalNames[key]):
					// The entry is the variable of that name.
					held := f.vars[key]
					v = Value{Complete: held.Complete, Sources: held.Sources, Safe: held.Safe, Expr: x}
				}
			}
			if variable.Name == "_FILES" && levels == 1 {
				v.upload = uploadEntry
			}
			if variable.Name == "_SERVER" {
				if !known {
					v = Value{Expr: x}
				} else if !(strings.HasPrefix(key, "HTTP_") || key == "REQUEST_URI" || key == "QUERY_STRING" || key == "PHP_SELF" || key == "PATH_INFO" || strings.HasPrefix(key, "PHP_AUTH_")) {
					v = Value{Expr: x, Complete: true}
				}
			}
			// The fields of an entry are one level below it; deeper values
			// are what their field holds (a list of names, of sizes).
			if variable.Name == "_FILES" && levels == 2 && known {
				switch key {
				case "tmp_name", "error", "size":
					v = Value{Expr: x, Complete: true}
				case "name", "type", "full_path":
				default:
					v = Value{Expr: x}
				}
			}
		}
	case *syntax.Array:
		for _, item := range n.Items {
			if item == nil {
				continue
			}
			e.eval(item.Key, f, r, depth+1)
			value := e.eval(item.Value, f, r, depth+1)
			v = merge(v, value)
			if item.Unpack || item.ByRef {
				v.Complete = false
			}
			if item.ByRef {
				f.alias(rootVariable(item.Value), aliasAny)
			}
		}
		v.Expr = x
	case *syntax.PropertyFetch:
		e.eval(n.Var, f, r, depth+1)
		v = Value{}
	case *syntax.Closure:
		// The body executes only on invocation, where it can write these names.
		for _, use := range n.Uses {
			if use.ByRef && use.Var != nil {
				f.alias(use.Var.Name, aliasHeld)
				if written := e.holderWrites(n, use.Var.Name); tainted(written) {
					if r.holders == nil {
						r.holders = map[string]Value{}
					}
					r.holders[use.Var.Name] = written
				}
			}
		}
	case *syntax.ArrowFunction: // The body executes only on invocation.
	case *syntax.Yield, *syntax.YieldFrom:
		v = Value{}
		r.complete = false
	case *syntax.Eval, *syntax.Include:
		f.vars = map[string]Value{}
		f.states = map[uint32]State{}
		v = Value{}
		r.complete = false
	default:
		first := true
		syntax.Children(x, func(child syntax.Node) {
			if expr, ok := child.(syntax.Expr); ok {
				cv := e.eval(expr, f, r, depth+1)
				if first {
					v = cv
					first = false
				} else {
					v = merge(v, cv)
				}
			}
		})
		v.Expr = x
		switch x.(type) {
		case *syntax.Instanceof, *syntax.Isset, *syntax.Empty:
			v.Safe |= HTML | SQL | Shell | Header
		}
	}
	// Only an expression that hands its operand on unchanged yields the
	// uploads table, one of its entries or an array of numbers.
	if v.upload != 0 || v.scalars {
		switch n := x.(type) {
		case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.ArrayDimFetch, *syntax.Unary, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.Arg:
		case *syntax.Binary:
			if n.Op.Kind != syntax.TCoalesce {
				v.upload, v.scalars = 0, false
			}
		default:
			v.upload, v.scalars = 0, false
		}
	}
	if old, ok := r.values[x]; ok {
		seen := merge(old, v)
		if _, read := x.(*syntax.Variable); read {
			// Every visit of the read adds the producers that reach it.
			seen.origins = origins(old, v)
		}
		r.values[x] = seen
	} else {
		r.values[x] = v
	}
	return v
}

// plainVariable names a target that is a variable written as such.
func plainVariable(x syntax.Expr) string {
	if v, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok && v.NameExpr == nil {
		return v.Name
	}
	return ""
}

// rebind handles a statement that makes a variable name stand for other
// storage (`unset`, `static`, `global`). Anything but a plain variable is a
// write to what the target names.
func rebind(target syntax.Expr, f *frame, r *scopeResult) {
	if name := plainVariable(target); name != "" {
		f.detach(name)
		return
	}
	kill(target, f, r)
}

// assign evaluates an assignment on the path that performs it.
func (e *Env) assign(n *syntax.Assign, f *frame, r *scopeResult, depth int) Value {
	var v Value
	target, plain := n.Var.(*syntax.Variable)
	plain = plain && target.NameExpr == nil && !n.ByRef
	if plain && n.Op.Kind != syntax.TEqual && !external(target.Name) {
		// A compound assignment combines the variable with the operand.
		old := e.eval(n.Var, f, r, depth+1)
		v = compound(n, old, e.eval(n.Value, f, r, depth+1), scalar(r, old.Expr, MaxDepth))
		f.write(r, target.Name, v)
		return v
	}
	v = e.eval(n.Value, f, r, depth+1)
	switch {
	case plain && n.Op.Kind == syntax.TEqual:
		// The variable holds the array with the elements tracked for it.
		bound := v
		bound.elements = e.elements(n.Value, f, r)
		f.write(r, target.Name, bound)
		if source := plainVariable(n.Value); source != "" && f.aliased[source] == aliasAny {
			// A copy of an array shares the slots references were taken of.
			f.alias(target.Name, aliasAny)
		}
	case n.ByRef:
		e.reference(n, v, f, r)
		v.Complete = false
	default:
		e.store(n, v, f, r)
		v.Complete = false
	}
	return v
}

// conditional reports a member access PHP may skip: a null-safe link, or
// anything chained after one, is not evaluated when the receiver is null.
func conditional(x syntax.Expr) bool {
	for {
		switch n := syntax.UnwrapParens(x).(type) {
		case *syntax.MethodCall:
			if n.NullSafe {
				return true
			}
			x = n.Var
		case *syntax.PropertyFetch:
			if n.NullSafe {
				return true
			}
			x = n.Var
		case *syntax.ArrayDimFetch:
			x = n.Var
		default:
			return false
		}
	}
}

func external(s string) bool {
	switch s {
	case "_GET", "_POST", "_REQUEST", "_COOKIE", "_FILES", "_SERVER":
		return true
	}
	return false
}

func (e *Env) call(n syntax.Node, args *syntax.ArgList, receiver syntax.Expr, f *frame, r *scopeResult, depth int) Value {
	var rv Value
	if receiver != nil {
		rv = e.eval(receiver, f, r, depth+1)
	}
	values := []Value{}
	written := []*syntax.Arg{}
	if args != nil {
		for _, x := range args.Args {
			values = append(values, e.eval(x, f, r, depth+1))
			if a, ok := x.(*syntax.Arg); ok {
				written = append(written, a)
			}
		}
	}
	name, builtin, result := e.callee(n)
	c := Call{Node: n, Name: name, Receiver: receiver, Arguments: values}
	if i, ok := r.callIndex[n]; ok {
		for j, v := range values {
			if j < len(r.calls[i].Arguments) {
				r.calls[i].Arguments[j] = merge(r.calls[i].Arguments[j], v)
			}
		}
	} else {
		r.callIndex[n] = len(r.calls)
		r.calls = append(r.calls, c)
	}
	v := Value{Complete: true, Expr: n.(syntax.Expr)}
	for _, a := range values {
		if a.scalars && builtin && (yieldsElements(name) || name == "array_chunk" && len(values) == 2) {
			// Only the numbers of the array reach the result.
			a.Safe |= HTML | SQL | Shell | Header
		}
		v = merge(v, a)
	}
	v.Expr = n.(syntax.Expr)
	v.Identity = 0
	// Arrays of numbers put together make one; anything else does not.
	v.upload, v.scalars = 0, v.scalars && builtin && (name == "array_merge" || name == "array_replace")
	bound, boundOK := e.bind(n, values)
	first := Value{}
	if boundOK && len(bound) > 0 {
		first = bound[0]
	} else if call, ok := n.(*syntax.FuncCall); ok && builtin && call.Args != nil {
		if ordered, known := e.guardArguments(call); known && len(ordered) > 0 && ordered[0] != nil {
			first = r.values[ordered[0]]
		}
	} else if !builtin && len(values) > 0 {
		first = values[0]
	}
	if s, ok := e.contract(e.summaryName(n)); ok {
		v = Value{Complete: s.Return.Complete, Expr: n.(syntax.Expr), Safe: s.Return.Safe}
		for _, source := range s.Return.Sources {
			if source.Parameter >= 0 {
				if boundOK && source.Parameter < len(bound) {
					v = merge(v, bound[source.Parameter])
				} else {
					v.Complete = false
				}
			} else {
				v.Sources, _ = union(v.Sources, []Source{source})
			}
		}
		v.Safe |= s.Return.Safe
		v.scalars = s.Return.scalars && len(v.Sources) > 0
		if s.Generator {
			v = Value{Complete: true, Expr: n.(syntax.Expr), Identity: n.Span().Start + 1}
		}
	} else if !builtin {
		// Nothing is known about what an unresolved callee returns: it may
		// well validate or escape its arguments.
		v = Value{Expr: v.Expr}
	}
	if builtin {
		if len(v.Sources) > 0 && scalarResult(result) {
			v.Safe |= HTML | SQL | Shell | Header
		}
		switch name {
		case "extract", "parse_str":
			// Each key of the array may replace a local binding of that
			// name, or create one, with an element of the array.
			if array, given := passed(args, values, "array", 0); given && name == "extract" || len(values) == 1 {
				if !given {
					array = values[0]
				}
				if array.scalars {
					// The variables hold the numbers; the keys name them.
					array.Safe |= HTML | SQL | Shell | Header
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
			if len(values) > 0 && e.scalarCallback(values[0].Expr) && len(v.Sources) > 0 {
				if len(values) == 2 && !e.list(values[1].Expr) {
					v.scalars = true
				} else {
					v.Safe |= HTML | SQL | Shell | Header
				}
			}
		}
		if len(values) > 0 && values[0].scalars && keepsElements(name) {
			// The result is that array with fewer elements: the other
			// arguments only choose among them.
			v = Value{Complete: v.Complete, Expr: v.Expr, Sources: values[0].Sources, Safe: values[0].Safe, scalars: true}
		}
		// Invoking a closure object runs project code like calling it.
		if e.exposed(f, r) && (name == "__invoke" || name == "call") {
			e.invalidateShared(f, r)
		}
		if e.exposed(f, r) && callbackParameter(e, n) {
			e.invalidateShared(f, r)
		}
		switch name {
		case "htmlspecialchars", "htmlentities":
			v.Safe |= HTML
		case "escapeshellarg":
			v.Safe |= Shell
		case "intval", "floatval":
			v.Safe |= HTML | SQL | Shell | Header
		case "md5", "sha1", "hash", "hash_hmac":
			position := 1
			switch name {
			case "hash":
				position = 2
			case "hash_hmac":
				position = 3
			}
			if boundOK && position < len(bound) {
				flag := bound[position]
				if flag.Complete && flag.Expr == nil {
					v.Safe |= HTML | SQL | Shell | Header
				} else {
					switch flag := syntax.UnwrapParens(flag.Expr).(type) {
					case *syntax.ConstFetch:
						if strings.EqualFold(flag.Name.Value, "false") {
							v.Safe |= HTML | SQL | Shell | Header
						}
					case *syntax.Literal:
						if flag.LitKind == syntax.LitInt && flag.Raw == "0" {
							v.Safe |= HTML | SQL | Shell | Header
						}
					}
				}
			}
		case "fopen", "tmpfile", "popen", "curl_init":
			v.Identity = n.Span().Start + 1
		case "prepare", "query":
			v = Value{Complete: true, Expr: n.(syntax.Expr), Identity: n.Span().Start + 1}
		}
		switch name {
		case "session_start", "session_write_close", "session_commit", "session_abort", "session_destroy":
			f.states[sessionID] = State{Operation: name, Span: n.Span()}
		case "http_response_code":
			f.states[responseID] = State{Operation: name, Span: n.Span()}
		case "flush", "ob_end_flush":
			if _, output := f.states[outputID]; output {
				if _, buffering := f.states[bufferID]; !buffering || name == "ob_end_flush" {
					f.states[committedID] = State{Operation: name, Span: n.Span()}
				}
			}
			if name == "ob_end_flush" {
				delete(f.states, bufferID)
			}
		case "ob_start":
			f.states[bufferID] = State{Operation: name, Span: n.Span()}
		case "ob_end_clean":
			delete(f.states, bufferID)
		case "header":
			if text, known := literalText(first.Expr); known && strings.HasPrefix(strings.ToLower(text), "set-cookie:") {
				f.states[cookieID] = State{Operation: name, Span: n.Span()}
			}
		}
	} else {
		e.invalidateEscaped(n, append(values, rv), f, r)
		e.invalidateShared(f, r)
		for id := range f.states {
			if id >= bufferID {
				delete(f.states, id)
			}
		}
	}
	if v.Identity != 0 {
		f.states[v.Identity] = State{Operation: name, Span: n.Span()}
	}
	id := rv.Identity
	if receiver == nil {
		id = first.Identity
	}
	if id != 0 {
		if builtin {
			f.states[id] = State{Operation: name, Span: n.Span()}
		} else {
			delete(f.states, id)
		}
	}
	for i, a := range written {
		if a.ByRef || a.Unpack {
			f.forget(r)
		}
		if i < len(values) && !builtin && values[i].Identity != 0 {
			delete(f.states, values[i].Identity)
		}
	}
	e.invalidateReferences(n, args, f, r)
	if builtin && args != nil && e.types.PHP.AtLeast(phpversion.PHP80) {
		e.accepted(n, args, f)
	}
	return v
}

func (e *Env) callName(n syntax.Node) (string, bool) {
	name, builtin, _ := e.callee(n)
	return name, builtin
}

// callee resolves a call's folded name, whether PHP itself declares it and,
// when it does, its declared native return type.
func (e *Env) callee(n syntax.Node) (string, bool, string) {
	switch n := n.(type) {
	case *syntax.FuncCall:
		if f := e.types.ResolveFunction(n); f != nil && f.Avail.In(e.types.PHP) {
			return strings.ToLower(f.FQN), f.Builtin, f.Return
		}
	case *syntax.MethodCall:
		if id, ok := n.Name.(*syntax.Identifier); ok {
			classes := e.types.TypeOf(n.Var).Classes()
			if len(classes) == 1 {
				if m := e.types.Index.FindMethod(classes[0], id.Value, e.types.PHP); m != nil && m.Avail.In(e.types.PHP) {
					return strings.ToLower(id.Value), m.Builtin, m.Return
				}
			}
			return strings.ToLower(id.Value), false, ""
		}
	case *syntax.StaticCall:
		if id, ok := n.Name.(*syntax.Identifier); ok {
			if class, ok := n.Class.(*syntax.Name); ok {
				fqn := e.types.Names.Class(class.Value, class.Span().Start)
				if m := e.types.Index.FindMethod(fqn, id.Value, e.types.PHP); m != nil && m.Avail.In(e.types.PHP) {
					return strings.ToLower(id.Value), m.Builtin, m.Return
				}
			}
			return strings.ToLower(id.Value), false, ""
		}
	}
	return "", false, ""
}

// invalidateEscaped separates operation histories across calls with unknown effects.
func (e *Env) invalidateEscaped(at syntax.Node, values []Value, f *frame, r *scopeResult) {
	identities := map[uint32]bool{}
	all := false
	for _, value := range values {
		if value.Identity != 0 {
			identities[value.Identity] = true
		}
		if value.Expr == nil {
			continue
		}
		syntax.Inspect(value.Expr, func(child syntax.Node) bool {
			if !r.complete {
				return false
			}
			r.operations++
			if r.operations > MaxTransfers {
				r.complete = false
				return false
			}
			if syntax.IsVariableScope(child) {
				all = true
				return false
			}
			if expr, ok := child.(syntax.Expr); ok {
				if v := r.values[expr]; v.Identity != 0 {
					identities[v.Identity] = true
				}
			}
			return true
		})
	}
	for name, value := range f.vars {
		if value.Identity != 0 && (all || identities[value.Identity]) {
			value.Invalidated = at.Span().Start + 1
			f.vars[name] = value
			delete(f.states, value.Identity)
		}
	}
}

// compound is the value a compound assignment leaves in its variable: the
// concatenation for `.=`, a number for the arithmetic operators, either
// operand's content otherwise. Its expression stays readable only when the
// previous value was known.
func compound(n *syntax.Assign, old, operand Value, scalar bool) Value {
	v := merge(old, operand)
	switch n.Op.Kind {
	case syntax.TPlusEqual:
		// Adding to a value that is not an array yields a number.
		if scalar {
			v.Safe |= HTML | SQL | Shell | Header
		}
	case syntax.TMinusEqual, syntax.TMulEqual, syntax.TDivEqual, syntax.TModEqual, syntax.TPowEqual, syntax.TSlEqual, syntax.TSrEqual:
		v.Safe |= HTML | SQL | Shell | Header
	}
	v = Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe}
	if old.Expr != nil && n.Op.Kind != syntax.TCoalesceEqual {
		v.Expr = n
	}
	return v
}

// kill discards the values a write target replaces. An element target
// changes part of its array, which keeps the request data it held; a
// property, static property or slot of either may be a reference to a local
// whose reference was stored there.
func kill(target syntax.Expr, f *frame, r *scopeResult) {
	switch n := syntax.UnwrapParens(target).(type) {
	case nil:
	case *syntax.Variable:
		if n.NameExpr != nil {
			f.forget(r)
		} else {
			f.write(r, n.Name, Value{})
		}
	case *syntax.ArrayDimFetch:
		switch name := rootVariable(n); name {
		case "":
			kill(n.Var, f, r)
		case "GLOBALS":
			global(n, Value{}, false, f, r)
		default:
			f.write(r, name, unsettled(f.vars[name], Value{}))
			// The slot may be a reference to a variable (`$a = [&$v]`).
			f.through(r, Value{})
		}
	case *syntax.Array, *syntax.List:
		for _, item := range targetItems(target) {
			if item != nil {
				kill(item.Value, f, r)
			}
		}
	default:
		f.through(r, Value{})
	}
}

// global applies a write to a globals table entry to the variable it stands
// for: every variable at file level, and the ones bound with `global` in a
// function. definite marks a plain assignment of the whole entry. A dynamic
// key may be any of them.
func global(target *syntax.ArrayDimFetch, v Value, definite bool, f *frame, r *scopeResult) {
	name := assigned(target)
	switch {
	case name == "":
		if r.global {
			f.unsettle(r, v)
		} else {
			r.operations += len(r.globalNames)
			for bound := range r.globalNames {
				f.write(r, bound, unsettled(f.vars[bound], v))
			}
		}
	case !r.global && !r.globalNames[name]:
	case definite:
		f.write(r, name, v)
	default:
		f.write(r, name, unsettled(f.vars[name], v))
	}
}

// reference handles `target = &source`.
func (e *Env) reference(n *syntax.Assign, v Value, f *frame, r *scopeResult) {
	target, plain := syntax.UnwrapParens(n.Var).(*syntax.Variable)
	source := rootVariable(n.Value)
	if plain && target.NameExpr == nil {
		// The two names now read one value.
		if direct, ok := syntax.UnwrapParens(n.Value).(*syntax.Variable); ok && direct.NameExpr == nil {
			f.bind(target.Name, direct.Name, n.Span().Start)
			f.vars[target.Name] = f.vars[direct.Name]
			return
		}
		f.alias(target.Name, aliasAny)
		f.alias(source, aliasAny)
		f.write(r, target.Name, unsettled(Value{}, v))
		return
	}
	// The reference is kept in an element, a property or a dynamic name. The
	// source keeps its value; an array given the slot changes.
	f.alias(source, aliasAny)
	switch slot := syntax.UnwrapParens(n.Var).(type) {
	case *syntax.Variable:
		f.forget(r)
	case *syntax.ArrayDimFetch:
		if name := rootVariable(slot); name != "" {
			f.alias(name, aliasAny)
			f.vars[name] = unsettled(f.vars[name], v)
		}
	}
}

// store handles a write to anything but a plain variable. Only the written
// targets change: an element write alters its array, a destructuring its
// slots, a property write no local binding unless a reference was stored
// there. What is written may carry request data to them.
func (e *Env) store(n *syntax.Assign, v Value, f *frame, r *scopeResult) {
	if v.Identity != 0 {
		// A handle kept elsewhere can be operated on without its variable.
		delete(f.states, v.Identity)
	}
	fetch, element := syntax.UnwrapParens(n.Var).(*syntax.ArrayDimFetch)
	name := rootVariable(n.Var)
	switch {
	case !element || name == "":
		kill(n.Var, f, r)
		f.through(r, v)
		if tainted(v) {
			spread(n.Var, v, f, r)
		}
	case name == "GLOBALS":
		if tainted(v) {
			e.globalWrite(n.Var, v, r)
		}
		_, whole := fetch.Var.(*syntax.Variable)
		global(fetch, v, whole && n.Op.Kind == syntax.TEqual, f, r)
	default:
		old := f.vars[name]
		array := unsettled(old, v)
		// A literal key directly in the array is tracked on its own; a
		// deeper or compound write adds to what the element held.
		first := fetch
		for {
			inner, nested := first.Var.(*syntax.ArrayDimFetch)
			if !nested {
				break
			}
			first = inner
		}
		if key, known := arrayKey(first.Dim); known && !external(name) && len(old.elements.all()) < MaxSources {
			own := Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe}
			if first != fetch || n.Op.Kind != syntax.TEqual {
				previous := old
				if tracked, ok := old.elements.lookup(key); ok {
					previous = tracked
				}
				own = unsettled(previous, v)
			}
			r.operations += len(old.elements.all())
			array.elements = old.elements.with(key, own)
		}
		f.write(r, name, array)
		f.through(r, v)
	}
}

// elements returns the tracked elements of the array a plain assignment
// stores: those of the variable, array literal or choice of either it reads.
// A long literal, or one with a key the analysis cannot name, is read as a
// whole.
func (e *Env) elements(value syntax.Expr, f *frame, r *scopeResult) *elementSet {
	switch n := syntax.UnwrapParens(value).(type) {
	case *syntax.Variable:
		if n.NameExpr == nil && !external(n.Name) {
			return f.vars[n.Name].elements
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TArrayCast {
			return e.elements(n.Expr, f, r)
		}
	case *syntax.Ternary:
		if n.Then != nil {
			return common(e.elements(n.Then, f, r), e.elements(n.Else, f, r))
		}
	case *syntax.Array:
		if len(n.Items) == 0 || len(n.Items) > MaxSources {
			return nil
		}
		next, counting := int64(0), true
		set := &elementSet{items: make(map[string]Value, len(n.Items))}
		for _, item := range n.Items {
			if item == nil || item.Unpack || item.ByRef {
				return nil
			}
			key, known := strconv.FormatInt(next, 10), counting
			if item.Key != nil {
				key, known = arrayKey(item.Key)
			}
			if !known {
				return nil
			}
			// PHP numbers an item without a key after the largest integer
			// key so far; what follows a negative one depends on its version.
			if index, err := strconv.ParseInt(key, 10, 64); err == nil && index >= next {
				next = index + 1
			} else if err == nil && index < 0 {
				counting = false
			}
			own := r.values[item.Value]
			set.items[key] = Value{Complete: own.Complete, Sources: own.Sources, Safe: own.Safe}
		}
		r.operations += len(set.items)
		return set
	}
	return nil
}

// arrayKey returns the key PHP uses for a constant array offset: a decimal
// integer for an integer in any notation, a float (truncated), a boolean and
// a string that spells an integer; the empty string for null.
func arrayKey(x syntax.Expr) (string, bool) {
	negative := false
	x = syntax.UnwrapParens(x)
	if sign, ok := x.(*syntax.Unary); ok && sign.Op.Kind == syntax.TMinus {
		negative, x = true, syntax.UnwrapParens(sign.Expr)
	}
	switch n := x.(type) {
	case *syntax.ConstFetch:
		if n.Name == nil || negative {
			break
		}
		switch strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`)) {
		case "true":
			return "1", true
		case "false":
			return "0", true
		case "null":
			return "", true
		}
	case *syntax.Literal:
		raw := strings.ReplaceAll(n.Raw, "_", "")
		switch {
		case n.LitKind == syntax.LitInt:
			if index, err := strconv.ParseInt(raw, 0, 64); err == nil {
				if negative {
					index = -index
				}
				return strconv.FormatInt(index, 10), true
			}
		case n.LitKind == syntax.LitFloat:
			if number, err := strconv.ParseFloat(raw, 64); err == nil && number < 1<<53 {
				if negative {
					number = -number
				}
				return strconv.FormatInt(int64(number), 10), true
			}
		case !negative:
			text, known := literalText(n)
			if index, err := strconv.ParseInt(text, 10, 64); err == nil && strconv.FormatInt(index, 10) == text {
				return text, true
			}
			return text, known
		}
	}
	return "", false
}

// lookup returns a tracked element.
func (s *elementSet) lookup(key string) (Value, bool) {
	if s == nil {
		return Value{}, false
	}
	v, ok := s.items[key]
	return v, ok
}

// all returns the tracked elements; none when nothing is tracked.
func (s *elementSet) all() map[string]Value {
	if s == nil {
		return nil
	}
	return s.items
}

// list reports an array a PHP function built and numbered itself: none of
// its keys comes from the data.
func (e *Env) list(array syntax.Expr) bool {
	call, ok := syntax.UnwrapParens(array).(*syntax.FuncCall)
	if !ok {
		return false
	}
	switch name, builtin := e.callName(call); {
	case !builtin:
	case name == "explode", name == "preg_split", name == "str_split", name == "mb_str_split", name == "array_values", name == "array_keys", name == "str_getcsv", name == "range":
		return true
	}
	return false
}

// yieldsElements reports a PHP function whose result is made of the values
// of the arrays it is given, without their keys.
func yieldsElements(name string) bool {
	switch name {
	case "implode", "join", "array_values", "array_sum", "array_product", "max", "min", "count", "sizeof", "in_array", "reset", "end", "current", "array_pop", "array_shift":
		return true
	}
	return false
}

// keepsElements reports a PHP function whose result is its first argument
// with some elements left out or reordered, under the same keys.
func keepsElements(name string) bool {
	switch name {
	case "array_unique", "array_filter", "array_slice", "array_reverse", "array_diff", "array_diff_key", "array_intersect", "array_intersect_key":
		return true
	}
	return false
}

// passed returns the value given to the parameter of that name and position.
func passed(list *syntax.ArgList, values []Value, name string, position int) (Value, bool) {
	at := 0
	for i, v := range values {
		if a, ok := list.Args[i].(*syntax.Arg); ok && a.Name != nil {
			if a.Name.Value == name {
				return v, true
			}
			continue
		}
		if at == position {
			return v, true
		}
		at++
	}
	return Value{}, false
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

// uploadField names the entries of an uploaded file the server chooses.
func uploadField(key string) bool {
	return key == "tmp_name" || key == "error" || key == "size"
}

// scalar reports an expression whose value is known not to be an array: a
// literal, or a step or compound assignment that started from one.
func scalar(r *scopeResult, x syntax.Expr, depth int) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Literal, *syntax.IncDec, *syntax.InterpolatedString:
		return true
	case *syntax.Assign:
		switch n.Op.Kind {
		case syntax.TPlusEqual, syntax.TCoalesceEqual:
			return depth > 0 && scalar(r, r.values[n.Var].Expr, depth-1)
		}
		return n.Op.Kind != syntax.TEqual
	}
	return false
}

// spread adds the request data of a written value to the bindings a target
// stores it in.
func spread(target syntax.Expr, v Value, f *frame, r *scopeResult) {
	if name := rootVariable(target); name != "" {
		// The names bound with the target read what it was given.
		f.write(r, name, unsettled(f.vars[name], v))
		return
	}
	for _, item := range targetItems(target) {
		if item != nil {
			spread(item.Value, v, f, r)
		}
	}
}

// exposed reports whether the frame holds any binding code outside the scope
// can replace. Most frames do not, which spares resolving the callee.
func (e *Env) exposed(f *frame, r *scopeResult) bool {
	return len(f.aliased) > 0 || r.global && e.globalsTable()
}

// invalidateShared discards the bindings code outside the scope can replace
// while a callee with unknown effects, or a callback, runs.
func (e *Env) invalidateShared(f *frame, r *scopeResult) {
	if r.global && e.globalsTable() {
		r.operations += len(f.vars)
		for name, old := range f.vars {
			f.vars[name] = unsettled(old, e.held(name, r))
		}
		return
	}
	// What a holder writes to one name is read under the names bound with it.
	r.operations += len(f.aliased)
	var groups map[uint32]Value
	for name, group := range f.aliased {
		if written := e.held(name, r); group >= aliasGroup && tainted(written) {
			if groups == nil {
				groups = map[uint32]Value{}
			}
			groups[group] = unsettled(groups[group], written)
		}
	}
	for name, group := range f.aliased {
		written, shared := groups[group]
		if !shared {
			written = e.held(name, r)
		}
		f.vars[name] = unsettled(f.vars[name], written)
	}
}

// held is the request data a reference holder of name writes to it: a
// closure of the scope capturing it by reference, or a function of the file
// that binds the same global variable.
func (e *Env) held(name string, r *scopeResult) Value {
	written := r.holders[name]
	if r.global || r.globalNames[name] {
		e.globals()
		if other, ok := e.globalTaint[name]; ok {
			written = merge(Value{Complete: true, Sources: written.Sources, Safe: written.Safe}, other)
		}
	}
	return written
}

// holderWrites is the request data a scope may leave in name, a variable it
// shares with its creator or the file: what the variable holds when the scope
// ends, and what the scope stores in the globals table entry of that name.
func (e *Env) holderWrites(scope syntax.Node, name string) Value {
	r := e.scope(scope)
	written := Value{Complete: true}
	for _, v := range []Value{r.exit[name], r.globalWrites[name]} {
		if tainted(v) {
			written = merge(written, Value{Complete: true, Sources: v.Sources, Safe: v.Safe})
		}
	}
	return written
}

// globalWrite records request data stored in a literal globals table entry.
func (e *Env) globalWrite(target syntax.Expr, v Value, r *scopeResult) {
	name := assigned(target)
	if name == "" {
		return
	}
	if r.globalWrites == nil {
		r.globalWrites = map[string]Value{}
	}
	r.globalWrites[name] = merge(Value{Complete: true, Sources: r.globalWrites[name].Sources, Safe: r.globalWrites[name].Safe}, Value{Complete: true, Sources: v.Sources, Safe: v.Safe})
}

// assigned names the variable a literal globals table entry stands for.
func assigned(target syntax.Expr) string {
	if t, ok := syntax.UnwrapParens(target).(*syntax.ArrayDimFetch); ok {
		if table, ok := t.Var.(*syntax.Variable); ok && table.Name == "GLOBALS" && table.NameExpr == nil {
			if key, known := literalText(t.Dim); known {
				return key
			}
		}
	}
	return ""
}

// callbackParameter reports a resolved builtin that may invoke user code.
func callbackParameter(e *Env, n syntax.Node) bool {
	params, _, _ := e.declaration(n)
	for _, p := range params {
		if strings.Contains(p.Type, "callable") {
			return true
		}
	}
	return false
}

// scalarResult reports a native type that only admits numbers, booleans and
// null. Such a value cannot carry markup, a query fragment or a command.
func scalarResult(declared string) bool {
	if declared == "" {
		return false
	}
	for declared != "" {
		var part string
		part, declared, _ = strings.Cut(declared, "|")
		switch strings.ToLower(strings.TrimPrefix(part, "?")) {
		case "int", "float", "bool", "false", "true", "null", "void", "never":
		default:
			return false
		}
	}
	return true
}

// nativeClass reports the construction of a class PHP itself declares, whose
// constructor runs no project code.
func (e *Env) nativeClass(n *syntax.New) bool {
	name, ok := n.Class.(*syntax.Name)
	if !ok {
		return false
	}
	class := e.types.Index.Class(e.types.Names.Class(name.Value, name.Span().Start), e.types.PHP)
	return class != nil && strings.HasPrefix(class.File, "stubs/")
}
