<?php
openssl_encrypt($data,"aes-256-gcm",<warning descr="Derive an encryption key from the password.">$_POST["password"]</warning>,OPENSSL_RAW_DATA,$iv,$tag);
