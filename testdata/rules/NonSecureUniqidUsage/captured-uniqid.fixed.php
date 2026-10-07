<?php
namespace Ids {
    function uniqid($prefix = '') { return $prefix; }

    $a = array_map(function ($value) { return \uniqid($value, true); }, ['a']);
    $b = uniqid('own');
}

namespace Plain {
    $c = array_map(function ($value) { return uniqid($value, true); }, ['b']);
}
