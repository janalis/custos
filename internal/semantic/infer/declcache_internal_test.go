package infer

import (
	"strconv"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// declEnv returns an Env over a class whose constant a doc type names, and
// its index.
func declEnv(t *testing.T) (*Env, *index.Index) {
	t.Helper()
	parse := func(path, src string) *syntax.File {
		f := syntax.Parse(path, []byte(src), syntax.Options{Version: phpversion.PHP84})
		if len(f.Errors) > 0 {
			t.Fatalf("parse %s: %v", path, f.Errors)
		}
		return f
	}
	f := parse("main.php", `<?php class C { const A = 'a'; const B = 'b'; }`)
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	return NewEnv(f, names.New(f), ix, phpversion.PHP84, Options{}), ix
}

// parsedCached reports whether e has text parsed in its parse cache.
func parsedCached(e *Env, text string) bool {
	_, ok := e.gapRoot().parsed[text]
	return ok
}

// TestParsedTypeSharedAndKept: a type text parsed alone does not depend on
// the index, so one parse serves an Env, its twins and its configured
// copies, and survives an index revision.
func TestParsedTypeSharedAndKept(t *testing.T) {
	env, ix := declEnv(t)
	env.parsedType("list<int>")
	ix.Add(index.Extract(syntax.Parse("d.php", []byte(`<?php class D {}`), syntax.Options{Version: phpversion.PHP84})))
	for name, view := range map[string]*Env{
		"env":        env,
		"native":     env.Native(),
		"proven":     env.Proven(),
		"configured": env.WithOptions(Options{Assertions: true}),
	} {
		if !parsedCached(view, "list<int>") {
			t.Errorf("%s: parse lost", name)
		}
	}
}

// TestDeclarationTypesUseParseCache: member, builtin member and T-rules
// declaration types parse their texts through the parse cache.
func TestDeclarationTypesUseParseCache(t *testing.T) {
	env, _ := declEnv(t)
	r := &TRules{Env: env}
	env.memberType("array", "list<int>")
	env.builtinMemberType("string|false", "non-empty-string")
	r.declaredAndDoc("int", "positive-int", false)
	r.declaredAndDoc("?float", "float", true)
	for _, text := range []string{"array", "list<int>", "string|false", "non-empty-string", "int", "positive-int", "?float", "float"} {
		if !parsedCached(env, text) {
			t.Errorf("%q not parsed through the cache", text)
		}
	}
}

// TestDeclaredAndDocCached: once typed, a declaration's T-rules type is
// read back without parsing or resolving its doc again, one naming a class
// constant included.
func TestDeclaredAndDocCached(t *testing.T) {
	env, _ := declEnv(t)
	r := &TRules{Env: env}
	for _, tc := range []struct {
		declared, doc string
		builtin       bool
	}{
		{"string", `\C::*`, false},
		{"string", `\C::*`, true},
		{"", "int[]", false},
	} {
		first := r.declaredAndDoc(tc.declared, tc.doc, tc.builtin).String()
		if allocs := testing.AllocsPerRun(10, func() { r.declaredAndDoc(tc.declared, tc.doc, tc.builtin) }); allocs != 0 {
			t.Errorf("%+v: %v allocs per cached read", tc, allocs)
		}
		if again := r.declaredAndDoc(tc.declared, tc.doc, tc.builtin).String(); again != first {
			t.Errorf("%+v: %s then %s", tc, first, again)
		}
	}
}

// TestParsedTypeBounded: past maxParsed texts, the parse cache starts over.
func TestParsedTypeBounded(t *testing.T) {
	env, _ := declEnv(t)
	for i := range maxParsed {
		env.parsedType(`\C` + strconv.Itoa(i))
	}
	if got := len(env.parsed); got != maxParsed {
		t.Fatalf("%d cached, want %d", got, maxParsed)
	}
	if got := env.Native().parsedType("int").String(); got != "int" || len(env.parsed) != 1 {
		t.Fatalf("past the bound: %s, %d cached", got, len(env.parsed))
	}
}
