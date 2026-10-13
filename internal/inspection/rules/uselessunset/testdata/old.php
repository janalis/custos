<?php
// Before PHP 7.0 func_get_args() returns the values originally passed.
function allArgs($first)
{
    <weak_warning descr="Unsetting a parameter only drops the local variable; this unset() is pointless.">unset($first);</weak_warning>
    return func_get_args();
}
