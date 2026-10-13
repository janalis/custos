<?php
namespace App {
    function strpos($h, $n) { return 0; }

    $a = strpos($url, 'https') === 0;
}

namespace {
    $b = <warning descr="Use 'strncmp($url, 'https', 5)' for a length-independent prefix check.">StrPos($url, 'https')</warning> === 0;
    $c = <warning descr="Use '\strncasecmp($url, 'ftp', 3)' for a length-independent prefix check.">\STRIPOS($url, 'ftp')</warning> !== 0;
}
