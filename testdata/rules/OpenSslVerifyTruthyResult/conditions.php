<?php
function check($d, $s, $k, $skip) {
    if ($skip) { return true; } elseif (<warning descr="Accept only verification result 1.">openssl_verify($d, $s, $k)</warning>) { return true; }
    if ($skip) { return true; } else if (<warning descr="Accept only verification result 1.">openssl_verify($d, $s, $k)</warning>) { return true; }
    if (!(<warning descr="Accept only verification result 1.">openssl_verify($d, $s, $k)</warning>)) { return false; }
    if ((!((<warning descr="Accept only verification result 1.">openssl_verify($d, $s, $k)</warning>)))) { return false; }
    do { $skip = !$skip; } while (<warning descr="Accept only verification result 1.">openssl_verify($d, $s, $k)</warning>);
    do { $skip = !$skip; } while (!<warning descr="Accept only verification result 1.">openssl_verify($d, $s, $k)</warning>);
    while ((<warning descr="Accept only verification result 1.">openssl_verify($d, $s, $k)</warning>)) { break; }
    return false;
}
