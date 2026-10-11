<?php
function check($d, $s, $k, $skip) {
    if ($skip) { return true; } elseif (openssl_verify($d, $s, $k) === 1) { return true; }
    if ($skip) { return true; } else if (openssl_verify($d, $s, $k) === 1) { return true; }
    if (openssl_verify($d, $s, $k) !== 1) { return false; }
    if ((openssl_verify($d, $s, $k) !== 1)) { return false; }
    do { $skip = !$skip; } while (openssl_verify($d, $s, $k) === 1);
    do { $skip = !$skip; } while (openssl_verify($d, $s, $k) !== 1);
    while ((openssl_verify($d, $s, $k) === 1)) { break; }
    return false;
}
