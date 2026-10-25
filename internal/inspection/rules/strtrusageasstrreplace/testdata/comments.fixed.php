<?php
function slugs($s)
{
    $a = strtr($s /* subject */, '-', '_');
    $b = strtr($s, /* from */ '-', '_');
    return [$a, $b];
}
