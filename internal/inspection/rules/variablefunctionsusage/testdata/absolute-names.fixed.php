<?php
namespace Shop {
    use Vendor as Lib;
    use function Tools\sum;

    function max(...$values) { return 0; }
    function call_user_func($f, ...$args) { return null; }

    \call_user_func('max', $a, $b);          // shadowed by Shop\max (D7a)
    \call_user_func('\max', $a, $b);         // shadowed by Shop\max (D7a)
    \call_user_func('sum', $a);              // shadowed by use function (D7a)
    min($a);
    \Pkg\Tool::run($a);
    \Lib\fmt($a);
    \Repo::find(1);
    \Other::make();
    self::make();
    \call_user_func($hook, $a);
}

namespace {
    Pkg\Tool::run($a);
    Repo::find(1);
}
