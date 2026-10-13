<?php
$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,"HEAD");curl_setopt($h,CURLOPT_NOBODY,true);curl_exec($h);$g=curl_init();curl_setopt($g,CURLOPT_CUSTOMREQUEST,"GET");curl_exec($g);curl_exec($unknown);strlen("HEAD");
