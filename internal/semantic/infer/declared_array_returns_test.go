package infer_test

import "testing"

// A callable declared to return a plain array, without @return, returns
// what its body returns when that is an array.
func TestDeclaredArrayReturnRefinedByBody(t *testing.T) {
	checkShape(t, `<?php
function rows(): array { return [['id' => 1]]; }
function maybe(bool $c): ?array { return $c ? ['a' => 1] : null; }
function strict(bool $c): array { if ($c) { return ['k' => 'v']; } }
function scalar(): array { return $GLOBALS['x']; }
/** @return list<string> */
function documented(): array { return [1]; }
function mixedBody(bool $c): array { return $c ? [1] : 'x'; }
final class Repo {
    public function all(): array { return ['n' => 1]; }
    private function own(): array { return ['p' => true]; }
    public function f() { t('own', $this->own()); }
}
class Open { public function all(): array { return ['n' => 1]; } }
t('rows', rows()); t('maybe', maybe(true)); t('strict', strict(true)); t('scalar', scalar());
t('documented', documented()); t('mixedBody', mixedBody(true));
t('final', (new Repo())->all()); t('open', (new Open())->all());
$cb = 'rows'; t('callable', call_user_func('rows'));
`, map[string]string{
		"rows": "int[][]<int[]<int<1, 1>>{id: int<1, 1>}>{0: int[]<int<1, 1>>{id: int<1, 1>}}", "maybe": "int[]|null<int<1, 1>>{a: int<1, 1>}", "strict": "string[]<string<len 1, 1>>{k: string<len 1, 1>}",
		"scalar": "array", "documented": "string[]", "mixedBody": "array", "final": "int[]<int<1, 1>>{n: int<1, 1>}",
		"open": "array", "own": "true[]{p: true}", "callable": "int[][]<int[]<int<1, 1>>{id: int<1, 1>}>{0: int[]<int<1, 1>>{id: int<1, 1>}}",
	})
}
