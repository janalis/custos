package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func hookTypes(t *testing.T, src string, native, trules bool, want map[string]string) {
	t.Helper()
	f := syntax.Parse("hooks.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	e := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	if native {
		e = e.Native()
	}
	typeOf := e.TypeOf
	if trules {
		typeOf = infer.NewTRules(e).TypeOf
	}
	got := map[string]string{}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			if name, ok := c.Name.(*syntax.Name); ok && name.Value == "t" {
				label := strings.Trim(string(f.Src[c.Args.Args[0].Span().Start:c.Args.Args[0].Span().End]), "'")
				got[label] = typeOf(c.Args.Args[1].(*syntax.Arg).Value).String()
			}
		}
		return true
	})
	for label, expected := range want {
		if got[label] != expected {
			t.Errorf("%s: got %q, want %q", label, got[label], expected)
		}
	}
}

func TestHookLocalScopes(t *testing.T) {
	src := `<?php
class Box {
 public int $n = 0;
 public int $x {
  get {
   t('getterValue', $value);
   t('sibling', $sibling);
   $local = 2;
   t('local', $local);
   $f = function () use ($local) { t('closure', $local); };
   $a = fn() => t('arrow', $local);
   t('this', $this);
   t('property', $this->n);
   /** @var string $documented */
   t('inline', $documented);
   return $local;
  }
  set {
   t('implicit', $value);
   $sibling = 'setter';
   t('setterLocal', $sibling);
   $value = 'changed';
   t('reassigned', $value);
  }
 }
 public string $short { SET => t('expression', $value); }
 public int $explicit { set(int|string $incoming) { t('explicit', $incoming); t('noImplicit', $value); } }
 public float $defaulted { set(mixed $incoming) { t('defaulted', $incoming); } }
 public int|string $guarded { set { if (!is_int($value)) { return; } t('narrow', $value); } }
 public array $shape { set { $xs = ['name' => 1]; $xs['name'] = 's'; t('write', $xs['name']); } }
 public function __construct(string $outside, public int $promoted { set { t('promoted', $value); t('outside', $outside); } }) {}
}
abstract class AbstractBox { abstract public int $n { get; set; } }
`
	want := map[string]string{"getterValue": "?unknown", "sibling": "?unknown", "local": "int", "closure": "int", "arrow": "int", "this": `\Box`, "property": "int", "inline": "string", "implicit": "int", "setterLocal": "string", "reassigned": "string", "expression": "string", "explicit": "int|string", "noImplicit": "?unknown", "defaulted": "mixed", "narrow": "int", "write": "int|string", "promoted": "int", "outside": "?unknown"}
	hookTypes(t, src, false, false, want)
	hookTypes(t, src, false, true, map[string]string{"implicit": "int", "reassigned": "string", "expression": "string", "explicit": "int|string", "defaulted": "mixed", "local": "int"})
}

func TestHookDocAndNativeTypes(t *testing.T) {
	src := `<?php
/** @phpstan-type Names string[] */
class Box {
 /** @var Names */
 public array $names { set { t('doc', $value); } }
 /** @var int[] */
 public array $explicit { /** @param string[] $incoming */ set(array $incoming) { t('param', $incoming); } }
 /** @var */
 public int $empty { set { t('empty', $value); } }
 /** @param int[] $items */
 public function __construct(public array $items { set { t('promotedDoc', $value); } }) {}
}

`
	hookTypes(t, src, false, false, map[string]string{"doc": "string[]", "param": "string[]", "empty": "int", "promotedDoc": "int[]"})
	hookTypes(t, src, true, false, map[string]string{"doc": "array", "param": "array", "empty": "int", "promotedDoc": "array"})
}

func TestHookClassTypes(t *testing.T) {
	src := `<?php
namespace App;
use Other\Value as Alias;
class Base {}
class Box extends Base {
 public Alias $alias { set { t('alias', $value); } }
 public self $same { set { t('self', $value); } }
 public parent $base { set { t('parent', $value); } }
 public ?parent $nullableBase { set { t('nullableParent', $value); } }
 public function __construct(public self $promoted { set { t('promotedSelf', $value); } }) {}
}
`
	want := map[string]string{"alias": `\Other\Value`, "self": `\App\Box`, "parent": `\App\Base`, "nullableParent": `\App\Base|null`, "promotedSelf": `\App\Box`}
	hookTypes(t, src, false, false, want)
	hookTypes(t, src, true, false, want)
}

func BenchmarkHookVariables(b *testing.B) {
	var src strings.Builder
	src.WriteString("<?php class Box {\n")
	for i := range 100 {
		fmt.Fprintf(&src, "public int|string $p%d { get { $v = 1; return $v; } set { if (!is_int($value)) { return; } $v = $value; $f = fn() => $v; } }\n", i)
	}
	src.WriteString("}")
	f := syntax.Parse("hooks.php", []byte(src.String()), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	b.ReportAllocs()
	for b.Loop() {
		e := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if v, ok := n.(*syntax.Variable); ok {
				e.TypeOf(v)
			}
			return true
		})
	}
}

// The parser accepts broken editor buffers; a missing parent cannot resolve.
func TestHookMissingParent(t *testing.T) {
	hookTypes(t, `<?php class InvalidBox { public parent $bad { set { t('missingParent', $value); } } }`, false, false, map[string]string{"missingParent": "?unknown"})
	hookTypes(t, `<?php class InvalidSetter { public int $n { set($incoming) { t('untyped', $incoming); } } }`, false, false, map[string]string{"untyped": "?unknown"})
	hookTypes(t, `<?php class InvalidSetter { public int $n { set($incoming) { t('untyped', $incoming); } } }`, false, true, map[string]string{"untyped": "?unknown"})
}

func TestPromotedHookInlineDocIsolation(t *testing.T) {
	src := `<?php
class Box {
 public function __construct(
  public int $first {
   get { /** @var string $outside */ t('getterDoc', $outside); return 1; }
   set { /** @var string $outside */ t('setterDoc', $outside); }
  },
  public int $second { set { /** @var float $other */ t('secondDoc', $other); } },
  int $plain = 0
 ) {
  t('outside', $outside);
  t('other', $other);
  /** @var bool $normal */
  t('normal', $normal);
  t('plain', $plain);
 }
}
`
	hookTypes(t, src, false, false, map[string]string{"getterDoc": "string", "setterDoc": "string", "secondDoc": "float", "outside": "?unknown", "other": "?unknown", "normal": "bool", "plain": "int"})
	hookTypes(t, src, true, false, map[string]string{"getterDoc": "?unknown", "setterDoc": "?unknown", "secondDoc": "?unknown", "outside": "?unknown", "other": "?unknown", "normal": "?unknown", "plain": "int"})
}
