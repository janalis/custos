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
	var values []syntax.Expr
	if v, ok := syntax.UnwrapParens(arg.Value).(*syntax.Variable); ok && syntax.EnclosingFuncLike(v) != nil {
		// Only the assignments reaching the call matter: a later
		// `$iv = base64_encode($iv)` does not feed the encryption.
		values, _ = util.PossibleValuesReaching(ctx.File, v)
	} else {
		values = util.DiscoverValues(ctx.Types(), arg.Value) // D3
	}
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
	indexed := false // custos: the index records wrappers of other files
	switch c := v.(type) {
	case *syntax.FuncCall:
		f := env.ResolveFunction(c)
		if d := util.FunctionDecl(ctx.File, f); d != nil {
			body = d.Body
		} else if f != nil {
			indexed = f.CSPRNG
		}
	case *syntax.MethodCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return false
		}
		for _, cls := range env.TypeOf(c.Var).Classes() {
			im := ix.FindMethod(strings.TrimPrefix(cls, `\`), id.Value, ctx.PHP)
			if m := util.MethodDecl(ctx.File, ix, im, ctx.PHP); m != nil {
				body = m.Body
				break
			}
			indexed = indexed || (im != nil && im.CSPRNG)
		}
	case *syntax.StaticCall:
		id, ok := c.Name.(*syntax.Identifier)
		if !ok {
			return false
		}
		if cls := ctx.Types().ClassRef(c.Class); cls != "" {
			im := ix.FindMethod(cls, id.Value, ctx.PHP)
			if m := util.MethodDecl(ctx.File, ix, im, ctx.PHP); m != nil {
				body = m.Body
			} else {
				indexed = im != nil && im.CSPRNG
			}
		}
	}
	if body == nil {
		return indexed
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
