<?php
function reviewValueCapture(string $x) {
    $y = gzencode($x);
    $read = function () use ($y, &$x) { return strlen($y); };
    $inner = function () { $y = 1; return function () use (&$y) { $y = 2; }; };
    return <warning descr="Match the decoder to the compression format.">gzinflate($y)</warning>;
}
$kept = gzcompress("sample payload", level: 9);
$peek = function () use ($kept) { return $kept; };
echo <warning descr="Match the decoder to the compression format.">gzdecode(($kept))</warning>;
