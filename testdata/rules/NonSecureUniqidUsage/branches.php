<?php
namespace Ids {
    function uniqid($prefix = '') { return $prefix . '1'; }
    $own = uniqid('x');
}

namespace {
    $f = uniqid(...);
    $s = <error descr="Pass more_entropy = true to uniqid() to reduce collisions.">uniqid(...$args)</error>;
    $p = array_filter($items, ...);
    $n = array_map(123, $items);
}
