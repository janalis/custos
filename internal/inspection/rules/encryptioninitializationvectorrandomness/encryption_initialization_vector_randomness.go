package encryptioninitializationvectorrandomness

import (
	"sort"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// encryptionInitializationVectorRandomness reports IVs passed to
// openssl_encrypt()/mcrypt_encrypt() that may not come from a
// cryptographically secure generator.
type encryptionInitializationVectorRandomness struct{}

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
// resolving to the global function (any case, D4), or getBytes() on a
// Random\Randomizer that keeps a secure engine.
func ivSecureCall(ctx *analysis.Context, v syntax.Expr) bool {
	switch c := v.(type) {
	case *syntax.FuncCall:
		return ivSecureName(ctx.GlobalFunctionName(c))
	case *syntax.MethodCall:
		return ivSecureRandomizer(ctx, c)
	}
	return false
}

// ivSecureRandomizer reports whether c is `$r->getBytes()` on a receiver
// typed Random\Randomizer that is not traced to a construction with an
// explicit engine other than `new Random\Engine\Secure` (the default
// engine is the system CSPRNG).
func ivSecureRandomizer(ctx *analysis.Context, c *syntax.MethodCall) bool {
	env := ctx.Types()
	id, ok := c.Name.(*syntax.Identifier)
	if !ok || !strings.EqualFold(id.Value, "getBytes") {
		return false
	}
	if cls := env.TypeOf(c.Var).Classes(); len(cls) != 1 || !strings.EqualFold(cls[0], `\Random\Randomizer`) {
		return false
	}
	sources := []syntax.Expr{syntax.UnwrapParens(c.Var)}
	if v, ok := sources[0].(*syntax.Variable); ok && syntax.EnclosingFuncLike(v) != nil {
		sources, _ = flowquery.PossibleValuesReaching(ctx.File, v)
	}
	for _, s := range sources {
		n, ok := s.(*syntax.New)
		if !ok || n.Args == nil || len(n.Args.Args) == 0 {
			continue
		}
		engine, ok := syntax.UnwrapParens(semanticquery.CallArgument(n.Args, 0, "engine")).(*syntax.New)
		if !ok || !strings.EqualFold(env.ClassRef(engine.Class), `Random\Engine\Secure`) {
			return false
		}
	}
	return true
}

// ivNoVectorMode reports whether the cipher (openssl) or mode (mcrypt)
// argument literally names ECB, which takes no IV.
func ivNoVectorMode(ctx *analysis.Context, mode syntax.Expr) bool {
	switch m := syntax.UnwrapParens(mode).(type) {
	case *syntax.Literal:
		s, _ := astquery.StringLiteralValue(m.Raw)
		s = strings.ToLower(s)
		return s == "ecb" || strings.HasSuffix(s, "-ecb")
	case *syntax.ConstFetch:
		return strings.TrimPrefix(ctx.Text(m), `\`) == "MCRYPT_MODE_ECB"
	}
	return false
}

// ivEmptyString reports whether v is the empty string literal.
func ivEmptyString(v syntax.Expr) bool {
	l, ok := v.(*syntax.Literal)
	if !ok {
		return false
	}
	s, ok := astquery.StringLiteralValue(l.Raw)
	return ok && s == ""
}

func (encryptionInitializationVectorRandomness) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	var gen string
	mode := 1
	modeName := "cipher_algo"
	switch ctx.GlobalFunctionName(call) { // D1
	case "openssl_encrypt":
		gen = "openssl_random_pseudo_bytes"
	case "mcrypt_encrypt":
		gen = "mcrypt_create_iv"
		mode, modeName = 3, "mode"
	default:
		return
	}
	// D2, E1: bound by position or by name; unknown behind a spread.
	iv := semanticquery.CallArgument(call.Args, 4, "iv")
	if iv == nil || iv.Span().Len() == 0 {
		return
	}
	var values []syntax.Expr
	var dead map[syntax.Expr]bool
	if v, ok := syntax.UnwrapParens(iv).(*syntax.Variable); ok && syntax.EnclosingFuncLike(v) != nil {
		// Only the assignments reaching the call matter: a later
		// `$iv = base64_encode($iv)` does not feed the encryption.
		values, _ = flowquery.PossibleValuesReaching(ctx.File, v)
		dead = ivReplacedValues(ctx, v)
	} else {
		values = semanticquery.DiscoverValues(ctx.Types(), iv) // D3
	}
	if len(values) == 0 {
		return
	}
	ecb := ivNoVectorMode(ctx, semanticquery.CallArgument(call.Args, mode, modeName))
	var r []string
	for _, v := range values { // D4
		if ivSecureCall(ctx, v) || dead[v] || (ecb && ivEmptyString(v)) {
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
	ctx.ReportNode(iv, "Generate the IV with "+gen+"(); it may come from: "+strings.Join(r, ", ")+".")
}

// ivReplacedValues returns the directly assigned values (and parameter
// default) of the local variable v that the structural reaching analysis
// keeps although a statement executed on every path to v replaces them: a
// try whose catches all assign or leave, a do-while body, or an `if` testing
// the variable for the very placeholder (null, or ” for a falsiness test).
func ivReplacedValues(ctx *analysis.Context, v *syntax.Variable) map[syntax.Expr]bool {
	scope := syntax.EnclosingFuncLike(v)
	type candidate struct {
		value syntax.Expr
		end   uint32
	}
	var cands []candidate
	defs, entry := flowquery.ReachingAssignmentsIn(ctx.File, scope, v, v.Name)
	if entry {
		params, _ := astquery.ScopeParts(scope)
		for _, p := range params {
			if p.Var != nil && p.Var.Name == v.Name && p.Default != nil {
				cands = append(cands, candidate{syntax.UnwrapParens(p.Default), 0})
			}
		}
	}
	for _, d := range defs {
		cands = append(cands, candidate{syntax.UnwrapParens(astquery.AssignedValue(d)), d.Span().End})
	}
	var dead map[syntax.Expr]bool
	at := v.Span().Start
	for n := syntax.Node(v); n != scope; n = n.Parent() {
		var list []syntax.Stmt
		switch l := n.(type) {
		case *syntax.Block:
			list = l.Stmts
		case *syntax.Case:
			list = l.Stmts
		}
		for _, s := range list {
			if s.Span().End > at {
				break
			}
			for _, c := range cands {
				if c.end <= s.Span().Start && ivReplaces(ctx, s, v.Name, c.value) {
					if dead == nil {
						dead = map[syntax.Expr]bool{}
					}
					dead[c.value] = true
				}
			}
		}
	}
	return dead
}

// ivReplaces reports whether control leaving s normally has always assigned
// $name when the variable held val before s.
func ivReplaces(ctx *analysis.Context, s syntax.Stmt, name string, val syntax.Expr) bool {
	switch x := s.(type) {
	case *syntax.Try:
		if !ivAssigns(x.Body, name) {
			return false
		}
		for _, c := range x.Catches {
			if !syntax.Terminates(c.Body) && !ivAssigns(c.Body, name) {
				return false
			}
		}
		return true
	case *syntax.DoWhile:
		return ivAssigns(x.Body, name)
	case *syntax.If:
		return ivGuards(ctx, x.Cond, name, val) && (syntax.Terminates(x.Body) || ivAssigns(x.Body, name))
	}
	return false
}

// ivAssigns reports whether s is, or is a block with, a plain assignment
// statement to $name that no earlier break/continue/goto can skip.
func ivAssigns(s syntax.Stmt, name string) bool {
	list := []syntax.Stmt{s}
	if b, ok := s.(*syntax.Block); ok {
		list = b.Stmts
	}
	for _, st := range list {
		if es, ok := st.(*syntax.ExprStmt); ok {
			if a, ok := syntax.UnwrapParens(es.Expr).(*syntax.Assign); ok && a.Op.Kind == syntax.TEqual && ivIsVar(a.Var, name) {
				return true
			}
		}
		jumps := false
		syntax.Inspect(st, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Break, *syntax.Continue, *syntax.Goto:
				jumps = true
			}
			return !jumps
		})
		if jumps {
			return false
		}
	}
	return false
}

// ivGuards reports whether cond holds whenever $name equals the placeholder
// val: a null comparison or is_null() for null, `!$name` / empty($name) for
// null and ”.
func ivGuards(ctx *analysis.Context, cond syntax.Expr, name string, val syntax.Expr) bool {
	null := syntax.IsNullConst(val)
	switch c := syntax.UnwrapParens(cond).(type) {
	case *syntax.Binary:
		if c.Op.Kind != syntax.TIsIdentical && c.Op.Kind != syntax.TIsEqual {
			return false
		}
		l, r := syntax.UnwrapParens(c.Left), syntax.UnwrapParens(c.Right)
		return null && (ivIsVar(l, name) && syntax.IsNullConst(r) || syntax.IsNullConst(l) && ivIsVar(r, name))
	case *syntax.FuncCall:
		return null && ctx.GlobalFunctionName(c) == "is_null" &&
			ivIsVar(semanticquery.CallArgument(c.Args, 0, "value"), name)
	case *syntax.Unary:
		return c.Op.Kind == syntax.TExclaim && ivIsVar(c.Expr, name) && (null || ivEmptyString(val))
	case *syntax.Empty:
		return ivIsVar(c.Expr, name) && (null || ivEmptyString(val))
	}
	return false
}

// ivIsVar reports whether e is the plain variable $name.
func ivIsVar(e syntax.Expr, name string) bool {
	v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
	return ok && v.NameExpr == nil && v.Name == name
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
		if d := semanticquery.FunctionDecl(ctx.File, f); d != nil {
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
			if m := semanticquery.MethodDecl(ctx.File, ix, im, ctx.PHP); m != nil {
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
			if m := semanticquery.MethodDecl(ctx.File, ix, im, ctx.PHP); m != nil {
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
