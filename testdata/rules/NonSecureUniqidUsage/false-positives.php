<?php
namespace Shop {
    $ok1 = uniqid('', true);
    $ok2 = uniqid(more_entropy: true, prefix: 'z');
    $ok3 = call_user_func('uniqid');
    $ok4 = array_map('strtoupper', ['a']);
    $ok5 = uniqid($p, false);
    $ok6 = array_map(fn($v) => uniqid($v, true), ['a']);
}

namespace Custom {
    function uniqid() { return 'x'; }
    $mine = uniqid();
    $mapped = array_map('Custom\\uniqid', ['a']);
}

namespace Wrapped {
    function array_map($cb, $items) { return $items; }
    $own = array_map('uniqid', ['a']);
}
