package semanticquery

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/types"
)

// UndeclaredEverywhere reports whether none of the class-likes (or, with
// function, functions) fqns can be declared anywhere the analyzed file may
// run. Every candidate must be missing from the index (stubs, project and
// vendor, at any PHP version), the index world must be complete for the
// file (see index.World), and no candidate may lie where the index cannot
// see: a namespace the project autoloads from unindexed directories, a
// namespace a PHP extension uses, a global function whose leading word no
// builtin function uses or a global class-like whose leading word (see
// index.ClassWord) no builtin class-like uses (an extension the stubs
// lack). A class-like is never reported undeclared when the project calls
// class_alias() with computed names, which may declare any class name.
func UndeclaredEverywhere(ctx *analysis.Context, function bool, fqns ...string) bool {
	ix := ctx.Index()
	world := ix.World()
	if len(fqns) == 0 || !world.Complete {
		return false
	}
	// A class_alias() with computed names may declare any class name.
	if !function && ix.DynamicAliases() {
		return false
	}
	if len(world.Hosts) > 0 && ctx.Shared(hostedKey{}, func() any { return importsFrom(ctx.File, world.Hosts) }).(bool) {
		return false
	}
	prefixes := ix.BuiltinPrefixes()
	for _, fqn := range fqns {
		fqn = strings.TrimPrefix(fqn, `\`)
		if fqn == "" || world.Unseen(fqn) {
			return false
		}
		if function && ix.Function(fqn, 0) != nil || !function && ix.Class(fqn, 0) != nil {
			return false
		}
		lower := strings.ToLower(fqn)
		if ns, _, namespaced := strings.Cut(lower, `\`); namespaced {
			if prefixes.Namespaces[ns] {
				return false
			}
			continue
		}
		if function {
			if word, _, ok := strings.Cut(lower, "_"); ok && !prefixes.Words[word] {
				return false
			}
		} else if !prefixes.ClassWords[index.ClassWord(fqn)] {
			return false
		}
	}
	return true
}

type hostedKey struct{}

// importsFrom reports whether file imports a name under one of namespaces
// (lower case, ending in a backslash): such a file runs within the package
// hosting them (see index.World.Hosts).
func importsFrom(file *syntax.File, namespaces []string) bool {
	found := false
	syntax.InspectFile(file, func(n syntax.Node) bool {
		u, ok := n.(*syntax.Use)
		if !ok {
			return !found
		}
		for _, item := range u.Items {
			name := item.Name.Value
			if u.Prefix != nil {
				name = u.Prefix.Value + `\` + name
			}
			name = strings.ToLower(strings.TrimPrefix(name, `\`))
			for _, ns := range namespaces {
				found = found || strings.HasPrefix(name, ns)
			}
		}
		return false
	})
	return found
}

// existenceChecks are the builtins whose first argument names a function
// or class-like the code asks about.
var existenceChecks = map[string]bool{
	"function_exists": true, "is_callable": true, "class_exists": true,
	"interface_exists": true, "trait_exists": true, "enum_exists": true,
}

// versionConstants are the constants whose presence in a condition makes
// the guarded code depend on the PHP version.
var versionConstants = map[string]bool{
	"php_version_id": true, "php_version": true, "php_major_version": true,
	"php_minor_version": true, "php_release_version": true,
}

type existenceKey struct{}

// ExistenceChecked reports whether the code at n may be meant to run only
// when one of the class-likes or functions fqns exists: the file asks for
// one of them by name (function_exists, is_callable, class_exists,
// interface_exists, trait_exists or enum_exists with a literal name or a
// `::class` fetch), or a check that concerns one of fqns (see guardsAny)
// guards n: it appears in a condition enclosing n (a conditional
// declaration or guarded code), or the conditions guarding n narrow its
// result, made earlier in n's scope (an early exit, a short circuit, a
// switch or match arm).
func ExistenceChecked(ctx *analysis.Context, n syntax.Node, fqns ...string) bool {
	asked := ctx.Shared(existenceKey{}, func() any { return askedNames(ctx) }).(map[string]bool)
	for _, fqn := range fqns {
		if asked[strings.ToLower(strings.TrimPrefix(fqn, `\`))] {
			return true
		}
	}
	for child, p := n, n.Parent(); p != nil; child, p = p, p.Parent() {
		var conds []syntax.Expr
		switch p := p.(type) {
		case *syntax.If:
			conds = append(conds, p.Cond)
			for _, e := range p.ElseIfs {
				conds = append(conds, e.Cond)
			}
		case *syntax.Ternary:
			conds = append(conds, p.Cond)
		case *syntax.Binary:
			if child == p.Right {
				conds = append(conds, p.Left)
			}
		}
		for _, cond := range conds {
			if cond != nil && guardIn(ctx, cond, fqns) {
				return true
			}
		}
	}
	if constructorGuards(ctx, n, fqns) {
		return true
	}
	// An earlier check whose result the conditions guarding n narrow: an
	// early exit, an else arm, a short circuit.
	at, ok := n.(syntax.Expr)
	return ok && NativeEarlierCallProves(ctx, at, func(x syntax.Expr) (types.Type, func(types.Type) bool) {
		if !guardsAny(ctx, x, fqns) {
			return types.Unknown, nil
		}
		return types.Bool, func(t types.Type) bool { return t.OnlyOf("true") || t.OnlyOf("false") }
	})
}

// constructorGuards reports whether n lies in an instance method of a class
// (the constructor aside) whose constructor throws or exits when
// extension_loaded() does not find an extension concerning one of fqns: no
// instance runs the method without the extension. A constructor returning
// early still makes the instance.
func constructorGuards(ctx *analysis.Context, n syntax.Node, fqns []string) bool {
	var method *syntax.Method
	for p := n.Parent(); p != nil; p = p.Parent() {
		switch p := p.(type) {
		case *syntax.Function:
			return false // callable without an instance
		case *syntax.Method:
			if method == nil {
				method = p
			}
		case *syntax.ClassLike:
			if method == nil || method.Modifiers.Has(syntax.TStatic) || strings.EqualFold(method.Name.Value, "__construct") {
				return false
			}
			for _, m := range p.Members {
				ctor, ok := m.(*syntax.Method)
				if !ok || ctor.Body == nil || !strings.EqualFold(ctor.Name.Value, "__construct") {
					continue
				}
				for _, s := range ctor.Body.Stmts {
					if check, ok := s.(*syntax.If); ok && aborts(check.Body) && extensionMissing(ctx, check.Cond, fqns) {
						return true
					}
				}
			}
			return false
		}
	}
	return false
}

// extensionMissing reports whether cond holds when extension_loaded() does
// not find an extension concerning one of fqns: it is `!extension_loaded()`
// of such an extension, or an `||` with one.
func extensionMissing(ctx *analysis.Context, cond syntax.Expr, fqns []string) bool {
	switch c := cond.(type) {
	case *syntax.Paren:
		return extensionMissing(ctx, c.Expr, fqns)
	case *syntax.Binary:
		if c.Op.Kind == syntax.TBooleanOr || c.Op.Kind == syntax.TOr {
			return extensionMissing(ctx, c.Left, fqns) || extensionMissing(ctx, c.Right, fqns)
		}
	case *syntax.Unary:
		call, ok := c.Expr.(*syntax.FuncCall)
		return ok && c.Op.Kind == syntax.TExclaim && ctx.GlobalFunctionName(call) == "extension_loaded" && guardsCall(ctx, call, fqns)
	}
	return false
}

// aborts reports whether body, an if statement's, ends by throwing or
// exiting.
func aborts(body syntax.Stmt) bool {
	if b, ok := body.(*syntax.Block); ok {
		if len(b.Stmts) == 0 {
			return false
		}
		body = b.Stmts[len(b.Stmts)-1]
	}
	if s, ok := body.(*syntax.ExprStmt); ok {
		switch s.Expr.(type) {
		case *syntax.Throw, *syntax.Exit:
			return true
		}
	}
	return false
}

// guardsAny reports whether x, a call or a comparison, checks something
// that may decide whether one of fqns exists: the PHP version (a version
// constant, version_compare(), phpversion() without argument), the
// existence of one of fqns by literal name or `::class` fetch, or an
// extension named like the leading word of one of fqns (extension_loaded()
// or phpversion() with its name, defined() with a constant whose leading
// word it is).
func guardsAny(ctx *analysis.Context, x syntax.Expr, fqns []string) bool {
	switch x.(type) {
	case *syntax.FuncCall, *syntax.Binary:
	default:
		return false
	}
	return guardIn(ctx, x, fqns)
}

// guardIn reports whether x holds a check that may decide whether one of
// fqns exists (see guardsAny).
func guardIn(ctx *analysis.Context, x syntax.Expr, fqns []string) bool {
	found := false
	syntax.Inspect(x, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.FuncCall:
			found = found || guardsCall(ctx, n, fqns)
		case *syntax.ConstFetch:
			found = found || versionConstants[strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`))]
		}
		return !found
	})
	return found
}

