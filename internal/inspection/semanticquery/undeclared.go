package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
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

// existenceChecks are the builtins whose first argument names a function
// or class-like the code asks about.
var existenceChecks = map[string]bool{
	"function_exists": true, "is_callable": true, "class_exists": true,
	"interface_exists": true, "trait_exists": true, "enum_exists": true,
}

// conditionChecks are the builtins whose presence in a condition makes the
// guarded code depend on the runtime: existence, extensions and versions.
var conditionChecks = map[string]bool{
	"function_exists": true, "is_callable": true, "class_exists": true,
	"interface_exists": true, "trait_exists": true, "enum_exists": true,
	"extension_loaded": true, "version_compare": true, "phpversion": true,
	"defined": true, "method_exists": true,
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
// `::class` fetch), a condition enclosing n checks existence, an extension
// or the PHP version (a conditional declaration or guarded code), or the
// conditions guarding n narrow the result of a check made earlier in its
// scope (an early exit, a short circuit, a switch or match arm) that
// concerns one of fqns (see guardsAny).
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
			if cond != nil && runtimeDependent(ctx, cond) {
				return true
			}
		}
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
	found := false
	syntax.Inspect(x, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.FuncCall:
			found = guardsCall(ctx, n, fqns)
		case *syntax.ConstFetch:
			found = versionConstants[strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`))]
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
			}
		}
	}
	return false
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

// runtimeDependent reports whether cond calls a condition check or reads a
// PHP version constant.
func runtimeDependent(ctx *analysis.Context, cond syntax.Expr) bool {
	found := false
	syntax.Inspect(cond, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.FuncCall:
			if conditionChecks[ctx.GlobalFunctionName(n)] {
				found = true
			}
		case *syntax.ConstFetch:
			name := strings.ToLower(strings.TrimPrefix(n.Name.Value, `\`))
			if versionConstants[name] {
				found = true
			}
		}
		return !found
	})
	return found
}
