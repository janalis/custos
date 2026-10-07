<?php
namespace Lib {
    function mktime() { return 0; }
    $a = mktime();
}

namespace App {
    use function Lib\mktime;

    function values($o)
    {
        $b = mktime();
        $c = \gmmktime(12);
        $d = \gmmktime(1, 2, 3, 4, 5, 2001);
        $e = \gmmktime(1, 2, 3, 4, 5, 2001, 0, 1);
        $f = MkTime();
        $g = $o->mktime();
        $h = Clock::gmmktime();
    }
}
