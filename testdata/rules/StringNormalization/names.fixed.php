<?php
namespace Text {
    function strtoupper($s) { return $s; }
    function rtrim($s) { return $s; }

    function tidy($code) {
        $r = [];
        $r[] = trim(strtoupper($code));
        $r[] = rtrim(\strtoupper($code));
        $r[] = \StrToUpper(\TRIM($code));
        return $r;
    }
}

namespace {
    function tidy($code) {
        $r = [];
        $r[] = StrToLower(LTrim($code));
        $r[] = strtolower($code);
        return $r;
    }
}
