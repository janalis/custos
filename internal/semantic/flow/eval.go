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
			if constantName(subject) == "" && (!v.plain && !textless(subject) || !r.values.at(subject).plain && !textless(n)) {
				e.implicit(n, f, r)
			}
		}
	case *syntax.Return:
		v := e.eval(n.Expr, f, r, 0)
		if n.Expr == nil {
			v = nullResult()
		}
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
		f.setState(outputID, State{Operation: "echo", Span: n.Span()})
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
				f.setVar(name, unsettled(f.vars.value(name), e.held(name, r), r))
			}
		}
	case *syntax.Unset:
		f.unmatchAll(r)
		leaving(f, r)
		for _, x := range n.Vars {
			e.operands(x, f, r, 0)
			// What is released may have a destructor, and a property may
			// have a magic method that unsets it.
			if name := rootVariable(x); name != "" {
				e.release(n, name, f, r)
			} else {
				e.implicit(n, f, r)
			}
			rebind(x, f, r)
		}
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
			f.setVar(name, Value{NonFalse: held && known.starts, entry: held, plain: known.plain, deep: known.deep})
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
	iterated := r.values.at(n.Expr)
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
			item.Safe |= everyContext
		}
		spread(n.Value, item, f, r)
		if name := plainVariable(n.Value); name != "" && iterated.upload == uploadTable {
			// Each element of the table is one uploaded file.
			entry := f.vars.value(name)
			entry.upload = uploadEntry
			f.setVar(name, entry)
		}
	}
	if name := plainVariable(n.Key); name != "" && iterated.plain {
		// A key of an array is a number or text: neither false nor null, and
		// no object. Anything else that is iterated chooses its keys.
		key := f.vars.value(name)
		key.NonFalse, key.NonNull, key.plain, key.deep = true, true, true, true
		f.setVar(name, key)
	}
	if name := plainVariable(n.Value); name != "" && iterated.deep && !n.ByRef {
		// An array that holds no object yields none.
		item := f.vars.value(name)
		item.plain, item.deep = true, true
		f.setVar(name, item)
	}
}

func remember(r *scopeResult, n syntax.Node, f frame) {
	key := syntax.Identity(n)
	if old, ok := r.before[key]; ok {
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
		r.before[key] = m
	}
}

