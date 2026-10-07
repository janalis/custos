<?php
function checks($role, $list, $map, $a, $b)
{
    $r = [];
    $r[] = <warning descr="Compare directly: '$role == 'admin''.">in_array($role, ['admin'])</warning>;
    $r[] = <warning descr="Compare directly: '$role === 'admin''.">in_array($role, array('admin'), TRUE)</warning>;
    $r[] = <warning descr="Compare directly: '$role == 'admin''.">in_array($role, ['k' => 'admin'], false)</warning>;
    $r[] = <warning descr="Compare directly: '$role != 'admin''.">!in_array($role, ['admin'])</warning>;
    $r[] = <warning descr="Compare directly: '$role !== 'admin''.">in_array($role, ['admin'], true) == false</warning>;
    $r[] = <warning descr="Compare directly: '$role == 'admin''.">false !== in_array($role, ['admin'])</warning>;
    $r[] = <warning descr="Compare directly: '$role != 'admin''.">true != in_array($role, ['admin'])</warning>;
    $r[] = $b || <warning descr="Compare directly: '$role == 'admin''.">in_array($role, ['admin'])</warning>;
    $r[] = <warning descr="Compare directly: '($a . $b) == 'ab''.">in_array($a . $b, ['ab'])</warning>;
    $r[] = <warning descr="Compare directly: '($a ? $b : 0) === 3'.">in_array($a ? $b : 0, [3], true)</warning>;
    $r[] = <warning descr="Compare directly: '($a ?? 0) != 7'.">in_array($a ?? 0, [7]) !== true</warning>;
    $r[] = <warning descr="Compare directly: '$a == 1'.">in_array($a, [1])</warning> . 'x';
    $r[] = !<warning descr="Look the key up directly with 'array_key_exists($role, $map)'.">in_array($role, array_keys($map), true)</warning>;
    $r[] = <warning descr="Look the key up directly with 'array_key_exists('id', $map)'.">\in_array('id', array_keys($map))</warning>;
    return $r;
}
