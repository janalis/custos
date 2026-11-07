package flow

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSummaryInventoryMatchesLazyDeclarations(t *testing.T) {
	for _, source := range []string{
		`namespace Example;function pass($x){return $x;}final class R{function run($x){return $x;}}`,
		`function pass($x){return $x;}function pass($x){return 'other';}class R{function run(){} function run(){}}`,
		`if($flag){function pass($x){return $x;}}`,
		`class R{} function R($x){return $x;}`,
	} {
		lazy, file := environment(t, source, nil)
		eager, _ := Summarize(file, lazy.types, nil, Options{})
		for _, name := range []string{"example\\pass", "example\\r::run", "pass", "r::run", "r", "example\\r", "missing"} {
			if eager.localBody(name) != lazy.localBody(name) {
				t.Fatalf("inventory changed declaration %s", name)
			}
		}
	}
}

func TestNestedValuesRemainOutOfSnapshots(t *testing.T) {
	e, file := environment(t, `function factory(){return ['safe'=>'value','nested'=>['field'=>'text']];}function send(){header(['nested'=>['field'=>'text']]);}`, nil)
	out := Extract(file, e.types.Index, e.types.PHP, nil, Options{})
	for _, sum := range out.Summaries {
		assertCompactElements(t, sum.Return.elements)
		for _, sink := range sum.Sinks {
			if sink.Argument.elements != nil {
				t.Fatal("sink retained local element structure")
			}
		}
	}
}

func assertCompactElements(t *testing.T, elements *elementSet) {
	t.Helper()
	if elements == nil {
		return
	}
	for _, value := range elements.items {
		if value.Expr != nil || value.Identity != 0 || value.origins != nil || value.targets != nil || value.returned != nil {
			t.Fatal("shape retained local execution metadata")
		}
		assertCompactElements(t, value.elements)
	}
}

func TestCompactShapeBudgetsAreDeterministic(t *testing.T) {
	input := &elementSet{items: map[string]Value{}}
	for _, key := range []string{"z", "a"} {
		nested := &elementSet{items: map[string]Value{}}
		for i := range 40 {
			nested.items[fmt.Sprint(i)] = Value{Complete: true}
		}
		input.items[key] = Value{Complete: true, elements: nested}
	}
	fuel := MaxSources
	want := compactElements(input, MaxDepth, &fuel)
	if want.items["a"].elements == nil || want.items["z"].elements != nil {
		t.Fatal("shape fuel did not follow stable key order")
	}
	for range 20 {
		fuel = MaxSources
		if got := compactElements(input, MaxDepth, &fuel); !reflect.DeepEqual(got, want) {
			t.Fatal("map iteration changed retained shape")
		}
	}
	fuel = MaxSources
	if compactElements(input, 0, &fuel) != nil {
		t.Fatal("shape depth budget ignored")
	}
}
