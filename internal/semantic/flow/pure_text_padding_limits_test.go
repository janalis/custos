package flow

import (
	"strings"
	"testing"

	"custos/internal/php/syntax"
)

func TestPureTextAndPaddingRuntimeBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	integer := func(i int64) Value { return scalarValue(constantScalar{k: 'i', i: i}) }
	text := func(s string) Value { return scalarValue(constantScalar{k: 's', s: s}) }
	deep := constructionRecord(text("x"))
	deep.elements.depth = MaxDepth - 1
	lower := &pureExpr{kind: pureCase, literal: &constantScalar{k: 's', s: "strtolower"}, left: &pureSlots[0]}
	trim := &pureExpr{kind: pureTrimChars, literal: &constantScalar{k: 's', s: "trim"}, left: &pureSlots[0], right: &pureSlots[1]}
	pad := &pureExpr{kind: pureStrPad, left: &pureSlots[0], right: &pureSlots[1], other: &pureSlots[2], args: []*pureExpr{&pureSlots[3]}}
	arrayPad := &pureExpr{kind: pureArrayPad, left: &pureSlots[0], right: &pureSlots[1], other: &pureSlots[2]}
	rangeExpr := &pureExpr{kind: pureRange, left: &pureSlots[0], right: &pureSlots[1]}
	a, b := text("ab"), text("-")
	for i := 0; i <= MaxSources; i++ {
		source := Source{Kind: "input", Parameter: i}
		if i < MaxSources/2 {
			a.Sources = append(a.Sources, source)
		} else {
			b.Sources = append(b.Sources, source)
		}
	}
	for i, tc := range []struct {
		expr  *pureExpr
		slots []Value
		r     scopeResult
	}{
		{lower, []Value{text(strings.Repeat("A", 64))}, scopeResult{operations: MaxTransfers - 8}},
		{trim, []Value{text(strings.Repeat("A", 64)), text("A")}, scopeResult{operations: MaxTransfers - 8}},
		{pad, []Value{text("ab"), integer(5), text("-"), integer(1)}, scopeResult{operations: MaxTransfers}},
		{pad, []Value{text("ab"), integer(5), text("-"), integer(1)}, scopeResult{copies: MaxCopied}},
		{pad, []Value{a, integer(5), b, integer(1)}, scopeResult{}},
		{arrayPad, []Value{constructionRecord(text("a")), integer(-3), deep}, scopeResult{}},
		{arrayPad, []Value{constructionRecord(text("a")), integer(3), deep}, scopeResult{}},
		{arrayPad, []Value{constructionRecord(deep), integer(3), text("z")}, scopeResult{}},
		{arrayPad, []Value{{Complete: true, plain: true, deep: true, elements: &elementSet{whole: true, width: 1, items: map[string]Value{"k": deep}, order: []string{"k"}}}, integer(3), text("z")}, scopeResult{}},
		{arrayPad, []Value{constructionRecord(text("a")), integer(3), text("z")}, scopeResult{copies: MaxCopied}},
		{rangeExpr, []Value{integer(0), integer(64)}, scopeResult{}},
	} {
		if _, ok := e.evalPure(tc.expr, tc.slots, &tc.r, MaxDepth); ok {
			t.Errorf("case%d accepted unsafe input/budget", i)
		}
	}
}

func TestPureTextAndPaddingCompilerBounds(t *testing.T) {
	e, _ := environment(t, ``, nil)
	c := pureCompiler{env: e}
	if c.padding(&syntax.FuncCall{}, "array_pad", nil) != nil || c.textShape(&syntax.FuncCall{}, "str_pad", nil) != nil {
		t.Fatal("invalid native metadata accepted")
	}
	for _, name := range []string{"strtolower", "str_pad"} {
		source := `function f($xs){return ` + name + `($xs);}`
		if name == "str_pad" {
			source = `function f($xs){return str_pad('a',5,'-',$xs);}`
		}
		e, file := environment(t, source, nil)
		call := callsNamed(file, name)[0]
		params, _, _ := e.declaration(call)
		for _, depth := range []bool{false, true} {
			c := pureCompiler{env: e}
			huge := pureSlots[0]
			if depth {
				huge.depth = MaxDepth
			} else {
				huge.cost = pureLimit
			}
			c.bind("xs", &huge)
			if c.textShape(call, name, params) != nil {
				t.Fatalf("%s: expanded text bound bypassed", name)
			}
		}
	}
	if strings.Contains(defaultTrimSet, "..") {
		t.Fatal("default trim set reads as a range")
	}
}
