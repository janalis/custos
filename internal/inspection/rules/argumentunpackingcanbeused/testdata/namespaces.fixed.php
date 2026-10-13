<?php
namespace Lib {
    function call_user_func_array($fn, $args) { return [$fn, $args]; }

    function local(array $args) {
        $a = call_user_func_array('max', $args);
        $b = \Lib\call_user_func_array('max', $args);
        $c = namespace\call_user_func_array('max', $args);
        $d = max(...$args);
    }
}

namespace App {
    use function Lib\call_user_func_array;

    function imported(array $args) {
        return call_user_func_array('min', $args);
    }
}

namespace Other {
    function fallback(array $args) {
        $a = Lib\call_user_func_array('min', $args);
        return min(...$args);
    }
}
