package infer_test

import "testing"

// An if without else whose body replaces the variable narrows it past the
// if by the failed condition only when the new value fails it too: a value
// the condition could never test (a string against `!== null`) is kept.
func TestGuardOverwriteKeepsNewValue(t *testing.T) {
	checkAnywhere(t, `<?php
/** @return array<string, int> */
function counts(): array { return []; }
function flag(): bool { return true; }
class Cache {
    private ?array $counts = null;
    private ?bool $flag = null;
    public function read() {
        if (is_null($this->counts)) {
            $this->counts = counts();
        }
        t('elements', $this->counts);
        if ($this->flag === null) {
            $this->flag = true;
        }
        t('bool', $this->flag);
    }
}
function f(?string $a, ?string $b, ?string $c, string|false $d) {
    if ($a !== null) {
        $a = strtolower($a);
    }
    t('notNull', $a);
    if (null !== $b) {
        $b = 'x' . $b;
    }
    t('yoda', $b);
    if ($c === null) {
        $c = strtolower('X');
    }
    t('isNull', $c);
    if (false === $d) {
        $d = strtolower('X');
    }
    t('isFalse', $d);
}
`, map[string]string{
		"notNull": "null|string", "yoda": "null|string", "isNull": "string", "isFalse": "string",
		"elements": "array", "bool": "bool",
	})
}
