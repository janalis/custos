<?php
$h=curl_init();curl_setopt($h,CURLOPT_POST,true);curl_setopt($h,<warning descr="Configure a consistent final HTTP request method.">CURLOPT_HTTPGET</warning>,true);curl_exec($h);
$g=curl_init();curl_setopt($g,CURLOPT_HTTPGET,true);curl_setopt($g,<warning descr="Configure a consistent final HTTP request method.">CURLOPT_POST</warning>,true);curl_exec($g);
