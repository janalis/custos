<?php
sodium_crypto_secretbox('alpha', random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES), $k); sodium_crypto_secretbox('beta', random_bytes(SODIUM_CRYPTO_SECRETBOX_NONCEBYTES), $k);
