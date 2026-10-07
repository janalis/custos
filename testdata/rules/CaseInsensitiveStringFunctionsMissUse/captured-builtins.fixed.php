<?php
namespace Paths {
    function strpos($h, $n) { return 0; }
    use function Other\strstr;

    $a = \strpos($path, '/');
    $b = \strstr($path, '/');
    $c = \strpos($path, '/');
    $d = strrpos($path, '/');
}
