package flow

import (
	"strings"

	"custos/internal/php/syntax"
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
		e.eval(n.Expr, f, r, 0)
		if n.ByRef || referenceTarget(n.Value, f) {
			// The loop variable stays bound to an element after the loop.
			f.alias(rootVariable(n.Value), true)
			f.alias(rootVariable(n.Expr), true)
		}
		kill(n.Key, f)
		kill(n.Value, f)
	case *syntax.Catch:
		kill(n.Var, f)
	case *syntax.Echo:
		for _, x := range n.Exprs {
			e.eval(x, f, r, 0)
		}
		f.states[outputID] = State{Operation: "echo", Span: n.Span()}
	case *syntax.Global:
		for _, x := range n.Vars {
			f.alias(rootVariable(x), false)
		}
		f.vars = map[string]Value{}
		f.states = map[uint32]State{}
	case *syntax.Unset, *syntax.StaticStmt:
		f.vars = map[string]Value{}
		f.states = map[uint32]State{}
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
	if r.operations > MaxTransfers || depth >= MaxDepth*8 {
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
		} else {
			v = f.vars[n.Name]
		}
	case *syntax.Paren:
		v = e.eval(n.Expr, f, r, depth+1)
	case *syntax.Assign:
		v = e.eval(n.Value, f, r, depth+1)
		if target, ok := n.Var.(*syntax.Variable); ok && target.NameExpr == nil && !n.ByRef && n.Op.Kind == syntax.TEqual {
			f.write(target.Name, v)
		} else {
			if n.ByRef {
				f.alias(rootVariable(n.Var), true)
				f.alias(rootVariable(n.Value), true)
			}
			f.vars = map[string]Value{}
			f.states = map[uint32]State{}
			v.Complete = false
		}
	case *syntax.FuncCall:
		if _, named := n.Name.(*syntax.Name); !named {
			// A callee expression runs first: an invoked closure literal
			// binds its reference captures before the call.
			e.eval(n.Name, f, r, depth+1)
		}
		v = e.call(n, n.Args, nil, f, r, depth)
	case *syntax.MethodCall:
		v = e.call(n, n.Args, n.Var, f, r, depth)
	case *syntax.StaticCall:
		v = e.call(n, n.Args, nil, f, r, depth)
	case *syntax.New:
		values := []Value{}
		if n.Args != nil {
			for _, a := range n.Args.Args {
				values = append(values, e.eval(a, f, r, depth+1))
			}
		}
		e.invalidateEscaped(n, values, f, r)
		e.invalidateReferences(n, n.Args, f)
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
			v.Complete = false
		}
	case *syntax.Binary:
		left := e.eval(n.Left, f, r, depth+1)
		var right Value
		switch n.Op.Kind {
		case syntax.TBooleanAnd, syntax.TAnd, syntax.TBooleanOr, syntax.TOr:
			// The right operand runs only on the left operand's deciding outcome.
			before := clone(*f)
			e.refine(n.Left, f, n.Op.Kind == syntax.TBooleanAnd || n.Op.Kind == syntax.TAnd)
			right = e.eval(n.Right, f, r, depth+1)
			*f = join(before, *f)
		case syntax.TCoalesce:
			before := clone(*f)
			right = e.eval(n.Right, f, r, depth+1)
			*f = join(before, *f)
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
		switch n.Op.Kind {
		case syntax.TExclaim, syntax.TIntCast, syntax.TDoubleCast, syntax.TBoolCast, syntax.TMinus, syntax.TPlus:
			v.Safe |= HTML | SQL | Shell | Header
		}
	case *syntax.IncDec:
		v = e.eval(n.Var, f, r, depth+1)
		if target, ok := n.Var.(*syntax.Variable); ok && target.NameExpr == nil {
			// The stepped variable no longer holds its assigned expression.
			next := Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe}
			f.write(target.Name, next)
			if n.Prefix {
				v = next
				v.Expr = x
			}
		} else {
			kill(n.Var, f)
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
		*f = join(*f, other)
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
				v, *f = merge(v, body), join(*f, branch)
			}
		}
		v.Expr = x
		v.Identity = 0
	case *syntax.ArrayDimFetch:
		v = e.eval(n.Var, f, r, depth+1)
		e.eval(n.Dim, f, r, depth+1)
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
		if variable, ok := root.(*syntax.Variable); ok {
			key, known := literalText(n.Dim)
			if variable.Name == "_SERVER" {
				if !known {
					v = Value{Expr: x}
				} else if !(strings.HasPrefix(key, "HTTP_") || key == "REQUEST_URI" || key == "QUERY_STRING" || key == "PHP_SELF" || key == "PATH_INFO" || strings.HasPrefix(key, "PHP_AUTH_")) {
					v = Value{Expr: x, Complete: true}
				}
			}
			if variable.Name == "_FILES" && levels >= 2 && known {
				switch key {
				case "tmp_name", "error", "size":
					v = Value{Expr: x, Complete: true}
				case "name", "type", "full_path":
				default:
					v.Complete = false
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
				f.alias(rootVariable(item.Value), true)
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
				f.alias(use.Var.Name, false)
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
	if old, ok := r.values[x]; ok {
		r.values[x] = merge(old, v)
	} else {
		r.values[x] = v
	}
	return v
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
		v = merge(v, a)
	}
	v.Expr = n.(syntax.Expr)
	v.Identity = 0
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
	if s, ok := e.snapshot.lookup(e.summaryName(n)); ok {
		v = Value{Complete: s.Return.Complete, Expr: n.(syntax.Expr), Safe: s.Return.Safe}
		for _, source := range s.Return.Sources {
			if source.Parameter >= 0 {
				if boundOK && source.Parameter < len(bound) {
					v = merge(v, bound[source.Parameter])
				} else {
					v.Complete = false
				}
			} else {
				v.Sources = append(v.Sources, source)
			}
		}
		v.Safe |= s.Return.Safe
		if s.Generator {
			v = Value{Complete: true, Expr: n.(syntax.Expr), Identity: n.Span().Start + 1}
		}
	} else if !builtin {
		v.Complete = false
	}
	if builtin {
		if len(v.Sources) > 0 && scalarResult(result) {
			v.Safe |= HTML | SQL | Shell | Header
		}
		switch name {
		case "extract", "parse_str":
			// Each key of the array may replace a local binding of that name.
			if len(values) > 0 && (name == "extract" || len(values) == 1) {
				for other, old := range f.vars {
					imported := merge(old, values[0])
					imported.Expr, imported.Identity = nil, 0
					f.vars[other] = imported
				}
			}
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
			f.vars = map[string]Value{}
			f.states = map[uint32]State{}
		}
		if i < len(values) && !builtin && values[i].Identity != 0 {
			delete(f.states, values[i].Identity)
		}
	}
	e.invalidateReferences(n, args, f)
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

// kill discards the values a write target replaces. A property or static
// property target changes no local binding.
func kill(target syntax.Expr, f *frame) {
	switch n := syntax.UnwrapParens(target).(type) {
	case *syntax.Variable:
		if n.NameExpr != nil {
			f.vars = map[string]Value{}
		} else {
			f.write(n.Name, Value{})
		}
	case *syntax.ArrayDimFetch:
		kill(n.Var, f)
	default:
		for _, item := range targetItems(target) {
			if item != nil {
				kill(item.Value, f)
			}
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
		for name, old := range f.vars {
			f.vars[name] = unsettled(old, e.held(name, r))
		}
		return
	}
	for name := range f.aliased {
		f.vars[name] = unsettled(f.vars[name], e.held(name, r))
	}
}

// held is the request data a reference holder of name writes to it.
func (e *Env) held(name string, r *scopeResult) Value {
	written := r.holders[name]
	if r.global {
		if other, ok := e.globalTaint[name]; ok {
			written = merge(Value{Complete: true, Sources: written.Sources, Safe: written.Safe}, other)
		}
	}
	return written
}

// holderWrites joins the request data that plain assignments inside scope
// give to name, a variable the scope shares with its creator or the file.
// The globals table entry of that name counts as the variable itself.
func (e *Env) holderWrites(scope syntax.Node, name string) Value {
	r := e.scope(scope)
	written := Value{Complete: true}
	syntax.Inspect(scope, func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		if a, ok := n.(*syntax.Assign); ok && assigned(a.Var) == name {
			if v := r.values[a.Value]; tainted(v) {
				written = merge(written, Value{Complete: true, Sources: v.Sources, Safe: v.Safe})
			}
		}
		return true
	})
	return written
}

// assigned names the variable a plain target or a literal globals table
// entry stands for.
func assigned(target syntax.Expr) string {
	switch t := syntax.UnwrapParens(target).(type) {
	case *syntax.Variable:
		if t.NameExpr == nil {
			return t.Name
		}
	case *syntax.ArrayDimFetch:
		if table, ok := t.Var.(*syntax.Variable); ok && table.Name == "GLOBALS" && table.NameExpr == nil {
			if key, known := literalText(t.Dim); known {
				return key
			}
		}
	}
	return ""
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
