<?php
function slugs($s)
{
    $a = <weak_warning descr="Use 'str_replace('-', '_', $s)' instead.">strtr($s /* subject */, '-', '_')</weak_warning>;
    $b = <weak_warning descr="Use 'str_replace('-', '_', $s)' instead.">strtr($s, /* from */ '-', '_')</weak_warning>;
    return [$a, $b];
}
