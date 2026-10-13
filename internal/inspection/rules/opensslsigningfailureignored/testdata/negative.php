<?php
if (openssl_sign($data, $signature, $key, OPENSSL_ALGO_SHA256)) { echo base64_encode($signature); }