// guardsCall is guardsAny for one call.
func guardsCall(ctx *analysis.Context, call *syntax.FuncCall, fqns []string) bool {
	fn := ctx.GlobalFunctionName(call)
	subject, named := checkedName(ctx, call)
	switch {
	case fn == "version_compare", fn == "phpversion" && call.Args != nil && len(call.Args.Args) == 0:
		return true
	case fn == "method_exists": // the other existence checks are askedNames'
		if named {
			for _, fqn := range fqns {
				if strings.EqualFold(subject, strings.TrimPrefix(fqn, `\`)) {
					return true
				}
			}
		}
	case fn == "extension_loaded", fn == "phpversion", fn == "defined":
		if named {
			word, _, _ := strings.Cut(subject, "_")
			for _, fqn := range fqns {
				if leading := leadingWords(fqn); leading[subject] || leading[word] {
					return true
				}
				if fn == "extension_loaded" && startsWord(strings.TrimPrefix(fqn, `\`), subject) {
					return true
				}
			}
		}
	}
	return false
}

// startsWord reports whether ext, an extension name, starts name up to a
// word boundary: `opendkim` starts `OpenDKIMSign` and `opendkim_sign`, but
// `mysql` does not start `mysqli_query`.
func startsWord(name, ext string) bool {
	if ext == "" || len(name) <= len(ext) || !strings.EqualFold(name[:len(ext)], ext) {
		return false
	}
	next := name[len(ext)]
	return next == '_' || next == '\\' || ('A' <= next && next <= 'Z')
}

// leadingWords are, lower case, the words an extension declaring fqn
// likely starts its name with: the first namespace segment of a namespaced
// name, else the word before an underscore and the first word of a
// CamelCase name.
func leadingWords(fqn string) map[string]bool {
	fqn = strings.TrimPrefix(fqn, `\`)
	if ns, _, namespaced := strings.Cut(fqn, `\`); namespaced {
		return map[string]bool{strings.ToLower(ns): true}
	}
	word, _, _ := strings.Cut(strings.ToLower(fqn), "_")
	return map[string]bool{word: true, index.ClassWord(fqn): true}
}

// checkedName is, lower case and without a leading backslash, the name
// the first argument of call gives as a literal string or a `::class`
// fetch.
func checkedName(ctx *analysis.Context, call *syntax.FuncCall) (string, bool) {
	if call.Args == nil || len(call.Args.Args) == 0 {
		return "", false
	}
	arg, ok := call.Args.Args[0].(*syntax.Arg)
	if !ok || arg.Value == nil {
		return "", false
	}
	switch v := arg.Value.(type) {
	case *syntax.Literal:
		if s, ok := syntax.LiteralStringValue(v); ok {
			return strings.ToLower(strings.TrimPrefix(s, `\`)), true
		}
	case *syntax.ClassConstFetch:
		name, named := v.Class.(*syntax.Name)
		id, ok := v.Name.(*syntax.Identifier)
		if named && ok && strings.EqualFold(id.Value, "class") {
			return strings.ToLower(ctx.Names().Class(name.Value, name.Span().Start)), true
		}
	}
	return "", false
}

// askedNames collects, lower case, the names the file's existence checks
// ask about.
func askedNames(ctx *analysis.Context) map[string]bool {
	asked := map[string]bool{}
	syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
		if call, ok := n.(*syntax.FuncCall); ok && existenceChecks[ctx.GlobalFunctionName(call)] {
			if name, ok := checkedName(ctx, call); ok {
				asked[name] = true
			}
		}
		return true
	})
	return asked
}
