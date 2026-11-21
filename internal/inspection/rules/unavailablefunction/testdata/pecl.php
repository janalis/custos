<?php
// The project requires ext-sodium: PECL builds it for PHP 7.1.
$box = sodium_crypto_box_keypair();
// mbstring has no such build.
<error descr="mb_str_pad() needs PHP 8.3; the project targets PHP 7.1.">mb_str_pad</error>('x', 3);
