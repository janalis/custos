package flow

import (
	"math"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestPureExpressionRefusals(t *testing.T) {
	for _, source := range []string{
		`function f($v){return +$v;}probe(f('text'));`,
		`function f($v){return -$v;}probe(f('text'));`,
		`function f($v){return $v+1;}probe(f(9223372036854775807));`,
		`function f($v){return $v+1;}probe(f(opaque()));`,
		`function f($v){return $v+1;}probe(f(1,2));`,
		`function f(int $v){return $v+1;}probe(f(1.5));`,
		`function f($v){return f($v);}probe(f(0));`,
		`function f($v){return $v;}function g($v){return f(nope:$v);}probe(g(0));`,
		`function f($v){return $v;}function g($v){return f(v:$v,v:0);}probe(g(0));`,
		`function f($v){return $v;}function g($v){return f(v:$v,0);}probe(g(0));`,
		`function f($v){return $v;}function g($v){return f();}probe(g(0));`,
		`function f($v){return $v;}function g($v){return f($v+opaque());}probe(g(0));`,
		`function f($v){return $v;}function g($v){return f(&$v);}probe(g(0));`,
		`function f($v){return -$v;}probe(f(-0.0));`,
	} {
		e, f := summarized(t, source)
		v := e.Value(argument(callsNamed(f, "probe")[0]))
		if source == `function f($v){return -$v;}probe(f(-0.0));` {
			got, ok := knownScalar(v)
			if !ok || got.k != 'f' || math.Signbit(got.f) {
				t.Fatal("negative zero negation lost")
			}
			continue
		}
		if _, known := knownScalar(v); known {
			t.Errorf("unsafe exact result: %s", source)
		}
	}
	// PHP before named arguments must not acquire a composed contract.
	file := syntax.Parse("old.php", []byte(`<?php function f($v){return $v;}function g($v){return f(v:$v);}`), syntax.Options{Version: phpversion.Max})
	ix := stubsIndexForResource(t, `<?php function f($v){return $v;}function g($v){return f(v:$v);}`, "")
	out := Extract(file, ix, phpversion.PHP74, nil, Options{})
	if out.Summaries["g"].pure != nil {
		t.Fatal("named call accepted before PHP 8")
	}
}

func TestPureVMWorkLimits(t *testing.T) {
	e, _ := summarized(t, `probe(1);`)
	one := pureConstant(constantScalar{k: 'i', i: 1})
	slot := &pureExpr{kind: pureSlot, slot: 0}
	invalid := &pureExpr{kind: pureSlot, slot: 1}
	add := func(a, b *pureExpr) *pureExpr {
		return &pureExpr{kind: pureBinary, op: syntax.TPlus, left: a, right: b}
	}
	call := func(c *pureContract, args ...*pureExpr) *pureExpr {
		return &pureExpr{kind: pureCall, call: c, args: args}
	}
	identity := &pureContract{params: []pureParam{{}}, expr: slot}
	for _, tc := range []struct {
		expr                      *pureExpr
		slots                     []Value
		depth, operations, copies int
	}{
		{one, nil, 0, 0, 0},
		{one, nil, MaxDepth, MaxTransfers, 0},
		{invalid, []Value{scalarValue(constantScalar{k: 'i'})}, MaxDepth, 0, 0},
		{call(identity, one), nil, MaxDepth, 0, MaxCopied},
		{call(identity, invalid), nil, MaxDepth, 0, 0},
		{call(identity, slot), []Value{{}}, MaxDepth, 0, 0},
		{call(&pureContract{params: []pureParam{{}}, expr: invalid}, one), nil, MaxDepth, 0, 0},
		{add(invalid, one), nil, MaxDepth, 0, 0},
		{add(slot, one), []Value{{Complete: true}}, MaxDepth, 0, 0},
		{add(one, invalid), nil, MaxDepth, 0, 0},
		{add(one, slot), []Value{{Complete: true}}, MaxDepth, 0, 0},
		{&pureExpr{kind: pureUnary, op: syntax.TPlus, left: pureConstant(constantScalar{k: 's', s: "x"})}, nil, MaxDepth, 0, 0},
		{&pureExpr{kind: pureUnary, op: syntax.TMinus, left: pureConstant(constantScalar{k: 'i', i: math.MinInt64})}, nil, MaxDepth, 0, 0},
		{&pureExpr{kind: pureBinary, op: syntax.TDot, left: pureConstant(constantScalar{k: 's', s: strings.Repeat("x", 4096)}), right: pureConstant(constantScalar{k: 's', s: "x"})}, nil, MaxDepth, 0, 0},
		{add(pureConstant(constantScalar{k: 'i', i: math.MaxInt64}), one), nil, MaxDepth, 0, 0},
	} {
		r := &scopeResult{operations: tc.operations, copies: tc.copies}
		if _, ok := e.evalPure(tc.expr, tc.slots, r, tc.depth); ok {
			t.Errorf("unsafe evaluation: %+v", tc)
		}
	}
	if _, ok := e.pureBind(identity, nil, true); ok {
		t.Fatal("wrong arity accepted")
	}
	r := &scopeResult{copies: MaxCopied}
	if _, _, ok := e.pureReturn(nil, identity, []Value{scalarValue(constantScalar{k: 'i'})}, nil, Value{}, nil, r); ok {
		t.Fatal("copy limit ignored")
	}
	if _, ok := e.runPure(&pureContract{expr: &pureExpr{kind: pureSequence, left: invalid, right: one}}, nil, &scopeResult{}, MaxDepth); ok {
		t.Fatal("discarded warning ignored")
	}
	// A record constructed inside a contract must not acquire a caller identity.
	record := Value{Complete: true, elements: &elementSet{whole: true, hasResources: true}}
	if _, _, ok := e.pureReturn(nil, &pureContract{outer: 1, expr: slot}, nil, []Value{record}, Value{}, nil, &scopeResult{}); ok {
		t.Fatal("unproven record accepted")
	}
}

func TestPureBindingBoundaries(t *testing.T) {
	e, _ := summarized(t, `probe(1);`)
	if _, ok := e.defaultArguments(nil, nil, []Value{{}}); ok {
		t.Fatal("unknown declaration defaulted")
	}
	f := syntax.Parse("call.php", []byte(`<?php function f($v=0){return $v;}f(bad:1);`), syntax.Options{Version: phpversion.Max})
	env := New(f, resourceTypes(f, stubsIndexForResource(t, `<?php function f($v=0){return $v;}f(bad:1);`, "")), nil, Options{})
	if _, ok := env.defaultArguments(callsNamed(f, "f")[0], []Value{scalarValue(constantScalar{k: 'i', i: 1})}, []Value{scalarValue(constantScalar{k: 'i'})}); ok {
		t.Fatal("invalid binding defaulted")
	}
	if pureCaptures(argument(callsNamed(f, "f")[0]), nil, 2) != nil {
		t.Fatal("invalid outer slot layout accepted")
	}
	// Inner parameter coercion uses the call site's definition, independently
	// of the callee's return coercion context.
	for _, tc := range []struct {
		defs, caller string
		known        bool
	}{
		{`<?php function zero(int $v){return $v;}`, `<?php function wrap($v){return zero($v);}probe(wrap('00'));`, true},
		{`<?php function zero(int $v){return $v;}`, `<?php declare(strict_types=1);function wrap($v){return zero($v);}probe(wrap('00'));`, false},
	} {
		ix := stubsIndexForResource(t, tc.defs, tc.caller)
		def := syntax.Parse("definitions.php", []byte(tc.defs), syntax.Options{Version: phpversion.Max})
		snap := NewSnapshot(Extract(def, ix, phpversion.Max, nil, Options{}))
		use := syntax.Parse("caller.php", []byte(tc.caller), syntax.Options{Version: phpversion.Max})
		_, published := Summarize(use, resourceTypes(use, ix), snap, Options{})
		env := New(use, resourceTypes(use, ix), snap.WithFile(published), Options{})
		_, known := knownScalar(env.Value(argument(callsNamed(use, "probe")[0])))
		if known != tc.known {
			t.Fatalf("inner call strictness lost: %s known=%v want=%v", tc.caller, known, tc.known)
		}
	}
}

func TestPureExpandedCallBudget(t *testing.T) {
	source := `function f($v){return $v+$v;}`
	for i := 0; i < 10; i++ {
		name := strings.Repeat("w", i+1)
		previous := "f"
		if i > 0 {
			previous = strings.Repeat("w", i)
		}
		source += `function ` + name + `($v){return ` + previous + `($v)+` + previous + `($v);}`
	}
	e, _ := summarized(t, source+`probe(1);`)
	for name, s := range Extract(e.file, e.types.Index, e.types.PHP, nil, Options{}).Summaries {
		if len(name) > 8 && s.pure != nil {
			t.Fatal("expanded call budget bypassed")
		}
	}
}

func TestPureContractsRejectPreviouslyPureRecursiveEdit(t *testing.T) {
	for _, tc := range []struct{ before, after string }{
		{`function f($x){return $x;}`, `function f($x){return f($x);}`},
		{`function f($x){return $x;}function g($x){return f($x);}`, `function f($x){return g($x);}function g($x){return f($x);}`},
	} {
		e, f := environment(t, tc.before, nil)
		snap := NewSnapshot(Extract(f, e.types.Index, e.types.PHP, nil, Options{}))
		changed := syntax.Parse(f.Path, []byte("<?php "+tc.after), syntax.Options{Version: phpversion.Max})
		out := Extract(changed, e.types.Index, e.types.PHP, snap, Options{})
		if out.Summaries["f"].pure != nil {
			t.Fatal("recursive edit reused a previously pure contract")
		}
	}
}
