<?php
function parts($path, $mode)
{
    $sep = '::';
    $a = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($path, '/')</weak_warning>;
    $b = <weak_warning descr="Needle has no letters; use 'strrpos(...)' instead.">strripos($path, '.', -2)</weak_warning>;
    $c = <weak_warning descr="Needle has no letters; use 'strstr(...)' instead.">\stristr($path, $sep)</weak_warning>;
    $d = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($path, "2024")</weak_warning>;
    return [$a, $b, $c, $d];
}
