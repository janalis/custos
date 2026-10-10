<?php
$key="0123456789012345"; $iv="0123456789012345"; $c=openssl_encrypt($data,"aes-128-cbc",$key,OPENSSL_RAW_DATA,$iv); openssl_decrypt($c,"aes-128-cbc",$key,OPENSSL_RAW_DATA,$iv);
