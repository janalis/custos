<?php
function verify(string $token, string $expected, array $row) {
    $a = <error descr="Both compared strings are the same expression; one of them is probably wrong.">hash_equals($token, $token)</error>;
    $b = <error descr="Both compared strings are the same expression; one of them is probably wrong.">\strcasecmp($row['name'], $row[ 'name' ])</error>;
    $c = <error descr="Both compared strings are the same expression; one of them is probably wrong.">substr_compare($token, $token, 0, 4)</error>;
    $d = strncmp($token, $expected, 4);
    $e = strnatcmp($token, ($token));
    return [$a, $b, $c, $d, $e];
}
