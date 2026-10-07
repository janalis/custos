<?php
function os_checks() {
    $win  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">stripos(PHP_OS, 'win')</weak_warning> === 0;
    $dar  = false !== <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">mb_strpos(PHP_OS, 'Darwin')</weak_warning>;
    $lin  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">strncasecmp(PHP_OS, 'LIN', 3)</weak_warning> == 0;
    $bsd  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">mb_strtoupper(mb_substr(PHP_OS, 0, 3))</weak_warning> != 'BSD';
    $sun  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">substr(PHP_OS, 0, 3)</weak_warning> === 'Sun';
    $neg  = <weak_warning descr="Compare PHP_OS_FAMILY instead of sniffing PHP_OS.">strpos(PHP_OS, 'X')</weak_warning> !== -1;

    // not reported
    $p1 = (strpos(PHP_OS, 'WIN')) === 0;
    $p2 = strpos(PHP_OS, 'WIN') > 0;
    $p3 = strpos(PHP_OS, 'WIN') === $offset;
    $p4 = substr(PHP_OS, 0, 3) === $prefix;
    $p5 = strpos(strtoupper(PHP_OS), 'WIN') === 0;
    $p6 = str_starts_with(PHP_OS, 'WIN') === true;
    $p7 = substr(PHP_OS, 0, 3) === 0;
    $p8 = strpos(PHP_OS, 'WIN') === 'x';
    $p9 = strpos(PHP_OS, 'WIN') <> 0;
}