// eval evaluates an expression on the frame: it charges the budget, hands the
// expression to the handler of its family, then normalizes what it yields by
// kind and records it. The epilogue stays inline: passing the value to a
// helper copies it on every evaluation and measurably slows the hot path.
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
	var v Value
	switch n := x.(type) {
	case *syntax.Variable:
		v = e.evalVariable(n, f, r, depth)
	case *syntax.Paren:
		v = e.eval(n.Expr, f, r, depth+1)
	case *syntax.Assign:
		v = e.evalAssign(n, f, r, depth)
	case *syntax.FuncCall:
		v = e.evalFuncCall(n, f, r, depth)
	case *syntax.MethodCall:
		v = e.evalMethodCall(n, f, r, depth)
	case *syntax.StaticCall:
		v = e.evalStaticCall(n, f, r, depth)
	case *syntax.New:
		v = e.evalNew(n, f, r, depth)
	case *syntax.Arg:
		v = e.evalArg(n, f, r, depth)
	case *syntax.Binary:
		v = e.evalBinary(n, f, r, depth)
	case *syntax.Unary:
		v = e.evalUnary(n, f, r, depth)
	case *syntax.IncDec:
		v = e.evalIncDec(n, f, r, depth)
	case *syntax.Ternary:
		v = e.evalTernary(n, f, r, depth)
	case *syntax.Match:
		v = e.evalMatch(n, f, r, depth)
	case *syntax.ArrayDimFetch:
		v = e.evalArrayDimFetch(n, f, r, depth)
	case *syntax.Array:
		v = e.evalArray(n, f, r, depth)
	case *syntax.PropertyFetch:
		v = e.evalPropertyFetch(n, f, r, depth)
	case *syntax.InterpolatedString:
		v = e.evalInterpolatedString(n, f, r, depth)
	case *syntax.Closure:
		v = e.evalClosure(n, f, r, depth)
	case *syntax.ArrowFunction:
		v = e.evalArrowFunction(n, f, r)
	case *syntax.Yield, *syntax.YieldFrom:
		e.suspend(x, f, r, depth)
		// What the consumer sends back is not known.
		v = Value{}
	case *syntax.Eval, *syntax.Include:
		v = evalDynamicCode(x, f, r)
	default:
		v = e.evalOther(x, f, r, depth)
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
	// Shape completeness describes an array value, not a scalar coercion or
	// predicate computed from that array. Only value-preserving forms retain it.
	if v.literal != nil || v.truth != 0 {
		switch x.(type) {
		case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.ArrayDimFetch, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		default:
			v.literal = nil
			v.truth = 0
		}
	}
	switch x.(type) {
	case *syntax.Binary, *syntax.Unary:
		v = evaluatedScalar(x, v, r)
	}
	if v.elements != nil {
		switch node := x.(type) {
		case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.Array, *syntax.ArrayDimFetch, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		case *syntax.Unary:
			if node.Op.Kind != syntax.TArrayCast && node.Op.Kind != syntax.TAt {
				v.elements = nil
			}
		case *syntax.Binary:
			if node.Op.Kind != syntax.TPlus && node.Op.Kind != syntax.TCoalesce {
				v.elements = nil
			}
		default:
			v.elements = nil
		}
	}
	if v.targets != nil || v.returned != nil {
		switch x.(type) {
		case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.Closure, *syntax.ArrowFunction, *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
		default:
			v.targets = nil
			v.returned = nil
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
	switch n := x.(type) {
	case *syntax.Variable, *syntax.Paren, *syntax.Assign, *syntax.Ternary, *syntax.Match, *syntax.Arg, *syntax.IncDec, *syntax.Binary, *syntax.Unary:
	case *syntax.ConstFetch:
		v.interval = e.constantInt(n)
	case *syntax.ClassConstFetch:
		v.interval = e.classConstantInt(n)
	default:
		v.interval = nil
	}
	if old, ok := r.values.get(x); ok {
		seen := merge(old, v)
		if _, read := x.(*syntax.Variable); read {
			// Every visit of the read adds the producers that reach it.
			seen.origins = origins(old, v, r)
		}
		r.values.set(x, seen)
	} else {
		r.values.set(x, v)
	}
	return v
}

// chainDepth charges nothing for a fluent link: each `->m()` of a chain wraps
// the one before it, so a long builder chain is deep without being nested
// code. Parsing bounds the tree depth, and operations still bound the work.
func chainDepth(receiver syntax.Expr) int {
	switch receiver.(type) {
	case *syntax.MethodCall, *syntax.PropertyFetch:
		return 0
	}
	return 1
}

// plainVariable names a target that is a variable written as such.
func plainVariable(x syntax.Expr) string {
	if v, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok && v.NameExpr == nil {
		return v.Name
	}
	return ""
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
	held, written := f.vars.lookup(name)
	v := Value{
		Complete: true, Expr: at, plain: true, deep: !written || held.deep,
		Sources: []Source{{Kind: name, Path: e.file.Path, Span: at.Span(), Parameter: -1}},
	}
	if name == "_FILES" {
		v.upload = uploadTable
	}
	return v
}

// searchedOnly reports whether argument x, the i-th of a call to builtin
// name, is the text or pattern searched for by a replacing function: none
// of it reaches the result.
func searchedOnly(name string, x syntax.Node, i int) bool {
	var param string
	switch name {
	case "str_replace", "str_ireplace":
		param = "search"
	case "preg_replace":
		param = "pattern"
	default:
		return false
	}
	a, ok := x.(*syntax.Arg)
	if !ok || a.Unpack {
		return false
	}
	if a.Name != nil {
		return a.Name.Value == param
	}
	return i == 0
}

func external(s string) bool {
	switch s {
	case "_GET", "_POST", "_REQUEST", "_COOKIE", "_FILES", "_SERVER":
		return true
	}
	return false
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

// invalidateEscaped separates operation histories across calls with unknown effects.
// escapeSet is what one call let escape: the identities reachable from its
// operands, and whether a nested scope may reach any local.
type escapeSet struct {
	identities []uint32
	all        bool
}

func (e *Env) invalidateEscaped(at syntax.Node, values []Value, f *frame, r *scopeResult) {
	identities := map[uint32]bool{}
	all := false
	for _, value := range values {
		collectRecordIdentities(value.elements, identities)
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
				if v := r.values.at(expr); v.Identity != 0 {
					identities[v.Identity] = true
				}
			}
			// An evaluated call already collected what its operands let
			// escape: its subtree holds nothing more.
			if memo, ok := r.escaped[child]; ok {
				all = all || memo.all
				for _, id := range memo.identities {
					identities[id] = true
				}
				return false
			}
			return true
		})
	}
	if r.escaped == nil {
		r.escaped = map[syntax.Node]escapeSet{}
	}
	memo := escapeSet{all: all}
	if len(identities) > 0 {
		memo.identities = make([]uint32, 0, len(identities))
		for id := range identities {
			memo.identities = append(memo.identities, id)
		}
	}
	r.escaped[at] = memo
	if all {
		for _, valueCell := range f.vars {
			value := valueCell.Value
			collectRecordIdentities(value.elements, identities)
		}
	}
	for id := range identities {
		delete(f.states, id)
	}
	for name, valueCell := range f.vars {
		value := valueCell.Value
		if value.Identity != 0 && (all || identities[value.Identity]) {
			value.Invalidated = at.Span().Start + 1
			f.setVar(name, value)
			delete(f.states, value.Identity)
		}
	}
}

