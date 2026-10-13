<?php
namespace Shop {
    use function Tools\sum;
    use Vendor as Lib;

    function max(...$values) { return 0; }

    function run(array $args) {
        $a = <warning descr="Call '\max(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('max', $args)</warning>;
        $b = <warning descr="Call '\sum(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('sum', $args)</warning>;
        $c = <warning descr="Call 'min(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('min', $args)</warning>;
        $d = <warning descr="Call '\Lib\fmt(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('Lib\fmt', $args)</warning>;
        $e = <warning descr="Call '\Util\fmt(...$args)' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('Util\fmt', $args)</warning>;
    }
}

namespace {
    $f = <warning descr="Call 'Util\fmt(...[1])' directly using argument unpacking (wrap with array_values() when keys are not sequential).">call_user_func_array('Util\fmt', [1])</warning>;
}
