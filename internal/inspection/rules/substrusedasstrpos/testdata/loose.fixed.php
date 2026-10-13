<?php
// Loose comparisons: numeric strings compare as numbers ("0 555" == "00"
// is false, but "0 " == "00" is true), so only non-numeric literals get a fix.
function dialPrefix(string $number, string $prefix) {
    $r = [];
    $r[] = substr($number, 0, 2) == '00';
    $r[] = substr($number, 0, 3) != '1e2';
    $r[] = substr($number, 0, strlen($prefix)) == $prefix;
    $r[] = strpos($number, 'tel:') === 0;
    $r[] = strpos($number, '00') === 0;
    return $r;
}
