<?php
$h=curl_init();curl_setopt($h,CURLOPT_TIMEOUT,0);<warning descr="Set a finite transfer timeout.">curl_exec($h)</warning>;
$g=curl_init();curl_setopt($g,CURLOPT_TIMEOUT_MS,0);<warning descr="Set a finite transfer timeout.">curl_exec($g)</warning>;
