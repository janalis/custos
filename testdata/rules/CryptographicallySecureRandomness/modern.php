<?php
function make() {
    // From PHP 7.4 the strength flag is always true: neither D2 nor D5.
    $a = <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>(16);
    $b = <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>(16, $strong);
    return \<weak_warning descr="Use random_bytes() for cryptographically secure randomness.">mcrypt_create_iv</weak_warning>(16, MCRYPT_DEV_RANDOM) . $a . $b . <weak_warning descr="Use random_bytes() for cryptographically secure randomness."><error descr="Pass the entropy source explicitly; its default differs between PHP versions.">mcrypt_create_iv</error></weak_warning>(4);
}
$top = <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>(4, $s);
openssl_random_pseudo_bytes(4, $s, 1);
$x->openssl_random_pseudo_bytes(4);
