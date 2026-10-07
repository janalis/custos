<?php
namespace Routing {
    function substr($s, $o, $l) { return $s; }
    function strtoupper($s) { return $s; }

    function match($uri, $base) {
        $r = [];
        $r[] = substr($uri, 0, \strlen($base)) == $base;
        $r[] = strtoupper(\substr($uri, 0, 3)) === 'GET';
        $r[] = \substr($uri, 0, \Str\strlen($base)) == $base;
        $r[] = \strpos($uri, $base) === 0;
        return $r;
    }
}

namespace {
    function match($uri, $base) {
        $r = [];
        $r[] = mb_strpos($uri, $base) !== 0;
        $r[] = stripos($uri, 'GET') === 0;
        return $r;
    }
}
