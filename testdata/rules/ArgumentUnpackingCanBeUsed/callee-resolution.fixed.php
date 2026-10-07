<?php
namespace Shop {
    use function Tools\sum;
    use Vendor as Lib;

    function max(...$values) { return 0; }

    function run(array $args) {
        $a = \max(...$args);
        $b = \sum(...$args);
        $c = min(...$args);
        $d = \Lib\fmt(...$args);
        $e = \Util\fmt(...$args);
    }
}

namespace {
    $f = Util\fmt(...[1]);
}
