package langmigration

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// instanceofCanBeUsed reports class checks through get_class()/is_a()/
// class_parents() & co. that instanceof expresses directly.
type instanceofCanBeUsed struct{}

func init() { register(instanceofCanBeUsed{}) }

func (instanceofCanBeUsed) ID() string { return "InstanceofCanBeUsed" }

func (instanceofCanBeUsed) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

func (instanceofCanBeUsed) Semantic() {}

// iocClassLiteral returns the FQN named by a class literal.
func iocClassLiteral(e syntax.Expr) (string, bool) {
	lit, ok := e.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitString || len(lit.Raw) < 2 || (lit.Raw[0] != '\'' && lit.Raw[0] != '"') {
		return "", false
	}
	content := lit.Raw[1 : len(lit.Raw)-1]
	if len(content) <= 3 || content == "__PHP_Incomplete_Class" {
		return "", false
	}
	return `\` + strings.ReplaceAll(content, `\\`, `\`), true
}

// iocGlobal reports whether call (named part, lower-cased) resolves to the
// global PHP function of that name, so user functions in a namespace are
// ignored.
func iocGlobal(ctx *analysis.Context, call *syntax.FuncCall, part string) bool {
	switch part {
	case "get_class", "get_parent_class", "is_a", "is_subclass_of", "in_array", "class_implements", "class_parents":
		return util.ResolvesToGlobalFunction(ctx.Names(), ctx.Index(), ctx.PHP, call, part)
	}
	return false
}

func iocNonString(ctx *analysis.Context, s syntax.Expr) bool {
	if lit, ok := s.(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
		return false
	}
	t := ctx.TypeOf(s)
	return !t.IsUnknown() && !t.Has("string")
}

func iocClassExists(ctx *analysis.Context, fqn string) bool {
	for _, c := range ctx.Index().ClassDecls(fqn, ctx.PHP) {
		if c.Kind == syntax.KindClass {
			return true
		}
	}
	return false
}

func iocInterfaceExists(ctx *analysis.Context, fqn string) bool {
	for _, c := range ctx.Index().ClassDecls(fqn, ctx.PHP) {
		if c.Kind == syntax.KindInterface {
			return true
		}
	}
	return false
}

// iocDeclaredExact reports whether fqn names a declaration of kind k spelled
// exactly as declared.
func iocDeclaredExact(ctx *analysis.Context, fqn string, k syntax.ClassKind) bool {
	for _, c := range ctx.Index().ClassDecls(fqn, ctx.PHP) {
		if c.Kind == k && strings.TrimPrefix(c.FQN, `\`) == strings.TrimPrefix(fqn, `\`) {
			return true
		}
	}
	return false
}

// iocObjectsOnly reports whether every component of s's type is a class.
func iocObjectsOnly(ctx *analysis.Context, s syntax.Expr) bool {
	t := ctx.TypeOf(s)
	for _, a := range t.Atoms() {
		if !strings.HasPrefix(a, `\`) {
			return false
		}
	}
	return !t.IsUnknown()
}

// iocFinalExact reports whether fqn names a final class declared with exactly
// this spelling: only then does `get_class($o) === 'X'` mean the same as
// `$o instanceof X` (F2).
func iocFinalExact(ctx *analysis.Context, fqn string) bool {
	for _, c := range ctx.Index().ClassDecls(fqn, ctx.PHP) {
		if c.Kind == syntax.KindClass && c.Final && strings.TrimPrefix(c.FQN, `\`) == strings.TrimPrefix(fqn, `\`) {
			return true
		}
	}
	return false
}

func (instanceofCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	_, part, ok := util.FuncNamePart(call)
	part = strings.ToLower(part) // PHP function names are case-insensitive
	if !ok || !iocGlobal(ctx, call, part) {
		return
	}
	args, ok := util.CallArgValues(call)
	if !ok {
		return
	}
	var (
		ctxNode syntax.Node
		subject syntax.Expr
		fqn     string
		negated bool
		exact   bool // F2: the rewrite is an exact equivalent
	)
	switch part {
	case "get_class", "get_parent_class": // D1
		if len(args) != 1 {
			return
		}
		bin, ok := call.Parent().(*syntax.Binary)
		if !ok {
			return
		}
		switch bin.Op.Kind {
		case syntax.TIsEqual, syntax.TIsIdentical:
		case syntax.TIsNotEqual, syntax.TIsNotIdentical:
			negated = true // D4
		default:
			return
		}
		other := bin.Left
		if bin.Left == syntax.Expr(call) {
			other = bin.Right
		}
		if fqn, ok = iocClassLiteral(other); !ok {
			return
		}
		subject, ctxNode = args[0], bin
		if !iocNonString(ctx, subject) || !iocClassExists(ctx, fqn) {
			return
		}
		if part == "get_class" && len(ctx.Index().ChildrenAll(fqn)) > 0 { // E1
			return
		}
		exact = part == "get_class" && iocFinalExact(ctx, fqn)
	case "is_a", "is_subclass_of": // D2
		switch len(args) {
		case 2:
		case 3:
			if v, ok := util.BoolConst(args[2]); !ok || v {
				return
			}
		default:
			return
		}
		if fqn, ok = iocClassLiteral(args[1]); !ok {
			return
		}
		subject, ctxNode = args[0], call
		if !iocNonString(ctx, subject) || !iocClassExists(ctx, fqn) {
			return
		}
		exact = part == "is_a"
	case "in_array": // D3
		if len(args) < 2 {
			return
		}
		inner, ok := args[1].(*syntax.FuncCall)
		if !ok {
			return
		}
		_, ip, ok := util.FuncNamePart(inner)
		ip = strings.ToLower(ip)
		if !ok || (ip != "class_implements" && ip != "class_parents") || !iocGlobal(ctx, inner, ip) {
			return
		}
		iargs, ok := util.CallArgValues(inner)
		if !ok || len(iargs) < 1 {
			return
		}
		if fqn, ok = iocClassLiteral(args[0]); !ok {
			return
		}
		subject, ctxNode = iargs[0], call
		if !iocNonString(ctx, subject) {
			return
		}
		if ip == "class_implements" && iocInterfaceExists(ctx, fqn) {
			// class_implements() lists interfaces: exact for objects when the
			// literal spells the interface as declared (in_array compares
			// case-sensitively, instanceof does not).
			exact = iocObjectsOnly(ctx, subject) && iocDeclaredExact(ctx, fqn, syntax.KindInterface)
		} else if !iocClassExists(ctx, fqn) {
			return
		}
	default:
		return
	}
	repl := ctx.Text(subject) + " instanceof " + fqn
	if negated {
		repl = "!" + repl
	}
	span := ctxNode.Span()
	if !exact { // F2: approximations are reported without a fix
		ctx.Report(span, "Consider '"+repl+"' (not an exact equivalent).")
		return
	}
	ctx.Report(span, "Prefer '"+repl+"'.", analysis.Fix{
		Title: "Use instanceof",
		Edits: func() []analysis.TextEdit { return []analysis.TextEdit{{Span: span, NewText: repl}} },
	})
}
