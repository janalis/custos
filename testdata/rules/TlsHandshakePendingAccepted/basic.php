<?php
if (stream_set_blocking($s, false)) { if (<warning descr="Require a completed TLS handshake before writing.">stream_socket_enable_crypto($s, true, STREAM_CRYPTO_METHOD_TLS_CLIENT)</warning> !== false) { fwrite($s, $secret); } }
