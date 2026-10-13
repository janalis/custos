<?php
function lookups74(array $map)
{
    // Before PHP 8 'admin' == 0 is true: no fix for the loose form.
    return <warning descr="Look the key up directly with 'array_key_exists('admin', $map)'.">in_array('admin', array_keys($map))</warning>;
}