// suspend hands control to the consumer of a generator, which may run any
// code before resuming it: like a call of unknown code given what is yielded.
// Locals keep their values, except storage code outside can reach. A
// generator returning by reference also hands out the variable it yields,
// which the consumer can then write whenever the generator is suspended.
func (e *Env) suspend(x syntax.Expr, f *frame, r *scopeResult, depth int) {
	var yielded []Value
	syntax.Children(x, func(child syntax.Node) {
		if expr, ok := child.(syntax.Expr); ok {
			yielded = append(yielded, e.eval(expr, f, r, depth+1))
		}
	})
	if y, ok := x.(*syntax.Yield); ok && y.Value != nil && byReference(syntax.EnclosingVariableScope(x)) {
		f.alias(rootVariable(y.Value), aliasHeld)
	}
	f.unmatchAll(r)
	leaving(f, r)
	e.invalidateEscaped(x, yielded, f, r)
	reenter(x, f, r)
	e.invalidateShared(x, f, r)
	for id := range f.states {
		if id >= bufferID {
			delete(f.states, id)
		}
	}
}

// byReference reports a function-like that returns by reference.
func byReference(scope syntax.Node) bool {
	switch s := scope.(type) {
	case *syntax.Function:
		return s.ByRef
	case *syntax.Method:
		return s.ByRef
	case *syntax.Closure:
		return s.ByRef
	}
	return false
}

