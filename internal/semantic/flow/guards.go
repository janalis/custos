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
	case *syntax.Unary:
		if n.Op.Kind == syntax.TExclaim {
			e.refine(n.Expr, f, !truth)
		}
	case *syntax.Binary:
		if _, variable := syntax.UnwrapParens(n.Right).(*syntax.Variable); variable {
			if _, constant := syntax.UnwrapParens(n.Left).(*syntax.ConstFetch); constant {
				copy := *n
				copy.Left, copy.Right = n.Right, n.Left
				n = &copy
			}
		}
		if (truth && (n.Op.Kind == syntax.TBooleanAnd || n.Op.Kind == syntax.TAnd)) || (!truth && (n.Op.Kind == syntax.TBooleanOr || n.Op.Kind == syntax.TOr)) {
			e.refine(n.Left, f, truth)
			e.refine(n.Right, f, truth)
		}
		if (truth && n.Op.Kind == syntax.TIsIdentical) || (!truth && n.Op.Kind == syntax.TIsNotIdentical) {
			variable, ok := syntax.UnwrapParens(n.Left).(*syntax.Variable)
			if ok {
				if _, literal := literalText(n.Right); literal {
					f.vars[variable.Name] = Value{Complete: true, Expr: n.Right}
				}
			}
		}
		if (!truth && n.Op.Kind == syntax.TIsIdentical) || (truth && n.Op.Kind == syntax.TIsNotIdentical) {
			variable, ok := syntax.UnwrapParens(n.Left).(*syntax.Variable)
			flag, constant := syntax.UnwrapParens(n.Right).(*syntax.ConstFetch)
			if ok && constant {
				v := f.vars[variable.Name]
				if strings.EqualFold(flag.Name.Value, "false") {
					v.NonFalse = true
					if s, exists := f.states[v.Identity]; exists {
						s.Successful = true
						f.states[v.Identity] = s
					}
				}
				if strings.EqualFold(flag.Name.Value, "null") {
					v.NonNull = true
				}
				f.vars[variable.Name] = v
			}
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
		arg := args[0]
		variable, ok := syntax.UnwrapParens(arg.Value).(*syntax.Variable)
		if !ok || variable.NameExpr != nil {
			return
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
