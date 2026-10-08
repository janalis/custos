<?php
function checks($role, $list, $map, $a, $b)
{
    $r = [];
    $r[] = $role == 'admin';
    $r[] = $role === 'admin';
    $r[] = $role == 'admin';
    $r[] = $role != 'admin';
    $r[] = $role !== 'admin';
    $r[] = $role == 'admin';
    $r[] = $role != 'admin';
    $r[] = $b || $role == 'admin';
    $r[] = ($a . $b) == 'ab';
    $r[] = ($a ? $b : 0) === 3;
    $r[] = ($a ?? 0) != 7;
    $r[] = ($a == 1) . 'x';
    $r[] = !in_array($role, array_keys($map), true);
    $r[] = array_key_exists('id', $map);
    return $r;
}
