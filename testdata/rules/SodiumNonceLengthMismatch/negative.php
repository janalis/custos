<?php
$c=sodium_crypto_secretbox($message,random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES),$key);
