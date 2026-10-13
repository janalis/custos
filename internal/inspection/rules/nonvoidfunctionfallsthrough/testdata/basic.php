<?php
function countItems(bool $ready): <error descr="Return a value on every reachable path.">int</error> { if ($ready) return 7; }

// The two tests leave zero uncovered.
function signOf(int $n): <error descr="Return a value on every reachable path.">int</error> { if ($n > 0) { return 1; } if ($n < 0) { return -1; } }
// The tested variable changes between the two branches.
function afterWrite($n, $m): <error descr="Return a value on every reachable path.">int</error> { if ($n > 0) { return 1; } $n = $m; if ($n <= 0) { return -1; } }
// Without a declared bool type, a third value passes both strict tests.
function looseFlag($flag): <error descr="Return a value on every reachable path.">int</error> { if ($flag === true) { return 1; } if ($flag === false) { return 0; } }
// One arm completes normally.
function partialMatch($e, $code): <error descr="Return a value on every reachable path.">int</error> { match ($code) { 1 => throw $e, default => null }; }
