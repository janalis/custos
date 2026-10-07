<?php
function scan($text, $word, $flag) {
    if (<warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>) {}
    if (<warning descr="Use 'stripos($text, '#') === false' instead; it avoids building a substring.">!stristr($text, '#')</warning>) {}
    while ($flag and <warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>) { $flag = false; }
    $a = <warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning> ? 1 : 2;
    $b = !(<warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>);
    $c = <warning descr="Use '\strpos($text, $word) === false' instead; it avoids building a substring.">FALSE == \strstr($text, $word, true)</warning>;
    $d = <warning descr="Use 'stripos($text, $word) !== false' instead; it avoids building a substring.">stristr($text, $word) != false</warning>;
    $e = <warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word) <> false</warning>;
    do {} while ((<warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>) || $flag);
    if ($flag) {} elseif (<warning descr="Use 'strpos($text, $word) !== false' instead; it avoids building a substring.">strstr($text, $word)</warning>) {}
    return [$a, $b, $c, $d, $e];
}
