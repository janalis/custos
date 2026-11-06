package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestPureQueryCompilationLimits(t *testing.T) {
	for _, source := range []string{
		`function f($v){return isset(opaque());}`,
		`function f($v){return isset($v,opaque());}`,
		`function f($v){return match(opaque()){1=>2};}`,
		`function f($v){return match($v){1=>opaque()};}`,
		`function f($v){return match($v){default=>1,default=>2};}`,
		`function f($xs){$s=0;foreach($xs as $x){$s=$s+$x;}opaque();}`,
		`function f($xs){$s=0;foreach($xs as $x){$s=$s+$x;}return 1;}`,
		`function f($xs){$s=0;foreach($xs as $s){$s=$s+1;}return $s;}`,
		`function f($xs){$s=0;foreach($xs as $x){return $s;}return $s;}`,

		`function f($xs){$s=0;foreach($xs as $x){$x=1;}return $s;}`,
		`function f($xs){$s=opaque();foreach($xs as $x){$s=$s+$x;}return $s;}`,
		`function f($xs){$s=0;foreach($xs as $x){$s=opaque();}return $s;}`,
		`function f($v){$a=$v+$v+$v+$v;return match($v){` + strings.Repeat(`1=>$a,`, 13) + `default=>2};}`,
		`function f($v){return isset($v,` + strings.Repeat(`$v,`, 50) + `$v);}`,
	} {
		e, file := environment(t, source, nil)
		_, summary := Summarize(file, e.types, nil, Options{})
		if summary.Summaries["f"].pure != nil {
			t.Errorf("unexpected contract: %s", source)
		}
	}
	source := `function f($v){return match($v){1=>2,default=>3};}`
	e, file := environment(t, source, nil)
	e.types.PHP = phpversion.PHP74
	_, summary := Summarize(file, e.types, nil, Options{})
	if summary.Summaries["f"].pure != nil {
		t.Fatal("match below PHP 8")
	}
}

func TestPureQueryEvaluationLimits(t *testing.T) {
	e, _ := environment(t, ``, nil)
	literal := pureConstant(constantScalar{k: 'i', i: 1})
	unknown := &pureExpr{kind: pureSlot, slot: 0, cost: 1, depth: 1}
	for _, expr := range []*pureExpr{
		{kind: pureMatch, left: unknown, args: []*pureExpr{literal, literal}},
		{kind: pureMatch, left: literal, args: []*pureExpr{unknown, literal}},
		{kind: pureReduce, left: &pureExpr{kind: pureArray}, right: unknown, other: literal},
	} {
		if _, ok := e.evalPure(expr, []Value{{}}, &scopeResult{}, MaxDepth); ok {
			t.Fatal("unknown value was proven")
		}
	}
	reduction := &pureExpr{kind: pureReduce, left: &pureExpr{kind: pureArray}, right: literal, other: literal}
	if _, ok := e.evalPure(reduction, nil, &scopeResult{copies: MaxCopied}, MaxDepth); ok {
		t.Fatal("copy budget bypassed")
	}
	if (&pureCompiler{env: e}).keyExists(&syntax.FuncCall{}, nil) != nil {
		t.Fatal("missing builtin roles")
	}
}

func TestPureOuterParameterLimits(t *testing.T) {
	for _, source := range []string{
		`function f(&$v){return fn()=>1;}`,
		`function f(...$v){return fn()=>1;}`,
		`function f(` + strings.TrimSuffix(strings.Repeat(`$v,`, 65), ",") + `){return fn()=>1;}`,
	} {
		e, file := environment(t, source, nil)
		if _, known := e.pureParameters(functionNamed(file, "f")); known {
			t.Fatal("unsupported outer parameter contract")
		}
	}
}
