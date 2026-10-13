<?php
<warning descr="Check signing success before publishing the signature.">openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256)</warning>; echo base64_encode($signature);
