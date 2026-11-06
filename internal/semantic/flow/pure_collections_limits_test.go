package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestPureCollectionsCompilerLimits(t *testing.T) {
	for _, source := range []string{
		`function f($x){return UNKNOWN_MODE;}`,
		`function f($x){return [,...$x];}`,
		`function f($x){return [&$x,...[]];}`,
		`function f($x){return [$x=>1,...[]];}`,
		`function f($x){return [1.5=>1,...[]];}`,
		`function f($x){return [9223372036854775807=>1,...[]];}`,
		`function f($x){return [opaque(),...[]];}`,
		`function f($x){return [...$x,` + strings.Repeat("1,", 65) + `];}`,
		`function f($x){return [...$x,` + strings.Repeat("1,", 32) + `];}`,
	} {
		env, file := environment(t, source, nil)
		_, out := Summarize(file, env.types, nil, Options{})
		if out.Summaries["f"].pure != nil {
			t.Errorf("unsupported contract: %s", source)
		}
	}
	env, _ := environment(t, ``, nil)
	compiler := pureCompiler{env: env}
	if compiler.stringPredicate(&syntax.FuncCall{}, "str_contains", nil) != nil {
		t.Fatal("missing metadata accepted")
	}
	env.types.PHP = phpversion.PHP73
	if compiler.spreadArray(&syntax.Array{}) != nil {
		t.Fatal("old spread accepted")
	}
	if compiler.stringPredicate(&syntax.FuncCall{}, "str_contains", nil) != nil {
		t.Fatal("old predicate accepted")
	}
}

func TestPureCollectionsEvaluationLimits(t *testing.T) {
	e, _ := environment(t, ``, nil)
	text := pureConstant(constantScalar{k: 's', s: "x"})
	slot := &pureExpr{kind: pureSlot, slot: 0}
	for _, tc := range []struct {
		expr  *pureExpr
		slots []Value
		r     scopeResult
	}{
		{&pureExpr{kind: pureStringPredicate, left: text, right: slot}, nil, scopeResult{}},
		{&pureExpr{kind: pureStringPredicate, left: text, right: pureConstant(constantScalar{k: 's', s: strings.Repeat("a", 4097)})}, nil, scopeResult{}},
		{&pureExpr{kind: pureSpreadArray, args: []*pureExpr{nil, slot}}, []Value{scalarValue(constantScalar{k: 'i'})}, scopeResult{}},
		{&pureExpr{kind: pureSpreadArray, weakArguments: true, args: []*pureExpr{nil, {kind: pureArray, args: []*pureExpr{text, text}}}}, nil, scopeResult{}},
		{&pureExpr{kind: pureSpreadArray, args: []*pureExpr{text, text}}, nil, scopeResult{copies: MaxCopied}},
		{&pureExpr{kind: pureSpreadArray, args: []*pureExpr{pureConstant(constantScalar{k: 's', s: "9223372036854775806"}), text, nil, {kind: pureArray}}}, nil, scopeResult{}},
	} {
		if _, ok := e.evalPure(tc.expr, tc.slots, &tc.r, MaxDepth); ok {
			t.Errorf("unsafe/budget result: %+v", tc.expr)
		}
	}
	empty := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{}}}
	if appendPureEntry(&empty, scalarValue(constantScalar{k: 'i'}), &scopeResult{copies: MaxCopied}) {
		t.Fatal("copy budget bypassed")
	}
	// A failed per-entry projection must not publish a partially built result.
	expr := &pureExpr{kind: pureBuild, slot: 1, left: &pureExpr{kind: pureArray, args: []*pureExpr{text, text}}, right: &pureExpr{kind: pureArray}, other: slot}
	if _, ok := e.evalPure(expr, nil, &scopeResult{}, MaxDepth); ok {
		t.Fatal("bad projection accepted")
	}
}

func TestPureSpreadRuntimeVersions(t *testing.T) {
	for _, tc := range []struct {
		php    phpversion.Version
		source string
		known  bool
	}{
		{phpversion.PHP74, `function f($xs){return [...$xs][0];}probe(f([4=>'yes']));`, true},
		{phpversion.PHP74, `function f($xs){return [...$xs]['a'];}probe(f(['a'=>'yes']));`, false},
		{phpversion.PHP80, `function f($xs){return [...$xs]['a'];}probe(f(['a'=>'yes']));`, false},
		{phpversion.PHP81, `function f($xs){return [...$xs]['a'];}probe(f(['a'=>'yes']));`, true},
	} {
		e, file := environment(t, tc.source, nil)
		e.types.PHP = tc.php
		snapshot := NewSnapshot(Extract(file, e.types.Index, tc.php, nil, Options{}))
		env := New(file, e.types, snapshot, Options{})
		got, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0])))
		if known != tc.known || known && got.s != "yes" {
			t.Errorf("version=%v source=%s got=%+v known=%v", tc.php, tc.source, got, known)
		}
	}
}
