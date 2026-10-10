<?php
openssl_encrypt($data,"aes-256-gcm",$_POST["key"],OPENSSL_RAW_DATA,$iv,$tag);
