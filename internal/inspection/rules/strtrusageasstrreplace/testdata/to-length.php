<?php
function codes($s, $sep = '--') {
    $a = strtr($s, '$', 'USD');
    $b = strtr($s, '-', '');
    $c = strtr($s, '-', $sep);
    $d = strtr($s, "\n", '<br>');
    $e = strtr($s, 'é', 'e');
    $f = strtr($s, 'e', 'é');
    $g = strtr($s, '-', "\\\\");
    $h = <weak_warning descr="Use 'str_replace('-', &quot;\n&quot;, $s)' instead.">strtr($s, '-', "\n")</weak_warning>;
    $i = <weak_warning descr="Use 'str_replace('/', '\\', $s)' instead.">strtr($s, '/', '\\')</weak_warning>;
    $one = '_';
    $j = <weak_warning descr="Use 'str_replace(' ', $one, $s)' instead.">strtr($s, ' ', $one)</weak_warning>;
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j];
}
