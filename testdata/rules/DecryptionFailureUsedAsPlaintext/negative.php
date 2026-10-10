<?php
$plain=openssl_decrypt($cipher,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag); if($plain===false){return;} file_put_contents($path,$plain);
