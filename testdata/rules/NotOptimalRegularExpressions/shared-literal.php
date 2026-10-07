<?php
function shared(string $s)
{
    $re = <error descr="The /e flag was removed from PCRE; use a callback replacement.">'/x+/e'</error>;
    preg_match($re, $s);
    preg_replace($re, 'y', $s);
    return preg_split($re, $s);
}

function quoted(string $s)
{
    $re = <warning descr="Pattern has no valid delimiters.">'k\d+'</warning>;
    preg_quote($re, '/');
    return preg_match($re, $s) + preg_match($re, $s . 'x');
}
