<?php
function lookups(array $map, int $id, string $name)
{
    return [
        // Safe: the needle cannot be confused with another key.
        array_key_exists($id, $map),
        array_key_exists('admin', $map),
        array_key_exists('admin', $map),
        // '5' is stored as the int key 5; '1.0' == 1 loosely: no fix.
        in_array('5', array_keys($map), true),
        in_array('1.0', array_keys($map)),
        in_array($name, array_keys($map), true),
        in_array($id, array_keys($map)),
    ];
}
