package propertycanbestatic

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

// propertyCanBeStatic reports non-public instance properties defaulting to
// a sizeable array literal.
type propertyCanBeStatic struct{}

func (propertyCanBeStatic) ID() string               { return "PropertyCanBeStatic" }
func (propertyCanBeStatic) Semantic()                {}
func (propertyCanBeStatic) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KProperty} }
func (propertyCanBeStatic) Check(ctx *analysis.Context, n syntax.Node) {
	prop := n.(*syntax.Property)
	if prop.Modifiers.Has(syntax.TStatic) || !(prop.Modifiers.Has(syntax.TPrivate) || prop.Modifiers.Has(syntax.TProtected)) { // D2
		return
	}
	if len(prop.Hooks) > 0 { // custos: neither a static property nor a constant can carry hooks
		return
	}
	cl := prop.Parent().(*syntax.ClassLike) // properties only appear in class-like bodies
	parent := ""
	if p := ctx.Names().ParentFQN(cl); p != "" {
		if pc := ctx.Index().Class(p, ctx.PHP); pc != nil {
			parent = pc.FQN
		}
	}
	for _, item := range prop.Props { // D1
		arr, ok := item.Default.(*syntax.Array) // D3
		if !ok || item.Var.Span().Len() == 0 {
			continue
		}
		if parent != "" && semanticquery.PropertyInChain(ctx.Index(), parent, item.Var.Name, ctx.PHP) != nil { // D4
			continue
		}
		count := 0
		for _, el := range arr.Items { // D5
			if el == nil || el.Unpack || el.Value == nil {
				continue
			}
			switch v := syntax.UnwrapParens(el.Value).(type) {
			case *syntax.Array:
				count++
			case *syntax.InterpolatedString:
				if !v.Backtick {
					count++
				}
			case *syntax.Literal:
				if v.LitKind == syntax.LitString {
					count++
				}
			}
			if count >= 3 {
				break
			}
		}
		if count < 3 {
			continue
		}
		if pcbsWritten(ctx, cl)[item.Var.Name] { // custos: per-instance state
			continue
		}
		msg := "Large array default on an instance property; consider a static property."
		if ctx.PHP >= phpversion.PHP56 {
			msg = "Large array default on an instance property; consider a static property or a class constant."
		}
		ctx.ReportNode(item.Var, msg)
	}
}

// pcbsWritten returns the names of the properties the class body can
// modify through `$this`: assignment, `++`/`--` and `unset` targets
// (`$this->p = …`, `$this->p['k'] = …`, `$this->p[] = …`), destructuring
// and foreach targets, and every place a reference to the property is
// taken (`=&`, `[&$this->p]`, by-reference foreach or destructuring, a
// by-reference return, an argument at a by-reference parameter). Such a
// property holds per-instance state (Roundcube's rcube_db options, set per
// connection; a queue consumed with array_shift), and a static one would
// be shared by every instance (custos).
func pcbsWritten(ctx *analysis.Context, cl *syntax.ClassLike) map[string]bool {
	return ctx.Memo("written:"+strconv.Itoa(int(cl.Span().Start)), func() any {
		out := map[string]bool{}
		mark := func(e syntax.Expr) {
			for {
				switch x := syntax.UnwrapParens(e).(type) {
				case *syntax.ArrayDimFetch:
					e = x.Var
					continue
				case *syntax.PropertyFetch:
					v, isVar := syntax.UnwrapParens(x.Var).(*syntax.Variable)
					if id, ok := x.Name.(*syntax.Identifier); ok && isVar && v.Name == "this" {
						out[id.Value] = true
					}
				}
				return
			}
		}
		// target marks a write target, descending into list()/[...]
		// destructuring; it reports whether an element is bound by
		// reference (`[&$x] = …`), which makes the source writable too.
		var target func(e syntax.Expr) bool
		items := func(list []*syntax.ArrayItem) bool {
			ref := false
			for _, it := range list {
				if it != nil {
					ref = target(it.Value) || it.ByRef || ref
				}
			}
			return ref
		}
		target = func(e syntax.Expr) bool {
			switch x := syntax.UnwrapParens(e).(type) {
			case *syntax.Array:
				return items(x.Items)
			case *syntax.List:
				return items(x.Items)
			}
			mark(e)
			return false
		}
		// args marks the arguments a call may take by reference: all of
		// them when the callee is unknown.
		args := func(list *syntax.ArgList, params []index.Param, known bool) {
			if list == nil {
				return
			}
			for pos, a := range list.Args {
				if arg, ok := a.(*syntax.Arg); ok && (!known || arg.ByRef || pcbsBindsByRef(arg, pos, params)) {
					mark(arg.Value)
				}
			}
		}
		call := func(n syntax.Node, list *syntax.ArgList) {
			callee, known := semanticquery.ResolveCallee(ctx, n)
			args(list, callee.Params, known)
		}
		// returned marks what a by-reference function body returns or
		// yields; nested functions are visited on their own.
		returned := func(body *syntax.Block) {
			syntax.Inspect(body, func(n syntax.Node) bool {
				switch x := n.(type) {
				case *syntax.Return:
					mark(x.Expr)
				case *syntax.Yield:
					mark(x.Value)
				case *syntax.Closure, *syntax.ArrowFunction:
					return false
				}
				return true
			})
		}
		syntax.Inspect(cl, func(n syntax.Node) bool {
			switch x := n.(type) {
			case *syntax.Assign:
				if target(x.Var) || x.ByRef {
					mark(x.Value)
				}
			case *syntax.IncDec:
				mark(x.Var)
			case *syntax.Unset:
				for _, v := range x.Vars {
					mark(v)
				}
			case *syntax.Foreach:
				target(x.Key)
				if target(x.Value) || x.ByRef {
					mark(x.Expr)
				}
			case *syntax.ArrayItem:
				if x.ByRef {
					mark(x.Value)
				}
			case *syntax.FuncCall:
				call(x, x.Args)
			case *syntax.MethodCall:
				call(x, x.Args)
			case *syntax.StaticCall:
				call(x, x.Args)
			case *syntax.New:
				var params []index.Param
				cls := semanticquery.StaticCallClass(ctx, x.Class)
				known := cls != "" && ctx.Index().Class(cls, ctx.PHP) != nil
				if m := ctx.Index().FindMethod(cls, "__construct", ctx.PHP); m != nil {
					params = m.Params
				}
				list := x.Args
				if anon, ok := x.Class.(*syntax.ClassLike); ok {
					list = anon.Args // unknown constructor: `new class($this->p) {…}`
				}
				args(list, params, known)
			case *syntax.Method:
				if x.ByRef && x.Body != nil {
					returned(x.Body)
				}
			case *syntax.Closure:
				if x.ByRef && x.Body != nil {
					returned(x.Body)
				}
			case *syntax.ArrowFunction:
				if x.ByRef {
					mark(x.Expr)
				}
			}
			return true
		})
		return out
	}).(map[string]bool)
}

// pcbsBindsByRef reports whether the argument at index pos of its list
// binds to a by-reference parameter: by name, by position (past the last
// parameter, to a variadic one) or, for a spread argument, to any
// parameter from its position on.
func pcbsBindsByRef(arg *syntax.Arg, pos int, params []index.Param) bool {
	for i, p := range params {
		if !p.ByRef {
			continue
		}
		switch {
		case arg.Name != nil:
			if strings.EqualFold(strings.TrimPrefix(p.Name, "$"), arg.Name.Value) {
				return true
			}
		case arg.Unpack:
			if i >= pos || p.Variadic {
				return true
			}
		case i == pos || p.Variadic && i < pos:
			return true
		}
	}
	return false
}
