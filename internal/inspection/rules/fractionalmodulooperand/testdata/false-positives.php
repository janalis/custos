<?php
function viaReferenceCapture(int $n): int {
    $width = 1.5;
    $widen = function () use (&$width) { $width = (int) ceil($width); };
    $widen();
    return $n % $width;
}
function viaChainedCapture(int $n): int {
    $base = 2.5;
    $widen = function () use (&$base) { $base = 3; };
    $widen();
    $width = $base;
    return $n % $width;
}
function viaExtract(int $n, array $settings): int {
    $width = 0.5;
    extract($settings);
    return $n % $width;
}
function viaParseStr(int $n, string $query): int {
    $width = 0.5;
    parse_str($query);
    return $n % ($width);
}
function viaLoop(int $n, array $settings): int {
    $width = 0.5;
    $sum = 0;
    while ($n-- > 0) {
        $sum += $n % $width;
        extract($settings);
    }
    return $sum;
}
function resetWidth(): void {
    global $width;
    $width = 4;
}
$width = 2.5;
resetWidth();
echo 10 % $width;
$depth = 2.5;
$registry = new Registry();
echo 10 % $depth;
$height = 2.5;
$registry->reset();
echo 10 % $height;
$length = 2.5;
Registry::reset();
echo 10 % $length;
