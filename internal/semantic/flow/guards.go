package flow

import (
	"strconv"
	"strings"

	"custos/internal/php/syntax"
)

func literalText(x syntax.Expr) (string, bool) {
	n, ok := syntax.UnwrapParens(x).(*syntax.Literal)
	if !ok {
		return "", false
	}
	if n.LitKind != syntax.LitString {
		return n.Raw, true
	}
	if len(n.Raw) < 2 {
		return "", false
	}
	if n.Raw[0] == '"' {
		s, err := strconv.Unquote(n.Raw)
		return s, err == nil
	}
	if n.Raw[0] == '\'' {
		return strings.NewReplacer(`\\`, `\`, `\'`, `'`).Replace(n.Raw[1 : len(n.Raw)-1]), true
	}
	return "", false
}

func (e *Env) refine(expr syntax.Expr, f *frame, truth bool) {
	if expr == nil {
		return
	}
	switch n := syntax.UnwrapParens(expr).(type) {
	case *syntax.Variable, *syntax.Assign:
		if truth {
			exclude(n, f, true, true)
		}
	case *syntax.Empty:
		if !truth {
			exclude(n.Expr, f, true, true)
			exclude(indexed(n.Expr), f, true, true)
		}
	case *syntax.Isset:
		if truth {
			for _, x := range n.Vars {
				exclude(x, f, false, true)
				exclude(indexed(x), f, true, true)
			}
		}
	case *syntax.Instanceof:
		if truth {
			exclude(n.Expr, f, true, true)
		}
	case *syntax.Unary:
		if n.Op.Kind == syntax.TExclaim {
			e.refine(n.Expr, f, !truth)
		}
	case *syntax.Binary:
		// A constant-first comparison states the same fact. It only pins the
		// variable to a literal in the written order, as it always did.
		mirrored := constantOperand(n.Left) && !constantOperand(n.Right)
		if mirrored {
			copy := *n
			copy.Left, copy.Right = n.Right, n.Left
			n = &copy
		}
		if (truth && (n.Op.Kind == syntax.TBooleanAnd || n.Op.Kind == syntax.TAnd)) || (!truth && (n.Op.Kind == syntax.TBooleanOr || n.Op.Kind == syntax.TOr)) {
			// The frame already holds what the right operand wrote: a fact
			// about the left operand describes a value that may be gone.
			if !e.rebinds(n.Right, n.Left) {
				e.refine(n.Left, f, truth)
			}
			e.refine(n.Right, f, truth)
		}
		flag := constantName(n.Right)
		if (truth && n.Op.Kind == syntax.TIsIdentical) || (!truth && n.Op.Kind == syntax.TIsNotIdentical) {
			if variable, ok := syntax.UnwrapParens(n.Left).(*syntax.Variable); ok && variable.NameExpr == nil && !mirrored {
				if _, literal := literalText(n.Right); literal {
					f.vars[variable.Name] = Value{Complete: true, Expr: n.Right}
				}
			}
		}
		if (!truth && n.Op.Kind == syntax.TIsIdentical) || (truth && n.Op.Kind == syntax.TIsNotIdentical) {
			exclude(n.Left, f, flag == "false", flag == "null")
		}
		// A value loosely different from false or null is neither of them.
		if ((!truth && n.Op.Kind == syntax.TIsEqual) || (truth && n.Op.Kind == syntax.TIsNotEqual)) && (flag == "false" || flag == "null") {
			exclude(n.Left, f, true, true)
		}
		if (truth && n.Op.Kind == syntax.TIsIdentical) || (!truth && n.Op.Kind == syntax.TIsNotIdentical) {
			if call, ok := n.Left.(*syntax.FuncCall); ok {
				name, builtin := e.callName(call)
				zero, known := literalText(n.Right)
				if args, valid := e.guardArguments(call); valid && builtin && name == "preg_match" && known && zero == "0" && len(args) >= 2 {
					pattern, subject := args[0], args[1]
					text, known := literalText(pattern.Value)
					variable, vok := subject.Value.(*syntax.Variable)
					if known && vok && (text == `/[\r\n]/` || text == `/[\n\r]/`) {
						v := f.vars[variable.Name]
						v.Safe |= Header
						v.headerMask = 3
						f.vars[variable.Name] = v
					}

				}
			}
			if call, ok := n.Left.(*syntax.FuncCall); ok && e.types.ResolveFunction(call) != nil {
				name, builtin := e.callName(call)
				if args, valid := e.guardArguments(call); valid && builtin && name == "strlen" && len(args) == 1 {
					arg := args[0]
					variable, ok := arg.Value.(*syntax.Variable)
					length, known := literalText(n.Right)
					if ok && known {
						if number, err := strconv.ParseInt(length, 10, 64); err == nil {
							v := f.vars[variable.Name]
							v.Length = number
							v.LengthKnown = true
							f.vars[variable.Name] = v
						}
					}

				}
			}
		}
	case *syntax.MethodCall:
		name, builtin := e.callName(n)
		if truth && builtin && name == "begintransaction" {
			if variable, ok := n.Var.(*syntax.Variable); ok {
				value := f.vars[variable.Name]
				if state, ok := f.states[value.Identity]; ok {
					state.Successful = true
					f.states[value.Identity] = state
				}
			}
		}
	case *syntax.FuncCall:
		name, builtin := e.callName(n)
		if !builtin {
			return
		}
		args, valid := e.guardArguments(n)
		if !valid || len(args) == 0 || args[0] == nil {
			return
		}
		if truth && name == "flock" {
			first := args[0]
			if variable, ok := first.Value.(*syntax.Variable); ok {
				value := f.vars[variable.Name]
				if state, ok := f.states[value.Identity]; ok {
					state.Successful = true
					f.states[value.Identity] = state
				}
			}

		}
		if truth && (name == "array_key_exists" || name == "key_exists") && len(args) == 2 && args[1] != nil {
			// Only an array (or an object) has keys.
			exclude(args[1].Value, f, true, true)
		}
		arg := args[0]
		variable, ok := syntax.UnwrapParens(arg.Value).(*syntax.Variable)
		if !ok || variable.NameExpr != nil {
			return
		}
		switch name {
		case "is_string", "is_array", "is_int", "is_integer", "is_long", "is_float", "is_double", "is_numeric",
			"is_object", "is_resource", "is_iterable", "is_countable", "is_callable":
			if truth {
				exclude(variable, f, true, true)
			}
		case "is_bool":
			if !truth {
				exclude(variable, f, true, false)
			}
		case "is_null":
			if !truth {
				exclude(variable, f, false, true)
			}
		}
		v := f.vars[variable.Name]
		if truth && (name == "ctype_digit" || name == "is_int" || name == "is_integer") {
			v.Safe |= SQL
			f.vars[variable.Name] = v
		}
		if truth && name == "in_array" && len(args) == 3 {
			list, strict := args[1], args[2]
			if list == nil || strict == nil {
				return
			}
			flag, ok := strict.Value.(*syntax.ConstFetch)
			if !ok || !strings.EqualFold(flag.Name.Value, "true") {
				return
			}
			array, ok := list.Value.(*syntax.Array)
			if !ok || len(array.Items) == 0 {
				return
			}
			for _, item := range array.Items {
				if item == nil || item.Unpack || item.ByRef {
					return
				}
				if _, ok := literalText(item.Value); !ok {
					return
				}
			}
			v.Sources = nil
			v.Safe = HTML | SQL | Shell | Header | URL | Path
			f.vars[variable.Name] = v
		}
		if !truth && name == "str_contains" && len(args) == 2 {
			arg := args[1]
			needle, ok := literalText(arg.Value)
			if !ok {
				return
			}
			if needle == "\r" {
				v.headerMask |= 1
			}
			if needle == "\n" {
				v.headerMask |= 2
			}
			if v.headerMask == 3 {
				v.Safe |= Header
			}
			f.vars[variable.Name] = v
		}
	}
}

// guarded names the variable a condition operand tests: the variable itself or
// the target of a plain assignment written inside the condition.
func guarded(x syntax.Expr) *syntax.Variable {
	switch n := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		if n.NameExpr == nil {
			return n
		}
	case *syntax.Assign:
		if target, ok := n.Var.(*syntax.Variable); ok && target.NameExpr == nil && !n.ByRef && n.Op.Kind == syntax.TEqual {
			return target
		}
	}
	return nil
}

// exclude records that the tested variable is not false and/or not null.
func exclude(x syntax.Expr, f *frame, nonFalse, nonNull bool) {
	variable := guarded(x)
	if variable == nil || !(nonFalse || nonNull) {
		return
	}
	// A variable the frame does not track has no value to qualify.
	v, tracked := f.vars[variable.Name]
	if !tracked {
		return
	}
	if nonFalse {
		v.NonFalse = true
		if s, exists := f.states[v.Identity]; exists {
			s.Successful = true
			f.states[v.Identity] = s
		}
	}
	if nonNull {
		v.NonNull = true
	}
	v.origins = v.origins.excluding(nonFalse, nonNull)
	f.vars[variable.Name] = v
}

// indexed returns the array variable an element access reads, at any depth:
// an element that is set, or a key that exists, proves it holds an array.
func indexed(x syntax.Expr) syntax.Expr {
	fetch, ok := syntax.UnwrapParens(x).(*syntax.ArrayDimFetch)
	if !ok {
		return nil
	}
	for {
		inner, nested := syntax.UnwrapParens(fetch.Var).(*syntax.ArrayDimFetch)
		if !nested {
			return fetch.Var
		}
		fetch = inner
	}
}

func constantOperand(x syntax.Expr) bool {
	switch syntax.UnwrapParens(x).(type) {
	case *syntax.ConstFetch, *syntax.Literal:
		return true
	}
	return false
}

// constantName returns the folded name of false or null.
func constantName(x syntax.Expr) string {
	if c, ok := syntax.UnwrapParens(x).(*syntax.ConstFetch); ok && c.Name != nil {
		switch name := strings.ToLower(strings.TrimPrefix(c.Name.Value, `\`)); name {
		case "false", "null":
			return name
		}
	}
	return ""
}

// rebinds reports that evaluating x may write a variable tested reads: the
// target of an assignment or a step, or an argument a callee may take by
// reference. An earlier fact about that variable then describes a value that
// may be gone.
func (e *Env) rebinds(x, tested syntax.Expr) bool {
	found := false
	mentions := func(written syntax.Expr) {
		if name := rootVariable(written); name != "" {
			syntax.Inspect(tested, func(read syntax.Node) bool {
				if v, ok := read.(*syntax.Variable); ok && v.NameExpr == nil && v.Name == name {
					found = true
				}
				return !found
			})
		}
	}
	syntax.Inspect(x, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Assign:
			mentions(n.Var)
		case *syntax.IncDec:
			mentions(n.Var)
		case *syntax.ArgList:
			params, _, known := e.declaration(n.Parent())
			position := 0
			for _, x := range n.Args {
				a, ok := x.(*syntax.Arg)
				if !ok {
					continue
				}
				// Only a resolved by-value parameter proves a read.
				read := false
				for i, param := range params {
					if a.Name == nil && i == position || a.Name != nil && param.Name == a.Name.Value {
						read = known && !param.ByRef && !a.Unpack && !a.ByRef
					}
				}
				if a.Name == nil {
					position++
				}
				if !read {
					mentions(a.Value)
				}
			}
		}
		return !found
	})
	return found
}
