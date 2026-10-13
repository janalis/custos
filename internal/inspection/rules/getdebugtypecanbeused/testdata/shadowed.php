<?php
namespace Lib {
    function gettype($x) { return 'x'; }
    $a = is_object($v) ? get_class($v) : gettype($v);
}
namespace App {
    use function Lib\gettype;
    $b = is_object($v) ? get_class($v) : gettype($v);
    $c = <weak_warning descr="Use 'get_debug_type($v)' instead (scalar type names differ from gettype()).">is_object($v) ? get_class($v) : \gettype($v)</weak_warning>;
}
