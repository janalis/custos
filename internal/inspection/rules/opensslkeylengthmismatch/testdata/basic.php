<?php
openssl_encrypt($data,"aes-256-gcm",<warning descr="Provide the cipher-required key length.">random_bytes(16)</warning>,OPENSSL_RAW_DATA,$iv,$tag);
