<?php
$h=curl_init();curl_setopt($h,CURLOPT_CUSTOMREQUEST,"HEAD");<warning descr="Use CURLOPT_NOBODY for a HEAD transfer.">curl_exec($h)</warning>;
$other=curl_init();curl_setopt($other,CURLOPT_CUSTOMREQUEST,"HEAD");curl_setopt($other,CURLOPT_NOBODY,false);<warning descr="Use CURLOPT_NOBODY for a HEAD transfer.">curl_exec($other)</warning>;
