<?php
function slug($title, $flag, $any) {
    $f = strtr($title, '--', '_');
    $g = strtr($title, '\r', '');
    $h = strtr($title, "\0", '');
    $i = strtr($title, ['-' => '_']);
    $j = strtr($title, '', '_');
    $k = strtr($title, $flag ? '-' : '+', '_');
    $l = strtr($title, $any, '_');
    $n = strtr($title, "\e", '_');
    $o = strtr($title, "$any", '_');
    $p = strtr($title, '-');
    return [$f, $g, $h, $i, $j, $k, $l, $m, $n, $o, $p];
}
$top = '-';
$q = strtr($title, $top, '_');
