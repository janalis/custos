<?php
// The project requires ext-sodium and ext-zend-opcache: PECL builds them
// for PHP 5.4.
$box = sodium_crypto_box_keypair();
opcache_reset();
// mbstring has no such build.
<error descr="mb_str_pad() needs PHP 8.3; the project targets PHP 5.4.">mb_str_pad</error>('x', 3);
