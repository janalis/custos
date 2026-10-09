package infer_test

import (
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
)

func TestGeneratorCompletion(t *testing.T) {
	check(t, `<?php
function completed() { yield 1; return 'done'; }
function maybe($b) { yield 1; if ($b) return 2; }
function bare() { yield 1; return; }
function unknown($x) { yield 1; return $x; }
function abrupt() { yield 1; throw new Exception(); return 'dead'; }
function loop() { yield 1; while (true) {} }
function dead() { return 1; yield 2; return 'dead'; }
function delegated() { return yield from completed(); }
function arrayDelegate() { return yield from []; }
function recursive() { return yield from recursive(); }
function nested() { yield 1; $x = function() { return 'nested'; }; return 2; }
final class Source { function items() { yield 1; return true; } }
function run() {
 t('done', completed()->getReturn()); t('maybe', maybe(true)->getReturn());
 t('bare', bare()->getReturn()); t('unknown', unknown(1)->getReturn());
 t('abrupt', abrupt()->getReturn()); t('loop', loop()->getReturn());
 t('dead', dead()->getReturn()); t('delegate', delegated()->getReturn());
 t('array', arrayDelegate()->getReturn()); t('rec', recursive()->getReturn());
 t('nested', nested()->getReturn()); t('method', (new Source())->items()->getReturn());
 $c = function() { yield 1; return 'closure'; }; t('closure', $c()->getReturn());
 $a = fn() => yield from completed(); t('arrow', $a()->getReturn());
}
`, map[string]string{"done": "string", "maybe": "int|null", "bare": "null", "unknown": "mixed", "abrupt": "never", "loop": "never", "dead": "int", "delegate": "string", "array": "null", "rec": "mixed", "nested": "int", "method": "true", "closure": "string", "arrow": "string"})
}

func TestGeneratorSendAndDelegation(t *testing.T) {
	check(t, `<?php
/** @return \Generator<int, string, float, bool> */
function documented() { t('sent', yield 'a'); return true; }
function plain() { t('unknownSend', yield 1); }
/** @return \Generator<int, int> */
function shortDoc() { t('shortSend', yield 1); }
t('globalSend', yield 1);
function badObject() { t('object', yield from new stdClass()); }
/** @param \Generator<int, string, float, bool>|int[] $g */
function union($g) { t('union', yield from $g); }
/** @param \Generator<int, string> $g */
function missing($g) { t('missing', yield from $g); }
/** @param \ArrayIterator $g */
function iterator($g) { t('iterator', yield from $g); }
/** @param \Traversable $g */
function interfaceOnly($g) { t('interface', yield from $g); }
function run($x) { t('invalid', yield from 1); t('unknown', yield from $x); t('sendReturn', documented()->send(1.5)); }
`, map[string]string{"shortSend": "?unknown", "globalSend": "?unknown", "object": "?unknown", "sent": "float|null", "unknownSend": "?unknown", "union": "bool|null", "missing": "?unknown", "iterator": "null", "interface": "?unknown", "invalid": "?unknown", "unknown": "?unknown", "sendReturn": "null|string"})
	f := syntax.Parse("native.php", []byte(`<?php /** @return \Generator<int,int,string,bool> */ function f() { yield 1; }`), syntax.Options{Version: phpver.PHP85})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP85).Native()
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if y, ok := n.(*syntax.Yield); ok && !env.TypeOf(y).IsUnknown() {
			t.Error("native yield used PHPDoc")
		}
		return true
	})
	checkVer(t, phpver.PHP55, false, `<?php function old() { yield 1; return; } t('old', old()->getReturn());`, map[string]string{"old": "?unknown"})
}

func TestGeneratorCompletionCrossFile(t *testing.T) {
	checkWith(t, map[string]string{"lib.php": `<?php function items() { yield 1; return 'done'; } final class Source { function items() { yield 1; return 2; } }`}, `<?php t('function', items()->getReturn()); t('method', (new Source())->items()->getReturn()); function delegate() { return yield from items(); } t('delegate', delegate()->getReturn());`, map[string]string{"function": "string", "method": "int", "delegate": "string"})
}
