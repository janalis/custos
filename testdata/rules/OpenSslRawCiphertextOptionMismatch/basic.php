<?php
$key="0123456789012345"; $iv="0123456789012345"; $c=openssl_encrypt($data,"aes-128-cbc",$key,OPENSSL_RAW_DATA,$iv); <warning descr="Match raw-ciphertext options when decrypting.">openssl_decrypt($c,"aes-128-cbc",$key,0,$iv)</warning>;
