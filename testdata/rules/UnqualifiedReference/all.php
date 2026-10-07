<?php
namespace {
    function legacy_hook() {}
}

namespace Shop\Cart {
    use function array_sum;
    use const E_ALL;

    function money($v) { return $v; }
    const SCALE = 2;

    $n = <weak_warning descr="Write '\count(...)' to allow compile-time binding.">count([1])</weak_warning>;
    $u = <weak_warning descr="Write '\ucfirst(...)' to allow compile-time binding.">ucfirst('x')</weak_warning>;
    $e = <weak_warning descr="Write '\PHP_EOL' to allow compile-time binding.">PHP_EOL</weak_warning>;
    \array_map(<weak_warning descr="Write '\legacy_hook' to allow compile-time binding.">'legacy_hook'</weak_warning>, []);
    \array_walk($list, <weak_warning descr="Write '\legacy_hook' to allow compile-time binding.">"legacy_hook"</weak_warning>);

    $ok = \count([1]) + money(1) + SCALE + array_sum([]) + E_ALL;
    $ok2 = true && null === __LINE__;
    \array_map('\legacy_hook', []);
    \array_map('Helper::run', []);
    \usort($list, 'legacy_hook');
    $ok3 = <weak_warning descr="Write '\StrLen(...)' to allow compile-time binding.">StrLen('x')</weak_warning> + namespace\strlen('x') + Sub\strlen('x');
}
