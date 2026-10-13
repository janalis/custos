<?php
function glued($x, $b)
{
    if ("a"==$x and $b) {}
    return null==$x;
}

function ordered($i, $o)
{
    $a = $i++ == "$i";
    $b = ($i = $o->id()) . '' === "$i";
    return $o->pos == "{$o->next()}";
}
