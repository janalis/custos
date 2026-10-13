<?php
function walk($r, $f, $m)
{
    while (($r = $r->next()) !== false) {}
    if (($x = $f()) === null) {}
    $y = ($m ?: 1) == 0;
    $z = $m . 'b' === 'a';
    return $m === true ?? $f;
}
