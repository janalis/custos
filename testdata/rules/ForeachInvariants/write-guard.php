<?php
function bump(&$pos) { $pos++; }
class Cursor {
    private $len;
    public function walk(array $items) {
        $this->len = count($items);
        for ($j = 0; $j < $this->len; $j++) {   // E8: property limit decremented
            echo $items[$j];
            $this->len--;
        }
    }
}
function scan(array $tokens) {
    $limit = count($tokens);
    for ($k = 0; $k < $limit; $k++) {          // E8: counter skipped ahead
        if ($tokens[$k] === '(') { $k += 2; }
    }
    for ($k = 0; $k < $limit; $k++) {          // E8: limit shrinks
        if ($tokens[$k] === '') { $limit--; }
    }
    for ($k = 0; $k < $limit; ++$k) {          // E8: by-reference argument
        echo $tokens[$k];
        bump($k);
    }
    for ($k = 0; $k < $limit; ++$k) {          // E8: overwritten by a nested foreach
        echo $tokens[$k];
        foreach ([1, 2] as $k) {}
    }
    for ($k = 0, $k = 1; $k < $limit; ++$k) {  // E8: second init write
        echo $tokens[$k];
    }
    for ($k = 0; $k < $limit; ++$k) {          // E8: destructuring
        echo $tokens[$k];
        [$k, $x] = [1, 2];
    }
    for ($k = 0; $k < $limit; ++$k) {          // E8: by-reference closure capture
        echo $tokens[$k];
        $f = function () use (&$k) { $k++; };
    }
    for ($k = 0; $k < $limit; ++$k) {          // E8: preg_match by-reference argument
        echo $tokens[$k];
        preg_match('/x/', 'x', $limit);
    }
}
function ok(array $tokens) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($k = 0, $size = count($tokens); $k < $size; ++$k) {
        echo $tokens[$k];                      // init assignment of the limit is standard
        $g = fn() => $k + 1;                    // arrow functions never write
        $h = function () use ($k) { $k++; };    // by-value capture is not a write
        strlen($k);
    }
}
