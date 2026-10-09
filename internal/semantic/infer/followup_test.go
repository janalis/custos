package infer_test

import "testing"

func TestAssignmentConditionNarrows(t *testing.T) {
	checkAnywhere(t, `<?php
class Job {}
class Q { public function next(): ?Job { return null; } }
function a(Q $q) {
    while ($job = $q->next()) { t('while', $job); }
    if ($x = $q->next()) { t('if', $x); } else { t('else', $x); }
    if (!$y = $q->next()) { return; }
    t('guard', $y);
    if (($z = $q->next()) && true) { t('and', $z); }
}
`, map[string]string{"while": `\Job`, "if": `\Job`, "else": `\Job|null`, "guard": `\Job`, "and": `\Job`})
}

func TestConditionalAssertionsNegated(t *testing.T) {
	checkAnywhere(t, `<?php
class Failure {}
/** @phpstan-assert-if-true Failure $m */
function is_failure($m): bool { return $m instanceof Failure; }
/** @psalm-assert-if-false Failure $m */
function is_ok($m): bool { return true; }
/** @phpstan-assert-if-true !null $m */
function present($m): bool { return true; }
function a(array|Failure $m, array|Failure $n, ?int $p) {
    if (is_failure($m)) { return; }
    t('afterGuard', $m);
    if (is_ok($n)) { t('okTrue', $n); } else { t('okFalse', $n); }
    if (!present($p)) { t('absent', $p); }
}
`, map[string]string{"afterGuard": "array", "okTrue": "array", "okFalse": `\Failure`, "absent": "null"})
}

func TestArrayReduceInitialValue(t *testing.T) {
	checkAnywhere(t, `<?php
class Box {}
function a(array $xs) {
    t('box', array_reduce($xs, fn(Box $b, $x) => $b, new Box()));
    t('noInit', array_reduce($xs, fn($c, $x) => 1));
    t('mixedInit', array_reduce($xs, fn($c, $x) => 1, $GLOBALS['x']));
    t('unknownCb', array_reduce($xs, 'nope', 0));
}
`, map[string]string{"box": `\Box`, "noInit": "int|null", "mixedInit": "?unknown", "unknownCb": "?unknown"})
}

func TestAbsentShapeKey(t *testing.T) {
	checkAnywhere(t, `<?php
/** @param array{a: int, ...} $open */
function a(string $k, array $open) {
    t('unsealed', $open['b']);
    t('literal', ['#markup' => 'x']['#attached']);
    $a = ['#markup' => 'x'];
    t('var', $a['#attached']);
    t('present', $a['#markup']);
    $b = ['x' => 'y'];
    $b[$k] = 'z';
    t('computedWrite', $b['other']);
}
`, map[string]string{"unsealed": "?unknown", "literal": "?unknown", "var": "?unknown", "present": "string", "computedWrite": "string"})
}
