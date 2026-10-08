<?php
// Loose comparisons: numeric strings compare as numbers ("0 555" == "00"
// is false, but "0 " == "00" is true), so only non-numeric literals get a fix.
function dialPrefix(string $number, string $prefix) {
    $r = [];
    $r[] = <weak_warning descr="Use 'strpos($number, '00') === 0' instead.">substr($number, 0, 2) == '00'</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($number, '1e2') !== 0' instead.">substr($number, 0, 3) != '1e2'</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($number, $prefix) === 0' instead.">substr($number, 0, strlen($prefix)) == $prefix</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($number, 'tel:') === 0' instead.">substr($number, 0, 4) == 'tel:'</weak_warning>;
    $r[] = <weak_warning descr="Use 'strpos($number, '00') === 0' instead.">substr($number, 0, 2) === '00'</weak_warning>;
    return $r;
}
