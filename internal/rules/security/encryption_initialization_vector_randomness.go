package security

import (
	"sort"
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// encryptionInitializationVectorRandomness reports IVs passed to
// openssl_encrypt()/mcrypt_encrypt() that may not come from a
// cryptographically secure generator.
type encryptionInitializationVectorRandomness struct{}

func init() { register(encryptionInitializationVectorRandomness{}) }

func (encryptionInitializationVectorRandomness) ID() string {
	return "EncryptionInitializationVectorRandomness"
}

func (encryptionInitializationVectorRandomness) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

// ivSecureName reports whether name (lower-case) is a secure generator.
func ivSecureName(name string) bool {
	return name == "random_bytes" || name == "openssl_random_pseudo_bytes" || name == "mcrypt_create_iv"
}

// ivSecureCall reports whether v is a call to a secure generator: a call
// resolving to the global function (any case, D4).
func ivSecureCall(ctx *analysis.Context, v syntax.Expr) bool {
	c, ok := v.(*syntax.FuncCall)
	return ok && ivSecureName(ctx.GlobalFunctionName(c))
}

func (encryptionInitializationVectorRandomness) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	var gen string
	switch ctx.GlobalFunctionName(call) { // D1
	case "openssl_encrypt":
		gen = "openssl_random_pseudo_bytes"
	case "mcrypt_encrypt":
		gen = "mcrypt_create_iv"
	default:
		return
	}
	if util.ArgCount(call) < 5 { // D2, E1
		return
	}
	arg, ok := call.Args.Args[4].(*syntax.Arg)
	if !ok || arg.Value == nil || arg.Value.Span().Len() == 0 {
		return
	}
	values := util.DiscoverValues(ctx.Types(), arg.Value) // D3
	if len(values) == 0 {
		return
	}
	var r []string
	for _, v := range values { // D4
		if ivSecureCall(ctx, v) {
			continue
		}
		r = append(r, ctx.Text(v))
	}
	if len(r) == 0 {
		return
	}
	if len(values) == 1 && ivSecureWrapper(ctx, values[0]) { // D5
		return
	}
	sort.Strings(r)
	ctx.ReportNode(arg.Value, "Generate the IV with "+gen+"(); it may come from: "+strings.Join(r, ", ")+".")
}

// ivSecureWrapper reports whether v is a call to a function/method declared
// with a body that calls one of the secure generators.
func ivSecureWrapper(ctx *analysis.Context, v syntax.Expr) bool {
	env := ctx.Types()
	ix := ctx.Index()
	var body *syntax.Block
	switch c := v.(type) {
	case *syntax.FuncCall:
		if d := util.FunctionDecl(ctx.File, env.ResolveFunction(c)); d != nil {
			body = d.Body
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return false
		}
		for _, cls := range env.TypeOf(c.Var).Classes() {
			if m := util.MethodDecl(ctx.File, ix, ix.FindMethod(strings.TrimPrefix(cls, `\`), id.Value, ctx.PHP), ctx.PHP); m != nil {
				body = m.Body
				break
			}
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return false
		}
		if cls := ivClassRef(ctx, c.Class); cls != "" {
			if m := util.MethodDecl(ctx.File, ix, ix.FindMethod(cls, id.Value, ctx.PHP), ctx.PHP); m != nil {
				body = m.Body
			}
		}
	}
	if body == nil {
		return false
	}
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if found {
			return false
		}
		var name string
		switch c := n.(type) {
		case *syntax.FuncCall:
			name = ctx.GlobalFunctionName(c)
		case *syntax.MethodCall:
			if id, ok := c.Name.(*syntax.Identifier); ok {
				name = strings.ToLower(id.Value)
			}
		case *syntax.StaticCall:
			if id, ok := c.Name.(*syntax.Identifier); ok {
				name = strings.ToLower(id.Value)
			}
		}
		found = ivSecureName(name)
		return !found
	})
	return found
}

func ivClassRef(ctx *analysis.Context, x syntax.Expr) string {
	if n, ok := x.(*syntax.Name); ok {
		switch strings.ToLower(n.Value) {
		case "self", "static":
			return util.ClassDeclFQN(ctx.Names(), enclosingClassLike(n))
		case "parent":
			return util.ParentFQN(ctx.Names(), enclosingClassLike(n))
		}
		return ctx.Names().Class(n.Value, n.Span().Start)
	}
	if cs := ctx.TypeOf(x).Classes(); len(cs) == 1 {
		return strings.TrimPrefix(cs[0], `\`)
	}
	return ""
}

func enclosingClassLike(n syntax.Node) *syntax.ClassLike {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if c, ok := p.(*syntax.ClassLike); ok {
			return c
		}
	}
	return nil
}
