<?php
openssl_decrypt($cipher,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,<warning descr="Validate the authentication tag length before decrypting.">$_POST["tag"]</warning>);
