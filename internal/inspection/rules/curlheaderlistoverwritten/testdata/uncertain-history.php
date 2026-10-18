<?php
$h=curl_init();if($flag){curl_setopt($h,CURLOPT_TIMEOUT,3);}curl_setopt($h,CURLOPT_HTTPHEADER,['Authorization: Bearer example']);curl_setopt($h,CURLOPT_HTTPHEADER,['Accept: application/json']);curl_exec($h);
