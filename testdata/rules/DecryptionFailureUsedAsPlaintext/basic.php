<?php
$plain=openssl_decrypt($cipher,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag); file_put_contents($path,<warning descr="Check decryption success before using the plaintext.">$plain</warning>);
