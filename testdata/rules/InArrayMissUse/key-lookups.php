<?php
function lookups(array $map, int $id, string $name)
{
    return [
        // Safe: the needle cannot be confused with another key.
        <warning descr="Look the key up directly with 'array_key_exists($id, $map)'.">in_array($id, array_keys($map), true)</warning>,
        <warning descr="Look the key up directly with 'array_key_exists('admin', $map)'.">in_array('admin', array_keys($map), true)</warning>,
        <warning descr="Look the key up directly with 'array_key_exists('admin', $map)'.">in_array('admin', array_keys($map))</warning>,
        // '5' is stored as the int key 5; '1.0' == 1 loosely: no fix.
        <warning descr="Look the key up directly with 'array_key_exists('5', $map)'.">in_array('5', array_keys($map), true)</warning>,
        <warning descr="Look the key up directly with 'array_key_exists('1.0', $map)'.">in_array('1.0', array_keys($map))</warning>,
        <warning descr="Look the key up directly with 'array_key_exists($name, $map)'.">in_array($name, array_keys($map), true)</warning>,
        <warning descr="Look the key up directly with 'array_key_exists($id, $map)'.">in_array($id, array_keys($map))</warning>,
    ];
}
