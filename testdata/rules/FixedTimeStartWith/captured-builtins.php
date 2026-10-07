<?php
namespace Urls {
    function strncmp($a, $b, $n) { return 0; }
    use function Other\strncasecmp;

    $a = <warning descr="Use '\strncmp($url, 'www.', 4)' for a length-independent prefix check.">strpos($url, 'www.')</warning> === 0;
    $b = <warning descr="Use '\strncasecmp($url, 'http', 4)' for a length-independent prefix check.">stripos($url, 'http')</warning> !== 0;
}
