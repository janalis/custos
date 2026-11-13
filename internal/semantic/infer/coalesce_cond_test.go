package infer_test

import "testing"

// A test of `$x['k'] ?? null` (truthy, or identical to a literal) shows
// the element was set, so the array is not null.
func TestNullDefaultedChainNarrowsBase(t *testing.T) {
	checkAnywhere(t, `<?php
class Q {
    private ?array $config = null;
    function f(?array $rows) {
        if ($rows['k'] ?? null) {
            t('truthy', $rows);
        }
        if (($rows['driver'] ?? null) === 'cloud') {
            t('identical', $rows);
        }
        if (($rows['driver'] ?? null) !== 'sqs') {
            return;
        }
        t('afterExit', $rows);
        if (($this->config['driver'] ?? null) === 'cloud') {
            t('property', $this->config);
        }
        if (($rows['driver'] ?? null) == 'x') {
            t('loose', $rows);
        }
    }
    function g(?array $rows, ?array $flags, ?array $other, $v) {
        if (($rows['driver'] ?? 'x') === 'x') {
            t('otherDefault', $rows);
        }
        if (($flags['on'] ?? null) === true) {
            t('true', $flags);
        }
        if (($other['on'] ?? null) === $v) {
            t('variable', $other);
        }
    }
}
`, map[string]string{"truthy": "non-empty array", "identical": "non-empty array", "afterExit": "non-empty array", "property": "non-empty array", "loose": "non-empty array", "otherDefault": "array|null", "true": "non-empty array", "variable": "array|null"})
}

// Storing `f() ?? []` in a property leaves it non-null, whatever f()
// returns.
func TestCoalescedPropertyWriteIsNonNull(t *testing.T) {
	checkAnywhere(t, `<?php
function cfg(): mixed { return null; }
class Q {
    private ?array $config = null;
    function f() {
        $this->config ??= cfg() ?? [];
        t('coalesceAssign', $this->config);
    }
    function g() {
        $this->config = cfg() ?? [];
        t('assign', $this->config);
    }
    function h() {
        $this->config = cfg() ?? null;
        t('nullDefault', $this->config);
    }
}
`, map[string]string{"coalesceAssign": "array", "assign": "array", "nullDefault": "array|null"})
}
