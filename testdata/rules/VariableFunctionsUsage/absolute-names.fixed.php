<?php
namespace Shop {
    use Vendor as Lib;
    use function Tools\sum;

    function max(...$values) { return 0; }
    function call_user_func($f, ...$args) { return null; }

    \max($a, $b);
    \sum($a);
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
