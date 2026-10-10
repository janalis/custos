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
		for _, x := range []syntax.Expr{n.Key, n.Value} {
			if v, ok := x.(*syntax.Variable); ok {
				f.vars[v.Name] = Value{}
			}
		}
	case *syntax.Echo:
		for _, x := range n.Exprs {
			e.eval(x, f, r, 0)
		}
		f.states[outputID] = State{Operation: "echo", Span: n.Span()}
	case *syntax.Global, *syntax.Unset, *syntax.StaticStmt:
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
			if f.states[id] != s {
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
			f.vars[target.Name] = v
		} else {
			f.vars = map[string]Value{}
			f.states = map[uint32]State{}
			v.Complete = false
		}
	case *syntax.FuncCall:
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
		before := clone(*f)
		right := e.eval(n.Right, f, r, depth+1)
		if n.Op.Kind == syntax.TBooleanAnd || n.Op.Kind == syntax.TBooleanOr || n.Op.Kind == syntax.TAnd || n.Op.Kind == syntax.TOr || n.Op.Kind == syntax.TCoalesce {
			*f = join(before, *f)
		}
		v = merge(left, right)
		v.Expr = x
		v.Identity = 0
		switch n.Op.Kind {
		case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TIsIdentical, syntax.TIsNotIdentical,
			syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual, syntax.TLess, syntax.TGreater,
			syntax.TSpaceship, syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor:
			v.Safe |= HTML | SQL | Shell | Header
		}
	case *syntax.Unary:
		v = e.eval(n.Expr, f, r, depth+1)
		v.Expr = x
		v.Identity = 0
		switch n.Op.Kind {
		case syntax.TExclaim, syntax.TIntCast, syntax.TDoubleCast, syntax.TBoolCast:
			v.Safe |= HTML | SQL | Shell | Header
		}
	case *syntax.Ternary:
		e.eval(n.Cond, f, r, depth+1)
		other := clone(*f)
		a := e.eval(n.Then, f, r, depth+1)
		b := e.eval(n.Else, &other, r, depth+1)
		*f = join(*f, other)
		v = merge(a, b)
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
		}
		v.Expr = x
	case *syntax.PropertyFetch:
		e.eval(n.Var, f, r, depth+1)
		v = Value{}
	case *syntax.Closure, *syntax.ArrowFunction: // Bodies execute only on invocation.
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
	name, builtin := e.callName(n)
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
	switch n := n.(type) {
	case *syntax.FuncCall:
		if f := e.types.ResolveFunction(n); f != nil && f.Avail.In(e.types.PHP) {
			return strings.ToLower(f.FQN), f.Builtin
		}
	case *syntax.MethodCall:
		if id, ok := n.Name.(*syntax.Identifier); ok {
			classes := e.types.TypeOf(n.Var).Classes()
			if len(classes) == 1 {
				if m := e.types.Index.FindMethod(classes[0], id.Value, e.types.PHP); m != nil && m.Avail.In(e.types.PHP) {
					return strings.ToLower(id.Value), m.Builtin
				}
			}
			return strings.ToLower(id.Value), false
		}
	case *syntax.StaticCall:
		if id, ok := n.Name.(*syntax.Identifier); ok {
			if class, ok := n.Class.(*syntax.Name); ok {
				fqn := e.types.Names.Class(class.Value, class.Span().Start)
				if m := e.types.Index.FindMethod(fqn, id.Value, e.types.PHP); m != nil && m.Avail.In(e.types.PHP) {
					return strings.ToLower(id.Value), m.Builtin
				}
			}
			return strings.ToLower(id.Value), false
		}
	}
	return "", false
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
