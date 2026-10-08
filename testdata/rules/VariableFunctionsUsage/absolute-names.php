<?php
namespace Shop {
    use Vendor as Lib;
    use function Tools\sum;

    function max(...$values) { return 0; }
    function call_user_func($f, ...$args) { return null; }

    \call_user_func('max', $a, $b);          // shadowed by Shop\max (D7a)
    \call_user_func('\max', $a, $b);         // shadowed by Shop\max (D7a)
    \call_user_func('sum', $a);              // shadowed by use function (D7a)
    <weak_warning descr="Call it directly: 'min($a)'.">\call_user_func('min', $a)</weak_warning>;
    <weak_warning descr="Call it directly: '\Pkg\Tool::run($a)'.">\call_user_func('Pkg\Tool::run', $a)</weak_warning>;
    <weak_warning descr="Call it directly: '\Lib\fmt($a)'.">\call_user_func('Lib\fmt', $a)</weak_warning>;
    <weak_warning descr="Call it directly: '\Repo::find(1)'.">\call_user_func(['Repo', 'find'], 1)</weak_warning>;
    <weak_warning descr="Call it directly: '\Other::make()'.">\call_user_func(['Base', 'Other::make'])</weak_warning>;
    <weak_warning descr="Call it directly: 'self::make()'.">\call_user_func('self::make')</weak_warning>;
    <weak_warning descr="Pass the arguments inline: '\call_user_func($hook, $a)'.">\call_user_func_array($hook, [$a])</weak_warning>;
}

namespace {
    <weak_warning descr="Call it directly: 'Pkg\Tool::run($a)'.">call_user_func('Pkg\Tool::run', $a)</weak_warning>;
    <weak_warning descr="Call it directly: 'Repo::find(1)'.">call_user_func(['Repo', 'find'], 1)</weak_warning>;
}
