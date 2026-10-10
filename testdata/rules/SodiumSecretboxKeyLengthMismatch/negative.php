<?php
sodium_crypto_secretbox($message, $nonce, sodium_crypto_secretbox_keygen());
