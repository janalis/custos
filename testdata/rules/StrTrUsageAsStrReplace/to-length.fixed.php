<?php
function codes($s, $sep = '--') {
    $a = strtr($s, '$', 'USD');
    $b = strtr($s, '-', '');
    $c = strtr($s, '-', $sep);
    $d = strtr($s, "\n", '<br>');
    $e = strtr($s, 'é', 'e');
    $f = strtr($s, 'e', 'é');
    $g = strtr($s, '-', "\\\\");
    $h = str_replace('-', "\n", $s);
    $i = str_replace('/', '\\', $s);
    $one = '_';
    $j = str_replace(' ', $one, $s);
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j];
}
