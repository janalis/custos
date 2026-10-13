<?php
function split_path($path)
{
    $a = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">STRIPOS($path, '/')</weak_warning>;
    $b = <weak_warning descr="Needle has no letters; use 'strrpos(...)' instead.">\StrRiPos($path, '.')</weak_warning>;
    $c = <weak_warning descr="Needle has no letters; use 'strstr(...)' instead.">Stristr($path, '::')</weak_warning>;
    $d = STRIPOS($path, 'x');
    return [$a, $b, $c, $d];
}
