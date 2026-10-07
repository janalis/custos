<?php
namespace {
    function legacy_hook() {}
}

namespace Shop\Cart {
    use function array_sum;
    use const E_ALL;

    function money($v) { return $v; }
    const SCALE = 2;

    $n = \count([1]);
    $u = \ucfirst('x');
    $e = \PHP_EOL;
    \array_map('\legacy_hook', []);
    \array_walk($list, "\\legacy_hook");

    $ok = \count([1]) + money(1) + SCALE + array_sum([]) + E_ALL;
    $ok2 = true && null === __LINE__;
    \array_map('\legacy_hook', []);
    \array_map('Helper::run', []);
    \usort($list, 'legacy_hook');
    $ok3 = \StrLen('x') + namespace\strlen('x') + Sub\strlen('x');
}
