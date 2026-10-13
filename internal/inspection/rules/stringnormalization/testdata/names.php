<?php
namespace Text {
    function strtoupper($s) { return $s; }
    function rtrim($s) { return $s; }

    function tidy($code) {
        $r = [];
        $r[] = trim(strtoupper($code));
        $r[] = rtrim(\strtoupper($code));
        $r[] = <weak_warning descr="Cut first, then change the case: '\StrToUpper(\TRIM($code))'.">\TRIM(\StrToUpper($code))</weak_warning>;
        return $r;
    }
}

namespace {
    function tidy($code) {
        $r = [];
        $r[] = <weak_warning descr="Cut first, then change the case: 'StrToLower(LTrim($code))'.">LTrim(StrToLower($code))</weak_warning>;
        $r[] = strtolower(<weak_warning descr="The inner 'STRTOLOWER(...)' call has no effect here.">STRTOLOWER($code)</weak_warning>);
        return $r;
    }
}
