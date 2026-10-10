<?php
$c=sodium_crypto_secretbox($message,<warning descr="Use the required secretbox nonce length.">random_bytes(12)</warning>,$key);
