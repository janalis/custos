<?php

namespace Env {
    function phpversion(): string { return '1.0'; }
    function strpos($h, $n) { return 0; }

    function probe()
    {
        // Env\phpversion() and Env\strpos() are user functions.
        $own = phpversion();
        $os = strpos(PHP_OS, 'WIN') === 0;
        $sapi = PHP_SAPI;
        $ver = PHP_VERSION;
        $old = PHP_VERSION_ID < 70300;
        $win = \StrToUpper(SUBSTR(PHP_OS, 0, 3)) === 'WIN';
    }
}
