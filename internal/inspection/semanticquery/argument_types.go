package semanticquery

import (
	_ "embed"
	"strings"
	"sync"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// stubTypeDriftList names the builtins of bundled extensions whose stub
// declares other native parameter types than PHP's reflection (PHP 8.0 to
// 8.5), or which PHP 8.5 does not declare (lower case, `class::method`).
//
//go:embed stub_type_drift.txt
var stubTypeDriftList string

var stubTypeDrift = sync.OnceValue(func() map[string]bool {
	out := map[string]bool{}
	for line := range strings.Lines(stubTypeDriftList) {
		if line = strings.TrimSpace(line); line != "" {
			out[line] = true
		}
	}
	return out
})

// TrustedBuiltinTypes reports whether the stub declaration of builtin callee
// (a function FQN or `Class::method`) in stub file declares the native
// parameter types PHP enforces at version ver: from PHP 8.0, where builtins
// declare their parameter types, in a bundled extension, and not known to
// drift from PHP.
func TrustedBuiltinTypes(file, callee string, ver phpversion.Version) bool {
	callee = strings.ToLower(strings.TrimPrefix(callee, `\`))
	return ver >= phpversion.PHP80 && BundledStub(file) && !stubTypeDrift()[callee]
}

// ArgumentMismatch is an argument that every declaration its call may bind
// to rejects.
type ArgumentMismatch struct {
	Arg *syntax.Arg
	// Callee names the nominal declaration (`f`, `Class::method`).
	Callee string
	// Param is the name of the parameter of the nominal declaration.
	Param string
	// Expected is the rejecting type of the nominal declaration, as declared.
	Expected string
	// Given is the argument's type.
	Given types.Type
	// Proven tells that the native parameter types reject the argument's
	// proven type (see infer.Env.ProvenTypeOf); otherwise a PHPDoc type
	// rejects it, or the native types reject only its PHPDoc-aware type.
	Proven bool
}

type argumentTypesKey struct{ call syntax.Node }

// ArgumentTypeMismatches lists the arguments of call that every declaration
// it may bind to (see CallTargets) rejects (see types.AllRejected), in
// source order. An argument is checked when it binds to a parameter in
// every declaration: positional arguments before any unpacked one, named
// arguments to a parameter of that name (never to a builtin's), and
// arguments collected by a variadic parameter, against its native type.
// By-reference parameters are not checked. Cached per call for the file.
func ArgumentTypeMismatches(ctx *analysis.Context, call syntax.Node) []ArgumentMismatch {
	return ctx.Shared(argumentTypesKey{call}, func() any { return argumentTypeMismatches(ctx, call) }).([]ArgumentMismatch)
}

// boundArg is a written argument and how it binds: by position or name.
type boundArg struct {
	arg  *syntax.Arg
	pos  int
	name string
}

type argumentCheckKey struct{}

func argumentTypeMismatches(ctx *analysis.Context, call syntax.Node) []ArgumentMismatch {
	args := bindableArgs(call)
	if len(args) == 0 || reboundReceiver(call) {
		return nil
	}
	callee, ok := ResolveCallee(ctx, call)
	if !ok {
		return nil
	}
	check := fileTypeCheck(ctx)
	// Every declaration must reject an argument, the nominal one first:
	// listing the overrides is left to the arguments it rejects.
	nominal := []CallTarget{{Params: callee.Params, Builtin: callee.Builtin}}
	var rejected []boundArg
	for _, b := range args {
		params, ok := bindAll(nominal, b)
		if !ok || reboundThis(b.arg.Value) {
			continue
		}
		if given := ctx.TypeOf(b.arg.Value); !given.IsUnknown() {
			if _, _, ok := check.rejected(nominal, params, given); ok {
				rejected = append(rejected, b)
			}
		}
	}
	if len(rejected) == 0 {
		return nil
	}
	label, targets, ok := CallTargets(ctx, call, callee, func(file, name string) bool {
		return TrustedBuiltinTypes(file, name, ctx.PHP)
	})
	if !ok {
		return nil
	}
	var out []ArgumentMismatch
	for _, b := range rejected {
		params, ok := bindAll(targets, b)
		if !ok {
			continue
		}
		given := ctx.TypeOf(b.arg.Value)
		expected, native, ok := check.rejected(targets, params, given)
		if !ok {
			continue
		}
		m := ArgumentMismatch{Arg: b.arg, Callee: label, Param: params[0].Name, Expected: expected, Given: given}
		// Proven: the native types also reject what the code alone shows.
		if native {
			if proven := ctx.Types().ProvenTypeOf(b.arg.Value); check.nativeRejects(targets, params, proven) {
				m.Proven, m.Given = true, proven
			}
		}
		out = append(out, m)
	}
	return out
}

// fileTypeCheck returns the acceptance state of the file of ctx, shared by
// the checks of arguments and property writes.
func fileTypeCheck(ctx *analysis.Context) *argumentCheck {
	return ctx.Shared(argumentCheckKey{}, func() any {
		classes := NewIndexRelations(ctx.Index(), ctx.PHP)
		return &argumentCheck{
			weak:     types.AcceptContext{Classes: classes, Strict: syntax.StrictTypes(ctx.File)},
			strict:   types.AcceptContext{Classes: classes, Strict: true},
			declared: map[declaredKey]types.Type{},
		}
	}).(*argumentCheck)
}

// bindableArgs lists the arguments of call that bind to a known parameter:
// those before the first unpacked argument; nil for a first-class callable.
func bindableArgs(call syntax.Node) []boundArg {
	var list *syntax.ArgList
	switch c := call.(type) {
	case *syntax.FuncCall:
		list = c.Args
	case *syntax.MethodCall:
		list = c.Args
	case *syntax.StaticCall:
		list = c.Args
	case *syntax.New:
		list = c.Args
	}
	if list == nil {
		return nil
	}
	var out []boundArg
	for i, node := range list.Args {
		arg, ok := node.(*syntax.Arg)
		if !ok {
			return nil // f(...)
		}
		if arg.Unpack || arg.Value == nil {
			break
		}
		b := boundArg{arg: arg, pos: i}
		if arg.Name != nil {
			b.pos, b.name = -1, arg.Name.Value
		}
		out = append(out, b)
	}
	return out
}

// reboundThis reports whether x reads $this in a closure or an arrow
// function, which may run bound to another object (Closure::bind, macros).
func reboundThis(x syntax.Expr) bool {
	switch syntax.EnclosingFuncLike(x).(type) {
	case *syntax.Closure, *syntax.ArrowFunction:
	default:
		return false
	}
	found := false
	syntax.Inspect(x, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.Name == "this" {
			found = true
		}
		return !found
	})
	return found
}

// reboundReceiver reports whether call is a method call whose receiver
// reads $this in a closure or an arrow function: the method may be another
// object's.
func reboundReceiver(call syntax.Node) bool {
	c, ok := call.(*syntax.MethodCall)
	return ok && reboundThis(c.Var)
}

// bindAll returns the parameter b binds to in each target; false when it
// binds to none in one of them, or to a by-reference parameter.
func bindAll(targets []CallTarget, b boundArg) ([]index.Param, bool) {
	params := make([]index.Param, 0, len(targets))
	for _, t := range targets {
		p, ok := bindParam(t, b)
		if !ok || p.ByRef {
			return nil, false
		}
		params = append(params, p)
	}
	return params, true
}

// bindParam returns the parameter of t that argument b binds to.
func bindParam(t CallTarget, b boundArg) (index.Param, bool) {
	if b.name != "" {
		if t.Builtin {
			return index.Param{}, false // stub names are not checked against PHP's
		}
		for _, p := range t.Params {
			if p.Name == b.name && !p.Variadic {
				return p, true
			}
		}
		return index.Param{}, false
	}
	for i, p := range t.Params {
		if p.Variadic || i == b.pos {
			return p, true
		}
	}
	return index.Param{}, false
}

// argumentCheck holds the acceptance contexts of one file (the file's own
// mode, and strict mode to compare declared types) and the declared types
// parsed so far.
type argumentCheck struct {
	weak, strict types.AcceptContext
	declared     map[declaredKey]types.Type
}

// declaredKey is a declared type and whether its default is null.
type declaredKey struct {
	text     string
	nullable bool
}

func (c *argumentCheck) at(t CallTarget) types.AcceptContext {
	ctx := c.weak
	ctx.Internal = t.Builtin
	return ctx
}

// rejected reports whether every target rejects given (the argument's
// PHPDoc-aware type) by the native type of its parameter or, for project
// code, its PHPDoc type; it returns the rejecting type of the first target
// and whether every target rejects given by its native type.
func (c *argumentCheck) rejected(targets []CallTarget, params []index.Param, given types.Type) (string, bool, bool) {
	expected, allNative := "", true
	for i, t := range targets {
		p := params[i]
		native := c.declaredType(p.Type, p.Default)
		var shown string
		switch {
		case types.AllRejected(native, given, c.at(t)):
			shown = p.Type
		case c.docRejects(t, p, native, given):
			shown, allNative = p.DocType, false
		default:
			return "", false, false
		}
		if i == 0 {
			expected = shown
		}
	}
	return expected, allNative, true
}

// docRejects reports whether the PHPDoc type of parameter p of project
// target t rejects given. A PHPDoc type is not consulted on a variadic
// parameter (it may describe the collected array or one element), when it
// mentions templates, or when the native type rejects it (a stale tag).
// The index keeps no integer range, string length or literal value of a
// PHPDoc type (see Type.DocString), so those are never compared.
func (c *argumentCheck) docRejects(t CallTarget, p index.Param, native, given types.Type) bool {
	if t.Builtin || p.Variadic || p.Gen != "" {
		return false
	}
	return c.docTypeRejects(p.DocType, p.Default, native, given, c.at(t))
}

// docTypeRejects reports whether PHPDoc type text, made nullable by a null
// default def, rejects given in acceptance context at; never for a type
// naming templates or one the native type rejects (a stale tag).
func (c *argumentCheck) docTypeRejects(text, def string, native, given types.Type, at types.AcceptContext) bool {
	if text == "" || strings.Contains(text, `\~`) {
		return false
	}
	doc := c.declaredType(text, def)
	if !native.IsUnknown() && types.AllRejected(native, doc, c.strict) {
		return false
	}
	return types.AllRejected(doc, given, at)
}

// nativeRejects reports whether every target's native parameter type
// rejects proven.
func (c *argumentCheck) nativeRejects(targets []CallTarget, params []index.Param, proven types.Type) bool {
	for i, t := range targets {
		if !types.AllRejected(c.declaredType(params[i].Type, params[i].Default), proven, c.at(t)) {
			return false
		}
	}
	return true
}

// declaredType parses a declared parameter type; a null default (`null` or
// `\null`, in any case) makes it nullable.
func (c *argumentCheck) declaredType(text, def string) types.Type {
	k := declaredKey{text, strings.EqualFold(strings.TrimPrefix(def, `\`), "null")}
	if t, ok := c.declared[k]; ok {
		return t
	}
	t := types.FromDoc(text, nil)
	if !t.IsUnknown() && k.nullable {
		t = types.Union(t, types.Of("null"))
	}
	c.declared[k] = t
	return t
}

// Message is the finding's text: the type to pass instead.
func (m ArgumentMismatch) Message() string {
	return "Pass " + declaredTypeName(m.Expected) + " to $" + m.Param + " of " + m.Callee + "(), not " + typeName(m.Given) + "."
}

// typeName prints t coarsely: its atoms, without value facts or the
// leading backslash of class names.
func typeName(t types.Type) string {
	atoms := t.Atoms()
	names := make([]string, len(atoms))
	for i, a := range atoms {
		names[i] = strings.TrimPrefix(a, `\`)
	}
	return strings.Join(names, "|")
}

// declaredTypeName prints a declared type as written, without the leading
// backslash of class names.
func declaredTypeName(text string) string {
	var b strings.Builder
	for i := 0; i < len(text); i++ {
		if text[i] == '\\' && (i == 0 || !isNameByte(text[i-1])) {
			continue
		}
		b.WriteByte(text[i])
	}
	return b.String()
}

func isNameByte(c byte) bool {
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
}
