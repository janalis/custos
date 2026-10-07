<?php
namespace App {
    function in_array($n, $h) { return false; }
    function array_keys($a) { return []; }

    $a = in_array($role, ['a']);
    $b = \in_array($role, array_keys($map));
}

namespace {
    $c = $role == 'a';
    $d = array_key_exists($role, $map);
}
