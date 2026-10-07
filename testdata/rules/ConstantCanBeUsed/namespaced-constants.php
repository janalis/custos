<?php

namespace Build {
    const PHP_VERSION_ID = 1;
    const M_PI = 3;

    function check()
    {
        $a = version_compare(Meta\PHP_VERSION, '8.1.0', '>=');
        $b = strpos(Meta\PHP_OS, 'WIN') === 0;
        $c = <weak_warning descr="Replace with '\PHP_VERSION_ID >= 80100'.">version_compare(PHP_VERSION, '8.1.0', '>=')</weak_warning>;
        $d = <weak_warning descr="Use the \M_PI constant instead of this call.">pi()</weak_warning>;
        $e = <weak_warning descr="Use the PHP_SAPI constant instead of this call.">php_sapi_name()</weak_warning>;
        $f = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">stripos(\PHP_OS, 'DAR')</weak_warning> === 0;
    }
}

namespace Shadow {
    const PHP_VERSION = '0.0.1';
    use const Other\PHP_OS;

    $g = version_compare(PHP_VERSION, '8.1.0', '>=');
    $h = strpos(PHP_OS, 'WIN') === 0;
}
