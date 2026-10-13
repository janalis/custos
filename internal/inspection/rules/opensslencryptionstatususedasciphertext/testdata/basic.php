<?php
$cipher = openssl_public_encrypt($data, $out, $key); echo <warning descr="Use the encryption output as ciphertext.">base64_encode($cipher)</warning>;
