<?php
$tag=$_POST["tag"]; if(strlen($tag)!==16){return;} openssl_decrypt($cipher,"aes-256-gcm",$key,OPENSSL_RAW_DATA,$iv,$tag);
