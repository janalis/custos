<?php
function bump(&$pos) { $pos++; }
function tail($first, &...$rest) {}
function plain($a) {}
class Cursor {
    private $len;
    public function walk(array $items) {
        $this->len = count($items);
        for ($j = 0; $j < $this->len; $j++) {   // E8: property limit decremented
            echo $items[$j];
            $this->len--;
        }
    }
    public static function shift(&$v) {}
    public function move(&$v) {}
}
function scan(array $tokens) {
    for ($k = 0; $k < count($tokens); $k++) {   // E8: counter skipped ahead
        if ($tokens[$k] === '(') { $k += 2; }
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: by-reference argument
        echo $tokens[$k];
        bump($k);
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: overwritten by a nested foreach
        echo $tokens[$k];
        foreach ([1, 2] as $k) {}
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: nested foreach key
        echo $tokens[$k];
        foreach ([1, 2] as $k => $unused) {}
    }
    for ($k = 0, $k = 1; $k < count($tokens); ++$k) {  // E8: second init write
        echo $tokens[$k];
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: destructuring
        echo $tokens[$k];
        [$k, $x] = [1, 2];
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: list() destructuring
        echo $tokens[$k];
        list($x, list($k)) = [1, [2]];
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: by-reference closure capture
        echo $tokens[$k];
        $f = function () use (&$k) { $k++; };
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: decrement
        echo $tokens[$k];
        $k--;
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: unset
        echo $tokens[$k];
        unset($k);
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: global
        echo $tokens[$k];
        global $k;
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: static
        echo $tokens[$k];
        static $k;
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: by-reference assignment source
        echo $tokens[$k];
        $alias = &$k;
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: variadic by-reference parameter
        echo $tokens[$k];
        tail(1, 2, $k);
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: named by-reference argument
        echo $tokens[$k];
        bump(pos: $k);
    }
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: static method by-reference argument
        echo $tokens[$k];
        Cursor::shift($k);
    }
    $c = new Cursor();
    for ($k = 0; $k < count($tokens); ++$k) {   // E8: method by-reference argument
        echo $tokens[$k];
        $c->move($k);
    }
}
function limitByRef(array $tokens) {
    $limit = count($tokens);
    for ($k = 0; $k < $limit; ++$k) {          // E8: preg_match by-reference argument
        echo $tokens[$k];
        preg_match('/x/', 'x', $limit);
    }
}
function limitForeach(array $tokens) {
    $limit = count($tokens);
    for ($k = 0; $k < $limit; ++$k) {          // E8: limit overwritten by a foreach
        echo $tokens[$k];
        foreach ($tokens as $limit) {}
    }
}
function limitInCondition(array $tokens) {
    $limit = count($tokens);
    for ($k = 0; $k < ($limit = count($tokens)); ++$k) {  // E8: write in the condition
        echo $tokens[$k];
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
function notWrites(array $tokens, $cb) {
    <warning descr="Iterate with foreach instead of a counter loop.">for</warning> ($k = 0; $k < count($tokens); ++$k) {
        echo $tokens[$k];
        unknownFn($k);                          // unresolved: not a write
        bump(other: $k);                        // no such parameter
        plain(1, $k);                           // extra argument, no variadic
        tail(1, ...$tokens);                    // unpacking
        $cb(...);                               // first-class callable
        [$a, [$b]] = [1, [2]];
        list($c) = [3];
        foreach ($tokens as $t) {}
        unset($t);
        static $s;
        $w = function () use (&$t) {};
    }
}
