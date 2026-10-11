<?php
// Before PHP 7.0 func_get_args() returns the values originally passed.
function argsReturned($label)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$label</weak_warning> .= '!';
    return func_get_args();
}
