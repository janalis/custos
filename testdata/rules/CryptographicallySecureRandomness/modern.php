<?php
function make() {
    $a = <weak_warning descr="Use random_bytes() for cryptographically secure randomness."><error descr="Pass a second argument to learn whether a strong algorithm was used.">openssl_random_pseudo_bytes</error></weak_warning>(16);
    return \<weak_warning descr="Use random_bytes() for cryptographically secure randomness.">mcrypt_create_iv</weak_warning>(16, MCRYPT_DEV_RANDOM);
}
$top = <weak_warning descr="Use random_bytes() for cryptographically secure randomness.">openssl_random_pseudo_bytes</weak_warning>(4, $s);
openssl_random_pseudo_bytes(4, $s, 1);
$x->openssl_random_pseudo_bytes(4);
