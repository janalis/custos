<?php
sodium_crypto_secretstream_xchacha20poly1305_push($s, 'more'); sodium_crypto_secretstream_xchacha20poly1305_push($s, 'end', '', SODIUM_CRYPTO_SECRETSTREAM_XCHACHA20POLY1305_TAG_FINAL);
