<?php
$h=curl_init();curl_setopt($h,CURLOPT_POSTFIELDS,"{}");curl_setopt($h,CURLOPT_HTTPHEADER,["Content-Type: application/json"]);curl_exec($h);$g=curl_init();curl_setopt($g,CURLOPT_POSTFIELDS,[]);curl_setopt($g,CURLOPT_HTTPHEADER,["Accept: application/json"]);curl_exec($g);curl_exec($unknown);strlen("json");
