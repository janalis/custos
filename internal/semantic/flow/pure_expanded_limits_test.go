package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestExpandedPureBudgets(t *testing.T) {
	e, _ := environment(t, ``, nil)
	one := pureConstant(constantScalar{k: 'i', i: 1})
	huge := *one
	huge.cost = pureLimit
	deep := *one
	deep.depth = MaxDepth
	if pureNode(pureChoice, &huge, one, one) != nil || pureNode(pureSequence, &deep, one, nil) != nil {
		t.Fatal("instruction budget bypassed")
	}
	for _, source := range []string{
		`function f($v){return ` + strings.Repeat("1+", 80) + `$v;}probe(f(1));`,
		`function f($v){return [` + strings.Repeat("1,", 65) + `];}probe(f(1));`,
		`function f($v){return [` + strings.Repeat("1,", 32) + `];}probe(f(1));`,
		`function f($v){return ` + strings.Repeat("[", MaxDepth+1) + `$v` + strings.Repeat("]", MaxDepth+1) + `;}probe(f(1));`,
		`function f($v){return strlen(opaque());}probe(f(1));`,
		`function id($v){return $v;}function f($v){return id(&$v,...[]);}probe(f(1));`,
		`function id($v){return $v;}function f($v){return id(...$v);}probe(f([1]));`,
	} {
		env, file := summarized(t, source)
		if _, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0]))); known {
			t.Fatalf("unsafe result: %s", source)
		}
	}
	file := syntax.Parse("test.php", []byte(`<?php return 1;`), syntax.Options{})
	compiler := pureCompiler{env: e, work: pureLimit - 1}
	if compiler.branch(file.Stmts[0], nil, MaxDepth) != nil {
		t.Fatal("branch compilation ignored accumulated work")
	}
	compiler.work = pureLimit
	if compiler.expression(file.Stmts[0].(*syntax.Return).Expr) != nil {
		t.Fatal("spent expression compiler continued")
	}
	if compiler.branch(nil, nil, MaxDepth) != nil {
		t.Fatal("spent branch compiler continued")
	}
	for _, tc := range []struct {
		source string
		php    phpversion.Version
	}{
		{`function f($v){return $v ?? 1;}`, phpversion.PHP56},
		{`function f($v=[1]){return $v;}`, phpversion.PHP53},
		{`function f($v){return strlen(string:$v);}`, phpversion.PHP74},
		{`function id($v){return $v;}function f($v){return id(...['v'=>$v]);}`, phpversion.PHP74},
		{`function id($v){return $v;}function f($v){return id(...[$v]);}`, phpversion.PHP55},
	} {
		env, file := environment(t, tc.source, nil)
		out := Extract(file, env.types.Index, tc.php, nil, Options{})
		if out.Summaries["f"].pure != nil {
			t.Fatalf("version guard ignored: %s", tc.source)
		}
	}
}

func TestExpandedPureVMRefusals(t *testing.T) {
	e, _ := environment(t, ``, nil)
	one := pureConstant(constantScalar{k: 'i', i: 1})
	slot := &pureExpr{kind: pureSlot, slot: 0}
	index := &pureExpr{kind: pureIndex, left: slot, right: pureConstant(constantScalar{k: 'f', f: 0.5})}
	array := &pureExpr{kind: pureArray, args: []*pureExpr{pureConstant(constantScalar{k: 's', s: "0"}), one}}
	for _, tc := range []struct {
		expr   *pureExpr
		slots  []Value
		copies int
	}{
		{array, nil, MaxCopied},
		{&pureExpr{kind: pureArray, args: []*pureExpr{one, slot}}, []Value{{Complete: true}}, 0},
		{index, []Value{{Complete: true, plain: true, deep: true, elements: emptyElements}}, 0},
		{index, []Value{{Complete: true}}, 0},
		{&pureExpr{kind: 255, left: one}, nil, 0},
	} {
		if _, known := e.evalPure(tc.expr, tc.slots, &scopeResult{copies: tc.copies}, MaxDepth); known {
			t.Fatal("invalid instruction or binding became known")
		}
	}
	for _, source := range []string{
		`function f($v){return $v && true;}probe(f(true));`,
		`function f($v){return $v || false;}probe(f(false));`,
		`function f($v){return $v and false;}probe(f(true));`,
		`function f($v){return $v or true;}probe(f(false));`,
	} {
		env, file := summarized(t, source)
		if _, known := knownScalar(env.Value(argument(callsNamed(file, "probe")[0]))); !known {
			t.Fatalf("boolean result lost: %s", source)
		}
	}
}

func TestStructuralPureInputs(t *testing.T) {
	e, _ := environment(t, ``, nil)
	key := pureConstant(constantScalar{k: 'i'})
	contract := &pureContract{params: []pureParam{{}, {}}, expr: &pureExpr{kind: pureIndex, projects: true, left: &pureExpr{kind: pureSlot, slot: 0}, right: key}}
	record := Value{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, items: map[string]Value{"0": scalarValue(constantScalar{k: 's', s: "safe"})}}}
	if _, _, known := e.pureReturn(nil, contract, []Value{record, {Complete: true}}, nil, Value{}, nil, &scopeResult{}); known {
		t.Fatal("unused mutable argument permitted a structural proof")
	}
	if pureReadOnlyInputs(nil, []Value{{Complete: true, plain: true, deep: true, Invalidated: 1}}) {
		t.Fatal("invalidated capture admitted")
	}
	if !pureReadOnlyInputs([]Value{record}, []Value{scalarValue(constantScalar{k: 'i'})}) {
		t.Fatal("immutable capture rejected")
	}
}
