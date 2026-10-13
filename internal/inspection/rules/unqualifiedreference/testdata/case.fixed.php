<?php
namespace App {
    use function COUNT;

    function array_map($f, $a) { return $a; }

    $a = \IS_INT($x);
    $b = count([]);
    $c = array_map('is_int', []);
    $d = \call_user_func('\Is_Int', 1);
}
