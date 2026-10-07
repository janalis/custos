<?php
namespace Shop {
    $id   = uniqid('', true);
    $ref  = \uniqid('ord_', true);
    $nam  = uniqid(prefix: 'q', more_entropy: true);
    $tags = array_map(function ($value) { return uniqid($value, true); }, ['a', 'b']);
    $one  = call_user_func_array(function ($value) { return uniqid($value, true); }, ['x']);

    function make($items)
    {
        $cb = 'uniqid';
        return array_filter($items, $cb);
    }
}

namespace Callbacks {
    $a = array_reduce([1, 2], 'uniqid', '');
    $b = array_walk($list, 'uniqid');
    $c = call_user_func(function ($value) { return uniqid($value, true); }, 'p_');
}
