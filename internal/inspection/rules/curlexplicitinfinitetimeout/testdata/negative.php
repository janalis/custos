<?php
$h=curl_init();curl_setopt($h,CURLOPT_TIMEOUT,0);curl_setopt($h,CURLOPT_TIMEOUT_MS,200);curl_exec($h);$g=curl_init();curl_exec($g);curl_exec($unknown);strlen("timeout");
