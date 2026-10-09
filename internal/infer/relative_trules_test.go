package infer_test

import (
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

func TestTRulesRelativeContractModes(t *testing.T) {
	f := syntax.Parse("relative.php", []byte(`<?php
class Root { public static function factory(): static {} }
class Base extends Root {
 public self $selfProperty;
 public parent $parentProperty;
 public function same(): self {}
 public function above(): parent {}
 public function next(): static {}
 public function run() { t('selfProperty', $this->selfProperty); t('parentProperty', $this->parentProperty); t('forward', parent::factory()); }
}

class Child extends Base {}
function run(Child $c) { t('self', $c->same()); t('parent', $c->above()); t('static', $c->next()); t('staticCall', Child::next()); }
`), syntax.Options{Version: phpver.PHP84})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	for _, strict := range []bool{false, true} {
		r := infer.NewTRules(infer.NewEnv(f, names.New(f), ix, phpver.PHP84))
		r.SpecOnly = strict
		want := map[string]string{"selfProperty": `\Base`, "parentProperty": `\Root`, "self": `\Base`, "parent": `\Root`, "static": `\Child`, "staticCall": `\Child`, "forward": `\Base`}
		if strict {
			want = map[string]string{"selfProperty": "self", "parentProperty": "parent", "self": "self", "parent": "parent", "static": "static", "staticCall": "static", "forward": "static"}
		}
		syntax.InspectFile(f, func(n syntax.Node) bool {
			call, ok := n.(*syntax.FuncCall)
			if !ok {
				return true
			}
			name, ok := call.Name.(*syntax.Name)
			if !ok || name.Value != "t" {
				return true
			}
			label := strings.Trim(call.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal).Raw, "'")
			got := r.TypeOf(call.Args.Args[1].(*syntax.Arg).Value).String()
			if got != want[label] {
				t.Errorf("strict=%v %s: got %s want %s", strict, label, got, want[label])
			}
			return true
		})
	}
}

func TestTRulesFirstClassCallableContracts(t *testing.T) {
	f := syntax.Parse("callables.php", []byte(`<?php
function value(): int { return 1; }
class Base {
 public function same(): self { return new Base(); }
 public static function factory(): static { return new static(); }
}
class Child extends Base {}
function inspect(Child $child) {
 t('function', value(...));
 t('method', $child->same(...));
 t('static', Child::factory(...));
}
`), syntax.Options{Version: phpver.PHP84})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	want := map[string]string{"function": "int", "method": `\Base`, "static": `\Child`}
	for _, native := range []bool{false, true} {
		for _, strict := range []bool{false, true} {
			e := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
			if native {
				e = e.Native()
			}
			r := infer.NewTRules(e)
			r.SpecOnly = strict
			syntax.InspectFile(f, func(n syntax.Node) bool {
				call, ok := n.(*syntax.FuncCall)
				if !ok {
					return true
				}
				name, ok := call.Name.(*syntax.Name)
				if !ok || name.Value != "t" {
					return true
				}
				label := strings.Trim(call.Args.Args[0].(*syntax.Arg).Value.(*syntax.Literal).Raw, "'")
				got := r.TypeOf(call.Args.Args[1].(*syntax.Arg).Value)
				if got.String() != `\Closure` || types.CallableReturn(got).String() != want[label] {
					t.Errorf("native=%v strict=%v %s: got %s", native, strict, label, got.DocString())
				}
				return true
			})
		}
	}
}
