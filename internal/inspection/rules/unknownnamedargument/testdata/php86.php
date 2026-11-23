<?php
// openssl_sign() and openssl_verify() gained $salt_length in PHP 8.6, which
// the bundled declarations do not list yet: their names are not checked.
openssl_sign($data, $signature, $key, salt_length: 32);
openssl_verify($data, $signature, $key, salt_length: 32);
// Other verified builtins keep their contract.
clamp(5, min: 1, max: 10);
clamp(5, 1, <error descr="Use a declared parameter name.">maximum: 10</error>);
// A named placeholder binds by name as an argument does.
function weigh(int $grams) {}
$w = weigh(<error descr="Use a declared parameter name.">gram: ?</error>);
$p = str_pad(<error descr="Use a declared parameter name.">strng: ?</error>, length: 4);
$q = weigh(grams: ?);
