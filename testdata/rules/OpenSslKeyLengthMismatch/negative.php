<?php
openssl_encrypt($data,"aes-256-gcm",random_bytes(32),OPENSSL_RAW_DATA,$iv,$tag);
