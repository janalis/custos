<?php
namespace App {
    function in_array($n, $h) { return false; }
    function array_keys($a) { return []; }

    $a = in_array($role, ['a']);
    $b = \in_array($role, array_keys($map));
}

namespace {
    $c = <warning descr="Compare directly: '$role == 'a''.">In_Array($role, ['a'])</warning>;
    $d = <warning descr="Look the key up directly with 'array_key_exists($role, $map)'.">IN_ARRAY($role, Array_Keys($map))</warning>;
}
