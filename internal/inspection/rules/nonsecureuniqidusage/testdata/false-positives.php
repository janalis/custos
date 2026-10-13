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

namespace Forwarded {
    // more_entropy already reaches uniqid() through the callback-taking call.
    $a = call_user_func('uniqid', 'p_', true);
    $b = call_user_func_array('uniqid', ['p_', true]);
    $c = call_user_func_array('uniqid', ['more_entropy' => true]);
    $d = call_user_func('uniqid', more_entropy: true);
    $e = array_map('uniqid', $prefixes, $flags);
    $f = call_user_func_array(callback: 'uniqid', args: ['p_', true]);
    $g = call_user_func_array('uniqid', [...$parts, 'p_', true]);
    $h = array_map(array: $l, cb: 'uniqid');
    $i = array_map(...$spread, ...$more);
}
