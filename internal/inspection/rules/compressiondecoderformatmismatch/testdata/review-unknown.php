<?php
function reviewSpread(array $args) {
    return gzuncompress(gzencode(...$args));
}
function reviewCapture(string $x) {
    $y = gzencode($x);
    $raw = function () use (&$y, $x) { $y = gzdeflate($x); };
    $raw();
    return gzinflate($y);
}
function reviewOwnCapture(string $x, $y) {
    return function () use (&$y, $x) {
        $y = gzencode($x);
        refresh();
        return gzinflate($y);
    };
}
function reviewCopiedCapture(string $x) {
    $y = gzencode($x);
    $raw = function () use (&$y, $x) { $y = gzdeflate($x); };
    $raw();
    $z = $y;
    $z = $z;
    return gzinflate($z);
}
$top = gzencode("sample payload");
$swap = function () use (&$top) { $top = gzdeflate("sample payload"); };
$swap();
echo gzinflate($top);
echo gzuncompress(gzencode("sample payload", 9, ...$rest));
