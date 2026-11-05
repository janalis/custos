package flow

import "testing"

func TestCompactAndSubstituteShapesAreImmutable(t *testing.T) {
	e, file := environment(t, `function row($value){return ['safe'=>'text','dirty'=>$value];}`, nil)
	out := Extract(file, e.types.Index, e.types.PHP, nil, Options{})
	template := out.Summaries["row"].Return
	fuel := MaxSources
	if got := compactElements(template.elements, MaxDepth, &fuel); got != template.elements {
		t.Fatal("immutable compact shape was cloned")
	}
	dirty := substitute(template, []Value{{Complete: true, Sources: []Source{{Kind: "_GET", Parameter: -1}}}}, true)
	clean := substitute(template, []Value{{Complete: true}}, true)
	if !tainted(dirty.elements.items["dirty"]) || tainted(clean.elements.items["dirty"]) {
		t.Fatal("caller-specific shape values crossed invocations")
	}
	if template.elements.items["dirty"].Sources[0].Parameter != 0 {
		t.Fatal("substitution mutated the template")
	}
	assertCompactElements(t, template.elements)
}

func TestEmptyShapesAreSharedWithoutMutation(t *testing.T) {
	for _, source := range []string{
		`$a=[];probe($a);`,
		`$a=array_map(null,[],[]);probe($a);`,
		`$a=array_filter([],fn($value)=>true);probe($a);`,
	} {
		e, file := summarized(t, source)
		v := e.Value(argument(callsNamed(file, "probe")[0]))
		if v.elements != emptyElements {
			t.Fatalf("empty shape was reallocated: %s", source)
		}
		copy := v.elements.with("key", Value{Complete: true})
		if len(copy.items) != 1 || len(emptyElements.items) != 0 || !emptyElements.whole {
			t.Fatal("write mutated the shared empty shape")
		}
	}
}
