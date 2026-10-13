<?php
function checks($role, $list, $map, $o)
{
    $r = [];
    $r[] = in_array($role, []);
    $r[] = in_array($role, ['admin', 'owner']);
    $r[] = in_array($role, $list);
    $r[] = in_array($role, array_keys($map, 1));
    $r[] = in_array($role);
    $r[] = in_array($role, ['a'], true, 1);
    $r[] = in_array($role, [...$list]);
    $r[] = $o->in_array($role, ['a']);
    $r[] = in_array($role, $o->keys());
    return $r;
}
