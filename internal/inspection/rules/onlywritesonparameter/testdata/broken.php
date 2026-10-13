<?php
// Error recovery: parameters and imports without a name; a duplicated
// parameter name (rejected by PHP) is reported once.
function noName($)
{
    return function () use ($) {};
}

function twice($p, $p)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$p[]</weak_warning> = 1;
}
