<?php

namespace Env {
    function phpversion(): string { return '1.0'; }
    function strpos($h, $n) { return 0; }

    function probe()
    {
        // Env\phpversion() and Env\strpos() are user functions.
        $own = phpversion();
        $os = strpos(PHP_OS, 'WIN') === 0;
        $sapi = <weak_warning descr="Use the PHP_SAPI constant instead of this call.">PHP_SAPI_NAME()</weak_warning>;
        $ver = <weak_warning descr="Use the PHP_VERSION constant instead of this call.">\PhpVersion()</weak_warning>;
        $old = <weak_warning descr="Replace with 'PHP_VERSION_ID < 70300'.">Version_Compare(PHP_VERSION, '7.3.0', '<')</weak_warning>;
        $win = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">\StrToUpper(SUBSTR(PHP_OS, 0, 3))</weak_warning> === 'WIN';
    }
}
