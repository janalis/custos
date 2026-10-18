package flow

import (
	"slices"
	"strconv"
	"strings"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

func (e *Env) transfer(n syntax.Node, f *frame, r *scopeResult) {
	remember(r, n, *f)
	switch n := n.(type) {
	case syntax.Expr:
		v := e.eval(n, f, r, 0)
		if label, ok := n.Parent().(*syntax.Case); ok && label.Cond == n {
			// A label is compared loosely with the subject of its switch.
			subject := label.Parent().(*syntax.Switch).Cond
			if constantName(subject) == "" && (!v.plain && !textless(subject) || !r.values[subject].plain && !textless(n)) {
				e.implicit(n, f, r)
			}
		}
	case *syntax.Return:
		v := e.eval(n.Expr, f, r, 0)
		if n.Expr != nil && !v.plain && returnsText(syntax.EnclosingVariableScope(n)) {
			// The declared type converts an object to text.
			r.runs = true
		}
		r.returns = append(r.returns, v)
	case *syntax.ExprStmt:
		e.eval(n.Expr, f, r, 0)
	case *syntax.Foreach:
		// The targets are given an element where the body is entered (see
		// each): the path that leaves the loop computes none of them.
		iterated := e.eval(n.Expr, f, r, 0)
		leaving(f, r)
		if !iterated.plain {
			// An object is iterated by the methods of its class.
			e.implicit(n.Expr, f, r)
		}
	case *syntax.Catch:
		// What the variable held is released, like by an assignment.
		f.unmatchAll(r)
		if n.Var != nil {
			e.release(n, n.Var.Name, f, r)
		}
		kill(n.Var, f, r)
	case *syntax.Echo:
		for _, x := range n.Exprs {
			if v := e.eval(x, f, r, 0); !v.plain {
				// An object is converted to text.
				e.implicit(x, f, r)
			}
		}
		leaving(f, r)
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
				f.vars[name] = unsettled(f.vars[name], e.held(name, r), r)
			}
		}
	case *syntax.Unset:
		f.unmatchAll(r)
		leaving(f, r)
		for _, x := range n.Vars {
			e.operands(x, f, r, 0)
			rebind(x, f, r)
		}
		// What is released may have a destructor.
		e.implicit(n, f, r)
	case *syntax.StaticStmt:
		scope := syntax.EnclosingVariableScope(n)
		for _, x := range n.Vars {
			name := x.Var.Name
			known, seen := r.statics[name]
			if seen && known.state == staticHeld {
				// Reached again, in a loop, the name reads what the function
				// left in the variable.
				r.holds(f, name)
			}
			rebind(x.Var, f, r)
			if !seen {
				if r.statics == nil {
					r.statics = map[string]static{}
				}
				known = e.declared(scope)[name]
				if known.state == staticLoose && !r.distrustAll && !r.distrust[name] {
					known.state = staticHeld
				}
				r.statics[name] = known
			}
			// The first invocation starts from the initial value, any other
			// from what the function left (see declared).
			held := known.state == staticHeld
			f.vars[name] = Value{NonFalse: held && known.starts, entry: held, plain: known.plain, deep: known.deep}
		}
	default:
		syntax.Children(n, func(child syntax.Node) {
			if x, ok := child.(syntax.Expr); ok {
				e.eval(x, f, r, 0)
			}
		})
	}
}

