<?php
$h=curl_init();$cb=function($ch,$line){if(true)return strlen($line);};curl_setopt($h,CURLOPT_HEADERFUNCTION,$cb);var_dump($cb($h,'abc'));
