<?php
$n = random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES); sodium_crypto_secretbox('alpha', $n, $k); <warning descr="Use a fresh nonce for each distinct message.">sodium_crypto_secretbox('beta', $n, $k)</warning>;
