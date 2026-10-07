<?php
namespace Paths {
    function strpos($h, $n) { return 0; }
    use function Other\strstr;

    $a = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">stripos($path, '/')</weak_warning>;
    $b = <weak_warning descr="Needle has no letters; use 'strstr(...)' instead.">stristr($path, '/')</weak_warning>;
    $c = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">\stripos($path, '/')</weak_warning>;
    $d = <weak_warning descr="Needle has no letters; use 'strrpos(...)' instead.">strripos($path, '/')</weak_warning>;
}
