<?php
function stored($d, $s, $k, $skip) {
    $ok = openssl_verify($d, $s, $k);
    $no = !(openssl_verify($d, $s, $k));
    if ($skip) { return $ok; } elseif (openssl_verify($d, $s, $k) === 1) { return true; }
    if (!(openssl_verify($d, $s, $k) === 1)) { return false; }
    do { $skip = !$skip; } while (openssl_verify($d, $s, $k) !== 1);
    if ($skip && openssl_verify($d, $s, $k)) { return $no; }
    if ($skip) { return true; } elseif ($d) { return (openssl_verify($d, $s, $k)); }
    return -openssl_verify($d, $s, $k);
}
