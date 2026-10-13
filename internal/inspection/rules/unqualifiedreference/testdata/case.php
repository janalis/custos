<?php
namespace App {
    use function COUNT;

    function array_map($f, $a) { return $a; }

    $a = <weak_warning descr="Write '\IS_INT(...)' to allow compile-time binding.">IS_INT($x)</weak_warning>;
    $b = count([]);
    $c = array_map('is_int', []);
    $d = \call_user_func(<weak_warning descr="Write '\Is_Int' to allow compile-time binding.">'Is_Int'</weak_warning>, 1);
}
