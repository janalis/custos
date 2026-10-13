<?php
openssl_encrypt($data,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag);
