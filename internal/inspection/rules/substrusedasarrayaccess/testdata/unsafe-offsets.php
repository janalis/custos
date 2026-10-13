<?php
// Offsets whose offset-access form would not match substr(): no report.
function grab(string $word, int $back, string $pos, $loose, int|string $either)
{
    $a = substr($word, -3, 1);      // strlen-3 may be negative on short strings
    $b = substr($word, -$back, 1);
    $c = substr($word, $pos, 1);    // non-integer string offset
    $d = substr($word, $loose, 1);  // offset type unknown
    $e = substr($word, $either, 1);
    $f = substr($word, 1.5, 1);
    return [$a, $b, $c, $d, $e, $f];
}

function decremented(string $s, int $i) {
    return substr($s, --$i, 1);
}

// An assigned offset that other code can still change is not a known value.
function viaReferenceCapture(string $s) {
    $off = 0;
    $dec = function () use (&$off) { $off -= 2; };
    $dec();
    return substr($s, $off, 1);
}

function viaGlobal(string $s) {
    global $pos;
    $pos = 0;
    rewindPosition();
    return substr($s, $pos, 1);
}

function viaStatic(string $s) {
    static $cursor;
    $cursor = 0;
    moveCursor();
    return substr($s, $cursor, 1);
}

function viaReferenceParameter(string $s, int &$shared) {
    $shared = 0;
    rewindPosition();
    return substr($s, $shared, 1);
}

function viaReferenceItem(string $s) {
    $slot = 0;
    $box = ['k' => &$slot];
    shiftBox($box);
    return substr($s, $slot, 1);
}

function viaDecrement(string $s) {
    $back = 0;
    $back--;
    --$back;
    return substr($s, ($back), 1);
}

$topLevel = 'abc';
$cursor = 0;
rewindPosition();
echo substr($topLevel, $cursor, 1);
