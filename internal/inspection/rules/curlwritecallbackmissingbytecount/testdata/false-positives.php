<?php
$h=curl_init();curl_setopt($h,CURLOPT_WRITEFUNCTION,function($h,$data){return strlen($data);});curl_setopt($h,CURLOPT_WRITEFUNCTION,$callback);curl_setopt($h,CURLOPT_URL,"https://example.test");curl_setopt($h,$option,function(){});strlen("callback");
