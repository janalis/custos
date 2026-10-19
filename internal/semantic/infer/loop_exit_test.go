package infer_test

import "testing"

// Past a loop left only through its condition, the condition is false.
func TestLoopConditionFalseAfterLoop(t *testing.T) {
	checkWith(t, nil, `<?php
function next_(): ?int { return 1; }
function whileNull(?int $a) {
    while ($a === null) { $a = next_(); }
    t('while', $a);
}
function doWhile() {
    do { $a = next_(); } while (null === $a);
    t('do', $a);
}
function forLast(?int $a) {
    for ($i = 0; $a === null; $i++) { $a = next_(); }
    t('for', $a);
}
function fetch(\SplFileObject $f) {
    while (($line = $f->fgets()) !== '') { }
    t('fetch', $line);
}
`, map[string]string{"while": "int", "do": "int", "for": "int", "fetch": "string"})
}

// A break (or a goto out) leaves with the condition still true; a write
// after the loop or a for without condition gives nothing.
func TestLoopConditionKeptOpen(t *testing.T) {
	checkWith(t, nil, `<?php
function next_(): ?int { return 1; }
function withBreak(?int $a, bool $b) {
    while ($a === null) { if ($b) { break; } $a = next_(); }
    t('break', $a);
}
function withGoto(?int $a, bool $b) {
    while ($a === null) { if ($b) { goto out; } $a = next_(); }
    out:
    t('goto', $a);
}
function noCond(?int $a) {
    for (;;) { if ($a !== null) { return; } }
}
function nestedBreak(?int $a, int $m) {
    while ($a === null) { switch ($m) { case 1: break; } $a = next_(); }
    t('nested', $a);
}
`, map[string]string{"break": "int|null", "goto": "int|null", "nested": "int"})
}

// A value assigned in a loop and carried around it by a back edge did not
// pass the conditions outside the loop again.
func TestBackEdgeValuesSkipOuterConditions(t *testing.T) {
	checkWith(t, nil, `<?php
class Cell {}
class X {
    function next(): ?Cell { return null; }
    function outer(?Cell $w, int $k) {
        if (null !== $w) {
            foreach ([1, 2] as $i) {
                if ($i > $k) { $w = $this->next(); continue; }
                t('outer', $w);
            }
        }
    }
    function inner(?Cell $w, int $k) {
        foreach ([1, 2] as $i) {
            if (null !== $w) { t('inner', $w); }
            $w = $this->next();
        }
    }
    function step(?Cell $w) {
        if (null !== $w) {
            for ($i = 0; $i < 3; $w = $this->next(), $i++) { t('step', $w); }
        }
    }
}
`, map[string]string{"outer": `\Cell|null`, "inner": `\Cell`, "step": `\Cell|null`})
	checkAnywhere(t, `<?php
class Cell {}
function short(?Cell $c) { return is_null($c) ?: t('short', $c); }
`, map[string]string{"short": `\Cell`})
}