// each starts a turn of a loop on the path that enters its body: the
// targets are evaluated, then given a key and an element of what the loop
// iterates.
func (e *Env) each(n *syntax.Foreach, f *frame, r *scopeResult) {
	iterated := r.values[n.Expr]
	if n.ByRef || referenceTarget(n.Value, f) {
		// The loop variable stays bound to an element after the loop.
		f.alias(rootVariable(n.Value), aliasAny)
		f.alias(rootVariable(n.Expr), aliasAny)
	}
	e.operands(n.Key, f, r, 0)
	e.operands(n.Value, f, r, 0)
	if len(targetItems(n.Value)) > 0 && !iterated.deep {
		// An element that is an object is destructured by a method of its
		// class.
		e.implicit(n, f, r)
	}
	kill(n.Key, f, r)
	kill(n.Value, f, r)
	// The keys and elements of request data are request data.
	if len(iterated.Sources) > 0 {
		spread(n.Key, projectKeys(Value{Complete: true, Sources: iterated.Sources, Safe: iterated.Safe}), f, r)
		item := projectValues(iterated)
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
	if name := plainVariable(n.Key); name != "" && iterated.plain {
		// A key of an array is a number or text: neither false nor null, and
		// no object. Anything else that is iterated chooses its keys.
		key := f.vars[name]
		key.NonFalse, key.NonNull, key.plain, key.deep = true, true, true, true
		f.vars[name] = key
	}
	if name := plainVariable(n.Value); name != "" && iterated.deep && !n.ByRef {
		// An array that holds no object yields none.
		item := f.vars[name]
		item.plain, item.deep = true, true
		f.vars[name] = item
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
		var m map[uint32]State
		if len(f.states) > 0 {
			m = make(map[uint32]State, len(f.states))
			for id, s := range f.states {
				m[id] = s
			}
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
	if len(r.statics) > 0 && !calm(x) {
		// It may also throw, which ends the function here.
		leaving(f, r)
	}
	v := Value{Complete: true, Expr: x}
	switch n := x.(type) {
	case *syntax.Variable:
		if n.NameExpr != nil {
			// The name may be the one of any variable, which a reference
			// to it or a callee given it can then write.
			if name := e.eval(n.NameExpr, f, r, depth+1); !name.plain {
				// An object given for the name is converted to text.
				e.implicit(x, f, r)
			}
			f.definite = false
			v = Value{}
			if name, known := literalText(n.NameExpr); known && external(name) {
				v = e.request(name, n, f)
			}
		} else if external(n.Name) {
			v = e.request(n.Name, n, f)
		} else {
			var tracked bool
			if v, tracked = f.vars[n.Name]; !tracked {
				// A variable nothing assigned is null, and the globals table
				// is an array. The session table is whatever the application
				// put there.
				v.plain, v.deep = n.Name == "GLOBALS" || f.definite && importable(n.Name), f.definite && importable(n.Name)
				if spread := f.vars[imported]; importable(n.Name) && tainted(spread) {
					// An import of request data may have created the variable.
					v = Value{Sources: spread.Sources, Safe: spread.Safe}
				}
			}
			v.elements = nil
			v.tested, v.unless, v.unlessFalse, v.unlessNull, v.skippedFalse, v.skippedNull, v.entry = nil, nil, false, false, false, false, false
		}
	case *syntax.Paren:
		v = e.eval(n.Expr, f, r, depth+1)
	case *syntax.Assign:
		if n.Op.Kind == syntax.TCoalesceEqual {
			// The offsets of the target are always evaluated; the right
			// side runs only when the target is null.
			e.operands(n.Var, f, r, depth+1)
			assigned := clone(*f)
			v = e.assign(n, &assigned, r, depth, true)
			*f = rejoin(*f, assigned, r)
			f.unmatchAll(r)
			break
		}
		v = e.assign(n, f, r, depth, false)
		f.unmatchAll(r)
	case *syntax.FuncCall:
		if _, named := n.Name.(*syntax.Name); !named {
			// A callee expression runs first: an invoked closure literal
			// binds its reference captures before the call.
			e.eval(n.Name, f, r, depth+1)
		}
		if !e.options.Assertions && e.types.PHP.AtLeast(phpversion.PHP70) && e.types.BuiltinAssert(n) {
			skipped := clone(*f)
			v = e.call(n, n.Args, nil, f, r, depth)
			*f = rejoin(skipped, *f, r)
		} else {
			v = e.call(n, n.Args, nil, f, r, depth)
			if e.options.Assertions {
				if condition := e.types.AssertGuard(n); condition != nil {
					e.refine(condition, f, true)
				}
			}
		}
		if e.types.LegacyAssertCode(n) {
			f.definite = false
			f.related = nil
			f.vars = map[string]Value{}
			f.states = map[uint32]State{}
			v = Value{}
			r.complete = false
		}
	case *syntax.MethodCall:
		if conditional(n) {
			// The name and the arguments are not evaluated when the
			// receiver is null.
			called := clone(*f)
			e.computed(n.Name, &called, r, depth)
			v = e.call(n, n.Args, n.Var, &called, r, depth)
			*f = rejoin(*f, called, r)
			break
		}
		e.computed(n.Name, f, r, depth)
		v = e.call(n, n.Args, n.Var, f, r, depth)
	case *syntax.StaticCall:
		e.computed(n.Class, f, r, depth)
		if conditional(n.Class) {
			// The call is not made when a null-safe link before it is null.
			called := clone(*f)
			e.computed(n.Name, &called, r, depth)
			v = e.call(n, n.Args, nil, &called, r, depth)
			*f = rejoin(*f, called, r)
			break
		}
		e.computed(n.Name, f, r, depth)
		v = e.call(n, n.Args, nil, f, r, depth)
	case *syntax.New:
		e.computed(n.Class, f, r, depth)
		values := []Value{}
		if n.Args != nil {
			for _, a := range n.Args.Args {
				values = append(values, e.eval(a, f, r, depth+1))
			}
		}
		f.unmatchAll(r)
		leaving(f, r)
		e.invalidateEscaped(n, values, f, r)
		e.invalidateReferences(n, n.Args, f, r)
		if !e.nativeClass(n) {
			reenter(n, f, r)
			if e.exposed(f, r) {
				e.invalidateShared(f, r)
			}
		} else if e.handed(n, "", n.Args, values) {
			e.implicit(n, f, r)
		}
		for id := range f.states {
			if id >= bufferID {
				delete(f.states, id)
			}
		}
		v.Identity = n.Span().Start + 1
	case *syntax.Arg:
		v = e.eval(n.Value, f, r, depth+1)
		if n.Unpack && !v.plain {
			// An object spread over the parameters is iterated.
			e.implicit(x, f, r)
		}
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
			if assignedInPlace(n.Right) != nil {
				// A later test of what the left operand read tells which
				// path was taken, while the values it read are in place.
				e.witness(n, &skipped)
				e.witness(n, f)
			}
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
				// Where this operand is skipped the variable keeps what it
				// held: a value that is not false or null there, joined with
				// one this operand's test vouches for, is neither (see
				// refineAssigned). A variable nothing assigned is null.
				// Nor is one that passed this test on an earlier turn, where
				// the left operand skips it since (see relation).
				name := stored.Var.(*syntax.Variable).Name
				kept, tracked := skipped.vars[name]
				held := skipped.relates(n)
				notFalse, notNull := kept.NonFalse || !tracked && skipped.definite || held&relatedFalse != 0, kept.NonNull || held&relatedNull != 0
				ran, left := uint8(0), uint8(0)
				if c, decides := compares(n.Left); decides && (c.right != nil || len(e.types.TypeOf(c.left).Classes()) == 0) {
					// Where the operand ran, the left operand does not skip
					// it while what it read stays in place.
					if !e.rebinds(n.Right, n.Left, f) {
						ran = relatedFalse | relatedNull
					}
					if notFalse {
						left |= relatedFalse
					}
					if notNull {
						left |= relatedNull
					}
				}
				skipped.relate(n, left)
				f.relate(n, ran)
				*f = rejoin(skipped, *f, r)
				if joined := f.vars[name]; notFalse || notNull {
					joined.unless, joined.unlessFalse, joined.unlessNull = n, false, false
					joined.skippedFalse, joined.skippedNull = notFalse, notNull
					f.vars[name] = joined
				}
				break
			}
			*f = rejoin(skipped, *f, r)
		case syntax.TCoalesce:
			before := clone(*f)
			right = e.eval(n.Right, f, r, depth+1)
			*f = rejoin(before, *f, r)
		default:
			right = e.eval(n.Right, f, r, depth+1)
		}
		if converted(n, left, right) {
			e.implicit(x, f, r)
		}
		v = merge(left, right)
		if n.Op.Kind != syntax.TCoalesce {
			v.interval = arithmetic(n.Op.Kind, left, right)
		}
		v.Expr = x
		v.Identity = 0
		// Only a concatenation, or a choice, yields the text it is given.
		v.matched = v.matched && (n.Op.Kind == syntax.TDot || n.Op.Kind == syntax.TCoalesce)
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
		v.interval = unaryInts(n.Op.Kind, v)
		if n.Op.Kind == syntax.TStringCast && !v.plain {
			e.implicit(x, f, r)
		}
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
			// Before PHP 8.3 a step leaves an object, or an array, as it is.
			strict := e.types.PHP.AtLeast(phpversion.PHP83)
			next := Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe, plain: v.plain || strict, deep: v.deep || strict}
			delta := int64(1)
			if n.Op.Kind == syntax.TDec {
				delta = -1
			}
			next.interval = arithmetic(syntax.TPlus, v, integerValue(delta))
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
		if truth, known := constantTruth(n.Cond); known {
			chosen := n.Else
			if truth {
				chosen = n.Then
			}
			e.refine(n.Cond, f, truth)
			v = cond
			if chosen != nil {
				v = e.eval(chosen, f, r, depth+1)
			}
			break
		}
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
		// Only an arm result becomes the value; the subject merely selects
		// it. The conditions are compared with it in the order written, each
		// once the ones before it failed, and the default arm, wherever it
		// is written, runs after all of them. Under `match (true)` a
		// condition that matches holds, and one that is a test and does not
		// match failed; `match (false)` reads them the other way round.
		e.eval(n.Cond, f, r, depth+1)
		subject := constantName(n.Cond)
		rest := clone(*f)
		var fallback *syntax.MatchArm
		arms := 0
		result := func(arm *syntax.MatchArm, branch frame) {
			body := e.eval(arm.Body, &branch, r, depth+1)
			if e.ends(arm.Body) {
				// An arm that throws or exits yields nothing and goes
				// nowhere after the match.
				return
			}
			if arms++; arms == 1 {
				v, *f = body, branch
			} else {
				v = merge(v, body)
				*f = rejoin(*f, branch, r)
			}
		}
		for _, arm := range n.Arms {
			if len(arm.Conds) == 0 {
				fallback = arm
				continue
			}
			var branch frame
			for i, c := range arm.Conds {
				e.eval(c, &rest, r, depth+1)
				r.copies += len(rest.vars) + len(rest.states)
				matched := clone(rest)
				if subject == "true" || subject == "false" {
					e.refine(c, &matched, subject == "true")
					if e.boolean(c) {
						e.refine(c, &rest, subject != "true")
					}
				}
				if i == 0 {
					branch = matched
				} else {
					branch = rejoin(branch, matched, r)
				}
			}
			result(arm, branch)
		}
		if fallback != nil {
			result(fallback, rest)
		}
		v.Expr = x
		v.Identity = 0
	case *syntax.ArrayDimFetch:
		if name := assigned(n); external(name) {
			e.eval(n.Var, f, r, depth+1)
			e.eval(n.Dim, f, r, depth+1)
			v = e.request(name, n, f)
			break
		}
		v = e.eval(n.Var, f, r, depth+1)
		if conditional(n.Var) {
			indexed := clone(*f)
			e.eval(n.Dim, &indexed, r, depth+1)
			*f = rejoin(*f, indexed, r)
		} else {
			e.eval(n.Dim, f, r, depth+1)
		}
		if !v.plain {
			// An offset of an object is read by a method of its class.
			e.implicit(x, f, r)
		}
		// An element is no object when its array holds none.
		base := v.deep
		v.Expr = x
		v.Identity = 0
		// What is known of the array says nothing of its element.
		v.NonFalse, v.NonNull = false, false
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
		v = projectValues(v)
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
			if f.matched(requestElement(x)) {
				// A switch found the element equal to a fixed text.
				v = Value{Complete: true, Expr: x, matched: true}
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
		v.plain, v.deep = base, base
	case *syntax.Array:
		v.deep = true
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
			if item.Unpack && !value.plain {
				// An object spread into the array is iterated.
				e.implicit(x, f, r)
			}
			if item.ByRef {
				f.alias(rootVariable(item.Value), aliasAny)
			}
		}
		v.Expr = x
	case *syntax.PropertyFetch:
		e.eval(n.Var, f, r, depth+1)
		if conditional(n) {
			// The name is not evaluated when the object is null.
			read := clone(*f)
			e.computed(n.Name, &read, r, depth)
			*f = rejoin(*f, read, r)
		} else {
			e.computed(n.Name, f, r, depth)
		}
		// A magic method or a hook may read the property. One that is
		// read as declared holds what its type admits.
		v = Value{}
		if r.idle() {
			break
		}
		if !e.backed(n) {
			e.implicit(x, f, r)
			break
		}
		property := e.types.Index.FindProperty(e.types.ClassFQN(syntax.EnclosingClass(n)), n.Name.(*syntax.Identifier).Value, e.types.PHP)
		if kind := e.kinds(property.Type, ""); kind&(kindKnown|kindObject|kindOpaque) == kindKnown {
			v.plain, v.deep = true, kind&kindArray == 0
		}
	case *syntax.InterpolatedString:
		// Each part is converted to text before the next one is read.
		for i, part := range n.Parts {
			value := e.eval(part, f, r, depth+1)
			if !value.plain {
				e.implicit(x, f, r)
			}
			if i == 0 {
				v = value
			} else {
				v = merge(v, value)
			}
		}
		v.Expr = x
	case *syntax.Closure:
		// The body executes only on invocation, where it can write these names.
		for _, use := range n.Uses {
			if use.ByRef && use.Var != nil {
				f.alias(use.Var.Name, aliasHeld)
				if written := e.holderWrites(n, use.Var.Name); retained(written) {
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
		r.gap("generator-suspension", x.Span())
	case *syntax.Eval, *syntax.Include:
		f.definite = false
		f.related = nil
		f.vars = map[string]Value{}
		f.states = map[uint32]State{}
		v = Value{}
		r.complete = false
		r.gap("dynamic-code", x.Span())
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
		switch n := x.(type) {
		case *syntax.Instanceof, *syntax.Isset, *syntax.Empty:
			v.Safe |= HTML | SQL | Shell | Header
		case *syntax.Literal, *syntax.StringPart, *syntax.ConstFetch, *syntax.MagicConst, *syntax.Name, *syntax.Identifier, *syntax.Throw:
		case *syntax.Print, *syntax.Exit:
			// The operand is converted to text.
			if !first && !v.plain {
				e.implicit(x, f, r)
			}
		case *syntax.ClassConstFetch:
			// `X::class` is the name written: nothing is loaded for it,
			// nor for a constant of a class that is loaded.
			if !className(n) && !r.idle() && !e.loaded(n) {
				e.implicit(x, f, r)
			}
		default:
			// A class loads on first use, and `clone` runs a method.
			e.implicit(x, f, r)
		}
	}
	// What the expression yields is no object when it computes a number, a
	// boolean, text, an array or a function; one that hands a value on
	// yields what it read.
	switch n := x.(type) {
	case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.ArrayDimFetch, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall,
		*syntax.PropertyFetch:
	case *syntax.Binary:
		// `+` may unite two arrays, with what they hold.
		if n.Op.Kind != syntax.TCoalesce {
			v.plain, v.deep = true, v.deep || n.Op.Kind != syntax.TPlus
		}
	case *syntax.Unary:
		// An array cast keeps what its operand holds, a text cast its text.
		switch n.Op.Kind {
		case syntax.TAt:
		case syntax.TObjectCast:
			v.plain, v.deep = false, false
		default:
			v.plain, v.deep = true, v.deep || n.Op.Kind != syntax.TArrayCast
			v.matched = v.matched && n.Op.Kind == syntax.TStringCast
		}
	case *syntax.Array:
		v.plain = true
	case *syntax.Literal, *syntax.StringPart, *syntax.InterpolatedString, *syntax.MagicConst:
		v.plain, v.deep = true, true
	case *syntax.ConstFetch:
		v.plain, v.deep = e.fixed(n), e.fixed(n)
	case *syntax.Isset, *syntax.Empty, *syntax.Instanceof, *syntax.Closure, *syntax.ArrowFunction, *syntax.Print:
		v.plain, v.deep, v.matched = true, true, false
	case *syntax.IncDec:
		// Since PHP 8.3 only a number or text can be stepped.
		strict := e.types.PHP.AtLeast(phpversion.PHP83)
		v.plain, v.deep, v.matched = v.plain || strict, v.deep || strict, false
	case *syntax.ClassConstFetch:
		v.plain, v.deep = className(n), className(n)
	default:
		v.plain, v.deep = false, false
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
	switch n := x.(type) {
	case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.IncDec, *syntax.Binary, *syntax.Unary:
	case *syntax.ConstFetch:
		v.interval = e.constantInt(n)
	case *syntax.ClassConstFetch:
		v.interval = e.classConstantInt(n)
	default:
		v.interval = nil
	}
	if old, ok := r.values[x]; ok {
		seen := merge(old, v)
		if _, read := x.(*syntax.Variable); read {
			// Every visit of the read adds the producers that reach it.
			seen.origins = origins(old, v, r)
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
		f.unmatch(r, name)
		f.detach(name)
		return
	}
	kill(target, f, r)
}

// assign evaluates an assignment on the path that performs it. targeted
// tells that the operands of the target were evaluated before.
func (e *Env) assign(n *syntax.Assign, f *frame, r *scopeResult, depth int, targeted bool) Value {
	var v Value
	target, plain := n.Var.(*syntax.Variable)
	plain = plain && target.NameExpr == nil && !n.ByRef
	if plain && n.Op.Kind != syntax.TEqual && !external(target.Name) {
		// A compound assignment combines the variable with the operand.
		old := e.eval(n.Var, f, r, depth+1)
		operand := e.eval(n.Value, f, r, depth+1)
		if n.Op.Kind == syntax.TConcatEqual && !(old.plain && operand.plain) {
			// Appending converts an object to text.
			e.implicit(n, f, r)
		}
		v = compound(n, old, operand, scalar(r, old.Expr, MaxDepth))
		v.interval = arithmetic(n.Op.Kind, old, operand)
		// `??=` keeps or takes a value, `+=` may unite two arrays.
		both := old.deep && operand.deep
		v.plain, v.deep = n.Op.Kind != syntax.TCoalesceEqual || old.plain && operand.plain, both || n.Op.Kind != syntax.TCoalesceEqual && n.Op.Kind != syntax.TPlusEqual
		f.write(r, target.Name, v)
		return v
	}
	if n.ByRef {
		// A reference to a request table or one of its elements reads what
		// the request sent, whatever a switch matched it with.
		f.unmatch(r, rootVariable(n.Value))
	}
	// PHP evaluates the offsets of an element or property target before
	// the value, and the keys of a destructuring after it.
	_, slots := syntax.UnwrapParens(n.Var).(*syntax.Array)
	if _, listed := syntax.UnwrapParens(n.Var).(*syntax.List); listed {
		slots = true
	}
	if !plain && !slots && !targeted {
		e.operands(n.Var, f, r, depth+1)
	}
	v = e.eval(n.Value, f, r, depth+1)
	if slots {
		e.operands(n.Var, f, r, depth+1)
	}
	if !plain && (slots && (!v.plain || !v.deep && nested(n.Var)) || propertyTarget(n.Var) || n.Op.Kind == syntax.TConcatEqual && !v.plain) {
		// A hook or a magic method writes the property; an object is
		// destructured by a method of its class, and appended as text.
		e.implicit(n, f, r)
	}
	switch {
	case plain && n.Op.Kind == syntax.TEqual:
		// What the variable held is released.
		e.release(n, target.Name, f, r)
		// The variable holds the array with the elements tracked for it.
		bound := v
		if _, held := r.statics[target.Name]; held {
			switch given := syntax.UnwrapParens(n.Value).(type) {
			case *syntax.Literal, *syntax.Array, *syntax.InterpolatedString, *syntax.New, *syntax.Closure, *syntax.ArrowFunction:
				// What a static variable is given here is neither false
				// nor null (see leaving).
				bound.NonFalse, bound.NonNull = true, true
			case *syntax.ConstFetch:
				// A cache emptied with null is not left false.
				bound.NonFalse, bound.NonNull = constantName(given) == "null" || constantName(given) == "true", constantName(given) == "true"
			}
		}
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
		if slots && v.deep {
			taken(n.Var, f)
		}
		v.Complete = false
	}
	return v
}

// taken marks the variables a destructuring fills from an array that holds
// no object: each is given an element of it, or null.
func taken(target syntax.Expr, f *frame) {
	for _, item := range targetItems(target) {
		if item == nil || item.ByRef {
			continue
		}
		if name := plainVariable(item.Value); name != "" {
			slot := f.vars[name]
			slot.plain, slot.deep = true, true
			f.vars[name] = slot
		} else {
			taken(item.Value, f)
		}
	}
}

// operands evaluates what a write target computes before the write: its
// offsets, the object and the name of a property, the keys of a
// destructuring. The variable written is not read.
func (e *Env) operands(target syntax.Expr, f *frame, r *scopeResult, depth int) {
	switch n := syntax.UnwrapParens(target).(type) {
	case *syntax.ArrayDimFetch:
		e.operands(n.Var, f, r, depth+1)
		e.computed(n.Dim, f, r, depth)
	case *syntax.PropertyFetch:
		e.eval(n.Var, f, r, depth+1)
		e.computed(n.Name, f, r, depth)
	case *syntax.StaticPropertyFetch:
		e.computed(n.Class, f, r, depth)
	case *syntax.Array, *syntax.List:
		for _, item := range targetItems(target) {
			if item != nil {
				e.computed(item.Key, f, r, depth)
				e.operands(item.Value, f, r, depth+1)
			}
		}
	case *syntax.Variable:
		if n.NameExpr != nil {
			e.eval(n, f, r, depth+1)
		}
	default:
		// The array or object written into is the result of an expression
		// (`$o->list()[0] = 1`).
		e.eval(target, f, r, depth+1)
	}
}

// computed evaluates a member, method or class name, or an offset of a write
// target, written as an expression.
func (e *Env) computed(name syntax.Expr, f *frame, r *scopeResult, depth int) {
	switch name.(type) {
	case nil, *syntax.Identifier, *syntax.Name, *syntax.ClassLike, *syntax.Literal, *syntax.ConstFetch:
		// Nothing is computed.
	default:
		e.eval(name, f, r, depth+1)
	}
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
		case *syntax.StaticCall:
			x = n.Class
		case *syntax.StaticPropertyFetch:
			x = n.Class
		case *syntax.ClassConstFetch:
			x = n.Class
		default:
			return false
		}
	}
}

// request constructs the same provenance for every spelling of a request table.
func (e *Env) request(name string, at syntax.Expr, f *frame) Value {
	held, written := f.vars[name]
	v := Value{
		Complete: true, Expr: at, plain: true, deep: !written || held.deep,
		Sources: []Source{{Kind: name, Path: e.file.Path, Span: at.Span(), Parameter: -1}},
	}
	if name == "_FILES" {
		v.upload = uploadTable
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
	if args != nil && len(args.Args) == 1 {
		if _, callable := args.Args[0].(*syntax.VariadicPlaceholder); callable {
			// `f(...)` only makes a closure: nothing is called.
			return Value{Complete: true, Expr: n.(syntax.Expr), plain: true, deep: true}
		}
	}
	followed := followTyped
	if receiver != nil {
		untest(receiver, f, true)
		followed = e.follows(receiver, f)
		if name := plainVariable(receiver); name != "" && name != "this" {
			// What the variable holds wherever the call is reached.
			if r.receivers == nil {
				r.receivers = map[syntax.Node]uint8{}
			}
			r.receivers[n] = max(r.receivers[n], followed)
		}
	}
	values := []Value{}
	var argumentStorage [8]*syntax.Arg
	written := argumentStorage[:0]
	if args != nil {
		values = make([]Value, 0, len(args.Args))
		for _, x := range args.Args {
			values = append(values, e.eval(x, f, r, depth+1))
			if a, ok := x.(*syntax.Arg); ok {
				written = append(written, a)
			}
		}
	}
	f.unmatchAll(r)
	leaving(f, r)
	name, builtin, result, documented := e.callee(n)
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
	// A PHP function that edits text makes, of matched labels and literals,
	// text of the program.
	labels, texts := 0, 0
	for _, a := range values {
		if builtin && !yieldsElements(name) && !keepsElements(name) && name != "array_merge" && name != "array_replace" && name != "array_map" && name != "array_chunk" {
			// An unaudited transform may expose keys in its result. A later
			// element projection cannot reinterpret those as original values.
			a = projectKeys(a)
		}
		if builtin && yieldsElements(name) {
			a = projectValues(a)
		}
		if a.scalars && builtin && (yieldsElements(name) || name == "array_chunk" && len(values) == 2) {
			// Only the numbers of the array reach the result.
			a.Safe |= HTML | SQL | Shell | Header
		}
		if a.matched {
			labels++
		} else if a.written() {
			texts++
		}
		v = merge(v, a)
	}
	derived := builtin && edits(name) && labels > 0 && labels+texts == len(values)
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
		v = substitute(s.Return, bound, boundOK)
		v.Expr = n.(syntax.Expr)
		v.Safe |= s.Return.Safe
		v.scalars = s.Return.scalars && len(v.Sources) > 0
		if s.Generator {
			v = Value{Complete: true, Expr: n.(syntax.Expr), Identity: n.Span().Start + 1}
		}
	} else if !builtin {
		r.gap("unresolved-dispatch", n.Span())
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
		if name == "__invoke" || name == "call" || e.callbackParameter(n, args, values) {
			reenter(n, f, r)
			if e.exposed(f, r) {
				e.invalidateShared(f, r)
			}
		} else if e.handed(n, name, args, values) {
			e.implicit(n, f, r)
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
		if !e.quiet(n, args, values) {
			reenter(n, f, r)
		}
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
	if args != nil {
		e.accepted(n, args, builtin, followed, f)
	}
	// What PHP itself returns is what it declares. An array one of its
	// functions returns holds what the function was given.
	v.matched = derived && len(v.Sources) == 0
	kind := e.kinds(result, documented)
	v.plain = builtin && kind&(kindKnown|kindObject|kindOpaque) == kindKnown
	v.deep = v.plain && (kind&kindArray == 0 || e.sealed(n, name, args, r))
	if builtin && e.keyed(n, name, args) {
		v.plain, v.deep = true, true
	}
	return v
}

// keyed reports a PHP function whose result holds no object whatever it is
// given: the keys of an array, its values turned into keys or counted, or
// what JSON text decodes to when objects are asked for as arrays.
func (e *Env) keyed(call syntax.Node, name string, list *syntax.ArgList) bool {
	if _, function := call.(*syntax.FuncCall); !function {
		return false
	}
	switch name {
	case "array_keys", "array_flip", "array_count_values":
		return true
	case "json_decode":
		if list == nil {
			return false
		}
		position := 0
		for _, x := range list.Args {
			arg, ok := x.(*syntax.Arg)
			if !ok || arg.Unpack {
				return false
			}
			if arg.Name == nil {
				position++
			}
			if arg.Name == nil && position == 2 || arg.Name != nil && arg.Name.Value == "associative" {
				return constantName(arg.Value) == "true"
			}
		}
	}
	return false
}

func (e *Env) callName(n syntax.Node) (string, bool) {
	name, builtin, _, _ := e.callee(n)
	return name, builtin
}

// callee resolves a call's folded name, whether PHP itself declares it and,
// when it does, its declared native return type and its documented one.
func (e *Env) callee(n syntax.Node) (string, bool, string, string) {
	switch n := n.(type) {
	case *syntax.FuncCall:
		if f := e.types.ResolveFunction(n); f != nil && f.Avail.In(e.types.PHP) {
			return strings.ToLower(f.FQN), f.Builtin, f.Return, f.DocReturn
		}
	case *syntax.MethodCall:
		if id, ok := n.Name.(*syntax.Identifier); ok {
			classes := e.types.TypeOf(n.Var).Classes()
			if len(classes) == 1 {
				if m := e.types.Index.FindMethod(classes[0], id.Value, e.types.PHP); m != nil && m.Avail.In(e.types.PHP) {
					return strings.ToLower(id.Value), m.Builtin, m.Return, m.DocReturn
				}
			}
			return strings.ToLower(id.Value), false, "", ""
		}
	case *syntax.StaticCall:
		if id, ok := n.Name.(*syntax.Identifier); ok {
			if class, ok := n.Class.(*syntax.Name); ok {
				if m := e.types.Index.FindMethod(e.named(class), id.Value, e.types.PHP); m != nil && m.Avail.In(e.types.PHP) {
					return strings.ToLower(id.Value), m.Builtin, m.Return, m.DocReturn
				}
			}
			return strings.ToLower(id.Value), false, "", ""
		}
	}
	return "", false, "", ""
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
	v = Value{Complete: v.Complete, Sources: v.Sources, Safe: v.Safe, matched: v.matched && n.Op.Kind == syntax.TConcatEqual}
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
	// A property written, or an element of one, leaves the variable the
	// value it is where PHP refuses to make an object for it.
	written := syntax.UnwrapParens(target)
	for {
		element, ok := written.(*syntax.ArrayDimFetch)
		if !ok {
			break
		}
		written = syntax.UnwrapParens(element.Var)
	}
	_, member := written.(*syntax.PropertyFetch)
	untest(target, f, r.modern && member)
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
			f.write(r, name, unsettled(f.vars[name], Value{}, r))
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
	// The request tables are entries of the globals table in every scope.
	f.unmatchAll(r)
	name := assigned(target)
	switch {
	case name == "":
		if r.global {
			f.unsettle(r, v)
		} else {
			r.operations += len(r.globalNames)
			for bound := range r.globalNames {
				f.write(r, bound, unsettled(f.vars[bound], v, r))
			}
		}
	case !r.global && !r.globalNames[name]:
	case definite:
		f.write(r, name, v)
	default:
		f.write(r, name, unsettled(f.vars[name], v, r))
	}
}

// reference handles `target = &source`.
func (e *Env) reference(n *syntax.Assign, v Value, f *frame, r *scopeResult) {
	untest(n.Value, f, false)
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
		if dynamic, ok := syntax.UnwrapParens(n.Value).(*syntax.Variable); ok && dynamic.NameExpr != nil {
			// The name stands for a variable chosen at run time.
			f.wild = true
		}
		f.write(r, target.Name, unsettled(Value{}, v, r))
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
			f.vars[name] = unsettled(f.vars[name], v, r)
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
		referenceTarget(n.Var, f)
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
		old, tracked := f.vars[name]
		// A table of the request, or a variable nothing assigned, is no object.
		held := old.plain || external(name) || !tracked && f.definite && importable(name)
		if !held {
			// An offset of an object is written by a method of its class.
			e.implicit(n, f, r)
		}
		array := unsettled(old, v, r)
		array.plain = held
		array.deep = held && v.deep && (old.deep || !tracked && (external(name) || f.definite && importable(name)))
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
				own = unsettled(previous, v, r)
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
		if truth, known := constantTruth(n.Cond); known {
			chosen := n.Else
			if truth {
				chosen = n.Then
				if chosen == nil {
					chosen = n.Cond
				}
			}
			return e.elements(chosen, f, r)
		}
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

// edits reports a PHP function whose result is its text arguments, changed
// in case, trimmed, cut, padded, replaced, formatted or escaped.
func edits(name string) bool {
	switch name {
	case "strtolower", "strtoupper", "ucfirst", "ucwords", "lcfirst", "trim", "ltrim", "rtrim", "substr", "str_pad", "str_replace", "sprintf", "strval",
		"htmlspecialchars", "htmlentities":
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
		f.write(r, names[i], r.values[item.Value])
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
	if dynamic, ok := syntax.UnwrapParens(target).(*syntax.Variable); ok && dynamic.NameExpr != nil {
		// Any variable, tracked or not, may be the one written.
		f.unsettle(r, v)
		f.importing(v)
		return
	}
	if name := rootVariable(target); name != "" {
		// The names bound with the target read what it was given.
		f.write(r, name, unsettled(f.vars[name], v, r))
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

// calm reports an expression that cannot end the function by throwing, its
// operands aside: a variable read, a constant, a plain assignment to a
// variable, an identity test or a boolean combination, an array literal, a
// choice between two values.
func calm(x syntax.Expr) bool {
	switch n := x.(type) {
	case *syntax.Literal, *syntax.ConstFetch, *syntax.Name, *syntax.Identifier, *syntax.Paren, *syntax.Array, *syntax.ArrayItem, *syntax.Ternary:
		return true
	case *syntax.Variable:
		return n.NameExpr == nil
	case *syntax.Assign:
		return n.Op.Kind == syntax.TEqual && !n.ByRef && plainVariable(n.Var) != ""
	case *syntax.Unary:
		return n.Op.Kind == syntax.TExclaim
	case *syntax.Binary:
		switch n.Op.Kind {
		case syntax.TIsIdentical, syntax.TIsNotIdentical, syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr:
			return true
		}
	}
	return false
}

// visible reports that a method is the one a call written here runs: a
// method that is not public is only reached from its own class or, when it
// is protected, from a class that extends it; anywhere else the call goes to
// `__call` or `__callStatic`, or fails. Each class on the way up from the
// one the call is written in must be declared once, with its parent.
func (e *Env) visible(call syntax.Node, m *index.Method) bool {
	within := strings.TrimPrefix(e.types.ClassFQN(syntax.EnclosingClass(call)), `\`)
	declaring := strings.TrimPrefix(m.Class, `\`)
	if m.Visibility == index.Public || strings.EqualFold(within, declaring) {
		return true
	}
	ix := e.types.Index
	for steps := 0; m.Visibility == index.Protected && steps < MaxDepth*4; steps++ {
		class := ix.Class(within, e.types.PHP)
		if class == nil || ix.ClassCount(within) != 1 || class.Parent == "" {
			break
		}
		if within = strings.TrimPrefix(class.Parent, `\`); strings.EqualFold(within, declaring) {
			return ix.ClassCount(within) == 1
		}
	}
	return false
}

// leaving checks, where the function can be entered again or ends, what is
// assumed of its static variables: each is not false once the function
// wrote it, and nothing else is bound to it. One the function did not write
// holds what it started with or what an invocation that returned left: it
// is left as it was found.
func leaving(f *frame, r *scopeResult) {
	for name, known := range r.statics {
		if known.state == staticHeld {
			r.holds(f, name)
		}
	}
}

// holds checks what is assumed of one static variable.
func (r *scopeResult) holds(f *frame, name string) {
	v, tracked := f.vars[name]
	if _, shared := f.aliased[name]; shared || tracked && !v.NonFalse && !v.entry {
		if r.broken == nil {
			r.broken = map[string]bool{}
		}
		r.broken[name] = true
	}
}

// idle reports a scope for which code that runs without a call written
// changes nothing more: it is known to run some, holds no static variable
// and matched no request table element.
func (r *scopeResult) idle() bool {
	return r.runs && len(r.statics) == 0 && !r.matching
}

// implicit applies code of the project that runs without a call written: a
// magic or hooked property, an offset of an object, a class that loads on
// first use, an object converted to text, compared with some or iterated.
// Such code can rewrite the request tables and enter the function again.
// Copying a value, testing its identity or its truth and computing with it
// run none.
func (e *Env) implicit(at syntax.Node, f *frame, r *scopeResult) {
	f.unmatchAll(r)
	reenter(at, f, r)
}

// reenter applies a call, or code that runs without one (see implicit), that
// may enter the function again to its static variables: an invocation
// assigns the ones the function assigns. What a
// guard established of them is gone, but that they are not false where
// that is assumed and checked (see leaving). The producer stays, as the
// value such a call leaves is one the function itself assigned; the value
// is marked as possibly another one.
func reenter(at syntax.Node, f *frame, r *scopeResult) {
	r.runs = true
	if len(r.statics) > 0 {
		f.related = nil
	}
	for name, known := range r.statics {
		v, tracked := f.vars[name]
		if !tracked || known.state == staticKept {
			continue
		}
		held := known.state == staticHeld && v.NonFalse
		v.NonFalse, v.NonNull, v.origins = held, false, v.origins.unguarded(r).excluding(held, false, r)
		v.tested, v.unless, v.unlessFalse, v.unlessNull, v.skippedFalse, v.skippedNull = nil, nil, false, false, false, false
		v.Invalidated = at.Span().Start + 1
		v.plain, v.deep = v.plain && known.plain, v.deep && known.deep
		f.vars[name] = v
	}
}

// quiet reports a call whose callee is exactly known and runs no code of the
// project but its own. Binding the arguments runs none either: no object is
// converted to text by a `string` parameter, no parameter looks a callable
// up, and no default value loads a class or makes an object.
func (e *Env) quiet(call syntax.Node, list *syntax.ArgList, values []Value) bool {
	summary, ok := e.contract(e.summaryName(call))
	if !ok || !summary.Quiet {
		return false
	}
	params, _, _ := e.declaration(call)
	var few [8]int
	positions := few[:0]
	if list != nil {
		positions = parameters(list, params, positions)
	}
	for at, p := range params {
		text, callable := coerced(p.Type)
		if callable || strings.Contains(p.Default, "::") || strings.Contains(strings.ToLower(p.Default), "new") {
			return false
		}
		for i, bound := range positions {
			if text && bound == at && !values[i].plain {
				return false
			}
		}
	}
	switch n := call.(type) {
	case *syntax.FuncCall:
		return true
	case *syntax.StaticCall:
		class, named := n.Class.(*syntax.Name)
		m := e.method(call)
		return named && m != nil && !strings.EqualFold(class.Value, "static") && e.visible(call, m)
	}
	m := e.method(call)
	return m != nil && (m.Final || m.Visibility == index.Private) && e.visible(call, m)
}

// coerced reports a declared parameter type that makes PHP run code of the
// project to bind an argument: `string` calls the `__toString` method of an
// object, `callable` loads the class a callable names.
func coerced(declared string) (text, callable bool) {
	declared = strings.TrimPrefix(declared, "?")
	for declared != "" {
		var part string
		part, declared, _ = strings.Cut(declared, "|")
		switch strings.ToLower(part) {
		case "string":
			text = true
		case "callable":
			callable = true
		}
	}
	return text, callable
}

// What a declared type admits, as far as running code of the project goes.
const (
	// kindKnown: the type was read.
	kindKnown uint8 = 1 << iota
	// kindText: it admits text, which an object is converted to.
	kindText
	// kindObject: it admits objects.
	kindObject
	// kindOpaque: it admits anything, or a callable.
	kindOpaque
	// kindArray: it admits arrays, which may hold objects.
	kindArray
)

// kinds reads a declared type, or the documented one where PHP declares
// none. Numbers, booleans, null, arrays and resources are none of the kinds.
func (e *Env) kinds(declared, documented string) uint8 {
	if declared == "" {
		declared = documented
	}
	if kind, read := e.kindOf[declared]; read {
		return kind
	}
	var kind uint8
	for _, atom := range types.FromDoc(declared, nil).Atoms() {
		kind |= kindKnown
		switch atom {
		case "int", "float", "bool", "true", "false", "null", "void", "never", "resource":
		case "string":
			kind |= kindText
		case "array":
			kind |= kindArray
		case "mixed", "callable":
			kind |= kindOpaque
		default:
			if strings.HasSuffix(atom, "[]") {
				kind |= kindArray
			} else {
				kind |= kindObject
			}
		}
	}
	if e.kindOf == nil {
		e.kindOf = map[string]uint8{}
	}
	e.kindOf[declared] = kind
	return kind
}

// handed reports a call of PHP itself, or the construction of one of its
// classes, that may run code of the project with what it is given: an
// object where it converts to text, counts, iterates, serializes or
// compares, a class it loads, an error it raises, a generator or a fiber it
// resumes, a method a class of the project replaces, a handler or a class
// it is given the name of (see runner, loads, invokes and builds). A
// constructor only keeps the objects it is given, and a type test only looks
// at its argument.
func (e *Env) handed(call syntax.Node, name string, list *syntax.ArgList, values []Value) bool {
	params, qualified, _ := e.declaration(call)
	_, constructed := call.(*syntax.New)
	if _, function := call.(*syntax.FuncCall); function {
		if runner(name) || loads(name, list) {
			return true
		}
		switch name {
		case "is_array", "is_string", "is_int", "is_integer", "is_long", "is_float", "is_double", "is_bool", "is_null", "is_numeric", "is_scalar",
			"is_object", "is_resource", "is_iterable", "is_countable", "gettype", "get_debug_type", "get_class", "spl_object_id", "spl_object_hash",
			"array_key_exists", "key_exists":
			return false
		case "array_shift", "array_pop":
			// The element taken out is destroyed when nothing keeps it.
			if _, dropped := call.Parent().(*syntax.ExprStmt); dropped && len(values) > 0 && !values[0].deep {
				return true
			}
		}
	} else if m := e.method(call); !constructed && m != nil && (m.Abstract || !m.Final && e.replaced(m.Class, m)) {
		return true
	}
	class, _, _ := strings.Cut(qualified, "::")
	if strings.HasPrefix(qualified, "generator::") || strings.HasPrefix(qualified, "fiber::") || !constructed && invokes(qualified, list) || constructed && builds(class, list) {
		return true
	}
	if list == nil {
		return false
	}
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		kind := uint8(0)
		if at >= 0 {
			kind = e.kinds(params[at].Type, params[at].DocType)
		}
		switch given := values[i]; {
		case given.deep:
		case kind&kindKnown == 0 || kind&kindOpaque != 0:
			return true
		case !given.plain && (kind&kindText != 0 || kind&kindObject != 0 && !constructed):
			return true
		case kind&kindArray != 0 && !constructed && !structural(name, len(values)):
			// The objects an array holds are converted or compared.
			return true
		}
	}
	return false
}

// structural reports a PHP function that counts, moves or picks the
// elements of an array without looking at what they are, called with that
// many arguments.
func structural(name string, given int) bool {
	switch name {
	case "array_keys":
		// With a value to search for, each element is compared with it.
		return given == 1
	case "count", "sizeof", "array_values", "array_key_first", "array_key_last", "array_merge", "array_slice", "array_reverse",
		"array_pop", "array_shift", "array_push", "array_unshift", "array_chunk", "array_is_list", "end", "reset", "current", "key", "next", "prev":
		return true
	}
	return false
}

// sealed reports a PHP function whose result holds no object: it makes none
// and is given none, in an argument that is an array or may be anything, nor
// a function to call, which can make some.
// The same holds of what it leaves in a reference argument.
func (e *Env) sealed(call syntax.Node, name string, list *syntax.ArgList, r *scopeResult) bool {
	if _, function := call.(*syntax.FuncCall); !function {
		return false
	}
	switch name {
	case "func_get_args", "func_get_arg", "get_defined_vars", "debug_backtrace", "compact", "get_object_vars", "get_class_vars":
		return false
	}
	params, _, _ := e.declaration(call)
	var few [8]int
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		if !ok || at < 0 || invoked(params[at]) && constantName(arg.Value) != "null" && !e.harmless(arg.Value) {
			return false
		}
		if kind := e.kinds(params[at].Type, params[at].DocType); !params[at].ByRef && !r.values[arg.Value].deep && (kind&kindKnown == 0 || kind&(kindObject|kindOpaque|kindArray) != 0) {
			return false
		}
	}
	return true
}

// className reports `X::class`: the name of a class, as text.
func className(n *syntax.ClassConstFetch) bool {
	name, ok := n.Name.(*syntax.Identifier)
	return ok && strings.EqualFold(name.Value, "class")
}

// converted reports a binary operation that calls the `__toString` method
// of an operand that is an object: a concatenation, or a loose comparison
// with a value that may be text.
func converted(n *syntax.Binary, left, right Value) bool {
	switch n.Op.Kind {
	case syntax.TDot:
		return !left.plain || !right.plain
	case syntax.TIsEqual, syntax.TIsNotEqual, syntax.TLess, syntax.TGreater, syntax.TIsSmallerOrEqual, syntax.TIsGreaterOrEqual, syntax.TSpaceship:
		return !left.plain && !textless(n.Right) || !right.plain && !textless(n.Left)
	}
	return false
}

// textless reports an operand written as a number, a boolean or null:
// comparing an object with it converts nothing to text.
func textless(x syntax.Expr) bool {
	x = syntax.UnwrapParens(x)
	if sign, ok := x.(*syntax.Unary); ok && (sign.Op.Kind == syntax.TMinus || sign.Op.Kind == syntax.TPlus) {
		x = syntax.UnwrapParens(sign.Expr)
	}
	literal, ok := x.(*syntax.Literal)
	return constantName(x) != "" || ok && literal.LitKind != syntax.LitString
}

// returnsText reports a function whose declared return type admits text:
// an object returned is converted to it.
func returnsText(scope syntax.Node) bool {
	var declared syntax.Expr
	switch n := scope.(type) {
	case *syntax.Function:
		declared = n.ReturnType
	case *syntax.Method:
		declared = n.ReturnType
	case *syntax.Closure:
		declared = n.ReturnType
	}
	return types.FromNode(declared, nil).Has("string")
}

// propertyTarget reports a write target that is a property, an element of
// one, or a destructuring with such a slot.
func propertyTarget(target syntax.Expr) bool {
	switch n := syntax.UnwrapParens(target).(type) {
	case *syntax.PropertyFetch, *syntax.StaticPropertyFetch:
		return true
	case *syntax.ArrayDimFetch:
		return propertyTarget(n.Var)
	case *syntax.Array, *syntax.List:
		for _, item := range targetItems(target) {
			if item != nil && propertyTarget(item.Value) {
				return true
			}
		}
	}
	return false
}

// declared returns the static variables of a scope, each with what the code
// of the scope does to it, read once however often the scope is analysed.
//
// state is staticLoose for a variable the scope may write, itself or one of
// its elements: it is the target of an assignment or a step, an argument (a
// callee may take it by reference), a slot of an array or of a
// destructuring, or named by `unset`, `catch` or `foreach`. Only an
// invocation that does so changes a static variable for the others.
//
// starts tells that its first value is not false: each of its declarations
// is written without an initializer, which starts it null, or with a
// literal, a signed number, an array, `null` or `true`.
//
// plain tells that it never holds an object: it starts from none, and every
// write the scope makes to it gives it a value that is none. Reading it then
// runs no method, whichever invocation wrote it last. deep tells that it
// never holds an array with an object in it either.
func (e *Env) declared(scope syntax.Node) map[string]static {
	if known, read := e.statics[scope]; read {
		return known
	}
	body := syntax.VariableScopeBody(scope)
	known := map[string]static{}
	syntax.Inspect(body, func(n syntax.Node) bool {
		if declared, ok := n.(*syntax.StaticVar); ok {
			name := declared.Var.Name
			fact, again := known[name]
			if !again {
				fact = static{starts: true, plain: true, deep: true}
			}
			initial := syntax.UnwrapParens(declared.Default)
			if sign, signed := initial.(*syntax.Unary); signed && (sign.Op.Kind == syntax.TMinus || sign.Op.Kind == syntax.TPlus) {
				initial = syntax.UnwrapParens(sign.Expr)
			}
			switch initial := initial.(type) {
			case nil, *syntax.Literal, *syntax.Array:
			case *syntax.ConstFetch:
				fact.starts = fact.starts && (constantName(initial) == "null" || constantName(initial) == "true")
			default:
				fact.starts = false
			}
			fact.plain, fact.deep = fact.plain && e.inert(declared.Default, name, false), fact.deep && e.inert(declared.Default, name, true)
			known[name] = fact
		}
		return true
	})
	syntax.Inspect(body, func(n syntax.Node) bool {
		v, ok := n.(*syntax.Variable)
		if !ok || v.NameExpr != nil {
			return true
		}
		fact, static := known[v.Name]
		if !static {
			return true
		}
		var written syntax.Expr = v
		parent := v.Parent()
		for {
			element, ok := parent.(*syntax.ArrayDimFetch)
			if !ok || element.Var != written {
				break
			}
			written, parent = element, element.Parent()
		}
		switch p := parent.(type) {
		case *syntax.Assign:
			switch {
			case p.ByRef:
				fact.state, fact.plain = staticLoose, false
			case p.Var != written:
			case p.Op.Kind != syntax.TEqual && p.Op.Kind != syntax.TCoalesceEqual && p.Op.Kind != syntax.TPlusEqual:
				fact.state = staticLoose
			case written == syntax.Expr(v):
				fact.state = staticLoose
				fact.plain, fact.deep = fact.plain && e.inert(p.Value, v.Name, false), fact.deep && e.inert(p.Value, v.Name, true)
			default:
				// An element is written: the variable stays an array.
				fact.state, fact.deep = staticLoose, fact.deep && e.inert(p.Value, v.Name, true)
			}
		case *syntax.IncDec:
			fact.state = staticLoose
		case *syntax.Arg:
			fact.state, fact.plain = staticLoose, fact.plain && e.byValue(p)
		case *syntax.ArrayItem:
			fact.state, fact.plain = staticLoose, fact.plain && !p.ByRef && !destructured(p)
		case *syntax.Foreach:
			fact.state, fact.plain = staticLoose, fact.plain && p.Expr == written
		case *syntax.ClosureUse:
			fact.state, fact.plain = staticLoose, fact.plain && !p.ByRef
		case *syntax.Unset, *syntax.Catch:
			fact.state, fact.plain = staticLoose, false
		case *syntax.Global:
			fact.plain = false
		}
		known[v.Name] = fact
		return true
	})
	// A write through a name chosen at run time, an import or code evaluated
	// in the scope may give any of them any value.
	open := len(known) > 0 && e.imports(body)
	for name, fact := range known {
		if open {
			fact.state, fact.plain = staticLoose, false
		}
		fact.deep = fact.deep && fact.plain
		known[name] = fact
	}
	if e.statics == nil {
		e.statics = map[syntax.Node]map[string]static{}
	}
	e.statics[scope] = known
	return known
}

// inert reports an expression that yields no object, whatever the variables
// hold but the one named: a literal, an array, a computation, a test, a
// function, what a PHP function declares to return, or a choice of those.
// With deep, an array it yields holds no object either.
func (e *Env) inert(x syntax.Expr, name string, deep bool) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case nil, *syntax.Literal, *syntax.InterpolatedString, *syntax.MagicConst, *syntax.Isset, *syntax.Empty, *syntax.Instanceof, *syntax.Closure, *syntax.ArrowFunction:
		return true
	case *syntax.ConstFetch:
		return e.fixed(n)
	case *syntax.IncDec:
		// Before PHP 8.3 a step leaves an object as it is.
		return e.types.PHP.AtLeast(phpversion.PHP83) || e.inert(n.Var, name, deep)
	case *syntax.Array:
		for _, item := range n.Items {
			if deep && item != nil && (item.ByRef || !e.inert(item.Value, name, true)) {
				return false
			}
		}
		return true
	case *syntax.Variable:
		return n.NameExpr == nil && n.Name == name
	case *syntax.Binary:
		if n.Op.Kind == syntax.TCoalesce || deep && n.Op.Kind == syntax.TPlus {
			return e.inert(n.Left, name, deep) && e.inert(n.Right, name, deep)
		}
		return true
	case *syntax.Unary:
		if n.Op.Kind == syntax.TAt || deep && n.Op.Kind == syntax.TArrayCast {
			return e.inert(n.Expr, name, deep)
		}
		return n.Op.Kind != syntax.TObjectCast
	case *syntax.Ternary:
		return (n.Then != nil || e.inert(n.Cond, name, deep)) && e.inert(n.Then, name, deep) && e.inert(n.Else, name, deep)
	case *syntax.Assign:
		switch n.Op.Kind {
		case syntax.TEqual:
			return !n.ByRef && e.inert(n.Value, name, deep)
		case syntax.TCoalesceEqual, syntax.TPlusEqual:
			return false
		}
		return true
	case *syntax.FuncCall:
		fname, builtin, result, documented := e.callee(n)
		kind := e.kinds(result, documented)
		if !builtin || kind&(kindKnown|kindObject|kindOpaque) != kindKnown {
			return false
		}
		return !deep || kind&kindArray == 0 || e.given(n, fname, name)
	}
	return false
}

// given is sealed for a call read without its values: each argument is
// written as an expression that yields no object and holds none.
func (e *Env) given(call *syntax.FuncCall, fname, name string) bool {
	switch fname {
	case "func_get_args", "func_get_arg", "get_defined_vars", "debug_backtrace", "compact", "get_object_vars", "get_class_vars":
		return false
	}
	params, _, _ := e.declaration(call)
	var few [8]int
	for i, at := range parameters(call.Args, params, few[:0]) {
		arg, ok := call.Args.Args[i].(*syntax.Arg)
		if !ok || at < 0 || arg.Unpack || invoked(params[at]) && !e.harmless(arg.Value) {
			return false
		}
		if kind := e.kinds(params[at].Type, params[at].DocType); !params[at].ByRef && !e.inert(arg.Value, name, true) && (kind&kindKnown == 0 || kind&(kindObject|kindOpaque|kindArray) != 0) {
			return false
		}
	}
	return true
}

// byValue reports an argument bound to a parameter that is known not to be
// a reference: the callee cannot write the variable given.
func (e *Env) byValue(arg *syntax.Arg) bool {
	list, ok := arg.Parent().(*syntax.ArgList)
	if !ok || arg.ByRef || arg.Unpack {
		return false
	}
	params, _, known := e.declaration(list.Parent())
	var few [8]int
	at := parameters(list, params, few[:0])[slices.Index(list.Args, syntax.Expr(arg))]
	return known && at >= 0 && !params[at].ByRef
}

// destructured reports an array item that is a slot of a destructuring: the
// target of an assignment or of a loop.
func destructured(item *syntax.ArrayItem) bool {
	for {
		array := item.Parent()
		if _, list := array.(*syntax.List); list {
			return true
		}
		switch p := array.Parent().(type) {
		case *syntax.ArrayItem:
			item = p
			continue
		case *syntax.Assign:
			return p.Var == array
		case *syntax.Foreach:
			return p.Value == array || p.Key == array
		}
		return false
	}
}

// invalidateShared discards the bindings code outside the scope can replace
// while a callee with unknown effects, or a callback, runs.
func (e *Env) invalidateShared(f *frame, r *scopeResult) {
	if r.global && e.globalsTable() {
		f.related = nil
		r.operations += len(f.vars)
		for name, old := range f.vars {
			f.vars[name] = unsettled(old, e.held(name, r), r)
		}
		return
	}
	// What a holder writes to one name is read under the names bound with it.
	if len(f.aliased) > 0 {
		f.related = nil
	}
	r.operations += len(f.aliased)
	var groups map[uint32]Value
	for name, group := range f.aliased {
		if written := e.held(name, r); group >= aliasGroup && retained(written) {
			if groups == nil {
				groups = map[uint32]Value{}
			}
			groups[group] = effects(groups[group], written, r)
		}
	}
	for name, group := range f.aliased {
		written, shared := groups[group]
		if !shared {
			written = e.held(name, r)
		}
		f.vars[name] = unsettled(f.vars[name], written, r)
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
			written = effects(written, other, r)
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
	left, captured := r.exit[received(name)]
	if !captured {
		left = r.exit[name]
	}
	for _, v := range []Value{left, r.globalWrites[name]} {
		if retained(v) {
			written = effects(written, v, r)
		}
	}
	written.origins = r.producers(written)
	written.Expr = nil
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
	r.globalWrites[name] = effects(r.globalWrites[name], v, r)
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

// callbackParameter reports a resolved builtin given a function it may
// call. None is given where the parameter is left out. A PHP function named
// by a literal, that only computes a number, a boolean or text, runs no code
// of the project when no argument of the call can hold an object for it to
// convert.
func (e *Env) callbackParameter(n syntax.Node, list *syntax.ArgList, values []Value) bool {
	params, _, _ := e.declaration(n)
	if list == nil || !slices.ContainsFunc(params, func(p index.Param) bool { return strings.Contains(p.Type, "callable") }) {
		return false
	}
	var few [8]int
	given, converts := false, false
	for i, at := range parameters(list, params, few[:0]) {
		arg, ok := list.Args[i].(*syntax.Arg)
		switch {
		case !ok || at < 0 || arg.Unpack || i >= len(values):
			return true
		case !strings.Contains(params[at].Type, "callable"):
			converts = converts || !values[i].deep
		case constantName(arg.Value) == "null":
		case !e.harmless(arg.Value):
			return true
		default:
			given = true
		}
	}
	return given && converts
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

// imports reports a body that can write a variable it does not name: through
// a name chosen at run time, `extract`, `parse_str` filling the scope, or
// code it evaluates or includes.
func (e *Env) imports(body syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Variable:
			found = found || n.NameExpr != nil
		case *syntax.Eval, *syntax.Include:
			found = true
		case *syntax.FuncCall:
			name, builtin := e.callName(n)
			found = found || builtin && (name == "extract" || name == "parse_str" && n.Args != nil && len(n.Args.Args) == 1)
		}
		return !found
	})
	return found
}

// invoked reports a parameter of a PHP function that takes a function to
// call.
func invoked(p index.Param) bool {
	return strings.Contains(strings.ToLower(p.Type), "callable") || strings.Contains(strings.ToLower(p.DocType), "callable")
}

// harmless reports a function to call, named by a string literal, that is a
// PHP function which only computes on text and numbers (see pure): it makes
// no object, hands none on, and runs no code of the project.
func (e *Env) harmless(callback syntax.Expr) bool {
	name, known := literalText(callback)
	if !known {
		return false
	}
	name = strings.ToLower(strings.TrimPrefix(name, `\`))
	f := e.types.Index.Function(name, e.types.PHP)
	return f != nil && f.Builtin && pure(name)
}

// nested reports a destructuring target with a slot that is one itself: the
// element it takes apart may be an object.
func nested(target syntax.Expr) bool {
	for _, item := range targetItems(target) {
		if item != nil && len(targetItems(item.Value)) > 0 {
			return true
		}
	}
	return false
}

// release applies the destructor that may run when a variable is given
// another value: what it held may be an object, or hold one, which nothing
// else refers to.
func (e *Env) release(at syntax.Node, name string, f *frame, r *scopeResult) {
	held, tracked := f.vars[name]
	if tracked && !held.deep || !tracked && !external(name) && !(f.definite && importable(name)) {
		e.implicit(at, f, r)
	}
}

// ends reports an expression that never completes: a throw, an exit or a
// call that does not return.
func (e *Env) ends(x syntax.Expr) bool {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Throw, *syntax.Exit:
		return true
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		return e.never(n)
	}
	return false
}

// fixed reports a constant that holds no object: `true`, `false` or `null`,
// one PHP itself defines, or any constant before PHP 8.1, since when the
// project can define one as an object.
func (e *Env) fixed(c *syntax.ConstFetch) bool {
	if constantName(c) != "" || !e.types.PHP.AtLeast(phpversion.PHP81) || c.Name == nil {
		return true
	}
	fqn, fallback := e.types.Names.Const(c.Name.Value, c.Span().Start)
	known := e.types.Index.Constant(fqn, e.types.PHP)
	if known == nil && fallback != "" {
		known = e.types.Index.Constant(fallback, e.types.PHP)
	}
	return known != nil && known.Builtin
}

// selfContained reports a class constant whose value names no other class:
// it is written without `::`, or only with `self::` before constants of the
// declaring class that are so themselves, a bounded number of steps deep.
func (e *Env) selfContained(constant *index.ClassConst, depth int) bool {
	if constant.Case || constant.Value == "" || depth == 0 {
		return false
	}
	named := func(c byte) bool {
		return c == '_' || c == '\\' || c >= '0' && c <= '9' || c|0x20 >= 'a' && c|0x20 <= 'z' || c >= 0x80
	}
	for rest := constant.Value; ; {
		at := strings.Index(rest, "::")
		if at < 0 {
			return true
		}
		from, to := at, at+2
		for from > 0 && named(rest[from-1]) {
			from--
		}
		for to < len(rest) && named(rest[to]) {
			to++
		}
		other := e.types.Index.FindConst(constant.Class, rest[at+2:to], e.types.PHP)
		if !strings.EqualFold(rest[from:at], "self") || other == nil || !strings.EqualFold(other.Class, constant.Class) || !e.selfContained(other, depth-1) {
			return false
		}
		rest = rest[to:]
	}
}

// rooted reports a class of the project whose ancestors are all known: each
// class up to the top is declared once, and so is each trait they use.
func (e *Env) rooted(fqn string) bool {
	ix := e.types.Index
	for steps := 0; steps < MaxDepth*4; steps++ {
		class := ix.Class(fqn, e.types.PHP)
		if class == nil || ix.ClassCount(fqn) != 1 {
			return false
		}
		for _, trait := range class.Traits {
			if ix.ClassCount(trait) != 1 {
				return false
			}
		}
		if fqn = class.Parent; fqn == "" {
			return true
		}
	}
	return false
}

// backed reports a property read that runs no code: `$this->name` in a
// method of a class that declares the property itself, without a hook, where
// no `__get` method can stand in for it. A class extending this one can only
// make the read run code by declaring `__get` (for a property removed with
// `unset`) or, unless the property is private, by declaring it again with a
// hook: the class is final, or no class of the project extending it has such
// a method and the property is private. An object PHP 8.4 made lazy runs its
// initializer on the first read: that is not followed.
func (e *Env) backed(n *syntax.PropertyFetch) bool {
	this, direct := n.Var.(*syntax.Variable)
	name, written := n.Name.(*syntax.Identifier)
	_, method := syntax.EnclosingVariableScope(n).(*syntax.Method)
	within := syntax.EnclosingClass(n)
	if !direct || this.Name != "this" || this.NameExpr != nil || !written || !method || within == nil || within.ClassKind != syntax.KindClass {
		return false
	}
	at := [2]string{e.types.ClassFQN(within), name.Value}
	known, read := e.properties[at]
	if !read {
		if e.properties == nil {
			e.properties = map[[2]string]bool{}
		}
		known = e.declares(at[0], at[1])
		e.properties[at] = known
	}
	return known
}

// declares is backed for a class and the name of a property.
func (e *Env) declares(fqn, name string) bool {
	ix := e.types.Index
	property := ix.FindProperty(fqn, name, e.types.PHP)
	if property == nil || property.Magic || property.Static || property.Hooked || property.ReadsRunCode || !strings.EqualFold(strings.TrimPrefix(property.Class, `\`), strings.TrimPrefix(fqn, `\`)) {
		return false
	}
	if !e.rooted(fqn) || ix.FindMethod(fqn, "__get", e.types.PHP) != nil {
		return false
	}
	return ix.Class(fqn, e.types.PHP).Final || property.Visibility == index.Private && !e.overrides(fqn, "__get", "")
}

// loaded reports a class constant read that loads no class: a constant of
// the enclosing class, named `self`, `parent` or by its own name, or
// `static` when no class can declare it again, whose value is written
// without a constant of another class (see selfContained). An enum case may
// name one.
func (e *Env) loaded(n *syntax.ClassConstFetch) bool {
	class, named := n.Class.(*syntax.Name)
	name, written := n.Name.(*syntax.Identifier)
	within := syntax.EnclosingClass(n)
	if !named || !written || within == nil || within.ClassKind == syntax.KindEnum {
		return false
	}
	own := e.types.ClassFQN(within)
	fqn := e.named(class)
	switch strings.ToLower(class.Value) {
	case "self", "parent", "static":
	default:
		if !strings.EqualFold(strings.TrimPrefix(fqn, `\`), strings.TrimPrefix(own, `\`)) {
			return false
		}
	}
	constant := e.types.Index.FindConst(fqn, name.Value, e.types.PHP)
	if constant == nil || !e.selfContained(constant, MaxDepth) || !e.rooted(own) {
		return false
	}
	if strings.EqualFold(class.Value, "static") {
		declared := e.types.Index.Class(own, e.types.PHP)
		return constant.Final || declared.Final
	}
	return true
}
