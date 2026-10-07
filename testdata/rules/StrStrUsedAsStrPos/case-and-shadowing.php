<?php
namespace App {
    function strstr($h, $n) { return ''; }

    if (strstr($text, $word)) {}
}

namespace {
    $j = <warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">StrStr($text, $word)</warning> ? 1 : 0;
    $k = <warning descr="Use '\stripos($text, $word) === false' instead; it avoids building a substring.">\STRISTR($text, $word) === false</warning>;
}
