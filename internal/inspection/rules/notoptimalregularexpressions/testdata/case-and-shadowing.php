<?php
namespace App {
    function preg_quote($s, $d = null) { return $s; }

    preg_quote('#a.b#');
}

namespace {
    <warning descr="Pass the delimiter to preg_quote() so it is escaped as well.">PREG_QUOTE('#a.b#')</warning>;
    Preg_Match('/^[a-z]+$/', <warning descr="Drop the case conversion and add the /i flag instead.">StrToLower($name)</warning>);
}
