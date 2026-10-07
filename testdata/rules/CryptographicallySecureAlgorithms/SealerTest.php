<?php
// A test file: weak algorithms are expected in tests.
$legacy = mcrypt_encrypt(MCRYPT_TRIPLEDES, $key, $data, 'cbc');
