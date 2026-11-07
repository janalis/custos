package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestCallbackBindingBoundaries(t *testing.T) {
	e, file := summarized(t, `array_map('strtolower',['safe']);array_filter(['safe'],'strtolower');`)
	r := e.scope(nil)
	for _, at := range []syntax.Node{nil, &syntax.MethodCall{}, callsNamed(file, "array_map")[0], callsNamed(file, "array_filter")[0]} {
		if _, known := e.mappedCallback(at, nil, nil, &frame{}, r); known {
			t.Fatal("absent mapping binding acquired a result")
		}
		if _, known := e.filteredCallback(at, nil, nil, &frame{}, r); known {
			t.Fatal("absent filter binding acquired a result")
		}
	}
	at := callsNamed(file, "array_map")[0]
	if _, known := e.mappedCallback(at, at.Args, nil, &frame{}, r); known {
		t.Fatal("missing mapping values acquired a result")
	}
	filter := callsNamed(file, "array_filter")[0]
	if _, known := e.filteredCallback(filter, filter.Args, nil, &frame{}, r); known {
		t.Fatal("missing filter values acquired a result")
	}
}

func TestIncompleteCallbackPacksAreRejected(t *testing.T) {
	e, file := summarized(t, `array_map(...['htmlspecialchars',['safe']]);`)
	at := callsNamed(file, "array_map")[0]
	r := e.scope(nil)
	values := append([]Value(nil), r.calls[r.callIndex[at]].Arguments...)
	elements := cloneElements(values[0].elements)
	for _, optional := range []bool{false, true} {
		elements.whole, elements.optional = optional, optional
		values[0].elements = elements
		if _, known := e.mappedCallback(at, at.Args, values, &frame{}, r); known {
			t.Fatal("partial or optional argument pack established complete mapping")
		}
	}
}

func TestCallbackBindingsRequireAResolvedTarget(t *testing.T) {
	e, file := environment(t, `missing(['safe'],'strtolower');`, nil)
	at := callsNamed(file, "missing")[0]
	r := e.scope(nil)
	if _, _, known := e.callbackBindings(at, at.Args, r.calls[r.callIndex[at]].Arguments, false); known {
		t.Fatal("unresolved target acquired parameter roles")
	}
}
