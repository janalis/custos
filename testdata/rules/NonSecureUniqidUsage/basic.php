<?php
namespace Shop {
    $id   = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">uniqid()</error>;
    $ref  = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">\uniqid('ord_')</error>;
    $nam  = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">uniqid(prefix: 'q')</error>;
    $tags = array_map(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">"uniqid"</error>, ['a', 'b']);
    $one  = call_user_func_array(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'\\uniqid'</error>, ['x']);

    function make($items)
    {
        $cb = 'uniqid';
        return array_filter($items, <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">$cb</error>);
    }
}

namespace Callbacks {
    $a = array_reduce([1, 2], <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, '');
    $b = array_walk($list, <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>);
    $c = call_user_func(<error descr="Pass more_entropy = true to uniqid() to reduce collisions.">'uniqid'</error>, 'p_');
}
