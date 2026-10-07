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