// elements returns the tracked elements of the array a plain assignment
// stores: those of the variable, array literal or choice of either it reads.
// A long literal, or one with a key the analysis cannot name, is read as a
// whole.
func (e *Env) elements(value syntax.Expr, f *frame, r *scopeResult) *elementSet {
	switch n := syntax.UnwrapParens(value).(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall, *syntax.ArrayDimFetch:
		return r.values.at(n).elements
	case *syntax.Variable:
		if n.NameExpr == nil && !external(n.Name) {
			return f.vars.value(n.Name).elements
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
		return r.values.at(n).elements
	}
	return nil
}

func (e *Env) literalElements(n *syntax.Array, r *scopeResult) *elementSet {
	if len(n.Items) == 0 {
		return emptyElements
	}
	if len(n.Items) > MaxSources {
		return nil
	}
	next, counting := int64(0), true
	set := &elementSet{items: make(map[string]Value, len(n.Items)), width: len(n.Items), whole: true}
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
		own := r.values.at(item.Value)
		if _, existed := set.items[key]; !existed {
			set.order = append(set.order, key)
		}
		entry := elementValue(own)
		if own.Complete && !own.replaced && own.Invalidated == 0 {
			entry.storedIdentity = own.Identity
		} else {
			entry.storedIdentity = 0
		}
		set.put(key, entry)
	}
	r.operations += len(set.items)
	return set
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

// uploadField names the entries of an uploaded file the server chooses.
func uploadField(key string) bool {
	return key == "tmp_name" || key == "error" || key == "size"
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
	v, tracked := f.vars.lookup(name)
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
		v, tracked := f.vars.lookup(name)
		if !tracked || known.state == staticKept {
			continue
		}
		held := known.state == staticHeld && v.NonFalse
		v.NonFalse, v.NonNull, v.origins = held, false, v.origins.unguarded(r).excluding(held, false, r)
		v.tested, v.unless, v.unlessFalse, v.unlessNull, v.skippedFalse, v.skippedNull = nil, nil, false, false, false, false
		v.Invalidated = at.Span().Start + 1
		v.plain, v.deep = v.plain && known.plain, v.deep && known.deep
		f.setVar(name, v)
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
		if kind := e.kinds(params[at].Type, params[at].DocType); !params[at].ByRef && !r.values.at(arg.Value).deep && (kind&kindKnown == 0 || kind&(kindObject|kindOpaque|kindArray) != 0) {
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
		e.depend(e.staticReads[scope])
		return known
	}
	reads := e.begin()
	defer e.end(reads)
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
	if reads != nil {
		if e.staticReads == nil {
			e.staticReads = map[syntax.Node]map[string]bool{}
		}
		e.staticReads[scope] = reads
	}
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
func (e *Env) invalidateShared(at syntax.Node, f *frame, r *scopeResult) {
	// What is left there was possibly written by the code at: a guard read
	// before it says nothing of a read after it.
	stamp := at.Span().Start + 1
	if r.global && e.globalsTable() {
		f.related = nil
		r.operations += len(f.vars)
		for name, oldCell := range f.vars {
			old := oldCell.Value
			v := unsettled(old, e.held(name, r), r)
			v.Invalidated, v.replaced = stamp, true
			f.setVar(name, v)
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
		v := unsettled(f.vars.value(name), written, r)
		v.Invalidated, v.replaced = stamp, true
		f.setVar(name, v)
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
	left, captured := r.exit.lookup(received(name))
	if !captured {
		left = r.exit.value(name)
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

// release applies the destructor that may run when a variable is given
// another value: what it held may be an object, or hold one, which nothing
// else refers to.
func (e *Env) release(at syntax.Node, name string, f *frame, r *scopeResult) {
	held, tracked := f.vars.lookup(name)
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
		// A constant defined under a computed name may be the namespaced
		// one, which PHP then reads instead of its own.
		if e.computedDefines() {
			return false
		}
		known = e.types.Index.Constant(fallback, e.types.PHP)
	}
	return known != nil && known.Builtin
}

// firstArg returns the first argument of a call written in order.
func firstArg(list *syntax.ArgList) (*syntax.Arg, bool) {
	if list == nil || len(list.Args) == 0 {
		return nil, false
	}
	arg, plain := list.Args[0].(*syntax.Arg)
	return arg, plain && !arg.Unpack && arg.Name == nil
}

// computedDefinesKey caches syntax only; no semantic option or index state.
type computedDefinesKey struct{}

// computedDefines reports a call of define in the file whose name is not
// literal text.
func (e *Env) computedDefines() bool {
	return e.file.Memo(computedDefinesKey{}, func() any {
		found := false
		syntax.InspectFile(e.file, func(n syntax.Node) bool {
			if found {
				return false
			}
			call, ok := n.(*syntax.FuncCall)
			if name, builtin := e.callName(n); ok && builtin && name == "define" {
				found = true
				if arg, plain := firstArg(call.Args); plain {
					_, literal := literalText(arg.Value)
					found = !literal
				}
			}
			return !found
		})
		return found
	}).(bool)
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
	if known, read := e.properties[at]; read {
		// The answer may have read whether a subclass declares __get.
		e.depend(known.reads)
		return known.holds
	}
	if e.properties == nil {
		e.properties = map[[2]string]cachedAnswer{}
	}
	reads := e.begin()
	holds := e.declares(at[0], at[1])
	e.end(reads)
	e.properties[at] = cachedAnswer{holds: holds, reads: reads}
	return holds
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

// completeSpread accepts only an evaluated literal-array producer. Completeness
// already accounts for nested spreads, references and unknown element values.
func (e *Env) completeSpread(v Value) bool {
	if !v.Complete || v.Expr == nil || e.types.PHP.Below(phpversion.PHP74) {
		return false
	}
	array, ok := syntax.UnwrapParens(v.Expr).(*syntax.Array)
	if !ok {
		return false
	}
	if e.types.PHP.AtLeast(phpversion.PHP81) {
		return true
	}
	// Before 8.1 string keys cannot be unpacked. Restrict this proof to
	// implicit numeric keys; explicit and nested keys remain conservative.
	for _, item := range array.Items {
		if item != nil && (item.Key != nil || item.Unpack) {
			return false
		}
	}
	return true
}
