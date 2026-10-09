package infer_test

import (
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

func TestLiteralMethodCallbacks(t *testing.T) {
	checkVer(t, phpversion.PHP85, false, `<?php
namespace App;
trait Values { public static function count(): int { return 1; } }
class Base { public static function text(): string { return ''; } }
final class C extends Base {
 use Values;
 public function value(): float { return 1.0; }
 public static function nothing(): void {}
 private static function hidden(): int { return 1; }
 protected static function secret(): int { return 1; }
 /** @template T @param T $x @return T */
 public static function identity($x) { return $x; }
 /** @return ($x is int ? int : string) */
 public static function condition($x) { return $x; }
 /** @return string */
 public static function documented() { return ''; }
}
/** @template T */
class Box { /** @return T */ public function get() {} }
/** @param Box<int> $box */
function run(C $c, $box, array $xs, $method, $unknown) {
 t('string', call_user_func('App\\C::text'));
 t('absolute', call_user_func('\\App\\C::text'));
 t('arraystring', call_user_func(['App\\C','text']));
 t('class', call_user_func([C::class,'count']));
 t('object', call_user_func([$c,'value']));
 t('generic', call_user_func([$box,'get']));
 t('void', call_user_func([C::class,'nothing']));
 t('doc', call_user_func([C::class,'documented']));
 t('map', array_map([C::class,'text'], $xs));
 t('reduce', array_reduce($xs, [C::class,'count'], 0));
 t('args', call_user_func_array([$c,'value'], []));
 t('pipe', 1 |> [C::class,'text']);
 t('objectmissing', call_user_func([$c,'missing']));
 t('function', call_user_func('strlen'));
 t('emptyfunction', call_user_func(''));
 t('closure', call_user_func(fn() => 1));
 t('missing', call_user_func('App\\C::missing'));
 t('relative', call_user_func('C::text'));
 t('private', call_user_func([C::class,'hidden']));
 t('protected', call_user_func([C::class,'secret']));
 t('nonstatic', call_user_func([C::class,'value']));
 t('template', call_user_func([C::class,'identity']));
 t('condition', call_user_func([C::class,'condition']));
 t('dynamic', call_user_func([$c,$method]));
 t('unknown', call_user_func([$unknown,'value']));
 t('scalar', call_user_func([1,'value']));
 t('badsize', call_user_func([C::class]));
 t('key', call_user_func([0 => C::class,'text']));
 t('ref', call_user_func([&$c,'value']));
 t('spread', call_user_func([...$xs,'text']));
 t('badmethod', call_user_func([C::class,1]));
 t('empty', call_user_func([C::class,'']));
 t('emptyclass', call_user_func(['','text']));
 t('badcolon', call_user_func('App\\C:text'));
 t('emptycolon', call_user_func('::text'));
 t('emptytail', call_user_func('App\\C::'));
 t('triple', call_user_func('App\\C:::text'));
 t('const', call_user_func([C::OTHER,'text']));
 t('dynamicclass', call_user_func([$c::class,'text']));
}
`, map[string]string{
		"string": "string", "absolute": "string", "arraystring": "string", "class": "int", "object": "float", "generic": "int", "void": "null", "doc": "string", "map": "string[]", "reduce": "int", "args": "float", "pipe": "string",
		"objectmissing": "mixed", "function": "int", "emptyfunction": "mixed", "closure": "int", "missing": "mixed", "relative": "mixed", "private": "mixed", "protected": "mixed", "nonstatic": "mixed", "template": "mixed", "condition": "mixed", "dynamic": "mixed", "unknown": "mixed", "scalar": "mixed", "badsize": "mixed", "key": "mixed", "ref": "mixed", "spread": "mixed", "badmethod": "mixed", "empty": "mixed", "emptyclass": "mixed", "badcolon": "mixed", "emptycolon": "mixed", "emptytail": "mixed", "triple": "mixed", "const": "mixed", "dynamicclass": "mixed",
	})
}

func BenchmarkLiteralMethodCallbacks(b *testing.B) {
	f := syntax.Parse("t.php", []byte(`<?php final class C { public static function text(): string { return ''; } public function value(): int { return 1; } } $c = new C(); call_user_func('C::text'); call_user_func([C::class, 'text']); call_user_func([$c, 'value']);`), syntax.Options{Version: phpversion.PHP85})
	ix := index.New(nil)
	ix.Add(index.Extract(f))
	ns := names.New(f)
	var calls []syntax.Expr
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.FuncCall); ok {
			calls = append(calls, c)
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		e := infer.NewEnv(f, ns, ix, phpversion.PHP85)
		for _, c := range calls {
			e.TypeOf(c)
		}
	}
}

func TestLiteralMethodCallbacksCrossFile(t *testing.T) {
	checkWith(t, map[string]string{"library.php": `<?php
namespace Library;
final class Source {
 public static function text() { return 'value'; }
 public function count() { return 1; }
}
final class Missing {}
`}, `<?php
use Library\Source;
function run(Source $source, Source|\Library\Missing $alternative, array $xs) {
 t('staticBody', call_user_func('Library\\Source::text'));
 t('staticClassBody', call_user_func([Source::class, 'text']));
 t('objectBody', call_user_func([$source, 'count']));
 t('union', call_user_func([$alternative, 'count']));
 t('mapBody', array_map([Source::class, 'text'], $xs));
}
`, map[string]string{
		"staticBody": "string", "staticClassBody": "string", "objectBody": "int", "union": "mixed", "mapBody": "string[]",
	})
}
