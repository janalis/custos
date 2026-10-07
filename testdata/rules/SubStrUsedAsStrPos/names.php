<?php
namespace Routing {
    function substr($s, $o, $l) { return $s; }
    function strtoupper($s) { return $s; }

    function match($uri, $base) {
        $r = [];
        $r[] = substr($uri, 0, \strlen($base)) == $base;
        $r[] = strtoupper(\substr($uri, 0, 3)) === 'GET';
        $r[] = \substr($uri, 0, \Str\strlen($base)) == $base;
        $r[] = <weak_warning descr="Use '\strpos($uri, $base) === 0' instead.">\SUBSTR($uri, 0, StrLen($base)) == $base</weak_warning>;
        return $r;
    }
}

namespace {
    function match($uri, $base) {
        $r = [];
        $r[] = <weak_warning descr="Use 'mb_strpos($uri, $base) !== 0' instead.">Mb_SubStr($uri, 0, MB_STRLEN($base)) != $base</weak_warning>;
        $r[] = <weak_warning descr="Use 'stripos($uri, 'GET') === 0' instead.">STRTOUPPER(SubStr($uri, 0, 3)) === 'GET'</weak_warning>;
        return $r;
    }
}
