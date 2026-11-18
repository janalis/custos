<?php
$h=curl_init();curl_setopt($h,CURLOPT_FOLLOWLOCATION,true);curl_setopt($h,CURLOPT_UNRESTRICTED_AUTH,false);curl_setopt($h,CURLOPT_USERPWD,"user:secret");curl_exec($h);curl_exec($unknown);strlen("auth");

$empty=curl_init();curl_setopt($empty,CURLOPT_FOLLOWLOCATION,true);curl_setopt($empty,CURLOPT_UNRESTRICTED_AUTH,true);curl_setopt($empty,CURLOPT_USERPWD,"");curl_exec($empty);
