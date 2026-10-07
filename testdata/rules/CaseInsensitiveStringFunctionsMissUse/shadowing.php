<?php
namespace App {
    function stripos($h, $n) { return 0; }

    $a = stripos($path, '/');
    $b = Other\strripos($path, '/');
}

namespace {
    $c = <weak_warning descr="Needle has no letters; use 'strpos(...)' instead.">StriPos($path, '/')</weak_warning>;
}
