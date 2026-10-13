<?php

namespace Build {
    const PHP_VERSION_ID = 1;
    const M_PI = 3;

    function check()
    {
        $a = version_compare(Meta\PHP_VERSION, '8.1.0', '>=');
        $b = strpos(Meta\PHP_OS, 'WIN') === 0;
        $c = \PHP_VERSION_ID >= 80100;
        $d = \M_PI;
        $e = PHP_SAPI;
        $f = stripos(\PHP_OS, 'DAR') === 0;
    }
}

namespace Shadow {
    const PHP_VERSION = '0.0.1';
    use const Other\PHP_OS;

    $g = version_compare(PHP_VERSION, '8.1.0', '>=');
    $h = strpos(PHP_OS, 'WIN') === 0;
}
