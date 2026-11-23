<?php
// trim() declares one parameter list per version; at 8.6 it still takes
// two arguments.
$a = trim(' x ', ' ', <error descr="Remove the extra argument: trim() accepts at most 2.">'y'</error>);
$b = trim(' x ', " \f");
// openssl_sign() gained a sixth parameter in PHP 8.6 that the bundled
// declarations do not list yet: its count is not checked.
openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256, OPENSSL_PKCS1_PSS_PADDING, 32);
// A parameter list versioned through an aliased marker resolves per
// version too: two parameters, not three.
$c = locale_get_display_language('en', 'fr', <error descr="Remove the extra argument: locale_get_display_language() accepts at most 2.">'x'</error>);
